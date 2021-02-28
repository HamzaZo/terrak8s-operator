### Installation Guide

### Prerequisites:
- To use terrak8s, you should:
  * Have a minimum kubernetes version of 1.15.
  * Have cluster admin permissions 
  * Have Helm 3 installed   


### Installation:
#### Deploying via Helm

A Helm chart exist in `chart/terrak8s`, you can deploy it via the following instruction:

```shell
$ helm upgrade terrak8s-operator --install chart/terrak8s/
```
Please note that this chart is only compatible with Helm v3.

### Uninstallation:

Run the following Helm command to uninstall terrak8s:
```shell
$  helm uninstall terrak8s-operator
```

Helm 3 will not cleanup terrak8s installed CRD. Run the following to uninstall terrak8s CRD:

```shell
$ kubectl delete crd postgresqls.sql.terrak8s.io
```