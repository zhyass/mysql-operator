# mysql-operator has moved

Get the new repository: https://github.com/radondb/radondb-mysql-kubernetes

## Quickstart

Install the operator named `test`:

```shell
helm install test https://github.com/zhyass/mysql-operator/releases/latest/download/mysql-operator.tgz
```

Then install the cluster named `sample`:

```shell
kubectl apply -f https://raw.githubusercontent.com/zhyass/mysql-operator/master/config/samples/mysql_v1_cluster.yaml
```

## Uninstall

Uninstall the cluster named `sample`:

```shell
kubectl delete clusters.mysql.radondb.io sample
```

Uninstall the operator name `test`:

```shell
helm uninstall test
```

Uninstall the crd:

```shell
kubectl delete customresourcedefinitions.apiextensions.k8s.io clusters.mysql.radondb.io
```
