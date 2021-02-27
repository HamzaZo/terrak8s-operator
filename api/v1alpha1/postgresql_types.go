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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ObjectPhase string

const (
	// PhaseRunning means that the sql resource is active and ready to receive traffic
	PhaseRunning ObjectPhase = "Running"
	// PhaseApplying means that the sql resource is currently applying
	PhaseApplying ObjectPhase = "Applying"
	// PhaseInitializing means that the sql resource is initializing
	PhaseInitializing ObjectPhase = "Initializing"
	// PhaseFailed means that the sql resource creation process is failed due to some reason
	PhaseFailed ObjectPhase = "Failed"
	// PhaseDestroying means that the sql resource are being destroying
	PhaseDestroying ObjectPhase = "Destroying"
)

// PostgreSqlSpec defines the desired state of PostgreSql
type PostgreSqlSpec struct {
	Project     PostgresqlInstanceProvider `json:"project"`
	RemoteState PostgresqlInstanceBackend  `json:"remoteState"`
	// +optional
	BucketConfig PostgresqlInstanceStorageBucket `json:"bucketConfig"`
	SqlInstance  PostgresqlInstanceSpec          `json:"sqlInstance"`
	Databases    []PostgresInstanceDatabases     `json:"databases"`
	Users        []PostgresInstanceDatabaseUsers `json:"users,omitempty"`
}

//PostgresqlInstanceSpec define the sql instance
type PostgresqlInstanceSpec struct {
	//DataBaseVersion define the PostgreSQL version to use
	DataBaseVersion string `json:"databaseVersion" tf:"database_version"`
	// +optional
	DeletionProtection bool `json:"deletionProtection" tf:"deletion_protection"`
	//The name of the Cloud SQL instance
	// +optional
	Name string `json:"name" tf:"name"`
	// Project the ID of the project in which the resource belongs
	// +optional
	Project string `json:"project" tf:"project"`
	// Region the instance will sit in
	// +optional
	Region string `json:"region" tf:"region"`
	// Settings to use to configure the database
	Settings []PostgresInstanceSettingsSpec `json:"settings" tf:"settings"`
}

//PostgresqlInstanceProvider define information about gcp tenant
type PostgresqlInstanceProvider struct {
	//Name define the project name
	Name string `json:"name" tf:"project"`
	// Region the instance will sit in
	// +optional
	Region string `json:"region" tf:"region"`
	// Zone define the preferred compute engine zone.
	// +optional
	Zone string `json:"zone" tf:"zone"`
}

//PostgresqlInstanceBackend define gcp bucket config
type PostgresqlInstanceBackend struct {
	//BucketName define the name of the GCS bucket
	BucketName string `json:"bucketName" tf:"bucket"`
	//BucketPrefix GCS prefix inside the bucket
	BucketPrefix string `json:"bucketPrefix" tf:"prefix"`
}

//PostgresqlInstanceStorageBucket define gcp bucket config
type PostgresqlInstanceStorageBucket struct {
	//Name define the name of the GCS bucket
	// +optional
	Name string `json:"name" tf:"name"`
	// Project the ID of the project in which the resource belongs
	// +optional
	Project string `json:"project" tf:"project"`
	//Location define the GCS bucket location
	// +optional
	Location string `json:"location" tf:"location"`
	//Destroy When deleting a bucket, this boolean option will delete all contained objects
	// +optional
	Destroy bool `json:"destroy" tf:"force_destroy"`
	//StorageClass define the storage class of the bucket
	// +optional
	StorageClass string `json:"storageClass" tf:"storage_class"`
	//LifecycleRules define the bucket Lifecycle Rules configuration
	// +optional
	LifecycleRule PostgresqlInstanceStorageBucketLifecycleRules `json:"lifecycleRule" tf:"lifecycle_rule"`
}

type PostgresqlInstanceStorageBucketLifecycleRules struct {
	//Condition define the Lifecycle Rule's condition configuration
	// +optional
	Condition map[string]int `json:"condition" tf:"condition"`
	//Action define the Lifecycle Rule's action configuration
	// +optional
	Action map[string]string `json:"action" tf:"action"`
}

//PostgresInstanceDatabases define databases config in sql instance
type PostgresInstanceDatabases struct {
	// Project the ID of the project in which the resource belongs
	// +optional
	Project string `json:"project" tf:"project"`
	//The name of the database in the Cloud SQL instance
	Name string `json:"name" tf:"name"`
	// The charset value
	// +optional
	Charset string `json:"charset" tf:"charset"`
	// +optional
	Collation string `json:"collation" tf:"collation"`
	//The name of the Cloud SQL instance
	// +optional
	Instance string `json:"instance" tf:"instance"`
}

//PostgresInstanceSettingsSpec define sql instance settings
type PostgresInstanceSettingsSpec struct {
	//The machine type to use
	// +optional
	MachineType string `json:"machineType" tf:"tier"`
	//The availability type of the Cloud SQL instance, high availability (REGIONAL) or single zone (ZONAL)
	// +optional
	AvailabilityType string `json:"availabilityType" tf:"availability_type"`
	//Configuration to increase storage size automatically.
	// +optional
	DiskAutoresize bool `json:"diskAutoresize" tf:"disk_autoresize"`
	//The type of data disk: PD_SSD or PD_HDD
	// +optional
	DiskType string `json:"diskType" tf:"disk_type"`
	//Specify when the instance should be active. Can be either ALWAYS, NEVER or ON_DEMAND
	// +optional
	ActivationPolicy string `json:"activationPolicy" tf:"activation_policy"`
	//A set of key/value user label pairs to assign to the instance.
	// +optional
	Labels map[string]string `json:"labels,omitempty" tf:"user_labels,omitempty"`
	// +optional
	DatabaseFlags []PostgresInstanceSettingsDatabaseFlags `json:"databaseFlags,omitempty" tf:"database_flags,omitempty"`
	//Network configuration
	IpConfiguration PostgresInstanceSettingsIpConfiguration `json:"ipConfiguration" tf:"ip_configuration"`
	//Backup configuration
	BackupConfiguration PostgresInstanceSettingsBackupConfiguration `json:"backupConfiguration" tf:"backup_configuration"`
	// +optional
	LocationPreference PostgresInstanceSettingsLocationPreference `json:"locationPreference" tf:"location_preference"`
	//Maintenance configuration for automatically apply maintenance to instance
	MaintenanceWindow PostgresInstanceSettingsMaintenanceWindow `json:"maintenanceWindow" tf:"maintenance_window"`
}

//PostgresInstanceDatabaseUsers contains database users spec
type PostgresInstanceDatabaseUsers struct {
	//The name of the user.
	Name string `json:"name" tf:"name"`
	// Project the ID of the project in which the resource belongs
	// +optional
	Project string `json:"project" tf:"project"`
	//The name of the Cloud SQL instance
	// +optional
	Instance string `json:"instance" tf:"instance"`
	//The Password of the Cloud SQL instance user
	Password PostgresInstanceDatabasePassword `json:"password" tf:"password"`
}

//PostgresInstanceDatabasePassword contains database password key
type PostgresInstanceDatabasePassword struct {
	//SecretKeyRef Selects a key of a secret in the pod's namespace.
	SecretKeyRef PostgresInstanceDatabasePasswordSpec `json:"secretKeyRef"`
}

//PostgresInstanceDatabasePasswordSpec holds password spec
type PostgresInstanceDatabasePasswordSpec struct {
	//The Name of the secret
	Name string `json:"name"`
	// The Key of the secret to select from.  Must be a valid secret key.
	Key string `json:"key"`
}

//PostgresInstanceSettingsDatabaseFlags define sql instance flags
type PostgresInstanceSettingsDatabaseFlags struct {
	// +optional
	Name string `json:"name,omitempty" tf:"name,omitempty"`
	// +optional
	Value string `json:"value,omitempty" tf:"value,omitempty"`
}

//PostgresInstanceSettingsIpConfiguration define instance network configuration
type PostgresInstanceSettingsIpConfiguration struct {
	//Whether this Cloud SQL instance should be assigned a public IPV4 address.
	// +optional
	Ipv4Enabled bool `json:"ipv4Enabled" tf:"ipv4_enabled"`
	//The VPC network from which the Cloud SQL instance is accessible
	PrivateNetwork string `json:"privateNetwork" tf:"private_network"`
	//Whether SSL connections over IP are enforced or not.
	// +optional
	RequireSSL bool `json:"requireSSL,omitempty" tf:"require_ssl,omitempty"`
}

//PostgresInstanceSettingsBackupConfiguration enable backup configuration
type PostgresInstanceSettingsBackupConfiguration struct {
	//Enabled backup configuration
	Enabled bool `json:"enabled" tf:"enabled"`
	//StartTime HH:MM format time indicating when backup configuration starts.
	StartTime string `json:"startTime" tf:"start_time"`
}

type PostgresInstanceSettingsLocationPreference struct {
	//A GAE application whose zone to remain in. Must be in the same region as this instance.
	// +optional
	FollowGaeApplication string `json:"followGaeApplication,omitempty" tf:"follow_gae_application,omitempty"`
	// Zone define the preferred compute engine zone.
	// +optional
	Zone string `json:"zone" tf:"zone"`
}

//PostgresInstanceSettingsMaintenanceWindow define instance maintenance config
type PostgresInstanceSettingsMaintenanceWindow struct {
	//Day of week (1-7)
	Day int64 `json:"day" tf:"day"`
	//Hour of day (0-23)
	Hour int64 `json:"hour" tf:"hour"`
}

//PostgresInstanceOutput define instance connection parameters
type PostgresInstanceOutput struct {
	//The connection name of the instance to be used in connection strings
	// +optional
	ConnectionName string `json:"connectionName,omitempty"`
	//The private IPv4 address assigned to the instance
	// +kubebuilder:default:=<pending>
	// +optional
	ConnectionIPAddress string `json:"connectionIPAddress,omitempty"`
}

// PostgreSqlStatus defines the observed state of PostgreSql
type PostgreSqlStatus struct {
	// +optional
	Phase ObjectPhase `json:"phase,omitempty"`
	// +optional
	Output PostgresInstanceOutput `json:"output,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="DatabaseVersion",type=string,JSONPath=`.spec.sqlInstance.databaseVersion`
// +kubebuilder:printcolumn:name="InstanceIP",type=string,JSONPath=`.status.output.connectionIPAddress`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:resource:shortName=pg

// PostgreSql is the Schema for the postgresqls API
type PostgreSql struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PostgreSqlSpec   `json:"spec,omitempty"`
	Status PostgreSqlStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PostgreSqlList contains a list of PostgreSql
type PostgreSqlList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PostgreSql `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PostgreSql{}, &PostgreSqlList{})
}
