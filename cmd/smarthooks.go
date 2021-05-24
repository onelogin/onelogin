package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/onelogin/onelogin-go-sdk/pkg/client"
	"github.com/onelogin/onelogin-go-sdk/pkg/oltypes"
	"github.com/onelogin/onelogin-go-sdk/pkg/services/smarthooks"
	smarthookenvs "github.com/onelogin/onelogin-go-sdk/pkg/services/smarthooks/envs"
	"github.com/onelogin/onelogin/clients"
	"github.com/onelogin/onelogin/menu"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	legalActions := map[string]interface{}{
		"new":    newHook,    // new up a hook request boilerplate and js file
		"list":   listHooks,  // list hook names/ids
		"get":    getHook,    // pull down a hook by id
		"deploy": deployHook, // create or update the hook depending on if id is given
		"delete": deleteHook, // deletes the smart hook
	}

	var (
		action         string
		smarthookName  *string
		oneloginClient *client.APIClient
	)

	availableHookTypes := []menu.Option{
		{
			Name: "Pre-Authentication",
			Value: &smarthooks.Options{
				RiskEnabled:          oltypes.Bool(false),
				LocationEnabled:      oltypes.Bool(false),
				MFADeviceInfoEnabled: oltypes.Bool(false),
			},
		},
		{
			Name:  "User-Migration",
			Value: &smarthooks.Options{},
		},
	}

	smarthooksCommand := &cobra.Command{
		Use:   "smarthooks",
		Short: "Assists in managing Smart Hooks in your OneLogin account",
		Long: `Creates a .js and .json file with the configuration needed for a Smart Hook and its backing javascript code.
		Available Actions:
			new                       => creates an empty hook.js file and hook.json file with empty required fields in the current working directory
			list                      => lists the hook IDs associated to your account
			deploy                    => deploys the smart hook defined in the hook.js and hook.json files in the current working directory via a create/update request to OneLogin API
			get     [id - required]   => retrieves the hook and saves it to a hook.js and hook.json file
			delete  [ids - required]  => accepts a list of IDs to be destroyed via a delete request to OneLogin API`,
		PreRun: func(cmd *cobra.Command, args []string) {
			action = strings.ToLower(args[0])
			if legalActions[action] == nil {
				log.Fatalf("Illegal Action!")
			}
			credsFile, err := os.OpenFile(viper.ConfigFileUsed(), os.O_RDWR, 0600)
			if err != nil {
				credsFile.Close()
				log.Println("Unable to open profiles file. Falling back to Environment Variables", err)
			}
			oneloginClient = clients.New(credsFile).OneLoginClient()
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
			} else if f, ok := legalActions[action].(func(name string, defaultHookConfig menu.Option)); ok {
				selectedHookType := menu.Run("Hook Type", "🎣", availableHookTypes)
				if len(args) < 2 {
					f(*smarthookName, selectedHookType)
				} else {
					f(args[1], selectedHookType)
				}
			} else {
				log.Fatalln("Unable to determine function to call")
			}
		},
	}
	smarthookName = smarthooksCommand.Flags().StringP("name", "n", "unnamed", "Smart Hook name")
	rootCmd.AddCommand(smarthooksCommand)
}

func newHook(name string, selectedHookType menu.Option) {
	menuOption := strings.ToLower(selectedHookType.Name)

	name = fmt.Sprintf("%s-%s", name, menuOption)
	workingDir, _ := os.Getwd()
	gitignore := filepath.Join(workingDir, fmt.Sprintf("%s/.gitignore", name))
	jsFileName := filepath.Join(workingDir, fmt.Sprintf("%s/hook.js", name))
	jsonFileName := filepath.Join(workingDir, fmt.Sprintf("%s/hook.json", name))
	envFileName := filepath.Join(workingDir, fmt.Sprintf("%s/.test-env", name))
	pkgFileName := filepath.Join(workingDir, fmt.Sprintf("%s/package.json", name))

	// #nosec G304 forcing the file to be created in the working directory
	err := os.Mkdir(name, 0750)
	if err != nil {
		log.Fatalln("Unable to create project folder")
	}
	os.Chdir(name)
	// #nosec G304 forcing the file to be created in the working directory
	hookJSONFile, err := os.OpenFile(jsonFileName, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		log.Fatalln("Unable to create hook.json ", err)
	}

	// #nosec G304 forcing the file to be created in the working directory
	hookScriptFile, err := os.OpenFile(jsFileName, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		log.Fatalln("Unable to create hook.js ", err)
	}

	// #nosec G304 forcing the file to be created in the working directory
	gitignoreFile, err := os.OpenFile(gitignore, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		log.Fatalln("Unable to create .gitignore ", err)
	}

	gitignoreFile.Write([]byte("test\n.test-env\nnode_modules"))

	_, err = os.Create(envFileName)
	if err != nil {
		log.Fatalln("Unable to create .test-env ", err)
	}

	// #nosec G304 forcing the file to be created in the working directory
	_, err = os.Create(pkgFileName)
	if err != nil {
		log.Fatalln("Unable to create package.json ", err)
	}

	// #nosec G204 running prescribed npm command
	if err := exec.Command("npm", "init", "-y").Run(); err != nil {
		log.Fatal("Problem executing npm init ", err)
	}

	hookCode := []byte(`exports.handler = async (context) => {
	// your code here
	return {
		success: true
	}
}`)

	h := smarthooks.SmartHook{
		Type:     oltypes.String(menuOption),
		Function: oltypes.String(string(hookCode)),
		Disabled: oltypes.Bool(true),
		Runtime:  oltypes.String("nodejs12.x"),
		Retries:  oltypes.Int32(0),
		Timeout:  oltypes.Int32(1),
		Options:  selectedHookType.Value.(*smarthooks.Options),
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
	gitignoreFile.Close()

	fmt.Printf("Created a new %s project.\n", menuOption)
	fmt.Println("To deploy your Smart Hook run 'onelogin smarthooks deploy' from the project directory")
}

func listHooks(client *client.APIClient) {
	hooks, err := client.Services.SmartHooksV1.Query(nil)
	if err != nil {
		log.Fatalln("Unable to query Smart Hooks", err)
	}
	for _, h := range hooks {
		fmt.Println(*h.Type, *h.ID)
	}
}

func getHook(id string, client *client.APIClient) {
	workingDir, _ := os.Getwd()

	// #nosec G304 forcing the file to be created in the working directory
	hookJSONFile, err := os.OpenFile(filepath.Join(workingDir, "hook.json"), os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		log.Fatalln("Unable to create hook.json ", err)
	}

	// #nosec G304 forcing the file to be created in the working directory
	hookScriptFile, err := os.OpenFile(filepath.Join(workingDir, "hook.js"), os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		log.Fatalln("Unable to create hook.js ", err)
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

	if err := exec.Command("npm", "init", "--y").Run(); err != nil {
		log.Fatalln("Problem executing npm init ", err)
	}

	for k, v := range h.Packages {
		// #nosec G204 function call argument is string printer for package name@version fed as argument to npm install
		if err := exec.Command("npm", "install", fmt.Sprintf("%s@%s", k, v)).Run(); err != nil {
			log.Fatalf("Problem executing npm install for %s %s", k, err)
		}
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
	type HookDeps struct {
		Dependencies map[string]string `json:"dependencies,omitempty"`
	}

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

	// #nosec G304 forcing the file to be created in the working directory
	hookPkg, err := ioutil.ReadFile(filepath.Join(workingDir, "package.json"))
	if err != nil {
		log.Fatalln("Unable to read hook.js ", err)
	}

	hook := smarthooks.SmartHook{}
	if err = json.Unmarshal(hookData, &hook); err != nil {
		log.Fatalln("unable to parse smart hook data", err)
	}

	hookDeps := HookDeps{}
	json.Unmarshal(hookPkg, &hookDeps)
	reg, err := regexp.Compile("[^0-9.]+")
	if err != nil {
		log.Fatal(err)
	}
	for k, v := range hookDeps.Dependencies {
		hook.Packages[k] = reg.ReplaceAllString(v, "") // must be exact version ("^0.21.0" must be "0.21.0")
	}
	fmt.Println("ASDF", hook.Packages)

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
