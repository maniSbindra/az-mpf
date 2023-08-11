package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	// "log"
	"net/http"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	log "github.com/sirupsen/logrus"
)

func readJson(path string) (map[string]interface{}, error) {
	templateFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	template := make(map[string]interface{})
	if err := json.Unmarshal(templateFile, &template); err != nil {
		return nil, err
	}

	return template, nil
}

// func (m *MinPermFinder) checkExistDeployment(ctx context.Context, resourceGroupName string, deploymentName string) (bool, error) {

// 	boolResp, err := m.DeploymentsClient.CheckExistence(ctx, resourceGroupName, deploymentName, nil)
// 	if err != nil {
// 		return false, err
// 	}

// 	return boolResp.Success, nil
// }

// func (m *MinPermFinder) createDeployment(deploymentName string, template, params map[string]interface{}) (*armresources.DeploymentExtended, error) {

// 	deploymentPollerResp, err := m.DeploymentsClient.BeginCreateOrUpdate(
// 		m.Ctx,
// 		m.ResourceGroupName,
// 		deploymentName,
// 		armresources.Deployment{
// 			Properties: &armresources.DeploymentProperties{
// 				Template:   template,
// 				Parameters: params,
// 				Mode:       to.Ptr(armresources.DeploymentModeIncremental),
// 			},
// 		},
// 		nil)

// 	if err != nil {
// 		return nil, fmt.Errorf("cannot create deployment: %v", err)
// 	}

// 	resp, err := deploymentPollerResp.PollUntilDone(m.Ctx, nil)
// 	if err != nil {
// 		return nil, fmt.Errorf("cannot get the create deployment future respone: %v", err)
// 	}

// 	return &resp.DeploymentExtended, nil
// }

// func (m *MinPermFinder) validateDeployment(ctx context.Context, resourceGroupName string, deploymentName string, template, params map[string]interface{}) (*armresources.DeploymentValidateResult, error) {

// 	pollerResp, err := m.DeploymentsClient.BeginValidate(
// 		ctx,
// 		resourceGroupName,
// 		deploymentName,
// 		armresources.Deployment{
// 			Properties: &armresources.DeploymentProperties{
// 				Template:   template,
// 				Parameters: params,
// 				Mode:       to.Ptr(armresources.DeploymentModeIncremental),
// 			},
// 		},
// 		nil)

// 	if err != nil {
// 		return nil, err
// 	}

// 	resp, err := pollerResp.PollUntilDone(ctx, nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &resp.DeploymentValidateResult, nil
// }

// Get parameters in standard format that is without the schema, contentVersion and parameters fields
func getParametersInStandardFormat(parameters map[string]interface{}) map[string]interface{} {
	// convert from
	// {
	// 	"$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentParameters.json#",
	// 	"contentVersion": "1.0.0.0",
	// 	"parameters": {
	// 	  "adminUsername": {
	// 		"value": "GEN-UNIQUE"
	// 	  },
	// 	  "adminPasswordOrKey": {
	// 		"value": "GEN-PASSWORD"
	// 	  },
	// 	  "dnsLabelPrefix": {
	// 		"value": "GEN-UNIQUE"
	// 	  }
	// 	}
	//   }

	// convert to
	// {
	// 	  "adminUsername": {
	// 		"value": "GEN-UNIQUE"
	// 	  },
	// 	  "adminPasswordOrKey": {
	// 		"value": "GEN-PASSWORD"
	// 	  },
	// 	  "dnsLabelPrefix": {
	// 		"value": "GEN-UNIQUE"
	// 	  }
	// 	}
	if parameters["$schema"] != nil {

		return parameters["parameters"].(map[string]interface{})

	}
	return parameters
}

func (m *MinPermFinder) DeployARMTemplate(deploymentName string) error {

	// jsonData, err := json.Marshal(properties)
	spCred, err := azidentity.NewClientSecretCredential(m.TenantID, m.SPClientID, m.SPClientSecret, nil)
	if err != nil {
		log.Fatal(err)
	}

	bearerToken, err := m.getBearerToken(spCred)
	if err != nil {
		return err
	}

	// read template and parameters
	template, err := readJson(m.TemplateFilePath)
	if err != nil {
		return err
	}

	parameters, err := readJson(m.ParametersFilePath)
	if err != nil {
		return err
	}

	// convert parameters to standard format
	parameters = getParametersInStandardFormat(parameters)

	fullTemplate := map[string]interface{}{
		"properties": map[string]interface{}{
			"mode":       "Incremental",
			"template":   template,
			"parameters": parameters,
		},
	}

	// convert bodyJSON to string
	fullTemplateJSONBytes, err := json.Marshal(fullTemplate)
	if err != nil {
		return err
	}

	fullTemplateJSONString := string(fullTemplateJSONBytes)

	log.Debugln()
	log.Debugln(fullTemplateJSONString)
	log.Debugln()
	// create JSON body with template and parameters

	url := fmt.Sprintf("https://management.azure.com/subscriptions/%s/resourcegroups/%s/providers/Microsoft.Resources/deployments/%s?api-version=2020-10-01", m.SubscriptionID, m.ResourceGroupName, deploymentName)

	client := &http.Client{}

	req, err := http.NewRequest("PUT", url, bytes.NewBufferString(fullTemplateJSONString))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Go HTTP Client")

	// add bearer token to header
	req.Header.Add("Authorization", "Bearer "+bearerToken)

	// make request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	respBody := string(body)

	// fmt.Println(respBody)
	log.Debugln(respBody)
	// print response body
	if strings.Contains(respBody, "Authorization") {
		return errors.New(respBody)
	}

	if strings.Contains(respBody, "InvalidTemplateDeployment") {
		// This indicates all Authorization errors are fixed
		// Sample error [{\"code\":\"PodIdentityAddonFeatureFlagNotEnabled\",\"message\":\"Provisioning of resource(s) for container service aks-24xalwx7i2ueg in resource group testdeployrg-Y2jsRAG failed. Message: PodIdentity addon is not allowed since feature 'Microsoft.ContainerService/EnablePodIdentityPreview' is not enabled.
		// Hence ok to proceed, and not return error in this condition
		log.Warnf("Non Authorizaton error occured: %s", respBody)
	}

	return nil

}

func (m *MinPermFinder) GetARMDeployment(deploymentName string) error {

	spCred, err := azidentity.NewClientSecretCredential(m.TenantID, m.SPClientID, m.SPClientSecret, nil)
	if err != nil {
		log.Fatal(err)
	}

	bearerToken, err := m.getBearerToken(spCred)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://management.azure.com/subscriptions/%s/resourcegroups/%s/providers/Microsoft.Resources/deployments/%s?api-version=2020-10-01", m.SubscriptionID, m.ResourceGroupName, deploymentName)

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Go HTTP Client")

	// add bearer token to header
	req.Header.Add("Authorization", "Bearer "+bearerToken)

	// make request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	respBody := string(body)

	// fmt.Println(respBody)
	log.Debugln(respBody)
	// print response body
	if strings.Contains(respBody, "Authorization") {
		return errors.New(respBody)
	}

	return nil
}

// Delete ARM deployment
func (m *MinPermFinder) CancelDeployment(deploymentName string) error {

	// Get deployments status. If status is "Running", cancel deployment, then delete deployment
	getResp, err := m.DeploymentsClient.Get(m.Ctx, m.ResourceGroupName, deploymentName, nil)
	if err != nil {
		// Error indicates deployment does not exist, so cancelling deployment not needed
		if strings.Contains(err.Error(), "DeploymentNotFound") {
			log.Infof("Could not get deployment %s: ,Error :%s \n", deploymentName, err)
			return nil
		}
	}

	log.Infof("Deployment status: %s\n", *getResp.DeploymentExtended.Properties.ProvisioningState)

	if *getResp.DeploymentExtended.Properties.ProvisioningState == armresources.ProvisioningStateRunning {

		retryCount := 0
		for _, err := m.DeploymentsClient.Cancel(m.Ctx, m.ResourceGroupName, deploymentName, nil); err != nil; {
			// cancel deployment
			if err != nil {
				// return err
				log.Warnf("Could not cancel deployment %s: %s, retrying in a bit", deploymentName, err)
				time.Sleep(5 * time.Second)
				retryCount++
				if retryCount >= 24 {
					log.Warnf("Could not cancel deployment %s: %s, giving up", deploymentName, err)
					return errors.New("Could Not Cancel Deployment")
				}

			}

		}
		log.Infof("Cancelled deployment %s", deploymentName)
		return nil
	}

	return nil
}

// create method to delete resource group
func (m *MinPermFinder) DeleteResourceGroup() error {

	_, err := m.ResourceGroupsClient.BeginDelete(m.Ctx, m.ResourceGroupName, nil)
	if err != nil {
		return err
	}

	return nil
}

// method to create resource group
func (m *MinPermFinder) CreateResourceGroup() error {

	rgParams := armresources.ResourceGroup{
		Location: &m.Location,
		Name:     &m.ResourceGroupName,
	}

	// create resource group
	_, err := m.ResourceGroupsClient.CreateOrUpdate(m.Ctx, m.ResourceGroupName, rgParams, nil)
	if err != nil {
		return err
	}

	return nil
}
