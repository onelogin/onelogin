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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/onelogin/onelogin-go-sdk/pkg/client"

	"log"
)

// Clients is a list of memoized instantiated clients
type Clients struct {
	OneLogin *client.APIClient
	AwsIam   *iam.IAM
	ClientConfigs
}

type ClientConfigs struct {
	AwsRegion                                           string
	OneLoginClientID, OneLoginClientSecret, OneLoginURL string
}

func New(clientConfigs ClientConfigs) *Clients {
	return &Clients{ClientConfigs: clientConfigs}
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
			log.Println("[Warning] Unable to connect to OneLogin!", err)
		}
		c.OneLogin = oneloginClient
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
		}
		c.AwsIam = iam.New(sess)
	}
	return c.AwsIam
}
