package tfimportables

import "github.com/onelogin/onelogin-go-sdk/pkg/services/apps"

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

type AppQuerier interface {
	Query(query *apps.AppsQuery) ([]apps.App, error)
}
