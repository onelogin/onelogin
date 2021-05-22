package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/manifoldco/promptui"
	"github.com/onelogin/onelogin-go-sdk/pkg/client"
	"github.com/onelogin/onelogin-go-sdk/pkg/oltypes"
	"github.com/onelogin/onelogin-go-sdk/pkg/services/smarthooks"
	smarthookenvs "github.com/onelogin/onelogin-go-sdk/pkg/services/smarthooks/envs"
	"github.com/onelogin/onelogin/profiles"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	legalActions := map[string]interface{}{
		"create": newHook,    // new up a hook request boilerplate and js file
		"list":   listHooks,  // list hook names/ids
		"get":    getHook,    // pull down a hook by id
		"deploy": deployHook, // create or update the hook depending on if id is given
		"delete": deleteHook, // deletes the smart hook
	}

	var action string
	var oneloginClient *client.APIClient

	rootCmd.AddCommand(&cobra.Command{
		Use:   "smarthooks",
		Short: "Assists in managing Smart Hooks in your OneLogin account",
		Long: `Creates a .js and .json file with the configuration needed for a Smart Hook and its backing javascript code.
		Available Actions:
			create                    => creates an empty hook.js file and hook.json file with empty required fields in the current working directory
			list                      => lists the hook IDs associated to your account
			deploy                    => deploys the smart hook defined in the hook.js and hook.json files in the current working directory via a create/update request to OneLogin API
			get     [id - required]   => retrieves the hook and saves it to a hook.js and hook.json file
			delete  [ids - required]  => accepts a list of IDs to be destroyed via a delete request to OneLogin API`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			action = strings.ToLower(args[0])

			if legalActions[action] == nil {
				log.Fatalf("Illegal Action!")
			}
			configFile, err := os.OpenFile(viper.ConfigFileUsed(), os.O_RDWR, 0600)
			if err != nil {
				configFile.Close()
				log.Println("Unable to open profiles file. Falling back to Environment Variables", err)
			}
			profile := profiles.ProfileService{
				Repository: profiles.FileRepository{
					StorageMedia: configFile,
				},
			}.GetActive()

			clientConfigs := client.APIClientConfig{}
			if profile == nil {
				fmt.Println("No active profile detected. Authenticating with environment variables")
				clientConfigs.ClientID = os.Getenv("ONELOGIN_CLIENT_ID")
				clientConfigs.ClientSecret = os.Getenv("ONELOGIN_CLIENT_SECRET")
				clientConfigs.Url = os.Getenv("ONELOGIN_OAPI_URL")
			} else {
				clientConfigs.ClientID = (*profile).ClientID
				clientConfigs.ClientSecret = (*profile).ClientSecret
				clientConfigs.Url = fmt.Sprintf("https://api.%s.onelogin.com", (*profile).Region)
			}

			oneloginClient, err = client.NewClient(&clientConfigs)
			if err != nil {
				log.Fatalln("Unable to initialize OneLogin client")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			// check function signature for action and ensure correct number of arguments given
			if f, ok := legalActions[action].(func(s string, client *client.APIClient)); ok {
				if len(args) < 2 {
					log.Fatalf("One argument is required for this action!")
				}
				f(args[1], oneloginClient)
			} else if f, ok := legalActions[action].(func(s []string, client *client.APIClient)); ok {
				if len(args) < 2 {
					log.Fatalf("At least one argument is required for this action!")
				}
				f(args[1:], oneloginClient)
			} else if f, ok := legalActions[action].(func(client *client.APIClient)); ok {
				f(oneloginClient)
			} else if f, ok := legalActions[action].(func()); ok {
				f()
			} else {
				log.Fatalln("Unable to determine function to call")
			}
		},
	})
}

func newHook() {
	workingDir, _ := os.Getwd()
	var existJSON bool
	var existJS bool

	if _, err := os.Stat(filepath.Join(workingDir, "hook.json")); err != nil {
		existJSON = true
	}
	if _, err := os.Stat(filepath.Join(workingDir, "hook.js")); err != nil {
		existJS = true
	}

	if existJSON && existJS {
		// #nosec G304 forcing the file to be created in the working directory
		hookJSONFile, err := os.OpenFile(filepath.Join(workingDir, "hook.json"), os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			log.Fatalln("Unable to read hook.json ", err)
		}

		// #nosec G304 forcing the file to be created in the working directory
		hookScriptFile, err := os.OpenFile(filepath.Join(workingDir, "hook.js"), os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			log.Fatalln("Unable to read hook.js ", err)
		}

		hookCode := []byte(`exports.handler = async (context) => {
	// your code here
	return {
		success: true
	}
}`)

		type hookType struct {
			Type           string
			Value          string
			DefaultOptions *smarthooks.Options
		}

		availableHookTypes := []hookType{
			hookType{
				Type: "Pre-Authentication",
				DefaultOptions: &smarthooks.Options{
					RiskEnabled:          oltypes.Bool(false),
					LocationEnabled:      oltypes.Bool(false),
					MFADeviceInfoEnabled: oltypes.Bool(false),
				},
			},
			hookType{
				Type:           "User-Migration",
				DefaultOptions: &smarthooks.Options{},
			},
		}

		templates := promptui.SelectTemplates{
			Active:   `🎣 {{ .Type | cyan | bold }}`,
			Inactive: `   {{ .Type | cyan }}`,
			Selected: `{{ "✔" | green | bold }} {{ "Hook Type" | bold }}: {{ .Type | cyan }}`,
		}

		list := promptui.Select{
			Label:     "Hook Type",
			Items:     availableHookTypes,
			Templates: &templates,
		}

		idx, _, err := list.Run()
		if err != nil {
			log.Fatalln(err)
		}

		h := smarthooks.SmartHook{
			Type:     oltypes.String(strings.ToLower(availableHookTypes[idx].Type)),
			Function: oltypes.String(string(hookCode)),
			Disabled: oltypes.Bool(false),
			Runtime:  oltypes.String("nodejs12.x"),
			Retries:  oltypes.Int32(0),
			Timeout:  oltypes.Int32(1),
			Options:  availableHookTypes[idx].DefaultOptions,
			EnvVars:  []smarthookenvs.EnvVar{},
			Packages: map[string]string{},
		}
		h.EncodeFunction()
		hook, _ := json.Marshal(h)

		if _, err = hookJSONFile.Write(hook); err != nil {
			hookJSONFile.Close()
			log.Fatal("Problem creating hook.json file", err)
		}

		if _, err = hookScriptFile.Write(hookCode); err != nil {
			hookScriptFile.Close()
			log.Fatal("Problem creating hook.js file", err)
		}
		hookJSONFile.Close()
		hookScriptFile.Close()
	} else {
		log.Println("Project already set up")
	}

}

func listHooks(client *client.APIClient) {
	hooks, err := client.Services.SmartHooksV1.Query(nil)
	if err != nil {
		log.Fatalln("Unable to query Smart Hooks", err)
	}
	for _, h := range hooks {
		fmt.Println(*h.ID)
	}
}

func getHook(id string, client *client.APIClient) {
	workingDir, _ := os.Getwd()

	// #nosec G304 forcing the file to be created in the working directory
	hookJSONFile, err := os.OpenFile(filepath.Join(workingDir, "hook.json"), os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		log.Fatalln("Unable to read hook.json ", err)
	}

	// #nosec G304 forcing the file to be created in the working directory
	hookScriptFile, err := os.OpenFile(filepath.Join(workingDir, "hook.js"), os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		log.Fatalln("Unable to read hook.js ", err)
	}

	h, err := client.Services.SmartHooksV1.GetOne(id)
	if err != nil {
		log.Fatalln("Unable to query Smart Hooks", err)
	}

	hook, _ := json.Marshal(h)

	if _, err = hookJSONFile.Write(hook); err != nil {
		hookJSONFile.Close()
		log.Fatal("Problem creating hook.json file", err)
	}

	h.DecodeFunction()

	hookCode := []byte(*h.Function)
	if _, err = hookScriptFile.Write(hookCode); err != nil {
		hookScriptFile.Close()
		log.Fatal("Problem creating hook.js file", err)
	}

	hookJSONFile.Close()
	hookScriptFile.Close()
}

func deployHook(client *client.APIClient) {
	workingDir, _ := os.Getwd()
	// #nosec G304 forcing the file to be created in the working directory
	hookData, err := ioutil.ReadFile(filepath.Join(workingDir, "hook.json"))
	if err != nil {
		log.Fatalln("Unable to read hook.json ", err)
	}

	// #nosec G304 forcing the file to be created in the working directory
	hookCode, err := ioutil.ReadFile(filepath.Join(workingDir, "hook.js"))
	if err != nil {
		log.Fatalln("Unable to read hook.js ", err)
	}

	hook := smarthooks.SmartHook{}
	if err = json.Unmarshal(hookData, &hook); err != nil {
		log.Fatalln("unable to parse smart hook data", err)
	}

	hook.Function = oltypes.String(string(hookCode))
	hook.EncodeFunction()

	hook.CreatedAt = nil
	hook.UpdatedAt = nil
	hook.Status = nil
	var savedHook *smarthooks.SmartHook
	if hook.ID != nil {
		savedHook, err = client.Services.SmartHooksV1.Update(&hook)
	} else {
		savedHook, err = client.Services.SmartHooksV1.Create(&hook)
	}
	if err != nil {
		log.Fatal("Problem saving the hook", err)
	}
	h, _ := json.Marshal(savedHook)

	ioutil.WriteFile(filepath.Join(workingDir, "hook.json"), h, 0600)

	savedHook.DecodeFunction()

	savedHookCode := []byte(*savedHook.Function)
	ioutil.WriteFile(filepath.Join(workingDir, "hook.js"), savedHookCode, 0600)
}

func deleteHook(ids []string, client *client.APIClient) {
	wg := sync.WaitGroup{}
	for _, id := range ids {
		wg.Add(1)
		go func(id string, wg *sync.WaitGroup) {
			defer wg.Done()
			if err := client.Services.SmartHooksV1.Destroy(id); err != nil {
				log.Println("Unable to delete hook with id: ", id)
			} else {
				log.Println("Successfully able to delete hook with id: ", id)
			}
		}(id, &wg)
	}
	wg.Wait()
	log.Println("Finished deleting hooks")
}
