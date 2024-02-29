package ARMTemplateWhatIf

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	// "log"
	"net/http"
	URL "net/url"
	"strings"

	"github.com/manisbindra/az-mpf/pkg/domain"
	"github.com/manisbindra/az-mpf/pkg/infrastructure/ARMTemplateShared"
	"github.com/manisbindra/az-mpf/pkg/infrastructure/azureAPI"
	"github.com/manisbindra/az-mpf/pkg/infrastructure/mpfSharedUtils"

	log "github.com/sirupsen/logrus"
)

type armWhatIfConfig struct {
	// mpfConfig   domain.MPFConfig
	armConfig   ARMTemplateShared.ArmTemplateAdditionalConfig
	azAPIClient *azureAPI.AzureAPIClients
}

func NewARMTemplateWhatIfAuthorizationChecker(subscriptionID string, armConfig ARMTemplateShared.ArmTemplateAdditionalConfig) *armWhatIfConfig {
	azAPIClient := azureAPI.NewAzureAPIClients(subscriptionID)
	return &armWhatIfConfig{
		azAPIClient: azAPIClient,
		armConfig:   armConfig,
	}

}

func (a *armWhatIfConfig) GetDeploymentAuthorizationErrors(mpfConfig domain.MPFConfig) (string, error) {
	return a.GetARMWhatIfAuthorizationErrors(a.armConfig.DeploymentName, mpfConfig)
}

func (a *armWhatIfConfig) CleanDeployment(mpfConfig domain.MPFConfig) error {
	log.Infoln("No additional cleanup needed in WhatIf mode")
	log.Infoln("*************************")

	return nil
}

// Get parameters in standard format that is without the schema, contentVersion and parameters fields

func (a *armWhatIfConfig) CreateEmptyDeployment(client *http.Client, deploymentName string, bearerToken string, mpfConfig domain.MPFConfig) error {

	deploymentUri := fmt.Sprintf("https://management.azure.com/subscriptions/%s/resourcegroups/%s/providers/Microsoft.Resources/deployments/%s?api-version=2020-10-01", mpfConfig.SubscriptionID, mpfConfig.ResourceGroup.ResourceGroupName, deploymentName)

	log.Info("Creating empty deployment...")
	log.Debug(deploymentUri)

	emptyTempl, err := mpfSharedUtils.ReadJson("../samples/templates/empty.json")
	if err != nil {
		return err
	}

	emptyTemplStdFmtMap := map[string]interface{}{
		"properties": map[string]interface{}{
			"mode":       "Incremental",
			"template":   emptyTempl,
			"parameters": map[string]interface{}{},
		},
	}

	// convert bodyJSON to string
	emptyTemplJSONBytes, err := json.Marshal(emptyTemplStdFmtMap)
	if err != nil {
		return err
	}

	emptyDeploymentJSONString := string(emptyTemplJSONBytes)

	deploymentReq, err := http.NewRequest("PUT", deploymentUri, bytes.NewBufferString(emptyDeploymentJSONString))
	if err != nil {
		return err
	}
	deploymentReq.Header.Set("Content-Type", "application/json")
	deploymentReq.Header.Set("Accept", "application/json")
	deploymentReq.Header.Set("User-Agent", "Go HTTP Client")

	// add bearer token to header
	deploymentReq.Header.Add("Authorization", "Bearer "+bearerToken)

	log.Debugf("%v", deploymentReq)

	// make deploymentReq
	deploymentResp, err := client.Do(deploymentReq)
	if err != nil {
		return err
	}
	log.Debugf("%v", deploymentResp)
	defer deploymentResp.Body.Close()

	return nil
}

func (a *armWhatIfConfig) GetARMWhatIfAuthorizationErrors(deploymentName string, mpfConfig domain.MPFConfig) (string, error) {

	bearerToken, err := a.azAPIClient.GetSPBearerToken(mpfConfig.TenantID, mpfConfig.SP.SPClientID, mpfConfig.SP.SPClientSecret)
	if err != nil {
		return "", err
	}

	// read template and parameters
	template, err := mpfSharedUtils.ReadJson(a.armConfig.TemplateFilePath)
	if err != nil {
		return "", err
	}

	parameters, err := mpfSharedUtils.ReadJson(a.armConfig.ParametersFilePath)
	if err != nil {
		return "", err
	}

	// convert parameters to standard format
	parameters = ARMTemplateShared.GetParametersInStandardFormat(parameters)

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
		return "", err
	}

	fullTemplateJSONString := string(fullTemplateJSONBytes)

	log.Debugln()
	log.Debugln(fullTemplateJSONString)
	log.Debugln()
	// create JSON body with template and parameters

	client := &http.Client{}

	// check if deplolyment exists if not create empty deployment

	// deploymentExists, err := a.CheckARMDeploymentExists(deploymentName, mpfConfig)
	// if !deploymentExists {
	// log.Info("MPF mode is whatif, creating empty deployment....")
	// err = a.CreateEmptyDeployment(client, deploymentName, bearerToken, mpfConfig)
	// if err != nil {
	// 	return "", err
	// }
	// }

	url := fmt.Sprintf("https://management.azure.com/subscriptions/%s/resourcegroups/%s/providers/Microsoft.Resources/deployments/%s/whatIf?api-version=2021-04-01", mpfConfig.SubscriptionID, mpfConfig.ResourceGroup.ResourceGroupName, deploymentName)
	reqMethod := "POST"

	req, err := http.NewRequest(reqMethod, url, bytes.NewBufferString(fullTemplateJSONString))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Go HTTP Client")

	// add bearer token to header
	req.Header.Add("Authorization", "Bearer "+bearerToken)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	whatIfRespLoc := resp.Header.Get("Location")
	log.Debugf("What if response Location: %s \n", whatIfRespLoc)

	_, err = URL.ParseRequestURI(whatIfRespLoc)
	if err != nil {
		return "", err
	}

	respBody, err := a.GetWhatIfResp(whatIfRespLoc, bearerToken)
	if err != nil {
		log.Infof("Could not fetch what if response: %v \n", err)
		return "", err
	}

	// fmt.Println(respBody)
	log.Debugln(respBody)
	// print response body
	if strings.Contains(respBody, "Authorization") {
		return respBody, nil
	}

	if strings.Contains(respBody, "InvalidTemplateDeployment") {
		// This indicates all Authorization errors are fixed
		// Sample error [{\"code\":\"PodIdentityAddonFeatureFlagNotEnabled\",\"message\":\"Provisioning of resource(s) for container service aks-24xalwx7i2ueg in resource group testdeployrg-Y2jsRAG failed. Message: PodIdentity addon is not allowed since feature 'Microsoft.ContainerService/EnablePodIdentityPreview' is not enabled.
		// Hence ok to proceed, and not return error in this condition
		log.Warnf("Non Authorizaton error occured: %s", respBody)
	}

	return "", nil

}

func (a *armWhatIfConfig) GetWhatIfResp(whatIfRespLoc string, bearerToken string) (string, error) {

	client := &http.Client{}

	req, err := http.NewRequest("GET", whatIfRespLoc, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Go HTTP Client")

	// add bearer token to header
	req.Header.Add("Authorization", "Bearer "+bearerToken)

	var respBody string
	maxRetries := 10
	retryCount := 0
	for {
		// make request
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		// read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		respBody = string(body)

		// If response body is not empty, break out of loop
		if respBody != "" {
			log.Infoln("Whatif Results Response Received..")
			break
		}

		retryCount++
		if retryCount == maxRetries {
			log.Warnf("Whatif Results Response Body is empty after %d retries, returning empty response body", maxRetries)
			return "", errors.New("Whatif Results Response Body is empty after 10 retries")
		}

		log.Infoln("Whatif Results Response Body is empty, retrying in a bit...")
		// Sleep for 500 milli seconds and try again
		time.Sleep(500 * time.Millisecond)

		// fmt.Println(respBody)
		// print response body
	}

	log.Debugln("Whatif Results Response Body:")
	log.Debugln(respBody)

	return respBody, nil

}

// func (a *armWhatIfConfig) CheckARMDeploymentExists(deploymentName string, mpfConfig domain.MPFConfig) (bool, error) {

// 	bearerToken, err := a.azAPIClient.GetSPBearerToken(mpfConfig.TenantID, mpfConfig.SP.SPClientID, mpfConfig.SP.SPClientSecret)
// 	if err != nil {
// 		return false, err
// 	}

// 	url := fmt.Sprintf("https://management.azure.com/subscriptions/%s/resourcegroups/%s/providers/Microsoft.Resources/deployments/%s?api-version=2020-10-01", mpfConfig.SubscriptionID, mpfConfig.ResourceGroup.ResourceGroupName, deploymentName)

// 	client := &http.Client{}

// 	req, err := http.NewRequest("GET", url, nil)
// 	if err != nil {
// 		return false, err
// 	}

// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Accept", "application/json")
// 	req.Header.Set("User-Agent", "Go HTTP Client")

// 	// add bearer token to header
// 	req.Header.Add("Authorization", "Bearer "+bearerToken)

// 	// make request
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return false, err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode == 404 {
// 		return false, nil
// 	}

// 	log.Debugf("Check Deployment exists response status code: %d \n", resp.StatusCode)

// 	return true, nil

// }
