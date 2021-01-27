package v1beta1

// Supported annotations by k8svault-controller
const (
	AnnotationVault     = "k8svault-controller.v1beta1.infra.doodle.com/vault"
	AnnotationPath      = "k8svault-controller.v1beta1.infra.doodle.com/path"
	AnnotationForce     = "k8svault-controller.v1beta1.infra.doodle.com/force"
	AnnotationFields    = "k8svault-controller.v1beta1.infra.doodle.com/fields"
	AnnotationTimestamp = "k8svault-controller.v1beta1.infra.doodle.com/reconciledAt"
)

// Mapping represents how a secret is mapped to vault. It is a representation
// of all supported annotations.
type Mapping struct {
	Vault  string
	Path   string
	Force  bool
	Fields map[string]string
}

// NewMapping creates a new mapping
func NewMapping() *Mapping {
	return &Mapping{
		Fields: make(map[string]string),
	}
}
