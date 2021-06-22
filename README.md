# OneLogin CLI
[![Go Report Card](https://goreportcard.com/badge/github.com/onelogin/onelogin)](https://goreportcard.com/report/github.com/onelogin/onelogin)
<a href='https://github.com/jpoles1/gopherbadger' target='_blank'>![gopherbadger-tag-do-not-edit](https://img.shields.io/badge/Go%20Coverage-82%25-brightgreen.svg?longCache=true&style=flat)</a>

## Description

The OneLogin CLI is your way to manage OneLogin resources such as Apps, Users, and Mappings via the Command Line.
<br/><br/>

## Get Started

### Install From Source - Requires Go
clone this repository
from inside the repository `go build ./...` to create a runnable binary
from inside the repository `go install .` to add a the runnable CLI to your GOPATH /bin directory

Alternatively you may run `make install` which just runs the above commands
<br/><br/>

### Install with Homebrew (Mac OS Only)
brew install onelogin/tap-onelogin/onelogin

Or brew tap onelogin/tap-onelogin and then brew install onelogin.
<br/><br/>

### Binaries
There are binaries available for the major platforms in this project's /build directory. Download the
binary for your system and add it to your /bin folder or run it directly per your system's requirements.

* `darwin-amd64`  => mac 64 bit and linux
* `windows-386`   => windows 32 bit
* `windows-amd64` => windows 64 bit
* `linux-386`     => linux 32 bit
* `linux-amd64`   => linux 64 bit

#### Install Binary on Mac
Download and extract the `darwin-amd64` package from the release artifacts list

`tar -xvf darwin-amd64.tar.gz && sudo mv build/darwin-amd64/onelogin-darwin-amd64 /usr/local/bin/onelogin` (you can specifiy a different name to invoke such as `usr/local/bin/ol`)

You'll likely get hit with a security warning when you try running `onelogin` for the first time. 

To fix, go to System Preferences > Security & Privacy

you'll be presented with the warning about the binary not being from an identified developer. Allow this app to run.

Try running the command again and click Open from the popup and you should be good to go.

#### Install Binary on Windows
Download and extract the `windows-amd64` package from the release artifacts list

Navigate to the extracted folder which should be in your Downloads folder and navigate to the .exe file (build > windows-amd64).

Create a folder in Program Files (Program Files \ Onelogin) and add the .exe to that folder.

Add Program Files \ Onelogin to your path by changing the environment variables

  Hit the window key and type path. Select "Edit the system environment variables"
  Toward the bottom on the Advanced tab, select "Environment Variables"
  In the System variables list, click the Path variable on the list and click "Edit"
  Click "New" and add `C:\Program Files\Onelogin`
  Click OK on all the windows

Open a `Cmd` window and start using `onelogin`


/usr/local/bin/onelogin

## Features
`onelogin profiles [action] <profile_name>`
Maintains a listing of accounts used by the CLI in a home/.onelogin/profiles file and facilitates creating, changing, deleting, indexing, and using known configurations. You are of course, free to go and edit the profiles file yourself and use this as a way to quickly switch out your environment.
Available Actions:
```
use             [name - required] => CLI will use this profile's credentials in all requests to OneLogin
show            [name - required] => shows information about the profile
edit   (update) [name - required] => edits selected profile information
remove (delete) [name - required] => removes selected profile
add    (create) [name - required] => adds profile to manage
list   (ls)     [name - optional] => lists managed profile that can be used. if name given, lists information about that profile
which  (current)                  => returns current active profile
```

`onelogin smarthooks [action] <id>`
Creates a .js and .json file with the configuration needed for a Smart Hook and its backing javascript code.
Available Actions:
```
new                                        => creates a new smart hook project in a sub-directory of the current working directory, with the given name and hook type.
list                                       => lists the hook IDs and types of hooks associated to your account.
deploy                                     => deploys the smart hook defined in the hook.js and hook.json files in the current working directory via a create/update request to OneLogin API.
test                                       => passes an example context defined in context.json to the hook code and runs it in lambda-local.
get         [id - required]                => creates a new smart hook project from an existing hook in OneLogin in current directory. ⚠️ Will overwrite existing project! To track changes or treat smart hook like a NodeJS project use a VCS.
delete      [ids - required]               => accepts a list of IDs to be destroyed via a delete request to OneLogin API.

env_vars                                   => lists the defined environment variable names. E.g. environment variables like FOO=bar BING=baz would turn up [FOO, BING].
put_env_vars [key=value pairs - required]  => creates or updates the environment variable with the given key. Must be given as FOO=bar BING=baz.
rm_env_vars  [key - required]              => deletes the environment variable with the given key.
```

`terraform-import <resource>`: Import your remote resources into a local Terraform State.
Running this command will do the following:
  1. Pull **all** your resources from the OneLogin API (remote)
  2. Establish a basic main.tf that represents all the apps in your account. Each app will get an empty Terraform resource "placeholder"
  3. Call `terraform import` for all the apps and update the `.tfstate`
  4. Using .tfstate, update main.tf to fill in the editable fields of the resource

<br/>

## Profiles
Add your OneLogin profiles with `onelogin profiles add <profile_name>`

You'll be prompted for your client_id and client_secret (obtained by creating a set of developer keys in the onelogin admin portal)

You can add as many profiles as you like, and you can switch the active profile with `onelogin profiles use <profile_name>` which will point the CLI at the active account.
<br/><br/>

## Smart Hooks
From an empty directory, where you plan to manage your Smart Hook run:
`onelogin smarthooks create`<br/>

Select the hook type from the propmpt and you'll be presented with some files 

`hook.json` - Config file for your Smart Hook where you can modify things like timeout and retries. 

&emsp; ⚠️ &nbsp; Do NOT modify the `function`, `env_vars`, `packages`, or `type`! This tool will handle that for you.</br>

`hook.js` - The good stuff. This is your Smart Hook code that gets run every time the triggering event happens.

&emsp; ⚠️ &nbsp; Do not remove the exports line. Smart Hooks use `exports.handler = async (context) => {}` as its `main` function.

&emsp; ⚠️ &nbsp; You must also return from your code an object with the `success` node defined. In a new project, this defaults to `return {success: true}` <br/>

`.env` - Where you manage environment variables as you would in other NodeJS projects

`.gitignore` - Lists things that are not recommended for check in to version control (like test artifacts, and `.env`)


<br/>

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

To run tests against your Smart Hook code:
* Create a `context.json` file (the input to the function). You can get one for [Pre-Authentication](https://github.com/onelogin/smarthooks-sdk/blob/master/src/preAuthentication/exampleContexts.js) or [User-Migration](https://github.com/onelogin/smarthooks-sdk/blob/master/src/userMigration/exampleContexts.js)
* Run `onelogin smarthooks test` from inside your Smart Hook Project
* Results will print to the screen

## Terraform Import
Import all OneLogin apps, create a main.tf file, and establish Terraform state.

From an empty directory, where you plan to manage your main.tf file run:
```sh
onelogin terraform-import onelogin_apps
```

If you have pre-existing resources defined in `main.tf` the tool is smart enough to merge those definitions. <br/><br/>

## Contributing
### Generally

Fork this repository, make your change and submit a PR to this repository against the `develop` branch.
<br/><br/>

### Adding Resources for Import - Terraform Importer
To add an importable resource, do these things:
1. Under the `terraform/importables` directory, add a file with the scheme <provider>_<resource>.go
2. Add a struct to represent your importable, add whatever filtering or special criteria fields you need.
OneLogin importables typically have at least a field for the resource's service from our SDK.
3. On that struct you just made, implement the `Importable` interface. this is where we pull all the resources from the remote/api and represent them as resources in terraform
4. Add structs that represent the fields you want to pull from tfstate into main.tf after the import for users to manage later. the state struct is how a resource is represented in .tfstate so in order for json marshalling to work, this struct has to look like your resource in tfstate.
5. Refer to this in `terraform/import/state.go` in the 'molds' section so the importer is aware of the fields that should be read from tfstate and will marshal the respective data.
6. in `cmd/terraform-import` add to the `importables` struct `<resource_name>: tfimportables.YourImportable{}` to register it
