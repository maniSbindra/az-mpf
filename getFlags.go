package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

func (m *MinPermFinder) ParseFlags() error {
	var (
		subscriptionID               string
		resourceGroupNamePfx         string
		location                     string
		deploymentNamePfx            string
		servicePrincipalClientID     string
		servicePrincipalObjectID     string
		servicePrincipalClientSecret string
		tenantID                     string
		templateFilePath             string
		parametersFilePath           string
		mpfMode                      string

		showDetailedOutput bool
		jsonOutput         bool
	)

	flag.StringVar(&subscriptionID, "subscriptionID", "", "Azure Subscription ID")
	flag.StringVar(&resourceGroupNamePfx, "resourceGroupName", "testdeployrg", "Resource Group Name Prefix")
	flag.StringVar(&deploymentNamePfx, "deploymentNamePfx", "testDeploy", "Deployment Name Prefix")
	flag.StringVar(&servicePrincipalClientID, "spClientID", "", "Service Principal Client ID")
	flag.StringVar(&servicePrincipalObjectID, "spObjectID", "", "Service Principal Object ID")
	flag.StringVar(&servicePrincipalClientSecret, "spClientSecret", "", "Service Principal Client Secret")
	flag.StringVar(&tenantID, "tenantID", "", "Azure Tenant ID")
	flag.StringVar(&templateFilePath, "templateFile", "", "Path to ARM Template File")
	flag.StringVar(&parametersFilePath, "parametersFile", "", "Path to Template Parameters File")

	// optional flags
	flag.BoolVar(&showDetailedOutput, "showDetailedOutput", false, "Show detailed output")
	flag.BoolVar(&jsonOutput, "jsonOutput", false, "Output results in JSON format")
	flag.StringVar(&location, "location", "eastus", "Azure Region to deploy to")
	flag.StringVar(&mpfMode, "mpfMode", "whatif", "Mode to run MinPermFinder in. Options: whatif, fullDeployment. default: whatif")

	flag.Parse()

	// log.Debug arguments

	// // print variables
	log.Debugln("subscriptionID:", subscriptionID)
	log.Debugln("resourceGroupName:", resourceGroupNamePfx)
	log.Debugln("deploymentNamePfx:", deploymentNamePfx)
	log.Debugln("servicePrincipalClientID:", servicePrincipalClientID)
	log.Debugln("servicePrincipalObjectID:", servicePrincipalObjectID)
	log.Debugln("tenantID:", tenantID)
	log.Debugln("templateFilePath:", templateFilePath)
	log.Debugln("parametersFilePath:", parametersFilePath)
	log.Debugln("mpfMode:", mpfMode)

	// // print values of environment variables
	log.Debugln("SUBSCRIPTION_ID:", os.Getenv("SUBSCRIPTION_ID"))
	log.Debugln("TENANT_ID:", os.Getenv("TENANT_ID"))
	log.Debugln("SP_CLIENT_ID:", os.Getenv("SP_CLIENT_ID"))
	log.Debugln("SP_OBJECT_ID:", os.Getenv("SP_OBJECT_ID"))
	log.Debugln("TEST_DEPLOYMENT_NAME_PFX:", os.Getenv("TEST_DEPLOYMENT_NAME_PFX"))
	log.Debugln("TEST_DEPLOYMENT_RESOURCE_GROUP_NAME:", os.Getenv("TEST_DEPLOYMENT_RESOURCE_GROUP_NAME"))
	log.Debugln("TEMPLATE_FILE:", os.Getenv("TEMPLATE_FILE"))
	log.Debugln("PARAMETERS_FILE:", os.Getenv("PARAMETERS_FILE"))

	// if arguments are not provided using flags, use environment variables if provided
	if subscriptionID == "" && os.Getenv("SUBSCRIPTION_ID") != "" {
		subscriptionID = os.Getenv("SUBSCRIPTION_ID")
	}

	if resourceGroupNamePfx == "" && os.Getenv("TEST_DEPLOYMENT_RESOURCE_GROUP_NAME_PFX") != "" {
		resourceGroupNamePfx = os.Getenv("TEST_DEPLOYMENT_RESOURCE_GROUP_NAME_PFX")
	}

	if deploymentNamePfx == "" && os.Getenv("TEST_DEPLOYMENT_NAME_PFX") != "" {
		deploymentNamePfx = os.Getenv("TEST_DEPLOYMENT_NAME_PFX")
	}

	if servicePrincipalClientID == "" && os.Getenv("SP_CLIENT_ID") != "" {
		servicePrincipalClientID = os.Getenv("SP_CLIENT_ID")
	}

	if servicePrincipalObjectID == "" && os.Getenv("SP_OBJECT_ID") != "" {
		servicePrincipalObjectID = os.Getenv("SP_OBJECT_ID")
	}

	if servicePrincipalClientSecret == "" && os.Getenv("SP_CLIENT_SECRET") != "" {
		servicePrincipalClientSecret = os.Getenv("SP_CLIENT_SECRET")
	}

	if tenantID == "" && os.Getenv("TENANT_ID") != "" {
		tenantID = os.Getenv("TENANT_ID")
	}

	if templateFilePath == "" && os.Getenv("TEMPLATE_FILE") != "" {
		templateFilePath = os.Getenv("TEMPLATE_FILE")
	}

	if parametersFilePath == "" && os.Getenv("PARAMETERS_FILE") != "" {
		parametersFilePath = os.Getenv("PARAMETERS_FILE")
	}

	if subscriptionID == "" || resourceGroupNamePfx == "" || deploymentNamePfx == "" ||
		servicePrincipalClientID == "" || servicePrincipalClientSecret == "" ||
		tenantID == "" || templateFilePath == "" || parametersFilePath == "" || servicePrincipalObjectID == "" {
		log.Debugln("Please provide all the required parameters using flags.")
		flag.Usage()
		// return error with values of all required parameters except secret
		vals := fmt.Sprintf("subscriptionID: %s, resourceGroupName: %s, deploymentNamePfx: %s, servicePrincipalClientID: %s, servicePrincipalObjectID: %s, tenantID: %s, templateFilePath: %s, parametersFilePath: %s", subscriptionID, resourceGroupNamePfx, deploymentNamePfx, servicePrincipalClientID, servicePrincipalObjectID, tenantID, templateFilePath, parametersFilePath)
		return errors.New("Values of of all required parameters not received. Values received: " + vals)
	}

	// set values in receiver
	m.SubscriptionID = subscriptionID
	m.ResourceGroupNamePfx = resourceGroupNamePfx
	m.SPClientID = servicePrincipalClientID
	m.SPObjectID = servicePrincipalObjectID
	m.SPClientSecret = servicePrincipalClientSecret
	m.TenantID = tenantID
	m.TemplateFilePath = templateFilePath
	m.ParametersFilePath = parametersFilePath
	m.DeploymentNamePfx = deploymentNamePfx

	// set optional values in receiver
	m.ShowDetailedOutput = showDetailedOutput
	m.JSONOutput = jsonOutput
	m.Location = location
	m.MPFMode = mpfMode

	return nil

}
