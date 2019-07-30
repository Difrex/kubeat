# Kubeat

A logger for the k8s without any logfiles.

## Why

In some situations such as not own K8S installation or restricted permissions, we can't use logging based on the files via DaemonSet.
But, K8S has a pods logging API. The Kubeat do per namespace logging simple.

## How to use


Setup it via the Helm.
```
cd helm/kubeat
emacs values.yaml
helm install .
```

Variables:

| Name                       | Default                     | Description                                                |
|:---------------------------|:----------------------------|:-----------------------------------------------------------|
| `disable_self_logging`     | `"yes"`                     | Do not log self output                                     |
| `rbac.create`              | `true`                      | Create an new role for the Kubeat                          |
| `serviceAccount.create`    | `true`                      | Create an new service account                              |
| `serviceAccount.name`      | `kubeat-logger`             | Name of the service account                                |
| `serviceAccount.namespace` | `default`                   | Namespace to use                                           |
| `configmap.type`           | `elasticsearch`             | Type of the logs receiver. Can be `elasticsearch` or `tcp` |
| `configmap.hosts`          | `["http://localhost:9200"]` | Hosts of the logs receiver.                                |
| `configmap.index`          | `kubeat`                    | Elasticsearch daily index prefix                           |
| `configmap.doc_type`       | `k8slog`                    | Elasticsearch document type                                |
| `configmap.limit`          | `1000`                      | Elasticsearch bucket soft limit                            |
| `secret.create`            | `true`                      | Create a secret with username and password                 |
| `secret.username`          | `"elastic"`                 | Elasticsearch username                                     |
| `secret.password`          | `"password"`                | Elasticsearch password                                     |

## Project status

In development, but it has deployed in the production clusters and all working fine.

## Known issues

* Only daily indices is supported
