package clients

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOneLoginClient(t *testing.T) {
	tests := map[string]struct {
		clients Clients
	}{
		"It initializes and memoizes the client": {
			clients: Clients{
				ClientConfigs: ClientConfigs{
					OneLoginClientID:     "test",
					OneLoginClientSecret: "test",
					OneLoginURL:          "test.com",
					AwsRegion:            "us",
				},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			test.clients.AwsIamClient()                  // instantiate and store address of aws client
			clnt := test.clients.OneLoginClient()        // retrieves that address
			assert.Equal(t, clnt, test.clients.OneLogin) // retrieved address should be the memoized address

		})
	}
}

func TestAwsIamClient(t *testing.T) {
	tests := map[string]struct {
		clients Clients
	}{
		"It initializes and memoizes the client": {
			clients: Clients{
				ClientConfigs: ClientConfigs{
					OneLoginClientID:     "test",
					OneLoginClientSecret: "test",
					OneLoginURL:          "test.com",
					AwsRegion:            "us",
				},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			test.clients.AwsIamClient()                // instantiate and store address of aws client
			clnt := test.clients.AwsIamClient()        // retrieves that address
			assert.Equal(t, clnt, test.clients.AwsIam) // retrieved address should be the memoized address
		})
	}
}

func TestOktaClient(t *testing.T) {
	tests := map[string]struct {
		clients Clients
	}{
		"It initializes and memoizes the client": {
			clients: Clients{
				ClientConfigs: ClientConfigs{
					OktaOrgName:  "test",
					OktaBaseURL:  "test.com",
					OktaAPIToken: "test",
				},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			test.clients.OktaClient()                // instantiate and store address of okta client
			clnt := test.clients.OktaClient()        // retrieves that address
			assert.Equal(t, clnt, test.clients.Okta) // retrieved address should be the memoized address
		})
	}
}
