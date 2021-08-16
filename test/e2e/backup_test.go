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

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	uuidgen uuid.UUID
)

// Test backup and restore of Kibishi using restic
var _ = Describe("[Restic] Velero tests on cluster using the plugin provider for object storage and Restic for volume backups", backup_restore_with_restic)

var _ = Describe("[Snapshot] Velero tests on cluster using the plugin provider for object storage and snapshots for volume backups", backup_restore_with_snapshots)

func backup_restore_with_snapshots() {
	backup_restore_test(true)
}

func backup_restore_with_restic() {
	backup_restore_test(false)
}

func backup_restore_test(useVolumeSnapshots bool) {
	var (
		backupName, restoreName string
		installParams           veleroInstallationParams
	)
	installParams.veleroCLI = veleroCLI
	installParams.veleroImage = veleroImage
	installParams.veleroNamespace = veleroNamespace
	installParams.cloudProvider = cloudProvider
	installParams.objectStoreProvider = objectStoreProvider
	installParams.useVolumeSnapshots = useVolumeSnapshots
	installParams.cloudCredentialsFile = cloudCredentialsFile
	installParams.bslBucket = bslBucket
	installParams.bslPrefix = bslPrefix
	installParams.bslConfig = bslConfig
	installParams.vslConfig = vslConfig
	installParams.crdsVersion = crdsVersion
	installParams.features = ""
	installParams.registryCredentialFile = registryCredentialFile

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
			err = veleroUninstall(context.Background(), client.kubebuilder, installVelero, veleroNamespace)
			Expect(err).To(Succeed())
		}
	})

	AfterEach(func() {
		// if installVelero {
		// 	err = veleroUninstall(context.Background(), client.kubebuilder, installVelero, veleroNamespace)
		// 	Expect(err).To(Succeed())
		// }
	})

	When("kibishii is the sample workload", func() {
		XIt("should be successfully backed up and restored to the default BackupStorageLocation", func() {
			backupName = "backup-" + uuidgen.String()
			restoreName = "restore-" + uuidgen.String()
			installParams.veleroImage = veleroImage
			if installVelero {
				Expect(veleroInstall(context.Background(), veleroImage, veleroNamespace, cloudProvider, objectStoreProvider, useVolumeSnapshots,
					cloudCredentialsFile, bslBucket, bslPrefix, bslConfig, vslConfig, crdsVersion, "", registryCredentialFile)).To(Succeed())
			}
			// Even though we are using Velero's CloudProvider plugin for object storage, the kubernetes cluster is running on
			// KinD. So use the kind installation for Kibishii.
			Expect(runKibishiiTests(client, cloudProvider, veleroCLI, veleroNamespace, backupName, restoreName, "", useVolumeSnapshots, registryCredentialFile)).To(Succeed(),
				"Failed to successfully backup and restore Kibishii namespace")
		})

		FIt("should be successfully backed up, upgraded and restored to the default BackupStorageLocation", func() {
			backupName = "backup-" + uuidgen.String()
			restoreName = "restore-" + uuidgen.String()
			// Even though we are using Velero's CloudProvider plugin for object storage, the kubernetes cluster is running on
			// KinD. So use the kind installation for Kibishii.
			installParams.veleroImage = "harbor-repo.vmware.com/harbor-ci/velero/velero:v1.6.3"
			if installVelero {
				Expect(veleroInstallNew(&installParams)).To(Succeed())
			}
			fmt.Println("---------runUpgradeTests------------")
			installParams.veleroImage = veleroImage
			Expect(runUpgradeTests(client, cloudProvider, veleroCLI, veleroNamespace, backupName, restoreName, "", useVolumeSnapshots, registryCredentialFile, &installParams)).To(Succeed(),
				"Failed to successfully backup and restore Kibishii namespace")
		})

		XIt("should successfully back up and restore to an additional BackupStorageLocation with unique credentials", func() {
			if additionalBSLProvider == "" {
				Skip("no additional BSL provider given, not running multiple BackupStorageLocation with unique credentials tests")
			}

			if additionalBSLBucket == "" {
				Skip("no additional BSL bucket given, not running multiple BackupStorageLocation with unique credentials tests")
			}

			if additionalBSLCredentials == "" {
				Skip("no additional BSL credentials given, not running multiple BackupStorageLocation with unique credentials tests")
			}

			Expect(veleroAddPluginsForProvider(context.TODO(), veleroCLI, veleroNamespace, additionalBSLProvider)).To(Succeed())

			installParams.veleroImage = veleroImage
			if installVelero {
				Expect(veleroInstall(context.Background(), veleroImage, veleroNamespace, cloudProvider, objectStoreProvider, useVolumeSnapshots,
					cloudCredentialsFile, bslBucket, bslPrefix, bslConfig, vslConfig, crdsVersion, "", registryCredentialFile)).To(Succeed())
			}

			// Create Secret for additional BSL
			secretName := fmt.Sprintf("bsl-credentials-%s", uuidgen)
			secretKey := fmt.Sprintf("creds-%s", additionalBSLProvider)
			files := map[string]string{
				secretKey: additionalBSLCredentials,
			}

			Expect(createSecretFromFiles(context.TODO(), client, veleroNamespace, secretName, files)).To(Succeed())

			// Create additional BSL using credential
			additionalBsl := fmt.Sprintf("bsl-%s", uuidgen)
			Expect(veleroCreateBackupLocation(context.TODO(),
				veleroCLI,
				veleroNamespace,
				additionalBsl,
				additionalBSLProvider,
				additionalBSLBucket,
				additionalBSLPrefix,
				additionalBSLConfig,
				secretName,
				secretKey,
			)).To(Succeed())

			bsls := []string{"default", additionalBsl}

			for _, bsl := range bsls {
				backupName = fmt.Sprintf("backup-%s", bsl)
				restoreName = fmt.Sprintf("restore-%s", bsl)
				// We limit the length of backup name here to avoid the issue of vsphere plugin https://github.com/vmware-tanzu/velero-plugin-for-vsphere/issues/370
				// We can remove the logic once the issue is fixed
				if bsl == "default" {
					backupName = fmt.Sprintf("%s-%s", backupName, uuidgen)
					restoreName = fmt.Sprintf("%s-%s", restoreName, uuidgen)
				}

				Expect(runKibishiiTests(client, cloudProvider, veleroCLI, veleroNamespace, backupName, restoreName, bsl, useVolumeSnapshots, registryCredentialFile)).To(Succeed(),
					"Failed to successfully backup and restore Kibishii namespace using BSL %s", bsl)
			}
		})
	})
}
