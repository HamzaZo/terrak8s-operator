/*
Copyright 2020 The Terrak8s-operator authors.

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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var postgresqllog = logf.Log.WithName("postgresql-resource")

func (r *PostgreSql) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-sql-terrak8s-io-v1alpha1-postgresql,mutating=true,failurePolicy=fail,groups=sql.terrak8s.io,resources=postgresqls,verbs=create;update,versions=v1alpha1,name=webhook-mutate.terrak8s.io

var _ webhook.Defaulter = &PostgreSql{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *PostgreSql) Default() {
	postgresqllog.Info("createdResource", "namespace", r.Namespace, "name", r.Name)

	if r.Spec.Project.Region == "" {
		r.Spec.Project.Region = "europe-west1"
	}
	if r.Spec.Project.Zone == "" {
		r.Spec.Project.Zone = "europe-west1-b"
	}

	SetDefaultBucketSpec(&r.Spec.BucketConfig, r.Spec.RemoteState.BucketName, r.Spec.Project.Name)

	for k := range r.Spec.SqlInstance.Settings {
		a := &r.Spec.SqlInstance.Settings[k]
		SetIpConfigurationDefaultSpec(&a.IpConfiguration)
		SetDatabaseInstanceSettingsSpec(a)
		SetLocationPreferenceDefaultSpec(&a.LocationPreference)
	}

	for k := range r.Spec.Users {
		x := &r.Spec.Users[k]
		SetDatabaseUserDefaultSpec(x, r.Name, r.Spec.Project.Name)
	}
	for k := range r.Spec.Databases {
		x := &r.Spec.Databases[k]
		SetDatabaseDefaultSpec(x, r.Name, r.Spec.Project.Name)
	}

	SetDefaultSqlInstanceSpec(&r.Spec.SqlInstance, r.Name, r.Spec.Project.Name)
}

func SetDefaultBucketSpec(obj *PostgresqlInstanceStorageBucket, name string, project string) {
	if obj.Name == "" {
		obj.Name = name
	}
	if obj.Project == "" {
		obj.Project = project
	}
	if obj.Location == "" {
		obj.Location = "europe-west1"
	}
	if obj.StorageClass == "" {
		obj.StorageClass = "STANDARD"
	}
	obj.Destroy = true

	p := PostgresqlInstanceStorageBucketLifecycleRules{}
	SetDefaultBucketLifecycleRules(&p)
	obj.LifecycleRule = p
}

func SetDefaultBucketLifecycleRules(obj *PostgresqlInstanceStorageBucketLifecycleRules) {
	if obj.Action == nil {
		obj.Action = map[string]string{
			"type": "Delete",
		}
	}
	if obj.Condition == nil {
		obj.Condition = map[string]int{
			"age": 3,
		}
	}
}

func SetDatabaseUserDefaultSpec(obj *PostgresInstanceDatabaseUsers, name string, project string) {
	if obj.Instance == "" {
		obj.Instance = name
	}
	if obj.Project == "" {
		obj.Project = project
	}
}

func SetDatabaseDefaultSpec(obj *PostgresInstanceDatabases, name string, project string) {
	if obj.Instance == "" {
		obj.Instance = name
	}
	if obj.Project == "" {
		obj.Project = project
	}
	if obj.Charset == "" {
		obj.Charset = "UTF8"
	}
	if obj.Collation == "" {
		obj.Collation = "en_US.UTF8"
	}
}

func SetLocationPreferenceDefaultSpec(obj *PostgresInstanceSettingsLocationPreference) {
	if obj.Zone == "" {
		obj.Zone = "europe-west1-b"
	}
}

func SetIpConfigurationDefaultSpec(obj *PostgresInstanceSettingsIpConfiguration) {
	obj.Ipv4Enabled = false
}

func SetDatabaseInstanceSettingsSpec(obj *PostgresInstanceSettingsSpec) {
	if obj.AvailabilityType == "" {
		obj.AvailabilityType = "ZONAL"
	}
	if obj.DiskType == "" {
		obj.DiskType = "PD_SSD"
	}
	if obj.ActivationPolicy == "" {
		obj.ActivationPolicy = "ALWAYS"
	}
	if obj.MachineType == "" {
		obj.MachineType = "db-f1-micro"
	}
	obj.DiskAutoresize = true
}

func SetDefaultSqlInstanceSpec(obj *PostgresqlInstanceSpec, name string, project string) {
	if obj.Region == "" {
		obj.Region = "europe-west1"
	}
	if obj.Project == "" {
		obj.Project = project
	}
	if obj.Name == "" {
		obj.Name = name
	}
	obj.DeletionProtection = false
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-sql-terrak8s-io-v1alpha1-postgresql,mutating=false,failurePolicy=fail,groups=sql.terrak8s.io,resources=postgresqls,versions=v1alpha1,name=webhook-validator.terrak8s.io

var _ webhook.Validator = &PostgreSql{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *PostgreSql) ValidateCreate() error {
	postgresqllog.Info("validate on create", "namespace", r.Namespace, "name", r.Name)

	return r.validatePostgresInstance()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *PostgreSql) ValidateUpdate(old runtime.Object) error {
	postgresqllog.Info("validate on update", "namespace", r.Namespace, "name", r.Name)

	return r.validatePostgresInstance()

}

// ValidateDelete implements webhook.Validator, we do nothing in ValidateDelete, since we don’t need
//to validate anything on deletion.
func (r *PostgreSql) ValidateDelete() error {
	postgresqllog.Info("validate on delete", "name", r.Name)

	return nil
}

func (r *PostgreSql) validatePostgresInstance() error {
	var allErrs field.ErrorList
	if err := r.validatePostgresInstanceName(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.validatePostgresInstanceSettings(); err != nil {
		allErrs = append(allErrs, err)
	}

	if len(allErrs) == 0 {
		return nil
	}

	return errors.NewInvalid(
		schema.GroupKind{Group: "sql.terrak8s.io", Kind: "v1alpha1"}, r.Name, allErrs)
}

func (r *PostgreSql) validatePostgresInstanceName() *field.Error {
	if len(r.Name) > validation.DNS1035LabelMaxLength {
		return field.Invalid(field.NewPath("metadata").Child("name"),
			r.Name, "Instance name must be less than 63 character")
	}
	return nil
}

func (r *PostgreSql) validatePostgresInstanceSettings() *field.Error {
	validateVersion := []string{"POSTGRES_9_6", "POSTGRES_10", "POSTGRES_11", "POSTGRES_12"}
	version := r.Spec.SqlInstance.DataBaseVersion
	if !ContainsVersion(validateVersion, version) {
		return field.Invalid(field.NewPath("spec").Child("sqlInstance").Child("databaseVersion"),
			version, "Invalid version, supported version are: POSTGRES_9_6, POSTGRES_10, POSTGRES_11, POSTGRES_12")
	}
	return nil

}

//ContainsVersion is helper func
func ContainsVersion(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
