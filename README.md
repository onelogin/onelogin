# OneLogin CLI

## Description

The OneLogin CLI is your way to manage OneLogin resources such as Apps, Users, and Mappings via the Command Line.

## Features
`terraform-import <resource>`: Import your OneLogin Apps into a local Terraform State.
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
Apps
