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
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/Azure/azure-pipeline-go/pipeline"
	storagemgmt "github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2019-06-01/storage"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/vmware-tanzu/velero/pkg/cmd/util/flag"
	"golang.org/x/net/context"
)

type AzureStorage string

const (
	subscriptionIDEnvVar = "AZURE_SUBSCRIPTION_ID"
	cloudNameEnvVar      = "AZURE_CLOUD_NAME"
	resourceGroupEnvVar  = "AZURE_RESOURCE_GROUP"
	storageAccountKey    = "AZURE_STORAGE_ACCOUNT_ACCESS_KEY"
	storageAccount       = "storageAccount"
	subscriptionID       = "subscriptionId"
	resourceGroup        = "resourceGroup"
)

func getStorageCredential(cloudCredentialsFile, bslConfig string) (string, string, error) {
	//"danfeng01" "/home/dd/cayman/azure/qiuming/credentials-velero"
	config := flag.NewMap()
	config.Set(bslConfig)
	accountName := config.Data()[storageAccount]
	// Account name must be provided in config
	if len(accountName) == 0 {
		return "", "", errors.New("Please provide bucket as Azure account name ")
	}
	subscriptionID := config.Data()[subscriptionID]
	resourceGroupCfg := config.Data()[resourceGroup]
	accountKey, err := getStorageAccountKey(cloudCredentialsFile, accountName, subscriptionID, resourceGroupCfg)
	if err != nil {
		return "", "", errors.Wrapf(err, "Fail to get storage key of bucket %s", accountName)
	}
	return accountName, accountKey, nil
}
func loadCredentialsIntoEnv(credentialsFile string) error {
	if credentialsFile == "" {
		return nil
	}

	if err := godotenv.Overload(credentialsFile); err != nil {
		return errors.Wrapf(err, "error loading environment from credentials file (%s)", credentialsFile)
	}
	return nil
}
func parseAzureEnvironment(cloudName string) (*azure.Environment, error) {
	if cloudName == "" {
		fmt.Println("cloudName is empty")
		return &azure.PublicCloud, nil
	}

	env, err := azure.EnvironmentFromName(cloudName)
	return &env, errors.WithStack(err)
}
func getStorageAccountKey(credentialsFile, accountName, subscriptionID, resourceGroupCfg string) (string, error) {
	if err := loadCredentialsIntoEnv(credentialsFile); err != nil {
		return "", err
	}
	storageKey := os.Getenv(storageAccountKey)
	if storageKey != "" {
		return storageKey, nil
	}
	if os.Getenv(cloudNameEnvVar) == "" {
		return "", errors.New("Credential file should contain AZURE_CLOUD_NAME")
	}
	var resourceGroup string
	if os.Getenv(resourceGroupEnvVar) == "" {
		if resourceGroupCfg == "" {
			return "", errors.New("Credential file should contain AZURE_RESOURCE_GROUP or AZURE_STORAGE_ACCOUNT_ACCESS_KEY")
		} else {
			resourceGroup = resourceGroupCfg
		}
	} else {
		resourceGroup = os.Getenv(resourceGroupEnvVar)
	}
	// get Azure cloud from AZURE_CLOUD_NAME, if it exists. If the env var does not
	// exist, parseAzureEnvironment will return azure.PublicCloud.
	env, err := parseAzureEnvironment(os.Getenv(cloudNameEnvVar))
	if err != nil {
		return "", errors.Wrap(err, "unable to parse azure cloud name environment variable")
	}

	// get subscription ID from object store config or AZURE_SUBSCRIPTION_ID environment variable
	if subscriptionID == "" {
		return "", errors.New("azure subscription ID not found in object store's config or in environment variable")
	}

	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err != nil {
		return "", errors.Wrap(err, "error getting authorizer from environment")
	}

	// get storageAccountsClient
	storageAccountsClient := storagemgmt.NewAccountsClientWithBaseURI(env.ResourceManagerEndpoint, subscriptionID)
	storageAccountsClient.Authorizer = authorizer

	// get storage key
	res, err := storageAccountsClient.ListKeys(context.TODO(), resourceGroup, accountName, storagemgmt.Kerb)
	if err != nil {
		return "", errors.WithStack(err)
	}
	if res.Keys == nil || len(*res.Keys) == 0 {
		return "", errors.New("No storage keys found")
	}

	for _, key := range *res.Keys {
		// uppercase both strings for comparison because the ListKeys call returns e.g. "FULL" but
		// the storagemgmt.Full constant in the SDK is defined as "Full".
		if strings.EqualFold(string(key.Permissions), string(storagemgmt.Full)) {
			storageKey = *key.Value
			break
		}
	}

	if storageKey == "" {
		return "", errors.New("No storage key with Full permissions found")
	}

	return storageKey, nil
}
func handleErrors(err error) {
	if err != nil {
		if serr, ok := err.(azblob.StorageError); ok { // This error is a Service-specific
			switch serr.ServiceCode() { // Compare serviceCode to ServiceCodeXxx constants
			case azblob.ServiceCodeContainerAlreadyExists:
				//fmt.Println("Received 409. Container already exists")
				return
			}
		}
		log.Fatal(err)
	}
}

func deleteBlob(p pipeline.Pipeline, accountName, containerName, blobName string) error {
	ctx := context.Background()

	URL_BLOB, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", accountName, containerName, blobName))
	if err != nil {
		return errors.Wrapf(err, "Fail to url.Parse")
	}
	blobURL := azblob.NewBlobURL(*URL_BLOB, p)
	_, err = blobURL.Delete(ctx, azblob.DeleteSnapshotsOptionNone, azblob.BlobAccessConditions{})
	return err
}
func (s AzureStorage) IsObjectsInBucket(cloudCredentialsFile, region, bslBucket, bslPrefix, bslConfig, backupObject string) (bool, error) {

	accountName, accountKey, err := getStorageCredential(cloudCredentialsFile, bslConfig)
	if err != nil {
		log.Fatal("Fail to get : accountName and accountKey, " + err.Error())
	}
	// Create a default request pipeline using your storage account name and account key.
	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		log.Fatal("Invalid credentials with error: " + err.Error())
	}
	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	containerName := bslBucket

	// From the Azure portal, get your storage account blob service URL endpoint.
	URL, _ := url.Parse(
		fmt.Sprintf("https://%s.blob.core.windows.net/%s", accountName, containerName))

	// Create a ContainerURL object that wraps the container URL and a request
	// pipeline to make requests.
	containerURL := azblob.NewContainerURL(*URL, p)

	// Create the container
	//fmt.Printf("Creating a container named %s\n", containerName)
	ctx := context.Background() // This example uses a never-expiring context
	_, err = containerURL.Create(ctx, azblob.Metadata{}, azblob.PublicAccessNone)
	handleErrors(err)

	// List the container that we have created above
	fmt.Printf("Finding backup %s blobs in Azure container/bucket %s\n", backupObject, containerName)
	for marker := (azblob.Marker{}); marker.NotDone(); {
		// Get a result segment starting with the blob indicated by the current Marker.
		listBlob, err := containerURL.ListBlobsFlatSegment(ctx, marker, azblob.ListBlobsSegmentOptions{})
		if err != nil {
			return false, errors.Wrapf(err, "Fail to create gcloud client")
		}
		marker = listBlob.NextMarker

		// Process the blobs returned in this result segment (if the segment is empty, the loop body won't execute)
		for _, blobInfo := range listBlob.Segment.BlobItems {
			//fmt.Print("	Blob name: " + blobInfo.Name + "\n")
			if strings.Contains(blobInfo.Name, backupObject) {
				fmt.Printf("Blob name: %s exist in %s\n", backupObject, blobInfo.Name)
				return true, nil
			}
		}
	}
	return false, nil
}

func (s AzureStorage) deleteObjectsInBucket(cloudCredentialsFile, region, bslBucket, bslPrefix, bslConfig, backupObject string) error {
	//"danfeng01" "/home/dd/cayman/azure/qiuming/credentials-velero"

	ctx := context.Background()
	/*config := flag.NewMap()
	config.Set(bslConfig)

	accountName := config.Data()[storageAccount]
	subscriptionID := config.Data()[subscriptionID]
	resourceGroupCfg := config.Data()[resourceGroup]
	accountKey, err := getStorageAccountKey(cloudCredentialsFile, accountName, subscriptionID, resourceGroupCfg)
	if len(accountName) == 0 {
		return errors.Wrapf(err, "Please provide bucket as Azure account name ")
	}
	fmt.Println(accountKey)
	if err != nil {
		return errors.Wrapf(err, "Fail to get storage key of bucket %s", accountName)
	}
	*/
	accountName, accountKey, err := getStorageCredential(cloudCredentialsFile, bslConfig)
	if err != nil {
		return errors.Wrapf(err, "Fail to get storage account name and  key of bucket %s", bslBucket)
	}
	// Create a default request pipeline using your storage account name and account key.
	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		log.Fatal("Invalid credentials with error: " + err.Error())
	}
	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	// Create a random string for the quick start container
	containerName := bslBucket

	// From the Azure portal, get your storage account blob service URL endpoint.
	URL, _ := url.Parse(
		fmt.Sprintf("https://%s.blob.core.windows.net/%s", accountName, containerName))

	// Create a ContainerURL object that wraps the container URL and a request
	// pipeline to make requests.
	containerURL := azblob.NewContainerURL(*URL, p)
	_, err = containerURL.Create(ctx, azblob.Metadata{}, azblob.PublicAccessNone)
	handleErrors(err)

	// List the container that we have created above
	fmt.Println("Listing the blobs in the container:")
	for marker := (azblob.Marker{}); marker.NotDone(); {
		// Get a result segment starting with the blob indicated by the current Marker.
		listBlob, err := containerURL.ListBlobsFlatSegment(ctx, marker, azblob.ListBlobsSegmentOptions{})
		if err != nil {
			return errors.Wrapf(err, "Fail to create gcloud client")
		}

		// ListBlobs returns the start of the next segment; you MUST use this to get
		// the next segment (after processing the current result segment).
		marker = listBlob.NextMarker
		// Process the blobs returned in this result segment (if the segment is empty, the loop body won't execute)
		for _, blobInfo := range listBlob.Segment.BlobItems {

			if strings.Contains(blobInfo.Name, bslPrefix+backupObject+"/") {
				deleteBlob(p, accountName, containerName, blobInfo.Name)
				if err != nil {
					log.Fatal("Invalid credentials with error: " + err.Error())
				}
				fmt.Printf("Deleted blob: %s according to backup resource %s\n", blobInfo.Name, bslPrefix+backupObject+"/")
			}
		}
	}
	return nil
}
