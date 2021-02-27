package terraform_test

import (
	sqlv1alpha1 "github.com/HamzaZo/terrak8s-operator/api/v1alpha1"
	"github.com/HamzaZo/terrak8s-operator/pkg/terraform"
	"github.com/HamzaZo/terrak8s-operator/pkg/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"path/filepath"
)

var _ = Describe("Terraform", func() {
	var (
		dir string
		cr  sqlv1alpha1.PostgreSql
		err error
		val map[string][]byte
		val1 map[string][]byte
	)
	testExpectedBucket := `
{
  "resource": {
    "google_storage_bucket": {
      "bucket": {
        "force_destroy": true,
        "lifecycle_rule": {
          "action": {
            "type": "Delete"
          },
          "condition": {
            "age": 3
          }
        },
        "location": "region-1",
        "name": "my-bucket",
        "project": "my-project",
        "storage_class": "STANDARD"
      }
    }
  }
} 
`
	testExpectedProvider := `
{
  "provider": {
    "google": {
      "project": "my-project",
      "region": "region-1",
      "zone": "zone-1"
    },
    "terraform": {
      "required_providers": {
        "google": {
          "source": "hashicorp/google",
          "version": "3.5.0"
        }
      }
    }
  }
}
`
	testExpectedBackend := `
{
  "terraform": {
    "backend": {
      "gcs": {
        "bucket": "my-bucket",
        "prefix": "test/tfstate"
      }
    }
  }
} 
`
	testExpectedInstanceWithMultipleDbAndUsers := `
{
  "resource": {
    "google_sql_database": {
      "database": {
        "charset": "UTF8",
        "collation": "en_US.UTF8",
        "instance": "my-instance",
        "name": "db-1",
        "project": "my-project"
      }
    }
  },
  "resource": {
    "google_sql_database": {
      "additional_databases": {
        "charset": "UTF8",
        "collation": "en_US.UTF8",
        "instance": "my-instance",
        "name": "db-2",
        "project": "my-project"
      }
    }
  },
  "resource": {
    "google_sql_user": {
      "default": {
        "depends_on": [
          "google_sql_database_instance.instance"
        ],
        "instance": "my-instance",
        "name": "user-1",
        "password": "jEnv2000!",
        "project": "my-project"
      }
    }
  },
  "resource": {
    "google_sql_user": {
      "additional_users": {
        "depends_on": [
          "google_sql_database_instance.instance"
        ],
        "instance": "my-instance",
        "name": "user-2",
        "password": "jEnv2001!",
        "project": "my-project"
      }
    }
  },
  "resource": {
    "google_sql_database_instance": {
      "instance": {
        "database_version": "POSTGRES_9_6",
        "deletion_protection": false,
        "name": "my-instance",
        "project": "my-project",
        "region": "region-1",
        "settings": [
          {
            "activation_policy": "ALWAYS",
            "availability_type": "ZONAL",
            "backup_configuration": {
              "enabled": true,
              "start_time": "21:09"
            },
            "database_flags": [
              {
                "name": "log_min_duration_statement",
                "value": "3000"
              }
            ],
            "disk_autoresize": true,
            "disk_type": "PD_SSD",
            "ip_configuration": {
              "ipv4_enabled": false,
              "private_network": "my-vpc"
            },
            "location_preference": {
              "zone": "zone-1"
            },
            "maintenance_window": {
              "day": 7,
              "hour": 4
            },
            "tier": "db-f1-micro",
            "user_labels": {
              "env": "test"
            }
          }
        ]
      }
    }
  }
}
`
	testExpectedInstance := `
{
  "resource": {
    "google_sql_database": {
      "database": {
        "charset": "UTF8",
        "collation": "en_US.UTF8",
        "instance": "my-instance",
        "name": "db",
        "project": "my-project"
      }
    }
  },
  "resource": {
    "google_sql_user": {
      "default": {
        "depends_on": [
          "google_sql_database_instance.instance"
        ],
        "instance": "my-instance",
        "name": "user-1",
        "password": "jEnv2000!",
        "project": "my-project"
      }
    }
  },
  "resource": {
    "google_sql_database_instance": {
      "instance": {
        "database_version": "POSTGRES_9_6",
        "deletion_protection": false,
        "name": "my-instance",
        "project": "my-project",
        "region": "region-1",
        "settings": [
          {
            "activation_policy": "ALWAYS",
            "availability_type": "ZONAL",
            "backup_configuration": {
              "enabled": true,
              "start_time": "21:09"
            },
            "database_flags": [
              {
                "name": "log_min_duration_statement",
                "value": "3000"
              }
            ],
            "disk_autoresize": true,
            "disk_type": "PD_SSD",
            "ip_configuration": {
              "ipv4_enabled": false,
              "private_network": "my-vpc"
            },
            "location_preference": {
              "zone": "zone-1"
            },
            "maintenance_window": {
              "day": 7,
              "hour": 4
            },
            "tier": "db-f1-micro",
            "user_labels": {
              "env": "test"
            }
          }
        ]
      }
    }
  }
}
`
	testExpectedInstanceWithTwoUsers := `
{
  "resource": {
    "google_sql_database": {
      "database": {
        "charset": "UTF8",
        "collation": "en_US.UTF8",
        "instance": "my-instance",
        "name": "db",
        "project": "my-project"
      }
    }
  },
  "resource": {
    "google_sql_user": {
      "default": {
        "depends_on": [
          "google_sql_database_instance.instance"
        ],
        "instance": "my-instance",
        "name": "user-1",
        "password": "jEnv2000!",
        "project": "my-project"
      }
    }
  },
  "resource": {
    "google_sql_user": {
      "additional_users": {
        "depends_on": [
          "google_sql_database_instance.instance"
        ],
        "instance": "my-instance",
        "name": "user-2",
        "password": "jEnv2001!",
        "project": "my-project"
      }
    }
  },
  "resource": {
    "google_sql_database_instance": {
      "instance": {
        "database_version": "POSTGRES_9_6",
        "deletion_protection": false,
        "name": "my-instance",
        "project": "my-project",
        "region": "region-1",
        "settings": [
          {
            "activation_policy": "ALWAYS",
            "availability_type": "ZONAL",
            "backup_configuration": {
              "enabled": true,
              "start_time": "21:09"
            },
            "database_flags": [
              {
                "name": "log_min_duration_statement",
                "value": "3000"
              }
            ],
            "disk_autoresize": true,
            "disk_type": "PD_SSD",
            "ip_configuration": {
              "ipv4_enabled": false,
              "private_network": "my-vpc"
            },
            "location_preference": {
              "zone": "zone-1"
            },
            "maintenance_window": {
              "day": 7,
              "hour": 4
            },
            "tier": "db-f1-micro",
            "user_labels": {
              "env": "test"
            }
          }
        ]
      }
    }
  }
}
`
	BeforeEach(func() {
		cr = sqlv1alpha1.PostgreSql{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-instance",
				Namespace: "default",
			},
			Spec: sqlv1alpha1.PostgreSqlSpec{
				Project: sqlv1alpha1.PostgresqlInstanceProvider{
					Name:   "my-project",
					Region: "region-1",
					Zone:   "zone-1",
				},
				RemoteState: sqlv1alpha1.PostgresqlInstanceBackend{
					BucketName:   "my-bucket",
					BucketPrefix: "test/tfstate",
				},
				BucketConfig: sqlv1alpha1.PostgresqlInstanceStorageBucket{
					Name:         "my-bucket",
					Project:      "my-project",
					Location:     "region-1",
					Destroy:      true,
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
						Project:   "my-project",
						Name:      "db",
						Charset:   "UTF8",
						Collation: "en_US.UTF8",
						Instance:  "my-instance",
					},
				},
				Users: []sqlv1alpha1.PostgresInstanceDatabaseUsers{
					{
						Name:     "user-1",
						Project:  "my-project",
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
					Project:            "my-project",
					Region:             "region-1",
					Settings: []sqlv1alpha1.PostgresInstanceSettingsSpec{
						{
							MachineType:      "db-f1-micro",
							ActivationPolicy: "ALWAYS",
							AvailabilityType: "ZONAL",
							DiskAutoresize:   true,
							DiskType:         "PD_SSD",
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
								Day:  7,
								Hour: 4,
							},
						},
					},
				},
			},
		}
		val = map[string][]byte{
			"password" : []byte("jEnv2000!"),
		}
		dir, err = util.CreateDirectory(cr.Namespace, cr.Name)
		Expect(err).ToNot(HaveOccurred(), "failed to create directory")
	})

	AfterEach(func() {
		err = util.HouseCleaning(dir)
		Expect(err).ToNot(HaveOccurred(), "failed to clean directory")
	})
	Context("Generate provider/backend", func() {
		It("Should write tf resources to files", func() {
			err = terraform.GenerateProviderAndBackendTF(&cr, filepath.Join(dir, "instance"))
			Expect(err).ToNot(HaveOccurred(), "failed to generate tf files")
			Expect(filepath.Join(dir, "instance") + "/" + "provider.tf.json").Should(BeARegularFile())
			Expect(filepath.Join(dir, "instance") + "/" + "backend.tf.json").Should(BeARegularFile())
		})
		It("Should generate provider tf from CR spec", func() {
			err = terraform.GenerateProviderAndBackendTF(&cr, filepath.Join(dir, "instance"))
			Expect(err).ToNot(HaveOccurred(), "failed to generate tf files")
			b, err := ioutil.ReadFile(filepath.Join(dir, "instance") + "/" + "provider.tf.json")
			Expect(err).ToNot(HaveOccurred(), "cannot read file")
			Expect(string(b)).Should(MatchJSON(testExpectedProvider))
		})
		It("Should generate backend tf from CR spec", func() {
			err = terraform.GenerateProviderAndBackendTF(&cr, filepath.Join(dir, "instance"))
			Expect(err).ToNot(HaveOccurred(), "failed to generate tf files")
			b, err := ioutil.ReadFile(filepath.Join(dir, "instance") + "/" + "backend.tf.json")
			Expect(err).ToNot(HaveOccurred(), "cannot read file")
			Expect(string(b)).Should(MatchJSON(testExpectedBackend))
		})


	})
	Context("Generate bucket", func() {
		It("Should write tf resources to files", func() {
			err = terraform.GenerateBucketTF(&cr, filepath.Join(dir, "bucket"))
			Expect(err).ToNot(HaveOccurred(), "failed to generate tf files")
			Expect(filepath.Join(dir, "bucket") + "/" + "provider.tf.json").Should(BeARegularFile())
			Expect(filepath.Join(dir, "bucket") + "/" + "bucket.tf.json").Should(BeARegularFile())
		})
		It("Should generate bucket tf from CR spec", func() {
			err = terraform.GenerateBucketTF(&cr, filepath.Join(dir, "bucket"))
			Expect(err).ToNot(HaveOccurred(), "failed to generate tf files")
			b, err := ioutil.ReadFile(filepath.Join(dir, "bucket") + "/" + "bucket.tf.json")
			Expect(err).ToNot(HaveOccurred(), "cannot read file")
			Expect(string(b)).Should(MatchJSON(testExpectedBucket))
		})
		It("Should generate provider tf from CR spec", func() {
			err = terraform.GenerateBucketTF(&cr, filepath.Join(dir, "bucket"))
			Expect(err).ToNot(HaveOccurred(), "failed to generate tf files")
			b, err := ioutil.ReadFile(filepath.Join(dir, "bucket") + "/" + "provider.tf.json")
			Expect(err).ToNot(HaveOccurred(), "cannot read file")
			Expect(string(b)).Should(MatchJSON(testExpectedProvider))
		})

	})

	Context("Generate output", func() {
		It("Should write tf resources to files ", func() {
			err = terraform.GenerateTFOutput(filepath.Join(dir, "instance"))
			Expect(err).ToNot(HaveOccurred(), "failed to generate tf files")
			Expect(filepath.Join(dir, "instance") + "/" + "output.tf").Should(BeARegularFile())
		})
	})

	Context("Generate instance", func() {
		It("Should write tf resources to files ", func() {
			err = terraform.GenerateTFInstance(&cr, filepath.Join(dir, "instance"), val)
			Expect(err).ToNot(HaveOccurred(), "failed to generate tf files")
			Expect(filepath.Join(dir, "instance") + "/" + "main.tf.json").Should(BeARegularFile())
		})
		It("Should generate instance tf with one database and user", func() {
			err = terraform.GenerateTFInstance(&cr, filepath.Join(dir, "instance"), val)
			Expect(err).ToNot(HaveOccurred(), "failed to generate tf files")
			b, err := ioutil.ReadFile(filepath.Join(dir, "instance") + "/" + "main.tf.json")
			Expect(err).ToNot(HaveOccurred(), "cannot read file")
			Expect(string(b)).Should(MatchJSON(testExpectedInstance))
		})

	})
	Context ("Generate instance with multiple users", func() {
		BeforeEach(func() {
			instance := sqlv1alpha1.PostgreSql{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-instance",
					Namespace: "default",
				},
				Spec: sqlv1alpha1.PostgreSqlSpec{
					Project: sqlv1alpha1.PostgresqlInstanceProvider{
						Name:   "my-project",
						Region: "region-1",
						Zone:   "zone-1",
					},
					RemoteState: sqlv1alpha1.PostgresqlInstanceBackend{
						BucketName:   "my-bucket",
						BucketPrefix: "test/tfstate",
					},
					BucketConfig: sqlv1alpha1.PostgresqlInstanceStorageBucket{
						Name:         "my-bucket",
						Project:      "my-project",
						Location:     "region-1",
						Destroy:      true,
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
							Project:   "my-project",
							Name:      "db",
							Charset:   "UTF8",
							Collation: "en_US.UTF8",
							Instance:  "my-instance",
						},
					},
					Users: []sqlv1alpha1.PostgresInstanceDatabaseUsers{
						{
							Name:     "user-1",
							Project:  "my-project",
							Instance: "my-instance",
							Password: sqlv1alpha1.PostgresInstanceDatabasePassword{
								SecretKeyRef: sqlv1alpha1.PostgresInstanceDatabasePasswordSpec{
									Name: "mypassword",
									Key:  "mykey1",
								},
							},
						},
						{
							Name:     "user-2",
							Project:  "my-project",
							Instance: "my-instance",
							Password: sqlv1alpha1.PostgresInstanceDatabasePassword{
								SecretKeyRef: sqlv1alpha1.PostgresInstanceDatabasePasswordSpec{
									Name: "mypassword",
									Key:  "mykey2",
								},
							},
						},
					},
					SqlInstance: sqlv1alpha1.PostgresqlInstanceSpec{
						DataBaseVersion:    "POSTGRES_9_6",
						DeletionProtection: false,
						Name:               "my-instance",
						Project:            "my-project",
						Region:             "region-1",
						Settings: []sqlv1alpha1.PostgresInstanceSettingsSpec{
							{
								MachineType:      "db-f1-micro",
								ActivationPolicy: "ALWAYS",
								AvailabilityType: "ZONAL",
								DiskAutoresize:   true,
								DiskType:         "PD_SSD",
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
									Day:  7,
									Hour: 4,
								},
							},
						},
					},
				},
			}
			val1 = map[string][]byte{
				"password" : []byte("jEnv2000!"),
				"password2": []byte("jEnv2001!"),
			}
			dir, err = util.CreateDirectory(instance.Namespace, instance.Name)
			Expect(err).ToNot(HaveOccurred(), "failed to create directory")
		})
		It("Should generate instance tf with two users and one database", func() {
			err = terraform.GenerateTFInstance(&cr, filepath.Join(dir, "instance"), val1)
			Expect(err).ToNot(HaveOccurred(), "failed to generate tf files")
			b, err := ioutil.ReadFile(filepath.Join(dir, "instance") + "/" + "main.tf.json")
			Expect(err).ToNot(HaveOccurred(), "cannot read file")
			Expect(string(b)).Should(MatchJSON(testExpectedInstanceWithTwoUsers))
		})
	})
	Context ("Generate instance with multiple users and databases", func() {
		BeforeEach(func() {
			instance2 := sqlv1alpha1.PostgreSql{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-instance",
					Namespace: "default",
				},
				Spec: sqlv1alpha1.PostgreSqlSpec{
					Project: sqlv1alpha1.PostgresqlInstanceProvider{
						Name:   "my-project",
						Region: "region-1",
						Zone:   "zone-1",
					},
					RemoteState: sqlv1alpha1.PostgresqlInstanceBackend{
						BucketName:   "my-bucket",
						BucketPrefix: "test/tfstate",
					},
					BucketConfig: sqlv1alpha1.PostgresqlInstanceStorageBucket{
						Name:         "my-bucket",
						Project:      "my-project",
						Location:     "region-1",
						Destroy:      true,
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
							Project:   "my-project",
							Name:      "db-1",
							Charset:   "UTF8",
							Collation: "en_US.UTF8",
							Instance:  "my-instance",
						},
						{
							Project:   "my-project",
							Name:      "db-2",
							Charset:   "UTF8",
							Collation: "en_US.UTF8",
							Instance:  "my-instance",
						},
					},
					Users: []sqlv1alpha1.PostgresInstanceDatabaseUsers{
						{
							Name:     "user-1",
							Project:  "my-project",
							Instance: "my-instance",
							Password: sqlv1alpha1.PostgresInstanceDatabasePassword{
								SecretKeyRef: sqlv1alpha1.PostgresInstanceDatabasePasswordSpec{
									Name: "mypassword",
									Key:  "mykey1",
								},
							},
						},
						{
							Name:     "user-2",
							Project:  "my-project",
							Instance: "my-instance",
							Password: sqlv1alpha1.PostgresInstanceDatabasePassword{
								SecretKeyRef: sqlv1alpha1.PostgresInstanceDatabasePasswordSpec{
									Name: "mypassword",
									Key:  "mykey2",
								},
							},
						},
					},
					SqlInstance: sqlv1alpha1.PostgresqlInstanceSpec{
						DataBaseVersion:    "POSTGRES_9_6",
						DeletionProtection: false,
						Name:               "my-instance",
						Project:            "my-project",
						Region:             "region-1",
						Settings: []sqlv1alpha1.PostgresInstanceSettingsSpec{
							{
								MachineType:      "db-f1-micro",
								ActivationPolicy: "ALWAYS",
								AvailabilityType: "ZONAL",
								DiskAutoresize:   true,
								DiskType:         "PD_SSD",
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
									Day:  7,
									Hour: 4,
								},
							},
						},
					},
				},
			}
			dir, err = util.CreateDirectory(instance2.Namespace, instance2.Name)
			Expect(err).ToNot(HaveOccurred(), "failed to create directory")
		})
		It("Should generate instance tf with two users and two database", func() {
			err = terraform.GenerateTFInstance(&cr, filepath.Join(dir, "instance"), val1)
			Expect(err).ToNot(HaveOccurred(), "failed to generate tf files")
			b, err := ioutil.ReadFile(filepath.Join(dir, "instance") + "/" + "main.tf.json")
			Expect(err).ToNot(HaveOccurred(), "cannot read file")
			Expect(string(b)).Should(MatchJSON(testExpectedInstanceWithMultipleDbAndUsers))
		})
	})

})

