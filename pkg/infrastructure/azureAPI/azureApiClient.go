package azureAPI

import (
	"context"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/authorization/mgmt/authorization"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v3"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	log "github.com/sirupsen/logrus"
)

type AzureAPIClients struct {
	RoleAssignmentsClient         authorization.RoleAssignmentsClient
	RoleAssignmentsDeletionClient *armauthorization.RoleAssignmentsClient
	// RoleDefinitionsClient authorization.RoleDefinitionsClient
	DeploymentsClient    *armresources.DeploymentsClient
	ResourceGroupsClient *armresources.ResourceGroupsClient

	// Default CLI Creds
	DefaultCred                         *azidentity.DefaultAzureCredential
	defaultAPIBearerToken               string
	defaultAPIBearerTokenLastCachedTime time.Time
	// SPCred                *azidentity.ClientSecretCredential
}

const defaultTokenCacheDuration = 10 * time.Minute

func NewAzureAPIClients(subscriptionID string) *AzureAPIClients {
	a := &AzureAPIClients{}
	a.SetApiClients(subscriptionID)
	return a
}

func getAuthorizer() (authorizer autorest.Authorizer, err error) {
	// Use the default Azure environment for authentication
	authorizer, err = auth.NewAuthorizerFromCLI()
	if err != nil {
		return nil, err
	}
	return authorizer, nil
}

type TokenProvider interface {
	GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error)
}

func (a *AzureAPIClients) getBearerToken(tp TokenProvider) (bearerToken string, err error) {
	opts := policy.TokenRequestOptions{Scopes: []string{"https://management.azure.com/.default"}}
	tok, err := tp.GetToken(context.Background(), opts)
	if err != nil {
		return "", err
	}

	return tok.Token, nil
}

func (a *AzureAPIClients) SetApiClients(subscriptionId string) error {
	authorizer, err := getAuthorizer()
	if err != nil {
		return err
	}

	a.DefaultCred, err = azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		// log.Fatal(err)
		log.Fatal(err)
	}

	// Set RoleAssignmentsClient
	a.RoleAssignmentsClient = authorization.NewRoleAssignmentsClient(subscriptionId)
	a.RoleAssignmentsClient.Authorizer = authorizer
	if err != nil {
		log.Fatal(err)
	}

	roleAssignmentsDeletionClientFactory, err := armauthorization.NewClientFactory(subscriptionId, a.DefaultCred, nil)
	if err != nil {
		log.Fatalf("failed to create role assignments deletion client factory: %v", err)
	}

	a.RoleAssignmentsDeletionClient = roleAssignmentsDeletionClientFactory.NewRoleAssignmentsClient()

	// Set RoleDefinitionsClient
	// a.RoleDefinitionsClient = authorization.NewRoleDefinitionsClient(subscriptionId)
	// a.RoleDefinitionsClient.Authorizer = authorizer

	resourcesClientFactory, err := armresources.NewClientFactory(subscriptionId, a.DefaultCred, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Set DeploymentsClient
	a.DeploymentsClient = resourcesClientFactory.NewDeploymentsClient()

	// Set ResourceGroupsClient
	a.ResourceGroupsClient, err = armresources.NewResourceGroupsClient(subscriptionId, a.DefaultCred, nil)
	if err != nil {
		log.Fatal(err)
	}

	return nil

}

func (a *AzureAPIClients) GetSPBearerToken(tenantID, spClientID, spClientSecret string) (string, error) {
	// Get the Service Principal creds
	spCred, err := azidentity.NewClientSecretCredential(tenantID, spClientID, spClientSecret, nil)
	if err != nil {
		log.Error(err)
		return "", err
	}

	bearerToken, err := a.getBearerToken(spCred)
	if err != nil {
		log.Error(err)
		return "", err
	}

	return bearerToken, nil

}

func (a *AzureAPIClients) GetDefaultAPIBearerToken() (bearerToken string, err error) {

	if a.defaultAPIBearerToken == "" || time.Since(a.defaultAPIBearerTokenLastCachedTime) > defaultTokenCacheDuration {
		bearerToken, err = a.getBearerToken(a.DefaultCred)
		if err != nil {
			return "", err
		}

		a.defaultAPIBearerToken = bearerToken
		a.defaultAPIBearerTokenLastCachedTime = time.Now()
		log.Infoln("Default API Bearer Token set")
	}

	return a.defaultAPIBearerToken, nil
}

// func (m *MinPermFinder) RefreshSPAPIAccessBearerToken() error {
// 	// Get the bearer token for the API access

// 	bearerToken, err := m.getBearerToken(m.SPCred)
// 	if err != nil {
// 		return err
// 	}

// 	m.SPCredBearerToken = bearerToken
// 	return nil
// }

// func (m *MinPermFinder) CreateCustomRoleWithInitialScopeAndPermissions() error {

// 	initialScope := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", m.SubscriptionID, m.ResourceGroupName)
// 	// initialScope := fmt.Sprintf("/subscriptions/%s", m.SubscriptionID)

// 	roleDefinition := authorization.RoleDefinition{
// 		RoleDefinitionProperties: &authorization.RoleDefinitionProperties{
// 			RoleName:    &m.RoleDefinitionName,
// 			Description: &m.RoleDefinitionName,
// 			AssignableScopes: &[]string{
// 				initialScope,
// 			},
// 			Permissions: &[]authorization.Permission{
// 				{
// 					Actions: to.StringSlicePtr([]string{
// 						// "Microsoft.Resources/deployments/read",
// 						// "Microsoft.Resources/deployments/write",
// 					}),
// 				},
// 			},
// 		},
// 	}

// 	// Create the custom role
// 	_, err := m.RoleDefinitionsClient.CreateOrUpdate(m.Ctx, initialScope, m.RoleDefinitionID, roleDefinition)
// 	return err
// }
