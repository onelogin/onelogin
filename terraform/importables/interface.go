package tfimportables

import (
	"github.com/onelogin/onelogin-go-sdk/pkg/services/apps"
	"github.com/onelogin/onelogin-go-sdk/pkg/services/user_mappings"
)

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

type AppQuerier interface {
	Query(query *apps.AppsQuery) ([]apps.App, error)
	GetOne(id int32) (*apps.App, error)
}

type UserMappingQuerier interface {
	Query(query *usermappings.UserMappingsQuery) ([]usermappings.UserMapping, error)
	GetOne(id int32) (*usermappings.UserMapping, error)
}
