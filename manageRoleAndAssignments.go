package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/authorization/mgmt/authorization"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

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

func (m *MinPermFinder) getBearerToken(tp TokenProvider) (bearerToken string, err error) {
	opts := policy.TokenRequestOptions{Scopes: []string{"https://management.azure.com/.default"}}
	tok, err := tp.GetToken(context.Background(), opts)
	if err != nil {
		return "", err
	}

	return tok.Token, nil
}

func (m *MinPermFinder) SetApiClients() error {
	authorizer, err := getAuthorizer()
	if err != nil {
		return err
	}

	m.DefaultCred, err = azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		// log.Fatal(err)
		log.Fatal(err)
	}

	// Set RoleAssignmentsClient
	m.RoleAssignmentsClient = authorization.NewRoleAssignmentsClient(m.SubscriptionID)
	m.RoleAssignmentsClient.Authorizer = authorizer
	if err != nil {
		log.Fatal(err)
	}

	// Set RoleDefinitionsClient
	m.RoleDefinitionsClient = authorization.NewRoleDefinitionsClient(m.SubscriptionID)
	m.RoleDefinitionsClient.Authorizer = authorizer

	resourcesClientFactory, err := armresources.NewClientFactory(m.SubscriptionID, m.DefaultCred, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Set DeploymentsClient
	m.DeploymentsClient = resourcesClientFactory.NewDeploymentsClient()

	// Set ResourceGroupsClient
	m.ResourceGroupsClient, err = armresources.NewResourceGroupsClient(m.SubscriptionID, m.DefaultCred, nil)
	if err != nil {
		log.Fatal(err)
	}

	return nil

}

func (m *MinPermFinder) SetDefaultAPIAccessBearerToken() error {
	// Get the bearer token for the API access
	if m.DefaultAPIBearerToken != "" {
		log.Infoln("Default API Bearer Token already set")
		return nil
	}

	// Get the bearer token
	log.Infoln("Getting new Default API Bearer Token")
	bearerToken, err := m.getBearerToken(m.DefaultCred)
	if err != nil {
		return err
	}

	m.DefaultAPIBearerToken = bearerToken
	return nil
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

func (m *MinPermFinder) CreateUpdateCustomRole(permissions []string) error {

	scope := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", m.SubscriptionID, m.ResourceGroupName)

	data := map[string]interface{}{
		"assignableScopes": []string{
			scope,
		},
		"description": m.RoleDefinitionName,
		"id":          m.RoleDefinitionResourceID,
		"name":        m.RoleDefinitionID,
		"permissions": []map[string]interface{}{
			{
				"actions":        permissions,
				"dataActions":    []string{},
				"notActions":     []string{},
				"notDataActions": []string{},
			},
		},
		"roleName": m.RoleDefinitionName,
		"roleType": "CustomRole",
		// "type":     "Microsoft.Authorization/roleDefinitions",
	}

	properties := map[string]interface{}{
		"properties": data,
	}
	// marshal data as json
	jsonData, err := json.Marshal(properties)
	if err != nil {
		return err
	}

	//convert to json string
	jsonString := string(jsonData)

	// log.Printf("jsonString: %s", jsonString)
	log.Debugf("jsonString: %s", jsonString)

	url := fmt.Sprintf("https://management.azure.com/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Authorization/roleDefinitions/%s?api-version=2018-01-01-preview", m.SubscriptionID, m.ResourceGroupName, m.RoleDefinitionID)

	client := &http.Client{}

	req, err := http.NewRequest("PUT", url, bytes.NewBufferString(jsonString))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Go HTTP Client")

	// add bearer token to header
	req.Header.Add("Authorization", "Bearer "+m.DefaultAPIBearerToken)

	// make request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// print response body
	// fmt.Println(string(body))
	log.Debugln(string(body))

	return nil
}

func (m *MinPermFinder) AssignRoleToSP() error {

	scope := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", m.SubscriptionID, m.ResourceGroupName)
	url := fmt.Sprintf("https://management.azure.com/%s/providers/Microsoft.Authorization/roleAssignments/%s?api-version=2022-04-01", scope, uuid.New().String())

	data := map[string]interface{}{
		"principalId":      m.SPObjectID,
		"principalType":    "ServicePrincipal",
		"roleDefinitionId": m.RoleDefinitionResourceID,
	}

	properties := map[string]interface{}{
		"properties": data,
	}

	// marshal data as json
	jsonData, err := json.Marshal(properties)
	if err != nil {
		return err
	}

	//convert to json string
	jsonString := string(jsonData)

	log.Debugf("jsonString: %s", jsonString)

	client := &http.Client{}

	req, err := http.NewRequest("PUT", url, bytes.NewBufferString(jsonString))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Go HTTP Client")

	// add bearer token to header
	req.Header.Add("Authorization", "Bearer "+m.DefaultAPIBearerToken)

	// make request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("Failed to assign role to SP. Status code: %s", string(body))
	}

	// print response body
	log.Debugln(string(body))
	return nil
}

// func (m *MinPermFinder) AssignRoleToSP(scope string) error {

// 	// armauthorization.NewClassicAdministratorsClient()
// 	clientFactory, err := armauthorization.NewClientFactory(m.SubscriptionID, m.DefaultCred, nil)
// 	if err != nil {
// 		log.Fatalf("failed to create client: %v", err)
// 	}

// 	res, err := clientFactory.NewRoleAssignmentsClient().Create(m.Ctx, m.SubscriptionID, uuid.New().String(), armauthorization.RoleAssignmentCreateParameters{
// 		Properties: &armauthorization.RoleAssignmentProperties{
// 			PrincipalID:      &m.SPObjectID,
// 			PrincipalType:    to.StringPtr(armauthorization.PrincipalTypeServicePrincipal),
// 			RoleDefinitionID: &m.RoleDefinitionResourceID,
// 		},
// 	}, nil)

// 	// rac, err := armauthorization.NewRoleAssignmentsClient(m.SubscriptionID, m.DefaultCred, nil)
// 	// if err != nil {
// 	// 	return err
// 	// }

// 	// rao := armauthorization.RoleAssignmentsClientCreateOptions{

// 	// }
// 	// roleAssignmentParams := authorization.RoleAssignmentCreateParameters{
// 	// 	Properties: &authorization.RoleAssignmentProperties{
// 	// 		PrincipalID:      &m.SPObjectID,
// 	// 		RoleDefinitionID: &m.RoleDefinitionResourceID,
// 	// 	},
// 	// }

// 	// roleAssignmentParams := armauthorization.RoleAssignmentCreateParameters{
// 	// 	Properties: &armauthorization.RoleAssignmentProperties{
// 	// 		PrincipalID:      &m.SPObjectID,
// 	// 		RoleDefinitionID: &m.RoleDefinitionResourceID,
// 	// 	},
// 	// }

// 	_, err = rac.Create(m.Ctx, scope, uuid.New().String(), roleAssignmentParams, nil)
// 	// _, err = m.RoleAssignmentsClient.Create(m.Ctx, scope, uuid.New().String(), roleAssignmentParams)

// 	if err != nil {
// 		if strings.Contains(err.Error(), "RoleAssignmentExists") {
// 			log.Infoln("Role assignment already exists. Skipping...")
// 			return nil
// 		}
// 		return err
// 	}

// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// Initialise detachRolesFromSP detaches all roles from the SP
func (m *MinPermFinder) DetachRolesFromSP() error {

	filter := fmt.Sprintf("assignedTo('%s')", m.SPObjectID)
	resp, err := m.RoleAssignmentsClient.List(m.Ctx, filter)

	if err != nil {
		return err
	}

	roleAssignments := resp.Values()

	// delete all these role assignments
	for _, roleAssignment := range roleAssignments {
		_, err := m.RoleAssignmentsClient.DeleteByID(m.Ctx, *roleAssignment.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

// Delete Role Definition
// func (m *MinPermFinder) DeleteCustomRoleDefinition() error {

// 	// detach all roles from the SP

// 	_, err := m.RoleDefinitionsClient.Delete(m.Ctx, m.ResourceGroupResourceID, m.RoleDefinitionID)

// 	if err != nil {
// 		log.Warnf("Could not delete role definition: %s\n", err)
// 	}

// 	log.Infoln("Role definition deleted successfully")
// 	return nil
// }

func (m *MinPermFinder) DeleteCustomRoleDefinition() error {
	url := fmt.Sprintf("https://management.azure.com/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Authorization/roleDefinitions/%s?api-version=2018-01-01-preview", m.SubscriptionID, m.ResourceGroupName, m.RoleDefinitionID)

	client := &http.Client{}

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Go HTTP Client")

	// add bearer token to header
	req.Header.Add("Authorization", "Bearer "+m.DefaultAPIBearerToken)

	// make request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Warnf("Could not delete role definition: %s\n", err)
	}

	log.Debugln(string(body))
	log.Infoln("Role definition deleted successfully")

	return nil
}
