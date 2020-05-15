package tfimportables

type Importable interface {
	ImportFromRemote() []ResourceDefinition
}

// ResourceDefinition represents the resource to be imported
type ResourceDefinition struct {
	Content  []byte
	Provider string
	Name     string
	Type     string
}