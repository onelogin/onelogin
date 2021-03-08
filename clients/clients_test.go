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
			clnts := New(test.Configs)
			clnts.OneLoginClient()                // instantiate and store address of aws client
			clnt := clnts.OneLoginClient()        // retrieves that address
			assert.Equal(t, clnt, clnts.OneLogin) // retrieved address should be the memoized address

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
			clnts := New(test.Configs)
			clnts.AwsIamClient()                // instantiate and store address of aws client
			clnt := clnts.AwsIamClient()        // retrieves that address
			assert.Equal(t, clnt, clnts.AwsIam) // retrieved address should be the memoized address
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
			clnts := New(test.Configs)
			clnts.OktaClient()                // instantiate and store address of okta client
			clnt := clnts.OktaClient()        // retrieves that address
			assert.Equal(t, clnt, clnts.Okta) // retrieved address should be the memoized address
		})
	}
}
