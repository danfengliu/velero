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
	"strings"

	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type GCSStorage string

func (s GCSStorage) IsObjectsInBucket(cloudCredentialsFile, region, bslBucket, bslPrefix, bslConfig, backupObject string) (bool, error) {
	bslPrefix = bslPrefix + "/"
	q := &storage.Query{
		Prefix: bslPrefix,
		//Delimiter: "/",
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(cloudCredentialsFile))
	if err != nil {
		return false, errors.Wrapf(err, "Fail to create gcloud client")
	}
	iter := client.Bucket(bslBucket).Objects(context.Background(), q)
	for {
		obj, err := iter.Next()
		if err == iterator.Done {
			fmt.Printf("Backup %s was not found under prefix %s \n", backupObject, bslPrefix)
			return false, nil
		}
		if err != nil {
			return false, errors.WithStack(err)
		}
		if obj.Name == bslPrefix {
			fmt.Println("Ignore GCS prefix itself")
			continue
		}
		if strings.Contains(obj.Name, bslPrefix+backupObject+"/") {
			fmt.Printf("Found delete-object %s of %s in bucket %s \n", backupObject, obj.Name, bslBucket)
			return true, nil
		}
	}
}
func (s GCSStorage) deleteObjectsInBucket(cloudCredentialsFile, region, bslBucket, bslPrefix, bslConfig, backupObject string) error {
	bslPrefix = bslPrefix + "/"
	q := &storage.Query{
		Prefix: bslPrefix,
		//Delimiter: "/",
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(cloudCredentialsFile))
	if err != nil {
		return errors.Wrapf(err, "Fail to create gcloud client")
	}
	bucket := client.Bucket(bslBucket)
	iter := bucket.Objects(context.Background(), q)
	for {
		obj, err := iter.Next()
		if err == iterator.Done {
			fmt.Println(err)
			return nil
		}
		if err != nil {
			return errors.WithStack(err)
		}
		if obj.Name == bslPrefix {
			fmt.Println("Ignore GCS prefix itself")
			continue
		}
		// Only support to delete folder named as backupObject
		if strings.Contains(obj.Name, bslPrefix+backupObject+"/") {
			if err = bucket.Object(obj.Name).Delete(ctx); err != nil {
				return errors.Wrapf(err, fmt.Sprintf("Fail to delete object %s in bucket %s", obj.Name, bslBucket))
			}
			fmt.Printf("delete %s", obj.Name)
		}
	}
}
