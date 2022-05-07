/*
 *
 * Copyright the Velero contributors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */

package backups

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/vmware-tanzu/velero/test/e2e"
	. "github.com/vmware-tanzu/velero/test/e2e/util/k8s"

	. "github.com/vmware-tanzu/velero/test/e2e/util/providers"
	. "github.com/vmware-tanzu/velero/test/e2e/util/velero"
)

type TTL struct {
	testNS     string
	backupName string
	ctx        context.Context
	ttl        time.Duration
}

func (b *TTL) Init() {
	rand.Seed(time.Now().UnixNano())
	UUIDgen, _ = uuid.NewRandom()
	b.testNS = "backu-ttl-test-" + UUIDgen.String()
	b.backupName = "backu-ttl-test-" + UUIDgen.String()
	b.ctx, _ = context.WithTimeout(context.Background(), time.Duration(time.Minute*10))
	b.ttl = time.Duration(1 * time.Minute)

}

func TTLTest() {
	test := new(TTL)
	client, err := NewTestClient()
	if err != nil {
		println(err.Error())
	}
	Expect(err).To(Succeed(), "Failed to instantiate cluster client for backup tests")
	t, _ := time.ParseDuration("1m0s")
	fmt.Println(t.Round(time.Minute).String())

	BeforeEach(func() {
		flag.Parse()
		if VeleroCfg.InstallVelero {
			Expect(VeleroInstall(context.Background(), &VeleroCfg, false)).To(Succeed())
		}
	})

	AfterEach(func() {
		if VeleroCfg.InstallVelero {
			if !VeleroCfg.Debug {
				Expect(VeleroUninstall(context.Background(), VeleroCfg.VeleroCLI, VeleroCfg.VeleroNamespace)).To(Succeed())
			}
		}
	})

	It("Backups in object storage should be synced to a new Velero successfully", func() {
		test.Init()
		By(fmt.Sprintf("Prepare workload as target to backup by creating namespace %s namespace", test.testNS))
		Expect(CreateNamespace(test.ctx, client, test.testNS)).To(Succeed(),
			fmt.Sprintf("Failed to create %s namespace", test.testNS))

		defer func() {
			Expect(DeleteNamespace(test.ctx, client, test.testNS, false)).To(Succeed(), fmt.Sprintf("Failed to delete the namespace %s", test.testNS))
		}()

		By(fmt.Sprintf("Backup the workload in %s namespace", test.testNS), func() {
			var BackupCfg BackupConfig
			BackupCfg.BackupName = test.backupName
			BackupCfg.Namespace = test.testNS
			BackupCfg.BackupLocation = ""
			BackupCfg.UseVolumeSnapshots = false
			BackupCfg.Selector = ""
			BackupCfg.TTL = test.ttl

			Expect(VeleroBackupNamespace(test.ctx, VeleroCfg.VeleroCLI, VeleroCfg.VeleroNamespace, BackupCfg)).To(Succeed(), func() string {
				RunDebug(context.Background(), VeleroCfg.VeleroCLI, VeleroCfg.VeleroNamespace, test.backupName, "")
				return "Fail to backup workload"
			})
		})

		By("Check TTL was set correctly", func() {
			ttl, err := GetBackupTTL(context.Background(), VeleroCfg.VeleroNamespace, test.backupName)
			Expect(err).NotTo(HaveOccurred(), "Fail to get Azure CSI snapshot checkpoint")
			t, _ := time.ParseDuration(strings.ReplaceAll(ttl, "'", ""))
			fmt.Println(t.Round(time.Minute).String())
			Expect(t).To(Equal(test.ttl))
		})

		By("Waiting period of time for removing backup ralated resources by GC", func() {
			time.Sleep(5 * time.Minute)
		})

		By("Check if backups are deleted as a result of sync from BSL", func() {
			Expect(WaitBackupDeleted(test.ctx, VeleroCfg.VeleroCLI, test.backupName, time.Minute*10)).To(Succeed(), fmt.Sprintf("Failed to check backup %s deleted", test.backupName))
		})

		By("Backup file from cloud object storage should be deleted", func() {
			Expect(ObjectsShouldNotBeInBucket(VeleroCfg.CloudProvider,
				VeleroCfg.CloudCredentialsFile, VeleroCfg.BSLBucket,
				VeleroCfg.BSLPrefix, VeleroCfg.BSLConfig, test.backupName,
				BackupObjectsPrefix, 5)).NotTo(HaveOccurred(), "Fail to get Azure CSI snapshot checkpoint")
		})

		By("PersistentVolume snapshots should be deleted", func() {
		})

		By("Associated Restores should be deleted", func() {
		})
	})
}
