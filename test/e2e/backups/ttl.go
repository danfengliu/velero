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
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/vmware-tanzu/velero/test/e2e"
	. "github.com/vmware-tanzu/velero/test/e2e/util/k8s"

	//. "github.com/vmware-tanzu/velero/test/e2e/util/providers"
	. "github.com/vmware-tanzu/velero/test/e2e/util/velero"
)

type TTL struct {
	testNS     string
	backupName string
	ctx        context.Context
}

func (b *TTL) Init() {
	rand.Seed(time.Now().UnixNano())
	UUIDgen, _ = uuid.NewRandom()
	b.testNS = "backu-ttl-test-" + UUIDgen.String()
	b.backupName = "backu-ttl-test-" + UUIDgen.String()
	b.ctx, _ = context.WithTimeout(context.Background(), time.Duration(time.Minute*10))
}

func TTLTest() {
	test := new(TTL)
	client, err := NewTestClient()
	if err != nil {
		println(err.Error())
	}
	Expect(err).To(Succeed(), "Failed to instantiate cluster client for backup tests")

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
			BackupCfg.TTL = time.Duration(1 * time.Minute)

			Expect(VeleroBackupNamespace(test.ctx, VeleroCfg.VeleroCLI, VeleroCfg.VeleroNamespace, BackupCfg)).To(Succeed(), func() string {
				RunDebug(context.Background(), VeleroCfg.VeleroCLI, VeleroCfg.VeleroNamespace, test.backupName, "")
				return "Fail to backup workload"
			})
		})

		By("Check all backups in object storage are synced to Velero", func() {

		})
	})
}
