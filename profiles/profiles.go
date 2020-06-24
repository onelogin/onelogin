package profiles

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type ProfileRepository struct {
	Source io.ReadWriter
}

type Profile struct {
	Name         string `json:"name"`
	Active       bool   `json:"active"`
	Region       string `json:"region"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func New() ProfileRepository {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln("Unable to find user home directory from $HOME or USERPROFILE environment variables")
	}
	p := filepath.Join(homeDir, ".onelogin")
	os.Mkdir(p, 0750)
	p = filepath.Join(p, "profiles.json")
	f, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		log.Fatalln("Unable to create profiles file")
	}
	return ProfileRepository{Source: f}
}

func (p ProfileRepository) Activate(name string) {
	profiles := p.Index()
	for n, prof := range profiles {
		if n == name {
			(*prof).Active = true
		} else {
			(*prof).Active = false
		}
	}
	p.persist(profiles)
}

func (p ProfileRepository) Find(name string) *Profile {
	profiles := p.Index()
	if profiles[name] != nil {
		return profiles[name]
	}
	return nil
}

func (p ProfileRepository) Index() map[string]*Profile {
	existingProfiles := map[string]*Profile{}
	fileData, err := ioutil.ReadAll(p.Source)
	if err != nil {
		log.Fatalln("Unable to read profiles", err)
	}
	if len(fileData) == 0 {
		return existingProfiles
	}
	err = json.Unmarshal(fileData, &existingProfiles)
	if err != nil {
		log.Fatalln("Unable to parse profiles file!")
	}
	return existingProfiles
}

func (p ProfileRepository) Create(name string) {
	existingProfiles := p.Index()
	profile := existingProfiles[name]
	if profile != nil {
		log.Fatalln("Profile with this name already exists!")
	}
	profile = &Profile{Name: name}
	collectProfileInput(profile)
	existingProfiles[(*profile).Name] = profile
	p.persist(existingProfiles)
}

func (p ProfileRepository) Update(name string) {
	existingProfiles := p.Index()
	profile := existingProfiles[name]
	if profile == nil {
		log.Fatalln("Profile does not exist!")
	}
	collectProfileInput(profile)
	existingProfiles[(*profile).Name] = profile
	p.persist(existingProfiles)
}

func (p ProfileRepository) Remove(name string) {
	existingProfiles := p.Index()
	delete(existingProfiles, name)
	p.persist(existingProfiles)
}

func collectProfileInput(p *Profile) {
	reader := bufio.NewReader(os.Stdin)
	var userInput string

	for {
		fmt.Printf("Add the profile's REGION (us or eu) [Enter to accept %s]: \n", p.Region)
		userInput, _ = reader.ReadString('\n')
		userInput = strings.ToLower(strings.TrimSuffix(userInput, "\n"))
		if (userInput == "us" || userInput == "eu") || (len(userInput) == 0 && p.Region != "") {
			break
		}
		fmt.Println("Invalid region given!")
	}
	p.Region = userInput

	fmt.Printf("Add the profile's CLIENT_ID [Enter to accept %s]: \n", p.ClientID)
	for {
		userInput, _ = reader.ReadString('\n')
		if userInput == "\n" && p.ClientID != "" {
			break
		}
		if userInput != "\n" {
			p.ClientID = strings.TrimSuffix(userInput, "\n")
			break
		}
		fmt.Println("Value cannot be blank!")
	}

	fmt.Printf("Add the profile's CLIENT_SECRET [Enter to accept %s]: \n", p.ClientSecret)
	for {
		userInput, _ = reader.ReadString('\n')
		if userInput == "\n" && p.ClientSecret != "" {
			break
		}
		if userInput != "\n" {
			p.ClientSecret = strings.TrimSuffix(userInput, "\n")
			break
		}
		fmt.Println("Value cannot be blank!")
	}
}

func (p ProfileRepository) persist(profiles map[string]*Profile) {
	data := map[string]Profile{}
	for n, prf := range profiles {
		data[n] = *prf
	}
	updatedProfiles, _ := json.Marshal(data)
	p.Source.(*os.File).Truncate(0)
	if _, err := p.Source.(*os.File).WriteAt(updatedProfiles, 0); err != nil {
		if err = p.Source.(*os.File).Close(); err != nil {
			log.Fatalln("Unable write profile", err)
		}
		log.Fatalln("Unable to persist", err)
	}
	if err := p.Source.(*os.File).Close(); err != nil {
		log.Fatalln("Unable write profile", err)
	}
}
