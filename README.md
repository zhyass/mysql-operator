# mysql-operator

## Quickstart

Install the operator named test:

```shell
helm install test https://github.com/zhyass/mysql-operator/releases/latest/download/mysql-operator.tgz
```

Then install the cluster:

```shell
kubectl apply -f https://raw.githubusercontent.com/zhyass/mysql-operator/master/config/samples/mysql_v1_cluster.yaml
```
