# k8svault-controller

A controller for Kubernetes with the ability to add k8s native secrets to hashicorp vault.
It might happen that you need secrets in vault however all tooling you are using is based on Kubernetes secrets.
This controllers makes sure that secrets and fields are available in vault as well.

It does this by adding certain annotations to a Secret resource.

## Example
Following a secret which fields must be placed into vault:

```yaml
apiVersion: v1
data:
  password: ZGUK
  username: MQo=
kind: Secret
metadata:
  name: my-secret
  namespace: default
type: Opaque
```

To bind both fields to our example vault path at `/secret/env/myapp` a binding might look like this:

```yaml
apiVersion: infra.doodle.com/v1beta1
kind: VaultBinding
metadata:
  name: my-secret
  namespace: default
spec:
  address: "vault:8200"
  path: "/secret/env/myapp"
  forceApply: true
  secret:
    name: my-secret
  fields:
  - name: password
  - name: username
    rename: root
  auth:
    role: default
```

**Note**: The field named  `password` gets written with the same name while the `username` field gets created as `root` in vault.

## Specify TLS settings

It is possible to set additional fields including TLS configuration for vault:

```yaml
apiVersion: infra.doodle.com/v1beta1
kind: VaultBinding
metadata:
  name: my-secret
  namespace: default
spec:
  address: "vault:8200"
  path: "/secret/env/myapp"
  forceApply: true
  secret:
    name: my-secret
  fields:
  - name: password
  - name: username
  tlsConfig:
    insecure: true
  auth:
    role: default
```

Other TLS settings:
```go
type VaultTLSSpec struct {
	CACert     string `json:"caCert,omitempty"`
	CAPath     string `json:"caPath,omitempty"`
	ClientCert string `json:"clientCert,omitempty"`
	ClientKey  string `json:"clientKey,omitempty"`
	ServerName string `json:"serverName,omitempty"`
	Insecure   bool   `json:"insecure,omitempty"`
}
```

## Helm chart

Please see [chart/k8svault-controller](https://github.com/DoodleScheduling/k8svault-controller) for the helm chart docs.

## Overwrite secrets in vault

The controller does not touch any other fields in vault besides the explicitly specified ones.
And existing ones with matching fields will only be overwritten if `spec.forceApply` is set.

**Note** If you want to keep the secret in sync with vault this must be enabled.

## Limitations

The controller only supports [Kubernetes authentication method](https://www.vaultproject.io/docs/auth/kubernetes) for now.
Currently there is no garbage collection implemented, meaning all the things created in vault are not removed if the binding gets deleted.

## Configure the controller

You may change base settings for the controller using env variables (or alternatively command line arguments).
It is possible to set defaults (fallback values) for the vault address and also all TLS settings.
