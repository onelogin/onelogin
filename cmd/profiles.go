package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/onelogin/onelogin/profiles"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
	"strings"
)

func init() {
	legalActions := map[string]interface{}{
		"add":     add,
		"list":    list,
		"ls":      list,
		"use":     use,
		"edit":    edit,
		"update":  edit,
		"remove":  remove,
		"delete":  remove,
		"which":   current,
		"current": current,
	}
	rootCmd.AddCommand(&cobra.Command{
		Use:   "profiles",
		Short: "Manage account settings for the CLI",
		Long: `Maintains a listing of accounts used by the CLI in a home/.onelogin/profiles file
		and facilitates creating, changing, deleting, indexing, and using known configurations.
		You are of course, free to go and edit the profiles file yourself and use this as a way to quickly switch out your environment.
		Available Actions:
			use             [name - required] => exports selected account information to environment
			edit (update)   [name - required] => edits selected account information
			remove (delete) [name - required] => removes selected account
			add             [name - required] => adds account to manage
			list (ls)       [name - optional] => lists managed accounts that can be used. if name given, lists information about that profile
			which (current) []                => returns current active profile`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				log.Fatalf("Must specify action to perform!")
			}
			action := strings.ToLower(args[0])
			if legalActions[action] == nil {
				log.Fatalf("Illegal Action!")
			}
			if (action == "add" || action == "use" || action == "edit" || action == "update" || action == "remove" || action == "delete") && len(args) < 2 {
				log.Fatalf("Profile Name is required for this action!")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			action := args[0]
			configFile, err := os.OpenFile(viper.ConfigFileUsed(), os.O_RDWR, 0600)
			if err != nil {
				configFile.Close()
				log.Fatalln("Unable to open profiles file", err)
			}
			profileService := profiles.ProfileService{
				Repository:  profiles.FileRepository{StorageMedia: configFile},
				InputReader: os.Stdin,
			}
			if f, ok := legalActions[action].(func(s string, pr profiles.ProfileService)); ok {
				if len(args) < 2 {
					f("", profileService)
				} else {
					profileName := args[1]
					f(profileName, profileService)
				}
				configFile.Close()
			} else if f, ok := legalActions[action].(func(pr profiles.ProfileService)); ok {
				f(profileService)
			} else {
				configFile.Close()
				log.Fatalf("Unexpected Error!")
			}
		},
	})
}

func add(name string, pr profiles.ProfileService) {
	pr.Create(name)
	fmt.Println("Successfully created:", name)
}

func list(name string, pr profiles.ProfileService) {
	if name != "" {
		out := pr.Find(name)
		if out != nil {
			printout, _ := json.MarshalIndent(out, "", " ")
			fmt.Println(string(printout))
		}
		return
	}
	out := pr.Index()
	profiles := make([]profiles.Profile, len(out))
	i := 0
	for _, p := range out {
		profiles[i] = *p
		i++
	}
	printout, _ := json.MarshalIndent(profiles, "", " ")
	fmt.Println(string(printout))
}

func current(pr profiles.ProfileService) {
	profiles := pr.Index()
	var active string
	for name, p := range profiles {
		if (*p).Active == true {
			active = name
			break
		}
	}
	fmt.Println("Current Profile:", active)
}

func use(name string, pr profiles.ProfileService) {
	pr.Activate(name)
	fmt.Println("Active profile:", name)
}

func edit(name string, pr profiles.ProfileService) {
	pr.Update(name)
	fmt.Println("Successfully updated:", name)
}

func remove(name string, pr profiles.ProfileService) {
	pr.Remove(name)
	fmt.Println("Successfully removed:", name)
}
