package v1beta1

// Supported annotations by k8svault-controller
const (
	AnnotationVault         = "k8svault-controller.v1beta1.infra.doodle.com/vault"
	AnnotationPath          = "k8svault-controller.v1beta1.infra.doodle.com/path"
	AnnotationForce         = "k8svault-controller.v1beta1.infra.doodle.com/force"
	AnnotationFields        = "k8svault-controller.v1beta1.infra.doodle.com/fields"
	AnnotationRole          = "k8svault-controller.v1beta1.infra.doodle.com/role"
	AnnotationTokenPath     = "k8svault-controller.v1beta1.infra.doodle.com/tokenPath"
	AnnotationReconciledAt  = "k8svault-controller.v1beta1.infra.doodle.com/reconciledAt"
	AnnotationTLSCACert     = "k8svault-controller.v1beta1.infra.doodle.com/tlsCACert"
	AnnotationTLSCAPath     = "k8svault-controller.v1beta1.infra.doodle.com/tlsCAPath"
	AnnotationTLSClientCert = "k8svault-controller.v1beta1.infra.doodle.com/tlsClientCert"
	AnnotationTLSClientKey  = "k8svault-controller.v1beta1.infra.doodle.com/tlsClientKey"
	AnnotationTLSServerName = "k8svault-controller.v1beta1.infra.doodle.com/tlsServerName"
	AnnotationTLSInsecure   = "k8svault-controller.v1beta1.infra.doodle.com/tlsInsecure"
)
