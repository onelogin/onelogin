package profiles

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	// "strings"
	// "os"
	"testing"
)

type MockFile struct {
	Content []byte
}

func (m *MockFile) Write(p []byte) (int, error) {
	m.Content = p
	return len(p), io.EOF
}

func (m *MockFile) Read(p []byte) (int, error) {
	for i, b := range m.Content {
		p[i] = b
	}
	return len(p), io.EOF
}

type MockCmdLineInput struct {
	Content []byte
}

func (m *MockCmdLineInput) Read(p []byte) (int, error) {
	for i, b := range m.Content {
		p[i] = b
	}
	return len(p), io.EOF
}

type MockRepository struct {
	StorageMedia *MockFile
}

func (p MockRepository) readAll() ([]byte, error) {
	return ioutil.ReadAll(p.StorageMedia)
}

func (p MockRepository) persist(profiles map[string]*Profile) {
	data := map[string]Profile{}
	for n, prf := range profiles {
		data[n] = *prf
	}
	updatedProfiles, _ := json.Marshal(data)
	p.StorageMedia.Write(updatedProfiles)
}

func TestIndex(t *testing.T) {
	tests := map[string]struct {
		MockStorage         *MockFile
		ExpectedReturnCount int
	}{
		"It lists profiles": {
			MockStorage:         &MockFile{Content: []byte(`{"t":{"name":"t","active":true,"region":"us","client_id":"ti","client_secret":"ts"}}`)},
			ExpectedReturnCount: 1,
		},
		"It lists nothing if no profiles": {
			MockStorage:         &MockFile{Content: []byte{}},
			ExpectedReturnCount: 0,
		},
		"It lists nothing if empty json": {
			MockStorage:         &MockFile{Content: []byte(`{}`)},
			ExpectedReturnCount: 0,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			profilesSvc := ProfileService{
				Repository: MockRepository{StorageMedia: test.MockStorage},
			}
			profiles := profilesSvc.Index()
			assert.Equal(t, test.ExpectedReturnCount, len(profiles))
		})
	}
}

func TestFind(t *testing.T) {
	tests := map[string]struct {
		MockStorage    *MockFile
		MockInput      string
		ExpectedReturn string
	}{
		"It gets profile with name": {
			MockStorage:    &MockFile{Content: []byte(`{"t":{"name":"t","active":true,"region":"us","client_id":"ti","client_secret":"ts"}}`)},
			MockInput:      "t",
			ExpectedReturn: "t",
		},
		"It lists nothing if no profiles exist with name": {
			MockStorage:    &MockFile{Content: []byte(`{"t":{"name":"t","active":true,"region":"us","client_id":"ti","client_secret":"ts"}}`)},
			MockInput:      "s",
			ExpectedReturn: "",
		},
		"It lists nothing if empty json": {
			MockStorage:    &MockFile{Content: []byte(`{}`)},
			MockInput:      "t",
			ExpectedReturn: "",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			profilesSvc := ProfileService{
				Repository: MockRepository{StorageMedia: test.MockStorage},
			}
			profile := profilesSvc.Find(test.MockInput)
			if test.ExpectedReturn == "" {
				assert.Nil(t, profile)
			} else {
				assert.Equal(t, test.ExpectedReturn, profile.Name)
			}
		})
	}
}

func TestActivate(t *testing.T) {
	tests := map[string]struct {
		MockStorage  *MockFile
		MockInput    string
		ExpectedEndT bool
		ExpectedEndS bool
	}{
		"It activates profile with given name": {
			MockStorage:  &MockFile{Content: []byte(`{"t":{"name":"t","active":true,"region":"us","client_id":"ti","client_secret":"ts"}, "s":{"name":"s","active":false,"region":"us","client_id":"si","client_secret":"ss"}}`)},
			MockInput:    "s",
			ExpectedEndT: false,
			ExpectedEndS: true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			profilesSvc := ProfileService{
				Repository: MockRepository{StorageMedia: test.MockStorage},
			}
			profilesSvc.Activate(test.MockInput)
			tProfile := profilesSvc.Find("t")
			sProfile := profilesSvc.Find("s")
			assert.Equal(t, test.ExpectedEndT, (*tProfile).Active)
			assert.Equal(t, test.ExpectedEndS, (*sProfile).Active)
		})
	}
}

func TestCreate(t *testing.T) {
	tests := map[string]struct {
		CmdLineInput         *MockCmdLineInput
		ProfileName          string
		MockStorage          *MockFile
		ExpectedProfile      Profile
		ExpectedProfileCount int
	}{
		"It creates and inserts a new profile": {
			CmdLineInput:         &MockCmdLineInput{Content: []byte("us\ntest\ntest\n")},
			ProfileName:          "test",
			MockStorage:          &MockFile{Content: []byte(`{"pre-existing":{"name":"pre-existing","active":false,"region":"us","client_id":"test","client_secret":"test"}}`)},
			ExpectedProfile:      Profile{Name: "test", Region: "us", ClientID: "test", ClientSecret: "test"},
			ExpectedProfileCount: 2,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			profilesSvc := ProfileService{
				Repository:  MockRepository{StorageMedia: test.MockStorage},
				InputReader: test.CmdLineInput,
			}
			profilesSvc.Create(test.ProfileName)
			profileCount := len(profilesSvc.Index())
			newProfile := profilesSvc.Find(test.ProfileName)
			assert.Equal(t, test.ExpectedProfileCount, profileCount)
			assert.Equal(t, test.ExpectedProfile, *newProfile)
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := map[string]struct {
		CmdLineInput         *MockCmdLineInput
		ProfileName          string
		MockStorage          *MockFile
		ExpectedProfile      Profile
		ExpectedProfileCount int
	}{
		"It updates a profile": {
			CmdLineInput:         &MockCmdLineInput{Content: []byte("\nupdate\nupdate\n")},
			ProfileName:          "test",
			MockStorage:          &MockFile{Content: []byte(`{"test":{"name":"test","active":false,"region":"us","client_id":"test","client_secret":"test"}}`)},
			ExpectedProfile:      Profile{Name: "test", Region: "us", ClientID: "update", ClientSecret: "update"},
			ExpectedProfileCount: 1,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			profilesSvc := ProfileService{
				Repository:  MockRepository{StorageMedia: test.MockStorage},
				InputReader: test.CmdLineInput,
			}
			profilesSvc.Update(test.ProfileName)
			profileCount := len(profilesSvc.Index())
			newProfile := profilesSvc.Find(test.ProfileName)
			assert.Equal(t, test.ExpectedProfileCount, profileCount)
			assert.Equal(t, test.ExpectedProfile, *newProfile)
		})
	}
}

func TestRemove(t *testing.T) {
	tests := map[string]struct {
		ProfileName          string
		MockStorage          *MockFile
		ExpectedProfile      Profile
		ExpectedProfileCount int
	}{
		"It removes a profile": {
			ProfileName:          "test",
			MockStorage:          &MockFile{Content: []byte(`{"test":{"name":"test","active":false,"region":"us","client_id":"test","client_secret":"test"}}`)},
			ExpectedProfileCount: 0,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			profilesSvc := ProfileService{
				Repository: MockRepository{StorageMedia: test.MockStorage},
			}
			profilesSvc.Remove(test.ProfileName)
			profileCount := len(profilesSvc.Index())
			assert.Equal(t, test.ExpectedProfileCount, profileCount)
		})
	}
}
