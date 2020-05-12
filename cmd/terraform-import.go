package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/onelogin/onelogin-cli/terraform"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "terraform-import",
		Short: "Import apps to local Terraform state",
		Long: `Uses Terraform Import to collect existing OneLogin apps
    and create new .tfstate and .tf files so you can begin managing
    existing resources with Terraform.`,
		Run: terraformImport,
	})
}

func terraformImport(cmd *cobra.Command, args []string) {
	fmt.Println("Terraform Import!")
	terraform.SetUpTerraformFile()
	log.Println("Initializing Terraform with 'terraform init'...")
	exec.Command("terraform", "init").Run()

	var shells []terraform.Shell

	switch strings.ToLower(args[0]) {
	case "apps":
		shells = terraform.CreateImportShells()
	default:
		fmt.Println("Must specify resource to import!")
		fmt.Println("Available resources: apps")
		os.Exit(1)
	}

	resources := terraform.ImportTFState(shells)
	terraform.WriteFinalMainTF(resources)
}
