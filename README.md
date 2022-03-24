# k8svault-controller

[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/5738/badge)](https://bestpractices.coreinfrastructure.org/projects/5738)
[![e2e](https://github.com/DoodleScheduling/k8svault-controller/workflows/e2e/badge.svg)](https://github.com/DoodleScheduling/k8svault-controller/actions)
[![report](https://goreportcard.com/badge/github.com/DoodleScheduling/k8svault-controller)](https://goreportcard.com/report/github.com/DoodleScheduling/k8svault-controller)
[![license](https://img.shields.io/github/license/DoodleScheduling/k8svault-controller.svg)](https://github.com/DoodleScheduling/k8svault-controller/blob/master/LICENSE)
[![release](https://img.shields.io/github/release/DoodleScheduling/k8svault-controller/all.svg)](https://github.com/DoodleScheduling/k8svault-controller/releases)

A controller for kubernetes for automating secret provisioning to hashicorp vault.
You may either provision secrets from kubernetes core secrets or from other vaults.

## Example VaultBinding

A `VaultBinding` binds a kubernetes core secret to a hashicorp vault path.
Following a secret which fields shall be placed into vault:

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
apiVersion: vault.infra.doodle.com/v1beta1
kind: VaultBinding
metadata:
  name: my-secret
  namespace: default
spec:
  address: "https://vault:8200"
  path: "/secret/env/myapp"
  forceApply: true
  secret:
    name: my-secret
  fields:
  - name: password
  - name: username
    rename: root
```
**Note**: The field named  `password` gets written with the same name while the `username` field gets created as `root` in vault.

## Example VaultMirror

A `VaultMirror` binds a source vault path to a destination vault path.
Following a secret which fields shall be placed into the destination vault:

```yaml
apiVersion: vault.infra.doodle.com/v1beta1
kind: VaultMirror
metadata:
  name: my-secret
  namespace: default
spec:
  source:
    address: "https://source-vault:8200"
    path: "/secret/env/myapp"
  destination:
    address: "https://source-vault:8200"
    path: "/secret/env/myapp"    
  forceApply: true
  interval: "0"
  fields:
  - name: password
  - name: username
```

## Specify Advanced TLS & Auth settings

It is possible to set additional fields including TLS configuration for vault:

```yaml
apiVersion: vault.infra.doodle.com/v1beta1
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

Other `tlsConfig` include:

*	CACert
*	CAPath
*	ClientCert
*	ClientKey
*	ServerName
*	Insecure

An example for `VaultMirror`:

```yaml
apiVersion: vault.infra.doodle.com/v1beta1
kind: VaultBinding
metadata:
  name: my-secret
  namespace: default
spec:
  source:
    address: "https://vault-source:8200"
    path: "/secret/env/myapp"
    tlsConfig:
      insecure: true
    auth:
      role: reader
  destination:
    address: "https://vault-dest:8200"
    path: "/secret/env/myapp"
    tlsConfig:
      insecure: true
    auth:
      role: writer
  forceApply: true
```

## Installation

### Helm

Please see [chart/k8svault-controller](https://github.com/DoodleScheduling/k8svault-controller/tree/master/chart/k8svault-controller) for the helm chart docs.

### Manifests/kustomize

Alternatively you may get the bundled manifests in each release to deploy it using kustomize or use them directly.

## Overwrite secrets in vault

The controller does not touch any other fields in vault besides the explicitly specified ones.
Existing fields will only be overwritten if `spec.forceApply` is set.

**Note** If you want to keep the secret in sync with vault this must be enabled.

## Limitations

Currently there are some limitations which we may address soon.

* There is only support for kubernetes vault auth
* There is only support for the secrets engine v1
* The controller only supports [Kubernetes authentication method](https://www.vaultproject.io/docs/auth/kubernetes) for now.
* Currently there is no garbage collection implemented, meaning all the things created in vault are not removed if the binding gets deleted.

## Configure the controller

You may change base settings for the controller using env variables (or alternatively command line arguments).
It is possible to set defaults (fallback values) for the vault address and also all TLS settings.

Available env variables:

| Name  | Description | Default |
|-------|-------------| --------|
| `METRICS_ADDR` | The address of the metric endpoint binds to. | `:9556` |
| `PROBE_ADDR` | The address of the probe endpoints binds to. | `:9557` |
| `ENABLE_LEADER_ELECTION` | Enable leader election for controller manager. | `false` |
| `LEADER_ELECTION_NAMESPACE` | Change the leader election namespace. This is by default the same where the controller is deployed. | `` |
| `NAMESPACES` | The controller listens by default for all namespaces. This may be limited to a comma delimted list of dedicated namespaces. | `` |
| `CONCURRENT` | The number of concurrent reconcile workers.  | `4` |
| `VAULT_ADDR` | Fallback vault address if no vault address is set in the VaultBinding. | `http://localhost:8200` |
| `VAULT_TOKEN_PATH` | Specify different path for the kubernetes ServiceAccount token file. Also acts as fallback and might be set in the VaultBinding as well. | `/var/run/secrets/kubernetes.io/serviceaccount/token` |
| `VAULT_ROLE` | Fallback vault authentication role used for authentication. Used if no role was specified in the VaultBinding. | `k8svault-controller` |

## Vault requirements
Vault needs to be configured to allow authenticate via kubernetes auth. A auth role must exists which maps against
the serviceAccount of this controller.
**Ensure** that the namespace and serviceaccount both matches the serviceaccount where this controller gets deployed.

Example role binding for kubernetes auth config:
```
vault write auth/kubernetes/role/k8svault-controller bound_service_account_names=k8svault-controller bound_service_account_namespaces=my-namespace-where-k8svault-controller-isinstalled policies=my-policy
```

Best practice is to create one for the controller on each vault you would like to manage secrets.
The auth role should be called `k8svault-controller`) which gets used by default in this controller. However you may also change the default one using the env `VAULT_ROLE`
or change it individually in each VaultBinding/VaultMirror.
