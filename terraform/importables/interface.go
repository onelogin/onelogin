package tfimportables

type Importable interface {
	ImportFromRemote() []ResourceDefinition
	HCLShape() interface{}
}

// ResourceDefinition represents the resource to be imported
type ResourceDefinition struct {
	Provider string
	Name     string
	Type     string
}
