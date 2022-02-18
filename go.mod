module github.com/DoodleScheduling/k8svault-controller

go 1.16

require (
	github.com/go-logr/logr v0.4.0
	github.com/hashicorp/errwrap v1.0.0
	github.com/hashicorp/vault/api v1.0.4
	github.com/hashicorp/vault/sdk v0.1.14-0.20200519221838-e0cfd64bc267 // indirect
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.15.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/testcontainers/testcontainers-go v0.12.0
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	sigs.k8s.io/controller-runtime v0.10.2
)
