package terraform

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/onelogin/onelogin-go-sdk/pkg/models"
)

// State is the in memory representation of tfstate.
type State struct {
	Resources []StateResource `json:"resources"`
}

type StateResource struct {
	Content   []byte
	Name      string             `json:"name"`
	Type      string             `json:"type"`
	Provider  string             `json:"provider"`
	Instances []ResourceInstance `json:"instances"`
}

type ResourceInstance struct {
	Data ResourceData `json:"attributes"`
}

type ResourceData struct {
	AllowAssumedSignin *bool                     `json:"allow_assumed_signin,omitempty"`
	ConnectorID        *int                      `json:"connector_id,omitempty"`
	Description        *string                   `json:"description,omitempty"`
	Name               *string                   `json:"name,omitempty"`
	Notes              *string                   `json:"notes,omitempty"`
	Visible            *bool                     `json:"visible,omitempty"`
	Provisioning       []models.AppProvisioning  `json:"provisioning,omitempty"`
	Parameters         []models.AppParameters    `json:"parameters,omitempty"`
	Configuration      []models.AppConfiguration `json:"configuration,omitempty"`
}

// Initialize inflates a State struct using the tfstate file
func (state *State) Initialize() {
	log.Println("Collecting State from tfstate File")

	path, _ := os.Getwd()
	p := filepath.Join(path, "/terraform.tfstate")
	data, err := ioutil.ReadFile(p)
	if err != nil {
		log.Fatal("Unable to Read tfstate")
	}

	if err := json.Unmarshal(data, state); err != nil {
		log.Fatal("Unable to Translate tfstate in Memory")
	}

}
