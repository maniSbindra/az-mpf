// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/authorization/mgmt/authorization"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

type MinPermFinder struct {

	// Context
	Ctx context.Context

	// Config Flags
	SubscriptionID       string
	ResourceGroupNamePfx string
	Location             string
	SPClientID           string
	SPObjectID           string
	SPClientSecret       string
	TenantID             string
	TemplateFilePath     string
	ParametersFilePath   string
	DeploymentNamePfx    string

	// ResourceGroupName value set from prefix
	ResourceGroupName string

	// Optional Flags
	ShowDetailedOutput bool
	JSONOutput         bool

	// API Clients using Default Creds
	RoleAssignmentsClient authorization.RoleAssignmentsClient
	RoleDefinitionsClient authorization.RoleDefinitionsClient
	DeploymentsClient     *armresources.DeploymentsClient
	ResourceGroupsClient  *armresources.ResourceGroupsClient

	// Default CLI Creds
	DefaultCred           *azidentity.DefaultAzureCredential
	DefaultAPIBearerToken string

	// Other variables
	RoleDefinitionID          string
	RoleDefinitionName        string
	RoleDefinitionDescription string
	RoleDefinitionResourceID  string
	ResourceGroupResourceID   string

	// The map from which the minimum permissions will be calculated
	scopePermissionMap map[string][]string
}

func main() {

	logLevel, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		logLevel = log.ErrorLevel
	}
	log.SetLevel(logLevel)

	mpf := &MinPermFinder{}
	ctx := context.Background()
	mpf.Ctx = ctx

	// Parse CLI Flags
	err = mpf.ParseFlags()
	if err != nil {
		log.Fatal(err)
	}

	// Set API Clients: Set Role Assignments Client, Role Definitions Client, Deployments Client, Resource Groups Client using CLI Default Creds
	err = mpf.SetApiClients()
	if err != nil {
		log.Fatal(err)
	}

	// Set API Access Bearer token using CLI Default Creds
	err = mpf.SetDefaultAPIAccessBearerToken()
	if err != nil {
		// log.Fatal(err)
		log.Fatal(err)
	}

	// Initialize other configuration values
	deploymentName := fmt.Sprintf("%s-%s", mpf.DeploymentNamePfx, generateRandomString(7))
	roleDefUUID, _ := uuid.NewRandom()
	mpf.RoleDefinitionID = roleDefUUID.String()
	mpf.RoleDefinitionName = fmt.Sprintf("tmp-rol-%s", generateRandomString(7))
	mpf.RoleDefinitionResourceID = fmt.Sprintf("/subscriptions/%s/providers/Microsoft.Authorization/roleDefinitions/%s", mpf.SubscriptionID, mpf.RoleDefinitionID)
	log.Infoln("roleDefinitionResourceID:", mpf.RoleDefinitionResourceID)
	mpf.ResourceGroupName = fmt.Sprintf("%s-%s", mpf.ResourceGroupNamePfx, generateRandomString(7))
	mpf.ResourceGroupResourceID = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", mpf.SubscriptionID, mpf.ResourceGroupName)

	mpf.scopePermissionMap = make(map[string][]string)

	// Create Resource Group
	log.Infof("Creating Resource Group: %s \n", mpf.ResourceGroupName)
	err = mpf.CreateResourceGroup()
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Resource Group: %s created successfully \n", mpf.ResourceGroupName)
	defer mpf.CleanUpResources(deploymentName)

	// Delete all existing role assignments for the service principal
	err = mpf.DetachRolesFromSP()
	if err != nil {
		log.Warnf("Unable to delete Role Assignments: %v\n", err)
		return
	}
	log.Info("Deleted all existing role assignments for service principal \n")

	// Initialize new custom role
	log.Infoln("Initializing Custom Role")
	err = mpf.CreateUpdateCustomRole([]string{})
	if err != nil {
		log.Warn(err)
		return
	}
	log.Infoln("Custom role initialized successfully")

	// Assign new custom role to service principal
	log.Infoln("Assigning new custom role to service principal")
	err = mpf.AssignRoleToSP()
	if err != nil {
		log.Warn(err)
		return
	}
	log.Infoln("New Custom Role assigned to service principal successfully")

	// This loop does the following
	// Try creating deployment, for any authorization error, get permission/action that resulted in error, add it to custom Role
	// Then Retry creation of deployment. After deployment created repeat the loop for getDeployment, and then if no error exit
	fnValidate := mpf.DeployARMTemplate
	validationPhase := "createDeployment"
	for {
		// deploymentName += "1"

		if validationPhase == "getDeployment" {
			fnValidate = mpf.GetARMDeployment
		}

		deploymentAuthorizationErr := fnValidate(deploymentName)

		if deploymentAuthorizationErr == nil {

			if validationPhase == "getDeployment" {
				log.Infoln("create deployment and get deployment successful...")
				log.Infoln("*************************")
				// authorizationErrors = false
				break
			}
			log.Infoln("create deployment successful, moving to get deployment...")
			log.Infoln("*************************")
			validationPhase = "getDeployment"
			continue
		}

		log.Debugln("Deployment Authorization Error:", deploymentAuthorizationErr)

		scpMp, err := parseDeploymentError(deploymentAuthorizationErr)
		if err != nil {
			log.Warnf("Could Not Parse Deployment Authorization Error: %v \n", err)
			return
		}

		log.Infoln("Successfully Parsed Deployment Authorization Error")
		log.Debugln("scope permissions found from deployment error:", scpMp)

		log.Infoln("Adding mising scopes/permissions to final result map...")
		for k, v := range scpMp {
			mpf.scopePermissionMap[k] = append(mpf.scopePermissionMap[k], v...)
			mpf.scopePermissionMap[mpf.ResourceGroupResourceID] = append(mpf.scopePermissionMap[mpf.ResourceGroupResourceID], v...)
		}

		// assign permission to role
		log.Infoln("Adding permission/scope to role...........")
		log.Debugln("Number of Permissions added to role:", len(mpf.scopePermissionMap[mpf.ResourceGroupResourceID]))

		err = mpf.CreateUpdateCustomRole(mpf.scopePermissionMap[mpf.ResourceGroupResourceID])
		if err != nil {
			log.Infoln("Error when adding permission/scope to role: \n", err)
			log.Warn(err)
			return
		}
		log.Infoln("Permission/scope added to role successfully")
	}

	// Format and print result
	mpf.FormatResult()
}

// Function to generate a random string of a given length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charsetLength := big.NewInt(int64(len(charset)))

	randomString := make([]byte, length)
	for i := range randomString {
		randomIndex, _ := rand.Int(rand.Reader, charsetLength)
		randomString[i] = charset[randomIndex.Int64()]
	}

	return string(randomString)
}

// This function cleans up temporary resources created including the resource group
func (m *MinPermFinder) CleanUpResources(deploymentName string) {
	log.Infoln("Cleaning up resources...")
	log.Infoln("*************************")

	// Cancel deployment. Even if cancelling deployment fails attempt to delete other resources
	_ = m.CancelDeployment(deploymentName)

	// Detach Roles from SP
	err := m.DetachRolesFromSP()
	if err != nil {
		log.Warnf("Could not detach roles from SP: %s\n", err)
	}

	// Delete Custom Role
	err = m.DeleteCustomRoleDefinition()
	if err != nil {
		log.Warnf("Could not delete custom role: %s\n", err)
	}

	// Delete Resource Group
	err = m.DeleteResourceGroup()
	if err != nil {
		log.Warnf("Error when deleting resource group: %s \n", err)
	}
	log.Infoln("Resource group deletion initiated successfully...")

}
