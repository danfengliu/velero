/*
Copyright the Velero contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/context"
	corev1api "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/vmware-tanzu/velero/pkg/builder"
)

// ensureClusterExists returns whether or not a kubernetes cluster exists for tests to be run on.
func ensureClusterExists(ctx context.Context) error {
	return exec.CommandContext(ctx, "kubectl", "cluster-info").Run()
}

func createSecretFromFiles(ctx context.Context, client testClient, namespace string, name string, files map[string]string) error {
	data := make(map[string][]byte)

	for key, filePath := range files {
		contents, err := ioutil.ReadFile(filePath)
		if err != nil {
			return errors.WithMessagef(err, "Failed to read secret file %q", filePath)
		}

		data[key] = contents
	}

	secret := builder.ForSecret(namespace, name).Data(data).Result()
	_, err := client.clientGo.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	return err
}

// waitForPods waits until all of the pods have gone to PodRunning state
func waitForPods(ctx context.Context, client testClient, namespace string, pods []string) error {
	timeout := 10 * time.Minute
	interval := 5 * time.Second
	err := wait.PollImmediate(interval, timeout, func() (bool, error) {
		for _, podName := range pods {
			checkPod, err := client.clientGo.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
			if err != nil {
				return false, errors.WithMessage(err, fmt.Sprintf("Failed to verify pod %s/%s is %s", namespace, podName, corev1api.PodRunning))
			}
			// If any pod is still waiting we don't need to check any more so return and wait for next poll interval
			if checkPod.Status.Phase != corev1api.PodRunning {
				fmt.Printf("Pod %s is in state %s waiting for it to be %s\n", podName, checkPod.Status.Phase, corev1api.PodRunning)
				return false, nil
			}
		}
		// All pods were in PodRunning state, we're successful
		return true, nil
	})
	if err != nil {
		return errors.Wrapf(err, fmt.Sprintf("Failed to wait for pods in namespace %s to start running", namespace))
	}
	return nil
}

type ObjectsInStorage interface {
	IsObjectsInBucket(cloudCredentialsFile, region, bslBucket, bslPrefix, bslConfig, backupObject string) (bool, error)
	deleteObjectsInBucket(cloudCredentialsFile, region, bslBucket, bslPrefix, bslConfig, backupObject string) error
}

func objectsShouldBeInBucket(cloudProvider, cloudCredentialsFile, region, bslBucket, bslPrefix, bslConfig, backupName, subPrefix string) error {
	exist, err := IsObjectsInBucket(cloudProvider, cloudCredentialsFile, region, bslBucket, bslPrefix, bslConfig, backupName, subPrefix)
	if !exist {
		return errors.Wrapf(err, "|| UNEXPECTED ||Backup object %s is not exist in object store after backup as expected", backupName)
	}
	fmt.Printf("|| EXPECTED || - Backup %s exsit in object storage bucket %s\n", backupName, bslBucket)
	return nil
}
func objectsShouldNotBeInBucket(cloudProvider, cloudCredentialsFile, region, bslBucket, bslPrefix, bslConfig, backupName, subPrefix string, retryTimes int) error {
	var err error
	var exist bool
	for i := 0; i < retryTimes; i++ {
		exist, err = IsObjectsInBucket(cloudProvider, cloudCredentialsFile, region, bslBucket, bslPrefix, bslConfig, backupName, subPrefix)
		if err != nil {
			return errors.Wrapf(err, "|| UNEXPECTED || - Failed to get backup %s in object store", backupName)
		}
		if !exist {
			fmt.Printf("|| EXPECTED || - Backup %s is not in object store\n", backupName)
			return nil
		}
		time.Sleep(1 * time.Minute)
	}
	return errors.New(fmt.Sprintf("|| UNEXPECTED ||Backup object %s still exist in object store after backup deletion", backupName))
}
func getProvider(cloudProvider string) (ObjectsInStorage, error) {
	var s ObjectsInStorage
	switch cloudProvider {
	case "aws", "vsphere":
		aws := AWSStorage("")
		s = &aws
	case "gcp":
		gcs := GCSStorage("")
		s = &gcs
	case "azure":
		az := AzureStorage("")
		s = &az
	default:
		return nil, errors.New(fmt.Sprintf("Cloud provider %s is not valid", cloudProvider))
	}
	return s, nil
}
func getFullPrefix(bslPrefix, subPrefix string) string {
	if bslPrefix == "" {
		bslPrefix = subPrefix + "/"
	} else {
		//subPrefix must have surfix "/", so that objects under it can be listed
		bslPrefix = strings.Trim(bslPrefix, "/") + "/" + strings.Trim(subPrefix, "/") + "/"
	}
	return bslPrefix
}
func IsObjectsInBucket(cloudProvider, cloudCredentialsFile, region, bslBucket, bslPrefix, bslConfig, backupName, subPrefix string) (bool, error) {
	bslPrefix = getFullPrefix(bslPrefix, subPrefix)
	s, err := getProvider(cloudProvider)
	if err != nil {
		return false, errors.Wrapf(err, fmt.Sprintf("Cloud provider %s is not valid", cloudProvider))
	}
	return s.IsObjectsInBucket(cloudCredentialsFile, region, bslBucket, bslPrefix, bslConfig, backupName)
}

func deleteObjectsInBucket(cloudProvider, cloudCredentialsFile, region, bslBucket, bslPrefix, bslConfig, backupName, subPrefix string) error {
	bslPrefix = getFullPrefix(bslPrefix, subPrefix)
	s, err := getProvider(cloudProvider)
	if err != nil {
		return errors.Wrapf(err, fmt.Sprintf("Cloud provider %s is not valid", cloudProvider))
	}
	err = s.deleteObjectsInBucket(cloudCredentialsFile, region, bslBucket, bslPrefix, bslConfig, backupName)
	if err != nil {
		return errors.Wrapf(err, fmt.Sprintf("Fail to delete %s", bslPrefix))
	}
	return nil
}
