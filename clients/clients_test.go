package clients

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOneLoginClient(t *testing.T) {
	tests := map[string]struct {
		Configs ClientConfigs
	}{
		"It initializes and memoizes the client": {
			Configs: ClientConfigs{
				AwsRegion:            "us-test-2",
				OneLoginClientID:     "test",
				OneLoginClientSecret: "test",
				OneLoginURL:          "test.com",
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			clients := New(test.Configs)
			clients.OneLoginClient()                // instantiate and store address of aws client
			clnt := clients.OneLoginClient()        // retrieves that address
			assert.Equal(t, clnt, clients.OneLogin) // retrieved address should be the memoized address

		})
	}
}

func TestAwsIamClient(t *testing.T) {
	tests := map[string]struct {
		Configs ClientConfigs
	}{
		"It initializes and memoizes the client": {
			Configs: ClientConfigs{
				AwsRegion:            "us-test-2",
				OneLoginClientID:     "test",
				OneLoginClientSecret: "test",
				OneLoginURL:          "test.com",
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			clients := New(test.Configs)
			clients.AwsIamClient()                // instantiate and store address of aws client
			clnt := clients.AwsIamClient()        // retrieves that address
			assert.Equal(t, clnt, clients.AwsIam) // retrieved address should be the memoized address
		})
	}
}

func TestOktaClient(t *testing.T) {
	tests := map[string]struct {
		Configs ClientConfigs
	}{
		"It initializes and memoizes the client": {
			Configs: ClientConfigs{
				OktaOrgName:  "org",
				OktaBaseURL:  "org.org",
				OktaAPIToken: "t0k3n",
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			clients := New(test.Configs)
			clients.OktaClient()                // instantiate and store address of okta client
			clnt := clients.OktaClient()        // retrieves that address
			assert.Equal(t, clnt, clients.Okta) // retrieved address should be the memoized address
		})
	}
}
