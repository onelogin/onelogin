# OneLogin CLI
[![Go Report Card](https://goreportcard.com/badge/github.com/onelogin/onelogin-cli)](https://goreportcard.com/report/github.com/onelogin/onelogin-cli)
<a href='https://github.com/jpoles1/gopherbadger' target='_blank'>![gopherbadger-tag-do-not-edit](https://img.shields.io/badge/Go%20Coverage-56%25-brightgreen.svg?longCache=true&style=flat)</a>
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

### API Credentials
The CLI uses the OneLogin API and therefore you'll need admin access to a OneLogin account where you can create API credentials. 

Create a set of API credentials with _manage all_ permission. 

* Export these credentials to your environment and the provider will read them in from there
```sh
export ONELOGIN_CLIENT_ID=<your client id>
export ONELOGIN_CLIENT_SECRET=<your client secret>
export ONELOGIN_OAPI_URL=<the api url for your region>
```

### Example
Import all OneLogin apps, create a main.tf file, and establish Terraform state.
```sh
onelogin-cli terraform-import onelogin_apps
```

### Install - Requires Go
```sh
go install github.com/onelogin/onelogin-cli
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
`onelogin-cli terraform-import onelogin_apps`

You'll be prompted to confirm the number of resources to import.
This will capture the state of your remote in its entirety

If you have some resources already set up in main.tf, this will merge your main.tf with resources from the remote

## Supported Resources
* `onelogin_apps` => returns all apps
* `onelogin_saml_apps` => returns saml apps only
* `onelogin_oidc_apps` => returns oidc apps only

## Contributing

### Terraform Importer
To add an importable resource, do these things:
1. Under the `terraform/importables` directory, add a file with the scheme <provider>_<resource>.go
2. add a struct to represent your importable, add whatever filtering or special criteria fields you need
3. on the struct, implement the `Importable` interface. this is where we pull all the resources from the remote/api and represent them as resources in terraform
4. in `terraform/import/state.go` add the fields you want to pull from tfstate into main.tf after the import for users to manage later. the state struct is how a resource is represented in .tfstate so in order for json marshalling to work, this struct has to look like your resource in tfstate.
5. in `cmd/terraform-import` add to the `importables` struct `<resource_name>: tfimportables.YourImportable{}` to register it
