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
	legalActions := map[string]interface{}{"add": add, "list": list, "use": use, "edit": edit, "remove": remove}
	rootCmd.AddCommand(&cobra.Command{
		Use:   "profiles",
		Short: "Manage account settings for the CLI",
		Long: `Maintains a listing of accounts used by the CLI in a home/.onelogin/profiles file
		and facilitates creating, changing, deleting, indexing, and using known configurations.
		You are of course, free to go and edit the profiles file yourself and use this as a way to quickly switch out your environment.
		Available Actions:
			use    [name - required] => exports selected account information to environment
			edit   [name - required] => edits selected account information
			remove [name - required] => removes selected account
			add    [name - required] => adds account to manage
			list   [name - optional] => lists managed accounts that can be used. if name given, lists information about that profile`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				log.Fatalf("Must specify action to perform!")
			}
			action := strings.ToLower(args[0])
			if legalActions[action] == nil {
				log.Fatalf("Illegal Action!")
			}
			if (action == "add" || action == "use" || action == "edit" || action == "remove") && len(args) < 2 {
				log.Fatalf("Profile Name is required for this action!")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			action := args[0]
			profileConfig := viper.ConfigFileUsed()
			f, err := os.OpenFile(profileConfig, os.O_RDWR, 0600)
			if err != nil {
				log.Fatalln("Unable to open profiles file")
			}
			profileService := profiles.ProfileService{
				Repository:  profiles.FileRepository{StorageMedia: f},
				InputReader: os.Stdin,
			}
			if f, ok := legalActions[action].(func(s string, pr profiles.ProfileService)); ok {
				if len(args) < 2 {
					f("", profileService)
				} else {
					profileName := args[1]
					f(profileName, profileService)
				}
			} else {
				log.Fatalf("Unexpected Error!")
			}
		},
	})
}

func add(name string, pr profiles.ProfileService) {
	pr.Create(name)
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

func use(name string, pr profiles.ProfileService) {
	pr.Activate(name)
}

func edit(name string, pr profiles.ProfileService) {
	pr.Update(name)
}

func remove(name string, pr profiles.ProfileService) {
	pr.Remove(name)
}
