package profiles

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type Repository interface {
	persist(profiles map[string]*Profile)
	readAll() ([]byte, error)
}

type FileRepository struct {
	StorageMedia *os.File
}

func (p FileRepository) readAll() ([]byte, error) {
	return ioutil.ReadAll(p.StorageMedia)
}

func (p FileRepository) persist(profiles map[string]*Profile) {
	data := map[string]Profile{}
	for n, prf := range profiles {
		data[n] = *prf
	}
	updatedProfiles, _ := json.Marshal(data)
	p.StorageMedia.Truncate(0)
	if _, err := p.StorageMedia.WriteAt(updatedProfiles, 0); err != nil {
		if err = p.StorageMedia.Close(); err != nil {
			log.Fatalln("Unable write profile", err)
		}
		log.Fatalln("Unable to persist", err)
	}
	if err := p.StorageMedia.Close(); err != nil {
		log.Fatalln("Unable write profile", err)
	}
}
