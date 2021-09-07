/*
Copyright © 2020 OneLogin Inc dominick.caponi@onelogin.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "onelogin",
	Short: "A CLI for managing IAM and Authentication resources",
	Long:  `The OneLogin CLI provides a convenient interface for managing OneLogin resources from the command line such as Apps and User Mappings. `,
	Run:   func(cmd *cobra.Command, args []string) { fmt.Println("Welcome to OneLogin") },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.onelogin.json)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// Add Version output command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number of OneLogin",
		Long:  `All software has versions. This is OneLogin's`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("v0.1.19")
		},
	})
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory with name ".onelogin" (without extension).
		viper.AddConfigPath(fmt.Sprintf("%s/.onelogin", home))
		viper.SetConfigName("profiles")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		p := filepath.Join(home, ".onelogin")
		os.Mkdir(p, 0750)
		p = filepath.Join(p, "profiles.json")
		_, err := os.Create(p)
		if err != nil {
			log.Fatalln("Unable to create config file!")
		}
		viper.ReadInConfig()
	}
}
