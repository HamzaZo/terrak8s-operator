package terraform_test

import (
	sqlv1alpha1 "github.com/HamzaZo/terrak8s-operator/api/v1alpha1"
	"github.com/HamzaZo/terrak8s-operator/pkg/terraform"
	"github.com/HamzaZo/terrak8s-operator/pkg/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"path/filepath"
)

var _ = Describe("Terraform", func() {
	var (
		str string
		cr  sqlv1alpha1.PostgreSql
		err error
	)
	BeforeEach(func() {
		cr = sqlv1alpha1.PostgreSql{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-instance",
				Namespace: "default",
			},
			Spec: sqlv1alpha1.PostgreSqlSpec{
				Project: sqlv1alpha1.PostgresqlInstanceProvider{
					Name:   "my-tenant",
					Region: "region-1",
					Zone:   "zone-1",
				},
				RemoteState: sqlv1alpha1.PostgresqlInstanceBackend{
					BucketName:   "my-bucket",
					BucketPrefix: "test/tfstate",
				},
				BucketConfig: sqlv1alpha1.PostgresqlInstanceStorageBucket{
					Name:         "my-bucket",
					Project:      "my-tenant",
					Location:     "region-1",
					Destroy:      false,
					StorageClass: "STANDARD",
					LifecycleRule: sqlv1alpha1.PostgresqlInstanceStorageBucketLifecycleRules{
						Condition: map[string]int{
							"age": 3,
						},
						Action: map[string]string{
							"type": "Delete",
						},
					},
				},
				Databases: []sqlv1alpha1.PostgresInstanceDatabases{
					{
						Project:   "my-tenant",
						Name:      "db",
						Charset:   "UTF8",
						Collation: "en_US.UTF8",
						Instance:  "my-instance",
					},
				},
				Users: []sqlv1alpha1.PostgresInstanceDatabaseUsers{
					{
						Name:     "user1",
						Project:  "my-tenant",
						Instance: "my-instance",
						Password: sqlv1alpha1.PostgresInstanceDatabasePassword{
							SecretKeyRef: sqlv1alpha1.PostgresInstanceDatabasePasswordSpec{
								Name: "mypassword",
								Key:  "mykey",
							},
						},
					},
				},
				SqlInstance: sqlv1alpha1.PostgresqlInstanceSpec{
					DataBaseVersion:    "POSTGRES_9_6",
					DeletionProtection: false,
					Name:               "my-instance",
					Project:            "my-tenant",
					Region:             "region-1",
					Settings: []sqlv1alpha1.PostgresInstanceSettingsSpec{
						{
							MachineType:      "db-f1-micro",
							ActivationPolicy: "ALWAYS",
							AvailabilityType: "ZONAL",
							DiskAutoresize:   true,
							DiskType:         "SSD",
							Labels: map[string]string{
								"env": "test",
							},
							DatabaseFlags: []sqlv1alpha1.PostgresInstanceSettingsDatabaseFlags{
								{
									Name:  "log_min_duration_statement",
									Value: "3000",
								},
							},
							IpConfiguration: sqlv1alpha1.PostgresInstanceSettingsIpConfiguration{
								Ipv4Enabled:    false,
								PrivateNetwork: "my-vpc",
								RequireSSL:     false,
							},
							BackupConfiguration: sqlv1alpha1.PostgresInstanceSettingsBackupConfiguration{
								Enabled:   true,
								StartTime: "21:09",
							},
							LocationPreference: sqlv1alpha1.PostgresInstanceSettingsLocationPreference{
								Zone: "zone-1",
							},
							MaintenanceWindow: sqlv1alpha1.PostgresInstanceSettingsMaintenanceWindow{
								Day:  GinkgoRandomSeed(),
								Hour: GinkgoRandomSeed(),
							},
						},
					},
				},
			},
		}
		str, err = util.CreateDirectory(cr.Namespace, cr.Name)
		Expect(err).ToNot(HaveOccurred(), "failed to create directory")
	})
	AfterEach(func() {
		err = util.HouseCleaning(str)
		Expect(err).ToNot(HaveOccurred(), "failed to clean directory")
	})
	Context("Generate provider/backend", func() {
		It("generate provider and backend tf from CR spec", func() {
			err = terraform.GenerateProviderAndBackendTF(&cr, filepath.Join(str, "instance"))
			Expect(err).ToNot(HaveOccurred(), "failed to generate tf files")
			Expect(filepath.Join(str, "instance") + "/" + "provider.tf.json").Should(BeARegularFile())
			Expect(filepath.Join(str, "instance") + "/" + "backend.tf.json").Should(BeARegularFile())
		})
	})
	Context("Generate bucket", func() {
		It("generate bucket tf from CR spec", func() {
			err = terraform.GenerateBucketTF(&cr, filepath.Join(str, "bucket"))
			Expect(err).ToNot(HaveOccurred(), "failed to generate tf files")
			Expect(filepath.Join(str, "bucket") + "/" + "provider.tf.json").Should(BeARegularFile())
			Expect(filepath.Join(str, "bucket") + "/" + "bucket.tf.json").Should(BeARegularFile())
		})
	})

	Context("Generate output", func() {
		It("generate output tf ", func() {
			err = terraform.GenerateTFOutput(filepath.Join(str, "instance"))
			Expect(err).ToNot(HaveOccurred(), "failed to generate tf files")
			Expect(filepath.Join(str, "instance") + "/" + "output.tf").Should(BeARegularFile())
		})
	})

	Context("Generate instance", func() {
		It("generate instance sql tf ", func() {
			val := map[string][]byte{
				"password": []byte("jEnv2000!"),
			}
			err = terraform.GenerateTFInstance(&cr, filepath.Join(str, "instance"), val)
			Expect(err).ToNot(HaveOccurred(), "failed to generate tf files")
			Expect(filepath.Join(str, "instance") + "/" + "main.tf.json").Should(BeARegularFile())
		})
	})

})
