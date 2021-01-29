# k8svault-controller

A controller for Kubernetes with the ability to add k8s native secrets to hashicorp vault.
It might happen that you need secrets in vault however all tooling you are using is based on Kubernetes secrets.
This controllers makes sure that secrets and fields are available in vault as well.

It does this by adding certain annotations to a Secret resource.

## Annotations

Annotation | Required | Description |
-----------| ---------|-------------|
k8svault-controller.v1beta1.infra.doodle.com/path | Required | The path where to apply values, example: `/secret/prod/myapp` |
k8svault-controller.v1beta1.infra.doodle.com/role | required | Specify a vault auth role which gets used while authenticating to vault.|
k8svault-controller.v1beta1.infra.doodle.com/vault | Optional | (If controller knows a default vault)| The vault host, example: `https://vault:8200` |
k8svault-controller.v1beta1.infra.doodle.com/force | Optional | If true any existing matching fields in vault will be overwritten, example: `true` |
k8svault-controller.v1beta1.infra.doodle.com/tokenPath | Optional | You may specify a different path to the kubernetes serviceAccount token path. |
k8svault-controller.v1beta1.infra.doodle.com/fields | Optional | Comma separated list of fields, example: `mysecret`. If not specified all fields are mapped to vault. |

Besides the common annotation there are advanced settings to drive TLS configuration:

Annotation | Required | Description |
-----------| ---------|-------------|
k8svault-controller.v1beta1.infra.doodle.com/tlsCACert | Optional | CACert is the path to a PEM-encoded CA cert file used to verify the Vault server SSL certificate. |
k8svault-controller.v1beta1.infra.doodle.com/tlsCAPath | Optional | CAPath is the path to a directory of PEM-encoded CA cert files to verify CAPath string. |
k8svault-controller.v1beta1.infra.doodle.com/tlsClientCert | Optional | The path to the certificate for Vault communication. |
k8svault-controller.v1beta1.infra.doodle.com/tlsClientKey | Optional | The private key for Vault communication. |
k8svault-controller.v1beta1.infra.doodle.com/tlsServerName | Optional | Used to set the SNI host when connecting to vault. |
k8svault-controller.v1beta1.infra.doodle.com/tlsInsecure | Optional | Allow insecure TLS communication to vault (no certificate validation). |

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

## Helm chart

Please see [chart/k8svault-controller](https://github.com/DoodleScheduling/k8svault-controller) for the helm chart docs.

## Vault values

The controller does not touch any other fields in vault besides the explicitly specified ones.
And existing ones with matching fields will only be overwritten if `k8svault-controller.v1beta1.infra.doodle.com/force` is set.

## Limitations

One secret can only be mapped to a single vault path. It is not possible to map fields to different vault paths right now.
The controller only supports (for now) vault using the [kubernetes authentication method](https://www.vaultproject.io/docs/auth/kubernetes).
Also with annotations we can't do proper garbage collection. For example if one changes the vault path, the previously filled one won't be re
moved.

There might be CRD at some point supporting these cases.

## Map fields

The annotation `k8svault-controller.v1beta1.infra.doodle.com/fields` may also be used to map fields to another vault field using a comma separated key=value list.
Spec: field [k8s-secret]: field [vault field]
If no mapping is done the field from the Secret gets applied to vault with the same field name.

## Configure the controller

You may change base settings for the controller using env variables (or alternatively command line arguments).
It is possible to set defaults (fallback values) for the vault address and also all TLS settings.
