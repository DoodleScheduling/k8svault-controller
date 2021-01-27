# k8svault-controller

A controller for Kubernetes with the ability to add k8s native secrets to hashicorp vault.
It might happen that you need secrets in vault however all existing toolchains are using Kubernetes secrets.
This controllers makes sure that secrets and fields are available in vault as well.

It does this by adding certain annotations to a Secret resource.

## Annotations

Annotation | Required | Description |
-----------| ---------|-------------|
k8svault-controller.v1beta1.infra.doodle.com/path | Required | The path where to apply values, example: `/secret/prod/myapp` |
k8svault-controller.v1beta1.infra.doodle.com/vault | Optional (If controller knows a default vault)| The vault host, example: `https://vault:8200` |
k8svault-controller.v1beta1.infra.doodle.com/force | Optional | If true any existing matching fields in vault will be overwritten, example: `true` |
k8svault-controller.v1beta1.infra.doodle.com/fields | Optional | Comma separated list of fields, example: `mysecret`. If not specified all fields are mapped to vault. |

## Map fields

The annotation `k8svault-controller.v1beta1.infra.doodle.com/fields` may also be used to map fields to another vault field using a comma separated key=value list.
Spec: field [k8s-secret]: field [vault field]
If no mapping is done the field from the Secret gets applied to vault with the same field name.


Example:
```yaml
apiVersion: v1
kind: Secret
metadata:
  annotations:
    k8svault-controller.v1beta1.infra.doodle.com/vault: https://vault:8200
    k8svault-controller.v1beta1.infra.doodle.com/path: /secret/prod/myapp
    k8svault-controller.v1beta1.infra.doodle.com/fields: username=appUsername,password=appPassword
type: Opaque
data:
  username: YWRtaW4=
  password: YWRtaW4=
```

## Vault values

The controller does not touch any other fields in vault besides the explicitly specified ones.
And existing ones with matching fields will only be overwritten if `k8svault-controller.v1beta1.infra.doodle.com/force` is set.


## Limitations

One secret can only be mapped to a single vault path. We might introduce a CRD in addition to annotations to allow further special cases.

## Helm chart
