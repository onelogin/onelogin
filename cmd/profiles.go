package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/onelogin/onelogin/profiles"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	legalActions := map[string]interface{}{
		"add":     add,
		"create":  add,
		"list":    list,
		"ls":      list,
		"show":    show,
		"use":     use,
		"edit":    edit,
		"update":  edit,
		"remove":  remove,
		"delete":  remove,
		"which":   current,
		"current": current,
	}
	rootCmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Creates a default OneLogin Profile",
		Long:  "Creates and activates the first OneLogin Profile",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			configFile, err := os.OpenFile(viper.ConfigFileUsed(), os.O_RDWR, 0600)
			if err != nil {
				configFile.Close()
				log.Fatalln("Unable to open profiles file", err)
			}
			profileService := profiles.ProfileService{
				Repository:  profiles.FileRepository{StorageMedia: configFile},
				InputReader: os.Stdin,
			}
			if len(profileService.Index()) > 0 {
				configFile.Close()
				log.Fatalln("Profiles already set up!")
			}
			configFile.Close()
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			configFile, err := os.OpenFile(viper.ConfigFileUsed(), os.O_RDWR, 0600)
			if err != nil {
				configFile.Close()
				log.Fatalln("Unable to open profiles file", err)
			}
			profileService := profiles.ProfileService{
				Repository:  profiles.FileRepository{StorageMedia: configFile},
				InputReader: os.Stdin,
			}
			profileService.Create("default")
			configFile.Close()
		},
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:   "profiles",
		Short: "Manage account settings for the CLI",
		Long: `Maintains a listing of accounts used by the CLI in a home/.onelogin/profiles file
		and facilitates creating, changing, deleting, indexing, and using known configurations.
		You are of course, free to go and edit the profiles file yourself and use this as a way to quickly switch out your environment.
		Available Actions:
			use             [name - required] => CLI will use this profile's credentials in all requests to OneLogin
			show            [name - required] => shows information about the profile
			edit   (update) [name - required] => edits selected profile information
			remove (delete) [name - required] => removes selected profile
			add    (create) [name - required] => adds profile to manage
			list   (ls)     [name - optional] => lists managed profile that can be used. if name given, lists information about that profile
			which  (current)                  => returns current active profile`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				log.Fatalf("Must specify action to perform!")
			}
			action := strings.ToLower(args[0])
			if legalActions[action] == nil {
				log.Fatalf("Illegal Action!")
			}
			switch action {
			case "show", "add", "use", "edit", "update", "remove", "delete":
				if len(args) < 2 {
					log.Fatalf("Profile Name is required for this action!")
				}
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
				profileName := args[1]
				f(profileName, profileService)
			} else if f, ok := legalActions[action].(func(pr profiles.ProfileService)); ok {
				f(profileService)
			} else {
				log.Fatalf("Unexpected Error!")
			}
			configFile.Close()
		},
	})
}

func add(name string, pr profiles.ProfileService) {
	pr.Create(name)
	fmt.Println("Successfully created:", name)
}

func list(pr profiles.ProfileService) {
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

func show(name string, pr profiles.ProfileService) {
	out := pr.Find(name)
	if out != nil {
		printout, _ := json.MarshalIndent(out, "", " ")
		fmt.Println(string(printout))
	}
}

func current(pr profiles.ProfileService) {
	profiles := pr.Index()
	var active string
	for name, p := range profiles {
		if (*p).Active {
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
