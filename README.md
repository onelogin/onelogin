# OneLogin CLI
[![Go Report Card](https://goreportcard.com/badge/github.com/onelogin/onelogin)](https://goreportcard.com/report/github.com/onelogin/onelogin)
<a href='https://github.com/jpoles1/gopherbadger' target='_blank'>![gopherbadger-tag-do-not-edit](https://img.shields.io/badge/Go%20Coverage-84%25-brightgreen.svg?longCache=true&style=flat)</a>
## Description

The OneLogin CLI is your way to manage OneLogin resources such as Apps, Users, and Mappings via the Command Line.

## Features
`onelogin profiles [action] <profile_name>`
Maintains a listing of accounts used by the CLI in a home/.onelogin/profiles file and facilitates creating, changing, deleting, indexing, and using known configurations. You are of course, free to go and edit the profiles file yourself and use this as a way to quickly switch out your environment.
Available Actions:
  * use             [name - required] => CLI will use this profile's credentials in all requests to OneLogin
  * show            [name - required] => shows information about the profile
  * edit   (update) [name - required] => edits selected profile information
  * remove (delete) [name - required] => removes selected profile
  * add    (create) [name - required] => adds profile to manage
  * list   (ls)     [name - optional] => lists managed profile that can be used. if name given, lists information about that profile
  * which  (current)                  => returns current active profile

`onelogin smarthooks [action] <id>`
Creates a .js and .json file with the configuration needed for a Smart Hook and its backing javascript code.
Available Actions:
  * create                    => creates an empty hook.js file and hook.json file with empty required fields in the current working directory
  * list                      => lists the hook IDs associated to your account
  * OneLogin API
  * get     [id - required]   => retrieves the hook and saves it to a hook.js and hook.json file
  * delete  [ids - required]  => accepts a list of IDs to be destroyed via a delete request to OneLogin API
  * 
`terraform-import <resource>`: Import your remote resources into a local Terraform State.
Running this command will do the following:
  1. Pull **all** your resources from the OneLogin API (remote)
  2. Establish a basic main.tf that represents all the apps in your account. Each app will get an empty Terraform resource "placeholder"
  3. Call `terraform import` for all the apps and update the `.tfstate`
  4. Using .tfstate, update main.tf to fill in the editable fields of the resource


## Profiles
Add your OneLogin profiles with `onelogin profiles add <profile_name>`

You'll be prompted for your client_id and client_secret (obtained by creating a set of developer keys in the onelogin admin portal)

You can add as many profiles as you like, and you can switch the active profile with `onelogin profiles use <profile_name>` which will point the CLI at the active account.

## Smart Hooks
From an empty directory, where you plan to manage your Smart Hook run:
`onelogin smarthooks create`<br/><br/>

⚠️ &nbsp; You'll need to do this in a new directory per hook as of `v0.1.10` <br/><br/>

Select the hook type from the propmpt and you'll be presented with 2 files `hook.json` and `hook.js`

You can add package definitions (similar to how you use a `package.json`) and environment variables in the `hook.json` file as well as modify other settings like timeout and retries.

`hook.js` is where the javascript code for your Smart Hook lives. You can use your favorite editor to update the code as you wish. <br/><br/>

⚠️ &nbsp; Do not remove the first line of this javascript. Smart Hooks use `exports.handler = async (context) => {}` as a `main` function.

⚠️ &nbsp; You must also return from your code an object with the `success` node defined. In a new project, this defaults to `return {success: true}` <br/><br/>

To apply changes to your Smart Hook, call the `onelogin smarthooks save` command from inside the directory containing `hook.js` and `hook.json`<br/><br/>

Create an empty [Smart Hook](https://developers.onelogin.com/api-docs/2/smart-hooks/overview) project
```sh
onelogin smarthooks create
```
<br/>

Update a [Smart Hook](https://developers.onelogin.com/api-docs/2/smart-hooks/overview) 
```sh
onelogin smarthooks save
```
<br/>

## Terraform Import
Import all OneLogin apps, create a main.tf file, and establish Terraform state.
```sh
onelogin terraform-import onelogin_apps
```


### Install From Source - Requires Go
clone this repository
from inside the repository `go build ./...` to create a runnable binary
from inside the repository `go install .` to add a the runnable CLI to your GOPATH /bin directory

Alternatively you may run `make install` which just runs the above commands<br/><br/>

### Binaries
There are binaries available for the major platforms in this project's /build directory. Download the
binary for your system and add it to your /bin folder or run it directly per your system's requirements.

* `darwin-amd64`  => mac 64 bit and linux
* `windows-386`   => windows 32 bit
* `windows-amd64` => windows 64 bit
* `linux-386`     => linux 32 bit
* `linux-amd64`   => linux 64 bit

<br/><br/>
## Terraform Importer

### Use
From an empty directory, where you plan to manage your main.tf file run:
`onelogin terraform-import onelogin_apps`

You'll be prompted to confirm the number of resources to import.
This will capture the state of your remote in its entirety

If you have some resources already set up in main.tf, this will merge your main.tf with resources from the remote
<br/><br/>
## Contributing

Fork this repository, make your change and submit a PR to this repository against the `develop` branch.
<br/>

### Terraform Importer
To add an importable resource, do these things:
1. Under the `terraform/importables` directory, add a file with the scheme <provider>_<resource>.go
2. Add a struct to represent your importable, add whatever filtering or special criteria fields you need.
OneLogin importables typically have at least a field for the resource's service from our SDK.
3. On that struct you just made, implement the `Importable` interface. this is where we pull all the resources from the remote/api and represent them as resources in terraform
4. Add structs that represent the fields you want to pull from tfstate into main.tf after the import for users to manage later. the state struct is how a resource is represented in .tfstate so in order for json marshalling to work, this struct has to look like your resource in tfstate.
5. Refer to this in `terraform/import/state.go` in the 'molds' section so the importer is aware of the fields that should be read from tfstate and will marshal the respective data.
6. in `cmd/terraform-import` add to the `importables` struct `<resource_name>: tfimportables.YourImportable{}` to register it
