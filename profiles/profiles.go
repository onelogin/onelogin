package profiles

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
)

type ProfileRepository struct {
	Repository  Repository
	InputReader io.Reader
}

type Profile struct {
	Name         string `json:"name"`
	Active       bool   `json:"active"`
	Region       string `json:"region"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func New(r Repository, rdr io.Reader) ProfileRepository {
	return ProfileRepository{Repository: r, InputReader: rdr}
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
	p.Repository.persist(profiles)
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
	fileData, err := p.Repository.readAll()
	if err != nil {
		log.Fatalln("Unable to read profiles", err)
	}
	if fileData[0] == 0 { // no data in file
		return existingProfiles
	}
	err = json.Unmarshal(bytes.Trim(fileData, "\x00"), &existingProfiles)
	if err != nil {
		log.Fatalln("Unable to parse profiles file!", err)
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
	collectProfileInput(profile, p.InputReader)
	existingProfiles[(*profile).Name] = profile
	p.Repository.persist(existingProfiles)
}

func (p ProfileRepository) Update(name string) {
	existingProfiles := p.Index()
	profile := existingProfiles[name]
	if profile == nil {
		log.Fatalln("Profile does not exist!")
	}
	collectProfileInput(profile, p.InputReader)
	existingProfiles[(*profile).Name] = profile
	p.Repository.persist(existingProfiles)
}

func (p ProfileRepository) Remove(name string) {
	existingProfiles := p.Index()
	delete(existingProfiles, name)
	p.Repository.persist(existingProfiles)
}

func collectProfileInput(p *Profile, rdr io.Reader) {
	var userInput string
	reader := bufio.NewReader(rdr)
	for {
		fmt.Printf("Add the profile's REGION (us or eu) [Enter to accept %s]: \n", p.Region)
		userInput, _ = reader.ReadString('\n')
		userInput = strings.ToLower(strings.TrimSuffix(userInput, "\n"))
		if userInput == "us" || userInput == "eu" || (len(userInput) == 0 && p.Region != "") {
			if len(userInput) == 0 && p.Region != "" {
				userInput = p.Region
			}
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
