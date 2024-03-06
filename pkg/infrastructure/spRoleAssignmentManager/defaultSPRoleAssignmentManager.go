package sproleassignmentmanager

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/manisbindra/az-mpf/pkg/domain"
	"github.com/manisbindra/az-mpf/pkg/infrastructure/azureAPI"

	log "github.com/sirupsen/logrus"
)

type SPRoleAssignmentManager struct {
	azAPIClient *azureAPI.AzureAPIClients
}

func NewSPRoleAssignmentManager(subscriptionID string) *SPRoleAssignmentManager {
	azAPIClient := azureAPI.NewAzureAPIClients(subscriptionID)
	return &SPRoleAssignmentManager{
		azAPIClient: azAPIClient,
	}
}

func (r *SPRoleAssignmentManager) CreateUpdateCustomRole(subscription string, role domain.Role, permissions []string) error {

	// rgScope := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscription, resourceGroupName)
	subScope := fmt.Sprintf("/subscriptions/%s", subscription)

	data := map[string]interface{}{
		"assignableScopes": []string{
			// rgScope,
			subScope,
		},
		"description": role.RoleDefinitionName,
		"id":          role.RoleDefinitionResourceID,
		"name":        role.RoleDefinitionID,
		"permissions": []map[string]interface{}{
			{
				"actions":        permissions,
				"dataActions":    []string{},
				"notActions":     []string{},
				"notDataActions": []string{},
			},
		},
		"roleName": role.RoleDefinitionName,
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

	url := fmt.Sprintf("https://management.azure.com/subscriptions/%s/providers/Microsoft.Authorization/roleDefinitions/%s?api-version=2018-01-01-preview", subscription, role.RoleDefinitionID)

	client := &http.Client{}

	req, err := http.NewRequest("PUT", url, bytes.NewBufferString(jsonString))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Go HTTP Client")

	// add bearer token to header
	req.Header.Add("Authorization", "Bearer "+r.azAPIClient.DefaultAPIBearerToken)

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

func (r *SPRoleAssignmentManager) AssignRoleToSP(subscription string, SPOBjectID string, role domain.Role) error {

	// scope := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscription, resourceGroupName)
	scope := fmt.Sprintf("/subscriptions/%s", subscription)
	url := fmt.Sprintf("https://management.azure.com/%s/providers/Microsoft.Authorization/roleAssignments/%s?api-version=2022-04-01", scope, uuid.New().String())

	data := map[string]interface{}{
		"principalId":      SPOBjectID,
		"principalType":    "ServicePrincipal",
		"roleDefinitionId": role.RoleDefinitionResourceID,
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
	req.Header.Add("Authorization", "Bearer "+r.azAPIClient.DefaultAPIBearerToken)

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

// func (r *SPRoleAssignmentManager) AssignRoleToSP(scope string) error {

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
// }z

// Initialise detachRolesFromSP detaches all roles from the SP
func (r *SPRoleAssignmentManager) DetachRolesFromSP(ctx context.Context, subscription string, SPOBjectID string, role domain.Role) error {

	filter := fmt.Sprintf("assignedTo('%s')", SPOBjectID)
	resp, err := r.azAPIClient.RoleAssignmentsClient.List(ctx, filter)

	if err != nil {
		return err
	}

	roleAssignments := resp.Values()

	// delete all these role assignments
	for _, roleAssignment := range roleAssignments {
		_, err := r.azAPIClient.RoleAssignmentsClient.DeleteByID(ctx, *roleAssignment.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *SPRoleAssignmentManager) DeleteCustomRole(subscription string, role domain.Role) error {
	url := fmt.Sprintf("https://management.azure.com/subscriptions/%s/providers/Microsoft.Authorization/roleDefinitions/%s?api-version=2018-01-01-preview", subscription, role.RoleDefinitionID)

	client := &http.Client{}

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Go HTTP Client")

	// add bearer token to header
	req.Header.Add("Authorization", "Bearer "+r.azAPIClient.DefaultAPIBearerToken)

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
