package terraform

import (
	"github.com/HamzaZo/structs"
	"github.com/HamzaZo/terrak8s-operator/pkg/util"
)

var (
	providerName         = "google"
	backendType          = "gcs"
	dataBaseResourceName = providerName + "_" + "sql_database"
	instanceResourceName = dataBaseResourceName + "_" + "instance"
	userResourceName     = providerName + "_" + "sql_user"
	bucketResourceName   = providerName + "_" + "storage_bucket"
)

const (
	providerVersion = "3.5.0"
)

func RenderDatabaseResource(databaseSpec interface{}) ([]byte, error) {
	res := []byte(`{"resource":{ "` + dataBaseResourceName + `":{ "database":`)
	mapD := structs.Map(databaseSpec)

	valD, err := util.ToJson(mapD)
	if err != nil {
		return nil, err
	}
	res = append(res, valD...)
	res = append(res, []byte("} } ")...)
	res = append(res, []byte(",")...)

	return res, nil
}

func RenderAdditionalDatabasesResource(databaseSpec interface{}) ([]byte, error) {
	res := []byte(`"resource":{ "` + dataBaseResourceName + `":{ "additional_databases":`)
	mapD := structs.Map(databaseSpec)

	valD, err := util.ToJson(mapD)
	if err != nil {
		return nil, err
	}
	res = append(res, valD...)
	res = append(res, []byte("} } ")...)
	res = append(res, []byte(",")...)

	return res, nil
}

func RenderSqlUserResource(userSpec interface{}, value interface{}) ([]byte, error) {
	res := []byte(`"resource":{ "` + userResourceName + `":{ "default":`)
	mapD := structs.Map(userSpec)

	mapD["depends_on"] = []string{
		instanceResourceName + ".instance",
	}
	mapD["password"] = value

	valD, err := util.ToJson(mapD)
	if err != nil {
		return nil, err
	}
	res = append(res, valD...)
	res = append(res, []byte("} }")...)
	res = append(res, []byte(",")...)

	return res, nil
}

func RenderAdditionalSqlUserResource(userSpec interface{}, value interface{}) ([]byte, error) {
	res := []byte(`"resource":{ "` + userResourceName + `":{ "additional_users":`)
	mapD := structs.Map(userSpec)

	mapD["depends_on"] = []string{
		instanceResourceName + ".instance",
	}
	mapD["password"] = value

	valD, err := util.ToJson(mapD)

	if err != nil {
		return nil, err
	}
	res = append(res, valD...)
	res = append(res, []byte("} } ")...)
	res = append(res, []byte(",")...)

	return res, nil
}

func RenderInstanceResource(instanceSpec interface{}) ([]byte, error) {
	res := []byte(`"resource":{ "` + instanceResourceName + `":{ "instance":`)
	mapD := structs.Map(instanceSpec)

	valD, err := util.ToJson(mapD)
	if err != nil {
		return nil, err
	}
	res = append(res, valD...)
	res = append(res, []byte("} } }")...)

	return res, nil
}

func RenderInstanceOutput() []byte {
	res := []byte(`output "connectionName" { value = google_sql_database_instance.instance.connection_name }`)
	res2 := []byte(`output "connectionIPAddress" {  value = google_sql_database_instance.instance.private_ip_address }`)
	res = append(res, []byte("\n")...)
	res = append(res, res2...)

	return res
}

func RenderBucketResource(bucketSpec interface{}) ([]byte, error) {
	b := []byte(`{ "resource":{ "` + bucketResourceName + `":{ "bucket":`)

	mapB := structs.Map(bucketSpec)

	valB, err := util.ToJson(mapB)
	if err != nil {
		return nil, err
	}
	b = append(b, valB...)
	b = append(b, []byte("} } } ")...)

	return b, nil
}

func RenderRemoteBackend(backendSpec interface{}) ([]byte, error) {
	b := []byte(`{ "terraform": { "backend": { "` + backendType + `": `)
	mapB := structs.Map(backendSpec)

	valB, err := util.ToJson(mapB)
	if err != nil {
		return nil, err
	}

	b = append(b, valB...)

	b = append(b, []byte("} } } ")...)

	return b, nil
}

func RenderProvider(providerSpec interface{}) ([]byte, error) {
	p := []byte(`{ "provider": { "` + providerName + `":`)
	t := []byte(`"terraform": { "required_providers": { "` + providerName + `": { "source": "hashicorp/google", "version": "` + providerVersion + `" } }`)

	mapP := structs.Map(providerSpec)
	valP, err := util.ToJson(mapP)
	if err != nil {
		return nil, err
	}
	p = append(p, valP...)
	p = append(p, []byte(",")...)

	p = append(p, t...)
	p = append(p, []byte("} } }")...)

	return p, nil
}
