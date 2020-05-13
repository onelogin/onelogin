# OneLogin CLI

## Description

The OneLogin CLI is your way to manage OneLogin resources such as Apps, Users, and Mappings via the Command Line.

## Features
`terraform-import <resource>`: Import your remote resources into a local Terraform State.
This will:
  1. Pull **all** your resources from the OneLogin API (remote)
  2. Establish a basic main.tf that represents all the apps in your account. Each app will get an empty Terraform resource "placeholder"
  3. Call `terraform import` for all the apps and update the `.tfstate`
  4. Using .tfstate, update main.tf to fill in the editable fields of the resource

## Usage
This assumes you have Terraform and Go installed.

`go install github.com/onelogin/onelogin-cli`

from an empty directory, where you plan to manage your main.tf file run:
`onelogin-cli terraform-import apps`

You'll be prompted to confirm importing apps. This will capture the state of your remote in its entirety

If you have some resources already set up in main.tf, this will merge your main.tf with resources from the remote

## Supported Resources
`onelogin_apps` => returns all apps
`onelogin_saml_apps` => returns saml apps only
`onelogin_oidc_apps` => returns oidc apps only

## Contributing

### Terraform Importer
To add an importable resource, do these things:
1. Under the `terraform` directory, add a file with the scheme <provider>_<resource>.go
2. add a struct to represent your importable, add whatever filtering or special criteria fields you need
3. on the struct, implement the `Importable` interface. this is where we pull all the resources from the remote/api and represent them as resources in terraform
4. in `terraform/state.go` add the fields you want to pull from tfstate into main.tf after the import for users to manage later. the state struct is how a resource is represented in .tfstate so in order for json marshalling to work, this struct has to look like your resource in tfstate.
5. in `cmd/terraform-import` add to the `importables` struct `<resource_name>: terraform.YourImportable{}`
