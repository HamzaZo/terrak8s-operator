### Local development

#### Requirement
* go 1.15.5
* kind v0.9.0 (kubernetes v1.19.1)
* controller-gen v0.2.5
* kubebuilder 2.3.1

#### Run Terrak8s locally

Assuming that you have already set up a [kind](https://kind.sigs.k8s.io/docs/user/quick-start/) cluster. We can test out the controller, and run it locally against the cluster. 
Before we do so, we'll need to install our CRDs by using controller-tools to update automatically the YAML manifests as follows :

Check [Prerequisites](Terrak8sGuide.md) before you continue

```shell
$ make crd-manifests  
$ kubectl apply -f config/crd/bases/sql.terrak8s.io_postgresqls.yaml
```

**Note:** bear in mind that you should disable the webhook by commenting the following code in `main.go` :

```go
  if err = (&sqlv1alpha1.PostgreSql{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "PostgreSql")
		os.Exit(1)
  }
```

then, run the controller and create postgreSql instance 
```shell
$ make run 
$ kubectl apply -f examples/complete-postgresql-instance.yaml
```

#### Deploy on cluster

Modify the value of `caBundle` in both **MutatingWebhookConfiguration** and **ValidatingWebhookConfiguration**

**Note:** keeps in mind that you need to uncomment the webhook code
```shell
$ kubectl create ns webhook-operator
$ ./config/manifests/script.sh --service webhook-service --namespace webhook-operator --secret webhook-server-cert
$ kubectl apply -f config/manifests/rbac
$ kubectl apply -f config/manifests/webhook
```
then, run
```shell
$ kubectl apply -f config/crd/bases/sql.terrak8s.io_postgresqls.yaml
$ make docker-build
$ make kind-load
$ make deploy-manager
```

Finally, deploy postgreSql instance 
```shell
$ kubectl apply -f examples/postgresql-instance.yaml
```