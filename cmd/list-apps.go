package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "list-apps",
		Short: "List OneLogin Apps",
		Long:  `Lists existing OneLogin Apps`,
		Run:   listApps,
	})
}

func listApps(cmd *cobra.Command, args []string) {
	fmt.Println("List Apps!")
}
