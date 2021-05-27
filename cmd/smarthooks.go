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
		// "test":         testHook,   // passes hook an example context and runs hook in lambda-local
		"delete":       deleteHook, // deletes the smart hook
		"env_vars":     listEnvs,   // lists the established environment variables for the account
		"put_env_vars": putEnvs,    // create or update an environment variable in the account
		"rm_env_vars":  rmEnvs,     // removes the environment variable from the account
	}

	var (
		action         string
		smarthookName  *string
		oneloginClient *client.APIClient
	)

	smarthooksCommand := &cobra.Command{
		Use:   "smarthooks",
		Short: "Assists in managing Smart Hooks in your OneLogin account",
		Long: `Creates a project with all the files and configuration needed for a Smart Hook and its backing javascript code. 
		Once your project is created, you should manage this with your favorite VCS as you would with any other NodeJS project.
		Available Actions:
			new                                        => creates a new smart hook project in a sub-directory of the current working directory, with the given name and hook type.
			list                                       => lists the hook IDs and types of hooks associated to your account.
			deploy                                     => deploys the smart hook defined in the hook.js and hook.json files in the current working directory via a create/update request to OneLogin API.
			test                                       => passes an example context defined in context.json to the hook code and runs it in lambda-local.
			get         [id - required]                => creates a new smart hook project from an existing hook in OneLogin in current directory. ⚠️ Will overwrite existing project! To track changes or treat smart hook like a NodeJS project use a VCS.
			delete      [ids - required]               => accepts a list of IDs to be destroyed via a delete request to OneLogin API.
			
			env_vars                                   => lists the defined environment variable names. E.g. environment variables like FOO=bar BING=baz would turn up [FOO, BING].
			put_env_vars [key=value pairs - required]  => creates or updates the environment variable with the given key. Must be given as FOO=bar BING=baz.
			rm_env_vars  [key - required]              => deletes the environment variable with the given key.`,
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
			} else if f, ok := legalActions[action].(func(name string)); ok {
				if len(args) < 2 {
					f(*smarthookName)
				} else {
					f(args[1])
				}
			} else {
				log.Fatalln("Unable to determine function to call")
			}
		},
	}
	smarthookName = smarthooksCommand.Flags().StringP("name", "n", "unnamed", "Smart Hook name")
	rootCmd.AddCommand(smarthooksCommand)
}

func createBlankFileAsync(name string, wg *sync.WaitGroup) {
	defer wg.Done()
	// #nosec G304 forcing the file to be managed in the working directory
	_, err := os.Create(name)
	if err != nil {
		log.Fatalf("Unable to create %s\n%s", name, err)
	}
}

func writeToFileAsync(name string, data []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	// #nosec G304 forcing the file to be managed in the working directory
	if err := ioutil.WriteFile(name, data, 0600); err != nil {
		log.Printf("Problem writing to %s\n%s", name, err)
	}
}

func execCommandAsync(wg *sync.WaitGroup, name string, args ...string) {
	defer wg.Done()
	// #nosec G204
	if err := exec.Command(name, args...).Run(); err != nil {
		log.Fatalf("Problem executing %s\n%s", name, err)
	}
}

func newHook(name string) {
	menuOption := menu.Run("Hook Type", "🎣", []menu.Option{
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
	})

	hookCode := []byte(`exports.handler = async (context) => {
	// your code here
	return {
		success: true
	}
}`)

	h := smarthooks.SmartHook{
		Type:     oltypes.String(menuOption.Name),
		Function: oltypes.String(string(hookCode)),
		Disabled: oltypes.Bool(true),
		Runtime:  oltypes.String("nodejs12.x"),
		Retries:  oltypes.Int32(0),
		Timeout:  oltypes.Int32(1),
		Options:  menuOption.Value.(*smarthooks.Options),
		EnvVars:  []smarthookenvs.EnvVar{},
		Packages: map[string]string{},
	}
	h.EncodeFunction()

	hookData, _ := json.Marshal(h)
	gitignoreData := []byte("test\n.test-env\nnode_modules")

	name = fmt.Sprintf("%s-%s", name, menuOption.Name)

	workingDir, _ := os.Getwd()
	gitignore := filepath.Join(workingDir, fmt.Sprintf("%s/.gitignore", name))
	hookjs := filepath.Join(workingDir, fmt.Sprintf("%s/hook.js", name))
	hookjson := filepath.Join(workingDir, fmt.Sprintf("%s/hook.json", name))
	testenv := filepath.Join(workingDir, fmt.Sprintf("%s/.test-env", name))

	// #nosec G304 forcing the file to be created in the working directory
	if err := os.Mkdir(name, 0750); err != nil {
		log.Fatalln("Unable to create project folder")
	}
	os.Chdir(name)

	wg := sync.WaitGroup{}
	wg.Add(5)
	go execCommandAsync(&wg, "npm", "init", "-y")
	go createBlankFileAsync(testenv, &wg)
	go writeToFileAsync(gitignore, gitignoreData, &wg)
	go writeToFileAsync(hookjson, hookData, &wg)
	go writeToFileAsync(hookjs, hookCode, &wg)
	wg.Wait()

	fmt.Printf("Created a new %s project.\n", menuOption.Name)
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
	h, err := client.Services.SmartHooksV1.GetOne(id)
	if err != nil {
		log.Fatalln("Unable to query Smart Hooks", err)
	}
	hook, err := json.Marshal(h)
	if err != nil {
		log.Fatalln("Unable to parse Smart Hook response")
	}

	h.DecodeFunction()
	hookCode := []byte(*h.Function)

	workingDir, _ := os.Getwd()
	jsFileName := filepath.Join(workingDir, fmt.Sprintf("%s/hook.js", id))
	jsonFileName := filepath.Join(workingDir, fmt.Sprintf("%s/hook.json", id))
	// #nosec G304 forcing the file to be created in the working directory
	if err = os.Mkdir(id, 0750); err != nil {
		log.Fatalln("Unable to create project folder")
	}
	os.Chdir(id)

	wg := sync.WaitGroup{}
	wg.Add(2)
	if err := exec.Command("npm", "init", "--y").Run(); err != nil {
		log.Fatalln("Problem executing npm init ", err)
	}

	for k, v := range h.Packages {
		wg.Add(1)
		go execCommandAsync(&wg, "npm", "install", fmt.Sprintf("%s@%s", k, v))
	}
	go writeToFileAsync(jsonFileName, hook, &wg)
	go writeToFileAsync(jsFileName, hookCode, &wg)
	wg.Wait()
}

func deployHook(client *client.APIClient) {
	type HookDeps struct {
		Dependencies map[string]string `json:"dependencies,omitempty"`
	}

	workingDir, _ := os.Getwd()
	// #nosec G304 forcing the file to be read from the working directory
	hookData, err := ioutil.ReadFile(filepath.Join(workingDir, "hook.json"))
	if err != nil {
		log.Fatalln("Unable to read hook.json ", err)
	}

	// #nosec G304 forcing the file to be read from the working directory
	hookCode, err := ioutil.ReadFile(filepath.Join(workingDir, "hook.js"))
	if err != nil {
		log.Fatalln("Unable to read hook.js ", err)
	}

	// #nosec G304 forcing the file to be read from the working directory
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

	h, err := json.Marshal(savedHook)
	if err != nil {
		log.Fatalln("Unable to parse Smart Hook response")
	}

	savedHook.DecodeFunction()
	savedHookCode := []byte(*savedHook.Function)

	wg := sync.WaitGroup{}
	go writeToFileAsync(filepath.Join(workingDir, "hook.json"), h, &wg)
	go writeToFileAsync(filepath.Join(workingDir, "hook.js"), savedHookCode, &wg)
	wg.Wait()
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

func listEnvs(client *client.APIClient) {
	vars, err := client.Services.SmartHooksEnvVarsV1.Query(nil)
	if err != nil {
		log.Fatalln("Unable to query Smart Hook Environment Variables", err)
	}
	for _, ev := range vars {
		fmt.Println(*ev.Name, *ev.ID)
	}
}

func putEnvs(vars []string, client *client.APIClient) {
	wg := sync.WaitGroup{}
	existingVarsResp, err := client.Services.SmartHooksEnvVarsV1.Query(nil)
	if err != nil {
		log.Fatalln("Unable to query Smart Hook Environment Variables", err)
	}
	existing := map[string]*smarthookenvs.EnvVar{} // map by name for easier lookup later
	for i, ev := range existingVarsResp {
		existing[*ev.Name] = &existingVarsResp[i]
	}

	for _, v := range vars {
		d := strings.Split(v, "=") // split FOO=bar to ["FOO", "bar"] tuples. Key is first value
		if len(d) != 2 {
			log.Fatalln("Malformatted environment variable key value pairs given")
		} else {
			wg.Add(1)
			if existing[d[0]] != nil {
				e := existing[d[0]]
				e.Value = oltypes.String(d[1])
				go func(ev *smarthookenvs.EnvVar, wg *sync.WaitGroup) {
					defer wg.Done()
					fmt.Println("Updating", *ev.Name, *ev.ID)
					if ev, err := client.Services.SmartHooksEnvVarsV1.Update(ev); err != nil {
						log.Println("Unable to update environment variable with id:", *ev.ID, err)
					} else {
						log.Println("Updated environment variable ", *ev.ID)
					}
				}(e, &wg)
			} else {
				r := smarthookenvs.EnvVar{Name: oltypes.String(d[0]), Value: oltypes.String(d[1])}
				go func(ev *smarthookenvs.EnvVar, wg *sync.WaitGroup) {
					defer wg.Done()
					fmt.Println("Creating", *r.Name)
					if ev, err := client.Services.SmartHooksEnvVarsV1.Create(ev); err != nil {
						log.Println("Unable to update environment variable with id:", *ev.ID, err)
					} else {
						log.Println("Updated environment variable", *ev.ID)
					}
				}(&r, &wg)
			}
		}
	}
	wg.Wait()
	log.Println("Finished updating environment variables")
}

func rmEnvs(vars []string, client *client.APIClient) {
	wg := sync.WaitGroup{}
	existingVarsResp, err := client.Services.SmartHooksEnvVarsV1.Query(nil)
	if err != nil {
		log.Fatalln("Unable to query Smart Hook Environment Variables", err)
	}

	existing := map[string]string{} // name: id
	for i, ev := range existingVarsResp {
		existing[*ev.Name] = *existingVarsResp[i].ID
	}

	for _, v := range vars {
		if existing[v] != "" {
			wg.Add(1)
			go func(id string, wg *sync.WaitGroup) {
				defer wg.Done()
				if err := client.Services.SmartHooksEnvVarsV1.Destroy(id); err != nil {
					log.Println("Unable to delete environment variable", id, err)
				} else {
					log.Println("Deleted environment variable", id)
				}
			}(existing[v], &wg)
		}
	}

	wg.Wait()
	log.Println("Finished removing environment variables")
}
