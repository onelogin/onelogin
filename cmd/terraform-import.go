package cmd

import (
	"fmt"

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
}
