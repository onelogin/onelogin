# OneLogin CLI
[![Go Report Card](https://goreportcard.com/badge/github.com/onelogin/onelogin)](https://goreportcard.com/report/github.com/onelogin/onelogin)
<a href='https://github.com/jpoles1/gopherbadger' target='_blank'>![gopherbadger-tag-do-not-edit](https://img.shields.io/badge/Go%20Coverage-86%25-brightgreen.svg?longCache=true&style=flat)</a>
## Description

The OneLogin CLI is your way to manage OneLogin resources such as Apps, Users, and Mappings via the Command Line.

## Features
`terraform-import <resource>`: Import your remote resources into a local Terraform State.
Running this command will do the following:
  1. Pull **all** your resources from the OneLogin API (remote)
  2. Establish a basic main.tf that represents all the apps in your account. Each app will get an empty Terraform resource "placeholder"
  3. Call `terraform import` for all the apps and update the `.tfstate`
  4. Using .tfstate, update main.tf to fill in the editable fields of the resource

## Usage
This assumes you have Terraform installed and the OneLogin provider side-loaded.
The OneLogin Terraform provider is still in beta. If you'd like to use the beta [see this guide](https://github.com/onelogin/onelogin-terraform-provider#onelogin-terraform-provider-sdk)

### Configuration
Add your OneLogin profiles with `onelogin profiles add <profile_name>`

You'll be prompted for your client_id and client_secret (obtained by creating a set of developer keys in the onelogin admin portal)

You can add as many profiles as you like, and you can switch the active profile with `onelogin profiles use <profile_name>` which will point the CLI at the active account.

### Example
Import all OneLogin apps, create a main.tf file, and establish Terraform state.
```sh
onelogin terraform-import onelogin_apps
```

### Install From Source - Requires Go
clone this repository
from inside the repository `go build ./...` to create a runnable binary
from inside the repository `go install .` to add a the runnable CLI to your GOPATH /bin directory

### Binaries
There are binaries available for the major platforms in this project's /build directory. Download the
binary for your system and add it to your /bin folder or run it directly per your system's requirements.

* `darwin-amd64`  => mac 64 bit and linux
* `windows-386`   => windows 32 bit
* `windows-amd64` => windows 64 bit
* `linux-386`     => linux 32 bit
* `linux-amd64`   => linux 64 bit

### Use
from an empty directory, where you plan to manage your main.tf file run:
`onelogin terraform-import onelogin_apps`

You'll be prompted to confirm the number of resources to import.
This will capture the state of your remote in its entirety

If you have some resources already set up in main.tf, this will merge your main.tf with resources from the remote

## Terraform Importer

### Supported Importable Resources
* `onelogin_apps` => returns all apps
* `onelogin_saml_apps` => returns saml apps only
* `onelogin_oidc_apps` => returns oidc apps only
* `onelogin_user_mappings` => returns all user mappings

## Contributing

### Terraform Importer
To add an importable resource, do these things:
1. Under the `terraform/importables` directory, add a file with the scheme <provider>_<resource>.go
2. Add a struct to represent your importable, add whatever filtering or special criteria fields you need.
OneLogin importables typically have at least a field for the resource's service from our SDK.
3. On that struct you just made, implement the `Importable` interface. this is where we pull all the resources from the remote/api and represent them as resources in terraform
4. Add structs that represent the fields you want to pull from tfstate into main.tf after the import for users to manage later. the state struct is how a resource is represented in .tfstate so in order for json marshalling to work, this struct has to look like your resource in tfstate.
5. Refer to this in `terraform/import/state.go` in the 'molds' section so the importer is aware of the fields that should be read from tfstate and will marshal the respective data.
6. in `cmd/terraform-import` add to the `importables` struct `<resource_name>: tfimportables.YourImportable{}` to register it
