package terraform

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/onelogin/onelogin-cli/utils"
	"github.com/onelogin/onelogin-go-sdk/pkg/models"
)

// Shell represents the resource to be imported
type Shell struct {
	ImportCommand *exec.Cmd
	Content       []byte
	Name          string
	Type          string
}

// PrepareTFImport creates the representation of the resource to be imported
// with identifying information and import commands
func (shell *Shell) PrepareTFImport() {
	shell.Content = append(shell.Content, []byte(fmt.Sprintf("resource %s %s {}\n", shell.Type, shell.Name))...)
	arg1 := fmt.Sprintf("%s.%s", shell.Type, shell.Name)
	pos := strings.Index(arg1, "-")
	id := arg1[pos+1 : len(arg1)]
	shell.ImportCommand = exec.Command("terraform", "import", arg1, id)
}

// State in memory representation of tfstate
type State struct {
	Resources []Resource `json:"resources"`
}

// Resource represents each resource type's list of resources
type Resource struct {
	Content   []byte
	Name      string             `json:"name"`
	Type      string             `json:"type"`
	Instances []resourceInstance `json:"instances"`
}

type resourceInstance struct {
	Data instanceData `json:"attributes"`
}

type instanceData struct {
	AllowAssumedSignin *bool                     `json:"allow_assumed_signin"`
	ConnectorID        *int                      `json:"connector_id"`
	Description        *string                   `json:"description"`
	Name               *string                   `json:"name"`
	Notes              *string                   `json:"notes"`
	Visible            *bool                     `json:"visible"`
	Provisioning       []models.AppProvisioning  `json:"provisioning"`
	Parameters         []models.AppParameters    `json:"parameters"`
	Configuration      []models.AppConfiguration `json:"configuration"`
}

func filterFulfilledShells(f *os.File, shells []Shell) []Shell {
	resourceMatch := regexp.MustCompile(`\b(\w*resource\w*)\b\s[a-zA-Z\_]*\s[a-zA-Z\_\-]*[0-9]*\s\{`)
	existing := make(map[string]int)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		t := scanner.Text()
		if resourceMatch.MatchString(t) {
			if existing[t] > 0 {
				existing[t]++
			} else {
				existing[t] = 1
			}
		}
	}
	var filtered []Shell
	for _, shell := range shells {
		if existing[strings.TrimSuffix(string(shell.Content), "}\n")] < 1 {
			filtered = append(filtered, shell)
		}
	}
	return filtered
}

// SetUpTerraformFile will create or overwrite main.tf
func SetUpTerraformFile() {
	path, _ := os.Getwd()
	p := filepath.Join(path, ("/main.tf"))
	f, err := os.OpenFile(p, os.O_RDWR, 0666)
	if err != nil {
		log.Println("Unable to open main.tf, creating instead")
		f, err = os.OpenFile(p, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Fatal("ERROR CREATING FILE", err)
		}
	}
	defer f.Close()

	log.Println("Creating Terraform Import File...")
	buffer := []byte("provider onelogin {}\n\n")
	f.Write(buffer)

}

// ImportTFState writes the resource shells to main.tf and calls each
// resource's terraform import command to update tfstate
func ImportTFState(shells []Shell) []Resource {
	path, _ := os.Getwd()
	p := filepath.Join(path, ("/main.tf"))
	f, err := os.OpenFile(p, os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Println("Unable to open main.tf")
	}
	defer f.Close()

	shells = filterFulfilledShells(f, shells)

	if len(shells) == 0 {
		fmt.Println("No new resources to import from remote")
		os.Exit(0)
	}

	fmt.Printf("This will import %d resources. Do you want to continue? (y/n): ", len(shells))
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	text := strings.ToLower(input.Text())
	if text != "y" && text != "yes" {
		fmt.Printf("User aborted operation!")
		os.Exit(0)
	}

	var buffer []byte
	for _, shell := range shells {
		buffer = append(buffer, shell.Content...)
	}
	if _, err := f.Write(buffer); err != nil {
		log.Fatal("Problem creating import file", err)
	}

	for i, shell := range shells {
		log.Printf("Importing resource %d of %d", i+1, len(shells))
		if err := shell.ImportCommand.Run(); err != nil {
			log.Fatal("Problem executing terraform import", shell.ImportCommand.Args, err)
		}
	}
	state := readTFStateJSON()
	return state.Resources
}

// WriteFinalMainTF reads .tfstate and updates the main.tf as if the tfstate was
// created with the ensuing main.tf file
func WriteFinalMainTF(resources []Resource) {
	path, _ := os.Getwd()
	p := filepath.Join(path, ("/main.tf"))
	f, err := os.OpenFile(p, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("Error finding or creating main.tf ", err)
	}
	defer f.Close()
	log.Println("Assembling main.tf...")
	buffer := []byte("provider onelogin {}\n\n")
	for _, resource := range resources {
		for _, instance := range resource.Instances {
			resource.Content = append(resource.Content, []byte(fmt.Sprintf("resource %s %s {\n", resource.Type, resource.Name))...)
			resource.Content = append(resource.Content, resourceBaseToHCL(instance.Data, 1)...)
			resource.Content = append(resource.Content, []byte("}\n\n")...)
		}
		buffer = append(buffer, resource.Content...)
	}
	_, err = f.WriteAt(buffer, 0)
	if err != nil {
		fmt.Println("ERROR Writing Final main.tf", err)
	}
}

func readTFStateJSON() State {
	log.Println("Collecting State from State File")

	path, _ := os.Getwd()
	p := filepath.Join(path, "/terraform.tfstate")
	data, err := ioutil.ReadFile(p)
	if err != nil {
		log.Fatal("Unable to Read tfstate")
	}
	v := State{}

	if err := json.Unmarshal(data, &v); err != nil {
		log.Fatal("Unable to Translate tfstate in Memory")
	}
	return v
}

func resourceBaseToHCL(input interface{}, indentLevel int) []byte {
	var out []byte

	tp := reflect.TypeOf(input)
	vl := reflect.ValueOf(input)

	for i := 0; i < tp.NumField(); i++ {
		line := make([]byte, indentLevel)
		for i := 0; i < indentLevel; i++ {
			line[i] = byte('\t')
		}

		field := vl.Field(i)
		if !field.IsZero() {
			switch field.Kind() {
			case reflect.Ptr:
				switch field.Elem().Kind() {
				case reflect.String:
					line = append(line, []byte(fmt.Sprintf("%s = \"%s\"\n", utils.ToSnakeCase(tp.Field(i).Name), field.Elem()))...)
					out = append(out, line...)
				case reflect.Bool, reflect.Int, reflect.Int32, reflect.Int64:
					line = append(line, []byte(fmt.Sprintf("%s = %v\n", utils.ToSnakeCase(tp.Field(i).Name), field.Elem()))...)
					out = append(out, line...)
				default:
					fmt.Println("Unable to Determine Type")
				}
			case reflect.Array, reflect.Slice:
				for j := 0; j < field.Len(); j++ {
					out = append(out, []byte(strings.ToLower(fmt.Sprintf("\n\t%s {\n", tp.Field(i).Name)))...)
					out = append(out, resourceBaseToHCL(field.Index(j).Interface(), indentLevel+1)...)
					out = append(out, []byte("\t}\n")...)
				}
			}
		}
	}
	return out
}
