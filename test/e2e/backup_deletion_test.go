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
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

const (
	deletionTest = "upgrade-workload"
)

// Test backup and restore of Kibishi using restic
var _ = Describe("[backup-resource][deletion][snapshots] Velero tests on cluster using the plugin provider for object storage and Restic for volume backups", backup_deletion_with_snapshots)
var _ = Describe("[backup-resource][deletion][restic] Velero tests on cluster using the plugin provider for object storage and Restic for volume backups", backup_deletion_with_restic)

func backup_deletion_with_snapshots() {
	backup_deletion_test(true)
}

func backup_deletion_with_restic() {
	backup_deletion_test(false)
}
func backup_deletion_test(useVolumeSnapshots bool) {
	var (
		backupName string
	)

	client, err := newTestClient()
	Expect(err).To(Succeed(), "Failed to instantiate cluster client for backup tests")

	BeforeEach(func() {
		if useVolumeSnapshots && cloudProvider == "kind" {
			Skip("Volume snapshots not supported on kind")
		}
		var err error
		flag.Parse()
		uuidgen, err = uuid.NewRandom()
		Expect(err).To(Succeed())
		if installVelero {
			Expect(veleroInstall(context.Background(), veleroCLI, veleroImage, resticHelperImage, plugins, veleroNamespace, cloudProvider, objectStoreProvider, useVolumeSnapshots,
				cloudCredentialsFile, bslBucket, bslPrefix, bslConfig, vslConfig, crdsVersion, "", registryCredentialFile)).To(Succeed())
		}
	})

	AfterEach(func() {
		/*if installVelero {
			err = veleroUninstall(context.Background(), veleroCLI, veleroNamespace)
			Expect(err).To(Succeed())
		}*/
	})

	When("kibishii is the sample workload", func() {
		It("should be successfully backed up and restored to the default BackupStorageLocation", func() {
			backupName = "backup-" + uuidgen.String()
			// Even though we are using Velero's CloudProvider plugin for object storage, the kubernetes cluster is running on
			// KinD. So use the kind installation for Kibishii.

			Expect(runBackupDeletionTests(client, veleroCLI, cloudProvider, veleroNamespace, backupName, "", useVolumeSnapshots, registryCredentialFile, "us-west-1", bslPrefix, bslConfig)).To(Succeed(),
				"Failed to successfully backup and restore Kibishii namespace")
		})
	})
}

// runUpgradeTests runs upgrade test on the provider by kibishii.
func runBackupDeletionTests(client testClient, veleroCLI, providerName, veleroNamespace, backupName, backupLocation string,
	useVolumeSnapshots bool, registryCredentialFile, region, bslPrefix, bslConfig string) error {
	/*//backupName = "ex-2-no-app"
	backupName = "aaa"
	err1 := objectsShouldBeInBucket(cloudProvider, cloudCredentialsFile, region, bslBucket, bslPrefix, bslConfig, backupName, backupObjectsPrefix)
	if err1 != nil {
		fmt.Println(errors.Wrapf(err1, "Failed to get object from bucket %q", backupName))
	}
	err1 = deleteObjectsInBucket(cloudProvider, cloudCredentialsFile, region, bslBucket, bslPrefix, bslConfig, backupName, backupObjectsPrefix)
	if err1 != nil {
		fmt.Println(errors.Wrapf(err1, "Failed to get object from bucket %q", backupName))
	}
	return nil
	*/
	oneHourTimeout, _ := context.WithTimeout(context.Background(), time.Minute*60)

	if err := createNamespace(oneHourTimeout, client, deletionTest); err != nil {
		return errors.Wrapf(err, "Failed to create namespace %s to install Kibishii workload", deletionTest)
	}
	defer func() {
		if err := deleteNamespace(context.Background(), client, deletionTest, true); err != nil {
			fmt.Println(errors.Wrapf(err, "failed to delete the namespace %q", deletionTest))
		}
	}()

	if err := kibishiiPrepareBeforeBackup(oneHourTimeout, client, providerName, deletionTest, registryCredentialFile); err != nil {
		return errors.Wrapf(err, "Failed to install and prepare data for kibishii %s", deletionTest)
	}
	err := objectsShouldNotBeInBucket(cloudProvider, cloudCredentialsFile, region, bslBucket, bslPrefix, bslConfig, backupName, backupObjectsPrefix, 1)
	if err != nil {
		fmt.Println(errors.Wrapf(err, "Failed to get object from bucket %q", backupName))
		return err
	}
	if err := veleroBackupNamespace(oneHourTimeout, veleroCLI, veleroNamespace, backupName, deletionTest, backupLocation, useVolumeSnapshots); err != nil {
		// TODO currently, the upgrade case covers the upgrade path from 1.6 to main and the velero v1.6 doesn't support "debug" command
		// TODO move to "runDebug" after we bump up to 1.7 in the upgrade case
		veleroBackupLogs(context.Background(), upgradeFromVeleroCLI, veleroNamespace, backupName)
		return errors.Wrapf(err, "Failed to backup kibishii namespace %s", deletionTest)
	}

	if providerName == "vsphere" && useVolumeSnapshots {
		// Wait for uploads started by the Velero Plug-in for vSphere to complete
		// TODO - remove after upload progress monitoring is implemented
		fmt.Println("Waiting for vSphere uploads to complete")
		if err := waitForVSphereUploadCompletion(oneHourTimeout, time.Hour, deletionTest); err != nil {
			return errors.Wrapf(err, "Error waiting for uploads to complete")
		}
	}
	err = objectsShouldBeInBucket(cloudProvider, cloudCredentialsFile, region, bslBucket, bslPrefix, bslConfig, backupName, backupObjectsPrefix)
	if err != nil {
		fmt.Println(errors.Wrapf(err, "Failed to get object from bucket %q", backupName))
		return err
	}
	err = deleteBackupResource(context.Background(), veleroCLI, backupName)
	if err != nil {
		fmt.Println(errors.Wrapf(err, "Failed to delete backup %q", backupName))
		return err
	}
	err = objectsShouldNotBeInBucket(cloudProvider, cloudCredentialsFile, region, bslBucket, bslPrefix, bslConfig, backupName, backupObjectsPrefix, 5)
	if err != nil {
		fmt.Println(errors.Wrapf(err, "Failed to get object from bucket %q", backupName))
		return err
	}
	backupName = "backup-1-" + uuidgen.String()
	if err := veleroBackupNamespace(oneHourTimeout, veleroCLI, veleroNamespace, backupName, deletionTest, backupLocation, useVolumeSnapshots); err != nil {
		// TODO currently, the upgrade case covers the upgrade path from 1.6 to main and the velero v1.6 doesn't support "debug" command
		// TODO move to "runDebug" after we bump up to 1.7 in the upgrade case
		veleroBackupLogs(context.Background(), upgradeFromVeleroCLI, veleroNamespace, backupName)
		return errors.Wrapf(err, "Failed to backup kibishii namespace %s", deletionTest)
	}
	err = deleteObjectsInBucket(cloudProvider, cloudCredentialsFile, region, bslBucket, bslPrefix, bslConfig, backupName, backupObjectsPrefix)
	if err != nil {
		fmt.Println(errors.Wrapf(err, "Failed to get object from bucket %q", backupName))
		return err
	}
	err = deleteBackupResource(context.Background(), veleroCLI, backupName)
	if err != nil {
		fmt.Println(errors.Wrapf(err, "|| UNEXPECTED || - Failed to delete backup %q", backupName))
	}
	fmt.Printf("|| EXPECTED || - Backup deletion test completed successfully\n")
	return nil
}
