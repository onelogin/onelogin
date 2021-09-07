// Package clients clients.go
// This module creates a list of client instances for any cloud provider so each client only gets
// instantiated once and can be shared among other callers.
//
// Adding Clients
// To add new clients, add to the Clients struct a field that represents how the client should be handled
// Then add a method to Clients that either creates + memoizes, or returns the memoized client per the
// client initialization procedure. These should be pulbic facing methods as they should be used to retrieve clients
package clients

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/onelogin/onelogin-go-sdk/pkg/client"
	"github.com/onelogin/onelogin/profiles"
)

// Clients is a list of memoized instantiated clients
type Clients struct {
	OneLogin *client.APIClient
	AwsIam   *iam.IAM
	Okta     *okta.Client
	ClientConfigs
}

type ClientConfigs struct {
	AwsRegion                                           string
	OneLoginClientID, OneLoginClientSecret, OneLoginURL string
	OktaOrgName, OktaBaseURL, OktaAPIToken              string
}

func New(credsFile *os.File) *Clients {
	profileService := profiles.ProfileService{
		Repository: profiles.FileRepository{
			StorageMedia: credsFile,
		},
	}
	profile := profileService.GetActive()
	clientConfigs := ClientConfigs{
		AwsRegion:            os.Getenv("AWS_REGION"),
		OktaOrgName:          os.Getenv("OKTA_ORG_NAME"),
		OktaBaseURL:          os.Getenv("OKTA_BASE_URL"),
		OktaAPIToken:         os.Getenv("OKTA_API_TOKEN"),
		OneLoginClientID:     os.Getenv("ONELOGIN_CLIENT_ID"),
		OneLoginClientSecret: os.Getenv("ONELOGIN_CLIENT_SECRET"),
		OneLoginURL:          os.Getenv("ONELOGIN_OAPI_URL"),
	}
	if profile == nil {
		fmt.Println("No active profile detected. Authenticating with environment variables")
	} else {
		fmt.Println("Using profile", (*profile).Name)
		clientConfigs.OneLoginClientID = (*profile).ClientID
		clientConfigs.OneLoginClientSecret = (*profile).ClientSecret
		clientConfigs.OneLoginURL = fmt.Sprintf("https://api.%s.onelogin.com", (*profile).Region)
	}
	return &Clients{ClientConfigs: clientConfigs}
}

func (c *Clients) OktaClient() *okta.Client {
	if c.Okta == nil {
		oktaURL := fmt.Sprintf("https://%s.%s", c.ClientConfigs.OktaOrgName, c.ClientConfigs.OktaBaseURL)
		_, oktaClient, err := okta.NewClient(
			context.TODO(),
			okta.WithOrgUrl(oktaURL),
			okta.WithToken(c.ClientConfigs.OktaAPIToken),
		)
		if err != nil {
			log.Fatalln("There was a problem configuring the Okta client. Ensure your Okta credentials are exported to your environment", err)
		} else {
			c.Okta = oktaClient
		}
	}
	return c.Okta
}

// OneLoginClient creates and returns an instance of the OneLogin API client if one does not exist
// Memoizes the OneLogin API client and returns that instance on every subsequent call
func (c *Clients) OneLoginClient() *client.APIClient {
	if c.OneLogin == nil {
		clientConfig := &client.APIClientConfig{
			Timeout:      5,
			ClientID:     c.ClientConfigs.OneLoginClientID,
			ClientSecret: c.ClientConfigs.OneLoginClientSecret,
			Url:          c.ClientConfigs.OneLoginURL,
		}
		oneloginClient, err := client.NewClient(clientConfig)
		if err != nil {
			log.Fatalln("There was a problem configuring the OneLogin client. Ensure your OneLogin credentials are exported to your environment", err)
		} else {
			c.OneLogin = oneloginClient
		}
	}
	return c.OneLogin
}

// AwsIamClient creates and returns an instance of the AWS API client if one does not exist
// Memoizes the AWS API client and returns that instance on every subsequent call
func (c *Clients) AwsIamClient() *iam.IAM {
	if c.AwsIam == nil {
		sess, err := session.NewSession(
			&aws.Config{
				Region: aws.String(c.ClientConfigs.AwsRegion),
			},
		)
		if err != nil {
			log.Fatalln("There was a problem configuring the AWS client. Ensure your AWS credentials are exported to your environment", err)
		} else {
			c.AwsIam = iam.New(sess)
		}
	}
	return c.AwsIam
}
