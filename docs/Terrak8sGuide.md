## How to use terrak8s

### Prerequisites:
Make sure that :
* You have access to k8s cluster.
* You create a k8s secret to store the GCP serviceAccount json key to authenticate against the GCP project, the secret key name should be `whatever_name_of_sa_key.json`.
* You create a k8s secret to hold database users passwords.
    * beer in mind that passwords should respect the following rules:
        - at least 7 letters
        - at least 1 number
        - at least 1 upper case
        - at least 1 special character

**Important Note:**
- Keep in mind that Google has made some restriction about cloud SQL instance name, you cannot reuse the same name of the
  deleted instance until one week from the deletion date.
- GCS bucket names are globally unique, which mean that no two buckets may hold the same name [here](https://cloud.google.com/storage/docs/naming-buckets). 

### Create Postgresql instance
The following is an example of a Postgresql. It creates a cloud sql postgres instance:

```yaml
apiVersion: sql.terrak8s.io/v1alpha1
kind: PostgreSql
metadata:
  name: my-instance
  namespace: demo
spec:
  project:
    name: "my-project"
  remoteState:
    bucketName: "my-bucket"
    bucketPrefix: "dev/tfstate"
  databases:
    - name: "sample-db1"
    - name: "sample-db2"
  users:
    - name: "user-1"
      password:
        secretKeyRef:
          name: cred-database
          key: database-password-1
    - name: "user-2"
      password:
        secretKeyRef:
          name: cred-database
          key: database-password-2
  sqlInstance:
    databaseVersion: "POSTGRES_9_6"
    settings:
      - machineType: "db-f1-micro"
        labels:
          name: "environment"
          value: "dev01"
        ipConfiguration:
          privateNetwork: "my-vpc"
        databaseFlags:
          - name: "log_min_duration_statement"
            value: "3000"
        backupConfiguration:
          startTime: "21:30"
          enabled: true
        maintenanceWindow:
          day: 7
          hour: 4
```
In this example:

* A PostgreSql named `my-instance` is created, indicated by the `.metadata.name` field.
* The `.spec.project.name` indicate google project name for Cloud SQL instance.
* The `.spec.remoteState` define gcp bucket that should be used to create tfstate.
    * **Note:** 
        - Terrak8s creates automatically GCS bucket to store Cloud SQL instance tfstate on it.
* The PostgreSql create a databases with name `sample-db1` and `sample-db2` indicated by `.spec.database.name`. Also, database users `user-1` and `user-2`, indicated by `.spec.users.name` field.
* The `.spec.sqlInstance.databaseVersion` define the postgres version, bear in mind that postgres supported version are **POSTGRES_9_6, POSTGRES_10, POSTGRES_11, POSTGRES_12**
* The `.spec.sqlInstance.settings` define Cloud SQL instance configuration which contains the following sub-fields:
    * The `.settings.machineType` indicate the machine type.
    * The `.settings.labels` field is a single {key,value}, that describe labels to assign the instance.
    * The `.settings.ipConfiguration.privateNetwork` represent the VPC network from which the Cloud SQL instance is accessible.
    * The `.settings.databaseFlags` field is a map of {key,value} pairs, that indicate Cloud SQL instance flags.
    * The `.settings.backupConfiguration` define backup configuration and when it should start.
    * The `.settings.maintenance` define maintenance configuration for automatically apply maintenance to the instance.
    
Follow the steps given below to create the above PostgreSql:

1.Create the PostgreSql in namespace `demo` by running the following command:
```shell
$ kubectl apply -f examples/postgresql-instance.yaml
```

2. Run `kubectl get postgresql -n demo` or (`pg` as shortName) to check if the PostgreSql was created.

If the PostgreSql is still being created, the output is similar to the following:

```shell
$ kubectl get postgresql -n demo 
NAME          PHASE      DATABASEVERSION   INSTANCEIP    AGE
my-instance   Applying    POSTGRES_9_6      <pending>    18s
```

When you inspect the PostgreSqls in your namespace, the following fields are displayed:

* `NAME` lists the names of the PostgreSqls in the namespace.
* `PHASE` displays the process creation phases.
* `DATABASEVERSION` displays the database version.
* `INSTANCEIP` displays the instance private ip for connection.
* `AGE` displays the amount of time that the application has been running.

The creation process takes a few minutes, depending on the C4 network. (*hints* use `watch -n 1 kubectl get pg -n demo` to track the progress of Postgresql creation)

3. Run `kubectl get postgresql -n demo` again a few minutes later. The output is similar to this:

```shell
$ kubectl get postgresql -n demo 
NAME          PHASE      DATABASEVERSION   INSTANCEIP    AGE
my-instance   Running     POSTGRES_9_6     192.168.0.12   9m
```

4. Get details of your PostgreSql by running `kubectl describe pg -n demo`

The output is similar to this:
```
Name:         my-instance
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  sql.terrak8s.io/v1alpha1
Kind:         PostgreSql
Metadata:
  Creation Timestamp:  2021-01-15T21:29:39Z
  Finalizers:
    sql.terrak8s.io
  Generation:  1
  Managed Fields:
  ...<skiped>
Spec:
  Project:
    Region:       europe-west1
    Name:         my-project
    Zone:         europe-west1-b
  Databases:
    Charset:    UTF8
    Collation:  en_US.UTF8
    Instance:   my-instance
    Name:       sample-db1
    Project:    my-project
    Charset:    UTF8
    Collation:  en_US.UTF8
    Instance:   my-instance
    Name:       sample-db1
    Project:    my-project
  Bucket Config:
    Destroy:  true
    Lifecycle Rule:
      Action:
        Type:  Delete
      Condition:
        Age:                      3
    Location:                     europe-west1
    Name:                         my-bucket
    Project:                      my-project
    Storage Class:                STANDARD
    Uniform Bucket Level Access:  true
  Remote State:
    Bucket Name:    my-bucket
    Bucket Prefix:  dev/tfstate
  Sql Instance:
    Database Version:     POSTGRES_9_6
    Deletion Protection:  false
    Name:                 my-instance
    Project:              my-project
    Region:               europe-west1
    Settings:
      Activation Policy:  ALWAYS
      Availability Type:  ZONAL
      Backup Configuration:
        Enabled:     true
        Start Time:  21:30
      Database Flags:
        Name:           log_min_duration_statement
        Value:          3000
      Disk Autoresize:  true
      Disk Type:        PD_SSD
      Ip Configuration:
        ipv4Enabled:      false
        Private Network:  my-vpc
      Labels:
        Name:   environment
        Value:  dev01
      Location Preference:
        Zone:        europe-west1-b
      Machine Type:  db-f1-micro
      Maintenance Window:
        Day:   7
        Hour:  4
  Users:
    Instance:  my-instance
    Name:      user-1
    Password:
      Secret Key Ref:
        Key:   database-password-1
        Name:  cred-database
    Project:   my-project
    Instance:  my-instance
    Name:      user-2
    Password:
      Secret Key Ref:
        Key:   database-password-2
        Name:  cred-database
    Project:   my-project
Status:
  Output:
    Connection IP Address:  192.168.0.12
    Connection Name:        my-project:europe-west1:my-instance
  Phase:                    Running
Events:  
  Type     Reason                Age                  From            Message
  ----     ------                ----                 ----            -------
  Warning  ApplyingFailed        35s                  sql-controller  failed to provision cloud sql instance "my-instance"
  Normal   SuccessfulApplying    28s (x2 over 9m30s)  sql-controller  successfully provision storage bucket "my-bucket"
  Normal   SuccessfulInitialize  27s (x2 over 9m27s)  sql-controller  successfully configured the remote backend "gcs" bucket
  Normal   SuccessfullyApplying  14s                  sql-controller  successfully creating cloud sql instance "my-instance"
  Normal   Synced                13s                  sql-controller  PostgreSql Resource synced successfully

```