package tfimportables

type Importable interface {
	ImportFromRemote(searchId *string) []ResourceDefinition // transforms resources from remote to an array ResourceDefinitions to be inserted into an HCL file
	HCLShape() interface{}                                  // dictates what fields on tfstate should be represented in HCL files
}

// ResourceDefinition represents basic information about the resource to be imported
// so it can be used in HCL file and set up terraform import command
type ResourceDefinition struct {
	Provider string // Name of provider Terraform will use to do import
	Name     string // Name of the resource as defined in HCL
	Type     string // Type of resource e.g. aws_iam_user
	ImportID string // ID used by Terraform provider to download the resource
}
