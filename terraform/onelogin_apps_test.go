package terraform

// ProviderClient.Services.AppsV2.GetApps

// oneloginClient, _ := client.NewClient(&client.APIClientConfig{
//   Timeout:      5,
//   ClientID:     os.Getenv("ONELOGIN_CLIENT_ID"),
//   ClientSecret: os.Getenv("ONELOGIN_CLIENT_SECRET"),
//   Url:          os.Getenv("ONELOGIN_OAPI_URL"),
// })

// type MockService struct{}
//
// func TestImportResourceDefinitionsFromRemote(t *testing.T) {
// 	tests := map[string]struct {
// 		InputImportable             Importable
// 		ExpectedResourceDefinitions []ResourceDefinition
// 	}{
// 		"it reaches out to the onelogin api for some resources": {
// 			InputImportable: OneloginAppsImportable{
// 				ProviderClient: &client.APIClient{
// 					Services: &client.Services{
// 						AppsV2: &services.AppsV2{},
// 					},
// 				},
// 				AppType: "onelogin_apps",
// 			},
// 		},
// 	}
// 	for name, test := range tests {
// 		t.Run(name, func(t *testing.T) {
//
// 			actual := test.InputImportable.ImportResourceDefinitionsFromRemote()
// 			assert.Equal(t, test.ExpectedResourceDefinitions, actual)
// 		})
// 	}
// }
