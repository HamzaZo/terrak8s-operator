package terraform

import (
	sqlv1alpha1 "github.com/HamzaZo/terrak8s-operator/api/v1alpha1"
	"github.com/HamzaZo/terrak8s-operator/pkg/util"
)

func GenerateProviderAndBackendTF(instance *sqlv1alpha1.PostgreSql, dir string) error {
	b, err := RenderRemoteBackend(instance.Spec.RemoteState)
	if err != nil {
		return err
	}
	output, err := util.GetPrettyJSON(b)
	if err != nil {
		return err
	}
	err = util.WriteToFile(output, dir, "backend.tf.json")
	if err != nil {
		return err
	}
	p, err := RenderProvider(instance.Spec.Project)
	if err != nil {
		return err
	}
	out, err := util.GetPrettyJSON(p)
	if err != nil {
		return err
	}
	err = util.WriteToFile(out, dir, "provider.tf.json")
	if err != nil {
		return err
	}
	return nil
}

func GenerateBucketTF(instance *sqlv1alpha1.PostgreSql, dir string) error {
	b, err := RenderBucketResource(instance.Spec.BucketConfig)
	if err != nil {
		return err
	}
	output, err := util.GetPrettyJSON(b)
	if err != nil {
		return err
	}
	err = util.WriteToFile(output, dir, "bucket.tf.json")
	if err != nil {
		return err
	}
	p, err := RenderProvider(instance.Spec.Project)
	if err != nil {
		return err
	}
	o, err := util.GetPrettyJSON(p)
	if err != nil {
		return err
	}
	err = util.WriteToFile(o, dir, "provider.tf.json")
	if err != nil {
		return err
	}
	return nil
}

func GenerateTFDatabases(instance *sqlv1alpha1.PostgreSql) ([]byte, error) {
	var output []byte
	for i, k := range instance.Spec.Databases {
		if i == 0 {
			d1, err := RenderDatabaseResource(k)
			if err != nil {
				return nil, err
			}
			output = append(output, d1...)
		} else {
			d2, err := RenderAdditionalDatabasesResource(k)
			if err != nil {
				return nil, err
			}
			output = append(output, d2...)
		}
	}

	return output, nil
}

func GenerateTFUsers(instance *sqlv1alpha1.PostgreSql, value map[string][]byte) ([]byte, error) {
	var output []byte

	for i, k := range instance.Spec.Users {
		if v, ok := value[k.Password.SecretKeyRef.Key]; ok {
			if i == 0 {
				u1, err := RenderSqlUserResource(k, string(v))
				if err != nil {
					return nil, err
				}
				output = append(output, u1...)
			} else {
				u2, err := RenderAdditionalSqlUserResource(k, string(v))
				if err != nil {
					return nil, err
				}
				output = append(output, u2...)
			}
		}
	}

	return output, nil
}

func GenerateTFInstance(instance *sqlv1alpha1.PostgreSql, dir string, value map[string][]byte) error {
	var res []byte

	db, err := GenerateTFDatabases(instance)
	if err != nil {
		return err
	}
	user, err := GenerateTFUsers(instance, value)
	if err != nil {
		return err
	}
	s, err := RenderInstanceResource(instance.Spec.SqlInstance)
	if err != nil {
		return err
	}

	res = append(res, db...)
	res = append(res, user...)
	res = append(res, s...)

	final, err := util.GetPrettyJSON(res)
	if err != nil {
		return err
	}
	err = util.WriteToFile(final, dir, "main.tf.json")
	if err != nil {
		return err
	}
	return nil
}

func GenerateTFOutput(dir string) error {
	err := util.WriteToFile(RenderInstanceOutput(), dir, "output.tf")
	if err != nil {
		return err
	}
	return nil

}
