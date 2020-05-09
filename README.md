# OneLogin CLI

## Description

The OneLogin CLI is your way to manage OneLogin resources such as Apps, Users, and Mappings via the Command Line.

## Features
`terraform-import`: Import your OneLogin Apps into a local Terraform State.
This will:
  1. Pull **all** your apps using the /apps endpoint
  2. Establish a basic main.tf that represents all the apps in your account
  3. Call `terraform import` for all the apps and update the `.tfstate`
  4. (W.I.P.) using .tfstate, update sub-components of apps in main.tf (e.g. parameters)

## Usage
This assumes you have Terraform and Go installed

`go install github.com/onelogin/onelogin-cli`

from an empty directory, where you plan to manage your main.tf file run:
`onelogin-cli terraform-import`
