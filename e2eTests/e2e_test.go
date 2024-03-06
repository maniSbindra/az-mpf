package e2etests

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/manisbindra/az-mpf/pkg/infrastructure/ARMTemplateShared"
	"github.com/manisbindra/az-mpf/pkg/infrastructure/authorizationCheckers/ARMTemplateDeployment"
	"github.com/manisbindra/az-mpf/pkg/infrastructure/authorizationCheckers/ARMTemplateWhatIf"
	"github.com/manisbindra/az-mpf/pkg/infrastructure/authorizationCheckers/terraform"
	mpfSharedUtils "github.com/manisbindra/az-mpf/pkg/infrastructure/mpfSharedUtils"
	resourceGroupManager "github.com/manisbindra/az-mpf/pkg/infrastructure/resourceGroupManager"
	sproleassignmentmanager "github.com/manisbindra/az-mpf/pkg/infrastructure/spRoleAssignmentManager"
	"github.com/manisbindra/az-mpf/pkg/presentation"
	"github.com/manisbindra/az-mpf/pkg/usecase"
	"github.com/stretchr/testify/assert"
)

func getTestingMPFArgs() (presentation.MpfCLIArgs, error) {

	subscriptionID := os.Getenv("SUBSCRIPTION_ID")
	servicePrincipalClientID := os.Getenv("SP_CLIENT_ID")
	servicePrincipalObjectID := os.Getenv("SP_OBJECT_ID")
	servicePrincipalClientSecret := os.Getenv("SP_CLIENT_SECRET")
	tenantID := os.Getenv("TENANT_ID")
	resourceGroupNamePfx := "e2eTest"
	deploymentNamePfx := "e2eTest"
	location := "eastus"

	if subscriptionID == "" || servicePrincipalClientID == "" || servicePrincipalObjectID == "" || servicePrincipalClientSecret == "" || tenantID == "" {
		return presentation.MpfCLIArgs{}, errors.New("required environment variables not set")
	}

	return presentation.MpfCLIArgs{
		SubscriptionID:       subscriptionID,
		ResourceGroupNamePfx: resourceGroupNamePfx,
		DeploymentNamePfx:    deploymentNamePfx,
		SPClientID:           servicePrincipalClientID,
		SPObjectID:           servicePrincipalObjectID,
		SPClientSecret:       servicePrincipalClientSecret,
		TenantID:             tenantID,
		Location:             location,
	}, nil

}

func TestARMTemplatWhatIfMultiResourceTemplate(t *testing.T) {

	mpfArgs, err := getTestingMPFArgs()
	if err != nil {
		t.Skip("required environment variables not set, skipping end to end test")
	}
	mpfArgs.TemplateFilePath = "../samples/templates/multi-resource-template.json"
	mpfArgs.ParametersFilePath = "../samples/templates/multi-resource-parameters.json"

	ctx := context.Background()

	mpfConfig := presentation.GetMPFConfig(mpfArgs)

	deploymentName := fmt.Sprintf("%s-%s", mpfArgs.DeploymentNamePfx, mpfSharedUtils.GenerateRandomString(7))
	armConfig := &ARMTemplateShared.ArmTemplateAdditionalConfig{
		TemplateFilePath:   mpfArgs.TemplateFilePath,
		ParametersFilePath: mpfArgs.ParametersFilePath,
		DeploymentName:     deploymentName,
	}

	// azAPIClient := azureAPI.NewAzureAPIClients(mpfArgs.SubscriptionID)
	var rgManager usecase.ResourceGroupManager
	var spRoleAssignmentManager usecase.ServicePrincipalRolemAssignmentManager
	rgManager = resourceGroupManager.NewResourceGroupManager(mpfArgs.SubscriptionID)
	spRoleAssignmentManager = sproleassignmentmanager.NewSPRoleAssignmentManager(mpfArgs.SubscriptionID)

	var deploymentAuthorizationCheckerCleaner usecase.DeploymentAuthorizationCheckerCleaner
	var mpfService *usecase.MPFService

	deploymentAuthorizationCheckerCleaner = ARMTemplateWhatIf.NewARMTemplateWhatIfAuthorizationChecker(mpfArgs.SubscriptionID, *armConfig)
	initialPermissionsToAdd := []string{"Microsoft.Resources/deployments/*", "Microsoft.Resources/subscriptions/operationresults/read"}
	permissionsToAddToResult := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}
	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, true, false, true)

	mpfResult, err := mpfService.GetMinimumPermissionsRequired()
	if err != nil {
		t.Error(err)
	}

	//check if mpfResult.RequiredPermissions is not empty and has 54 permissions for scope ResourceGroupResourceID
	// Microsoft.Authorization/roleAssignments/read
	// Microsoft.Authorization/roleAssignments/write
	// Microsoft.Compute/virtualMachines/extensions/read
	// Microsoft.Compute/virtualMachines/extensions/write
	// Microsoft.Compute/virtualMachines/read
	// Microsoft.Compute/virtualMachines/write
	// Microsoft.ContainerRegistry/registries/read
	// Microsoft.ContainerRegistry/registries/write
	// Microsoft.ContainerService/managedClusters/read
	// Microsoft.ContainerService/managedClusters/write
	// Microsoft.Insights/actionGroups/read
	// Microsoft.Insights/actionGroups/write
	// Microsoft.Insights/activityLogAlerts/read
	// Microsoft.Insights/activityLogAlerts/write
	// Microsoft.Insights/diagnosticSettings/read
	// Microsoft.Insights/diagnosticSettings/write
	// Microsoft.KeyVault/vaults/read
	// Microsoft.KeyVault/vaults/write
	// Microsoft.ManagedIdentity/userAssignedIdentities/read
	// Microsoft.ManagedIdentity/userAssignedIdentities/write
	// Microsoft.Network/ApplicationGatewayWebApplicationFirewallPolicies/read
	// Microsoft.Network/ApplicationGatewayWebApplicationFirewallPolicies/write
	// Microsoft.Network/applicationGateways/read
	// Microsoft.Network/applicationGateways/write
	// Microsoft.Network/bastionHosts/read
	// Microsoft.Network/bastionHosts/write
	// Microsoft.Network/natGateways/read
	// Microsoft.Network/natGateways/write
	// Microsoft.Network/networkInterfaces/read
	// Microsoft.Network/networkInterfaces/write
	// Microsoft.Network/networkSecurityGroups/read
	// Microsoft.Network/networkSecurityGroups/write
	// Microsoft.Network/privateDnsZones/read
	// Microsoft.Network/privateDnsZones/virtualNetworkLinks/read
	// Microsoft.Network/privateDnsZones/virtualNetworkLinks/write
	// Microsoft.Network/privateDnsZones/write
	// Microsoft.Network/privateEndpoints/privateDnsZoneGroups/read
	// Microsoft.Network/privateEndpoints/privateDnsZoneGroups/write
	// Microsoft.Network/privateEndpoints/read
	// Microsoft.Network/privateEndpoints/write
	// Microsoft.Network/publicIPAddresses/read
	// Microsoft.Network/publicIPAddresses/write
	// Microsoft.Network/publicIPPrefixes/read
	// Microsoft.Network/publicIPPrefixes/write
	// Microsoft.Network/virtualNetworks/read
	// Microsoft.Network/virtualNetworks/write
	// Microsoft.OperationalInsights/workspaces/read
	// Microsoft.OperationalInsights/workspaces/write
	// Microsoft.OperationsManagement/solutions/read
	// Microsoft.OperationsManagement/solutions/write
	// Microsoft.Resources/deployments/read
	// Microsoft.Resources/deployments/write
	// Microsoft.Storage/storageAccounts/read
	// Microsoft.Storage/storageAccounts/write
	assert.NotEmpty(t, mpfResult.RequiredPermissions)
	assert.Equal(t, 54, len(mpfResult.RequiredPermissions[mpfConfig.ResourceGroup.ResourceGroupResourceID]))
}

func TestARMTemplatWhatIfAksPrivateSubnetTemplate(t *testing.T) {

	mpfArgs, err := getTestingMPFArgs()
	if err != nil {
		t.Skip("required environment variables not set, skipping end to end test")
	}
	mpfArgs.TemplateFilePath = "../samples/templates/aks-private-subnet.json"
	mpfArgs.ParametersFilePath = "../samples/templates/aks-private-subnet-parameters.json"

	ctx := context.Background()

	mpfConfig := presentation.GetMPFConfig(mpfArgs)

	deploymentName := fmt.Sprintf("%s-%s", mpfArgs.DeploymentNamePfx, mpfSharedUtils.GenerateRandomString(7))
	armConfig := &ARMTemplateShared.ArmTemplateAdditionalConfig{
		TemplateFilePath:   mpfArgs.TemplateFilePath,
		ParametersFilePath: mpfArgs.ParametersFilePath,
		DeploymentName:     deploymentName,
	}

	// azAPIClient := azureAPI.NewAzureAPIClients(mpfArgs.SubscriptionID)
	var rgManager usecase.ResourceGroupManager
	var spRoleAssignmentManager usecase.ServicePrincipalRolemAssignmentManager
	rgManager = resourceGroupManager.NewResourceGroupManager(mpfArgs.SubscriptionID)
	spRoleAssignmentManager = sproleassignmentmanager.NewSPRoleAssignmentManager(mpfArgs.SubscriptionID)

	var deploymentAuthorizationCheckerCleaner usecase.DeploymentAuthorizationCheckerCleaner
	var mpfService *usecase.MPFService

	deploymentAuthorizationCheckerCleaner = ARMTemplateWhatIf.NewARMTemplateWhatIfAuthorizationChecker(mpfArgs.SubscriptionID, *armConfig)
	initialPermissionsToAdd := []string{"Microsoft.Resources/deployments/*", "Microsoft.Resources/subscriptions/operationresults/read"}
	permissionsToAddToResult := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}
	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, true, false, true)

	mpfResult, err := mpfService.GetMinimumPermissionsRequired()
	if err != nil {
		t.Error(err)
	}

	//check if mpfResult.RequiredPermissions is not empty and has 8	 permissions for scope ResourceGroupResourceID
	// Microsoft.ContainerService/managedClusters/read
	// Microsoft.ContainerService/managedClusters/write
	// Microsoft.Network/virtualNetworks/read
	// Microsoft.Network/virtualNetworks/subnets/read
	// Microsoft.Network/virtualNetworks/subnets/write
	// Microsoft.Network/virtualNetworks/write
	// Microsoft.Resources/deployments/read
	// Microsoft.Resources/deployments/write
	assert.NotEmpty(t, mpfResult.RequiredPermissions)
	assert.Equal(t, 8, len(mpfResult.RequiredPermissions[mpfConfig.ResourceGroup.ResourceGroupResourceID]))
}

func TestARMTemplateFullDeploymentMultiResourceTemplate(t *testing.T) {

	mpfArgs, err := getTestingMPFArgs()
	if err != nil {
		t.Skip("required environment variables not set, skipping end to end test")
	}
	mpfArgs.TemplateFilePath = "../samples/templates/multi-resource-template.json"
	mpfArgs.ParametersFilePath = "../samples/templates/multi-resource-parameters.json"

	ctx := context.Background()

	mpfConfig := presentation.GetMPFConfig(mpfArgs)

	deploymentName := fmt.Sprintf("%s-%s", mpfArgs.DeploymentNamePfx, mpfSharedUtils.GenerateRandomString(7))
	armConfig := &ARMTemplateShared.ArmTemplateAdditionalConfig{
		TemplateFilePath:   mpfArgs.TemplateFilePath,
		ParametersFilePath: mpfArgs.ParametersFilePath,
		DeploymentName:     deploymentName,
	}

	// azAPIClient := azureAPI.NewAzureAPIClients(mpfArgs.SubscriptionID)
	var rgManager usecase.ResourceGroupManager
	var spRoleAssignmentManager usecase.ServicePrincipalRolemAssignmentManager
	rgManager = resourceGroupManager.NewResourceGroupManager(mpfArgs.SubscriptionID)
	spRoleAssignmentManager = sproleassignmentmanager.NewSPRoleAssignmentManager(mpfArgs.SubscriptionID)

	var deploymentAuthorizationCheckerCleaner usecase.DeploymentAuthorizationCheckerCleaner
	var mpfService *usecase.MPFService

	deploymentAuthorizationCheckerCleaner = ARMTemplateDeployment.NewARMTemplateDeploymentAuthorizationChecker(mpfArgs.SubscriptionID, *armConfig)
	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, []string{}, []string{}, true, false, true)

	mpfResult, err := mpfService.GetMinimumPermissionsRequired()
	if err != nil {
		t.Error(err)
	}

	//check if mpfResult.RequiredPermissions is not empty and has 54 permissions for scope
	// Microsoft.Authorization/roleAssignments/read
	// Microsoft.Authorization/roleAssignments/write
	// Microsoft.Compute/virtualMachines/extensions/read
	// Microsoft.Compute/virtualMachines/extensions/write
	// Microsoft.Compute/virtualMachines/read
	// Microsoft.Compute/virtualMachines/write
	// Microsoft.ContainerRegistry/registries/read
	// Microsoft.ContainerRegistry/registries/write
	// Microsoft.ContainerService/managedClusters/read
	// Microsoft.ContainerService/managedClusters/write
	// Microsoft.Insights/actionGroups/read
	// Microsoft.Insights/actionGroups/write
	// Microsoft.Insights/activityLogAlerts/read
	// Microsoft.Insights/activityLogAlerts/write
	// Microsoft.Insights/diagnosticSettings/read
	// Microsoft.Insights/diagnosticSettings/write
	// Microsoft.KeyVault/vaults/read
	// Microsoft.KeyVault/vaults/write
	// Microsoft.ManagedIdentity/userAssignedIdentities/read
	// Microsoft.ManagedIdentity/userAssignedIdentities/write
	// Microsoft.Network/ApplicationGatewayWebApplicationFirewallPolicies/read
	// Microsoft.Network/ApplicationGatewayWebApplicationFirewallPolicies/write
	// Microsoft.Network/applicationGateways/read
	// Microsoft.Network/applicationGateways/write
	// Microsoft.Network/bastionHosts/read
	// Microsoft.Network/bastionHosts/write
	// Microsoft.Network/natGateways/read
	// Microsoft.Network/natGateways/write
	// Microsoft.Network/networkInterfaces/read
	// Microsoft.Network/networkInterfaces/write
	// Microsoft.Network/networkSecurityGroups/read
	// Microsoft.Network/networkSecurityGroups/write
	// Microsoft.Network/privateDnsZones/read
	// Microsoft.Network/privateDnsZones/virtualNetworkLinks/read
	// Microsoft.Network/privateDnsZones/virtualNetworkLinks/write
	// Microsoft.Network/privateDnsZones/write
	// Microsoft.Network/privateEndpoints/privateDnsZoneGroups/read
	// Microsoft.Network/privateEndpoints/privateDnsZoneGroups/write
	// Microsoft.Network/privateEndpoints/read
	// Microsoft.Network/privateEndpoints/write
	// Microsoft.Network/publicIPAddresses/read
	// Microsoft.Network/publicIPAddresses/write
	// Microsoft.Network/publicIPPrefixes/read
	// Microsoft.Network/publicIPPrefixes/write
	// Microsoft.Network/virtualNetworks/read
	// Microsoft.Network/virtualNetworks/write
	// Microsoft.OperationalInsights/workspaces/read
	// Microsoft.OperationalInsights/workspaces/write
	// Microsoft.OperationsManagement/solutions/read
	// Microsoft.OperationsManagement/solutions/write
	// Microsoft.Resources/deployments/read
	// Microsoft.Resources/deployments/write
	// Microsoft.Storage/storageAccounts/read
	// Microsoft.Storage/storageAccounts/write
	assert.NotEmpty(t, mpfResult.RequiredPermissions)
	assert.Equal(t, 54, len(mpfResult.RequiredPermissions[mpfConfig.ResourceGroup.ResourceGroupResourceID]))
}

func TestARMTemplateFullDeploymentAksPrivateSubnetTemplate(t *testing.T) {

	mpfArgs, err := getTestingMPFArgs()
	if err != nil {
		t.Skip("required environment variables not set, skipping end to end test")
	}
	mpfArgs.TemplateFilePath = "../samples/templates/aks-private-subnet.json"
	mpfArgs.ParametersFilePath = "../samples/templates/aks-private-subnet-parameters.json"

	ctx := context.Background()

	mpfConfig := presentation.GetMPFConfig(mpfArgs)

	deploymentName := fmt.Sprintf("%s-%s", mpfArgs.DeploymentNamePfx, mpfSharedUtils.GenerateRandomString(7))
	armConfig := &ARMTemplateShared.ArmTemplateAdditionalConfig{
		TemplateFilePath:   mpfArgs.TemplateFilePath,
		ParametersFilePath: mpfArgs.ParametersFilePath,
		DeploymentName:     deploymentName,
	}

	// azAPIClient := azureAPI.NewAzureAPIClients(mpfArgs.SubscriptionID)
	var rgManager usecase.ResourceGroupManager
	var spRoleAssignmentManager usecase.ServicePrincipalRolemAssignmentManager
	rgManager = resourceGroupManager.NewResourceGroupManager(mpfArgs.SubscriptionID)
	spRoleAssignmentManager = sproleassignmentmanager.NewSPRoleAssignmentManager(mpfArgs.SubscriptionID)

	var deploymentAuthorizationCheckerCleaner usecase.DeploymentAuthorizationCheckerCleaner
	var mpfService *usecase.MPFService

	deploymentAuthorizationCheckerCleaner = ARMTemplateDeployment.NewARMTemplateDeploymentAuthorizationChecker(mpfArgs.SubscriptionID, *armConfig)
	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, []string{}, []string{}, true, false, true)

	mpfResult, err := mpfService.GetMinimumPermissionsRequired()
	if err != nil {
		t.Error(err)
	}

	//check if mpfResult.RequiredPermissions is not empty and has 8
	// Microsoft.ContainerService/managedClusters/read
	// Microsoft.ContainerService/managedClusters/write
	// Microsoft.Network/virtualNetworks/read
	// Microsoft.Network/virtualNetworks/subnets/read
	// Microsoft.Network/virtualNetworks/subnets/write
	// Microsoft.Network/virtualNetworks/write
	// Microsoft.Resources/deployments/read
	// Microsoft.Resources/deployments/write
	assert.NotEmpty(t, mpfResult.RequiredPermissions)
	assert.Equal(t, 8, len(mpfResult.RequiredPermissions[mpfConfig.ResourceGroup.ResourceGroupResourceID]))
}

func TestTerraformACI(t *testing.T) {

	mpfArgs, err := getTestingMPFArgs()
	if err != nil {
		t.Skip("required environment variables not set, skipping end to end test")
	}
	mpfArgs.MPFMode = "terraform"

	var tfpath string
	if os.Getenv("MPF_TFPATH") == "" {
		t.Skip("Terraform Path TF_PATH not set, skipping end to end test")
	}
	tfpath = os.Getenv("MPF_TFPATH")

	_, filename, _, _ := runtime.Caller(0)
	curDir := path.Dir(filename)
	wrkDir := path.Join(curDir, "../samples/terraform/aci")
	varsFile := path.Join(curDir, "../samples/terraform/aci/dev.vars.tfvars")

	ctx := context.Background()

	mpfConfig := presentation.GetMPFConfig(mpfArgs)

	// azAPIClient := azureAPI.NewAzureAPIClients(mpfArgs.SubscriptionID)
	var rgManager usecase.ResourceGroupManager
	var spRoleAssignmentManager usecase.ServicePrincipalRolemAssignmentManager
	rgManager = resourceGroupManager.NewResourceGroupManager(mpfArgs.SubscriptionID)
	spRoleAssignmentManager = sproleassignmentmanager.NewSPRoleAssignmentManager(mpfArgs.SubscriptionID)

	var deploymentAuthorizationCheckerCleaner usecase.DeploymentAuthorizationCheckerCleaner
	var mpfService *usecase.MPFService

	initialPermissionsToAdd := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}
	permissionsToAddToResult := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}
	deploymentAuthorizationCheckerCleaner = terraform.NewTerraformAuthorizationChecker(wrkDir, tfpath, varsFile)
	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, false, true, false)

	mpfResult, err := mpfService.GetMinimumPermissionsRequired()
	if err != nil {
		t.Error(err)
	}

	//check if mpfResult.RequiredPermissions is not empty and has 54 permissions for scope ResourceGroupResourceID
	assert.NotEmpty(t, mpfResult.RequiredPermissions)
	assert.Equal(t, 8, len(mpfResult.RequiredPermissions[mpfConfig.ResourceGroup.ResourceGroupResourceID]))
}

//
// Below test can take more than 10 mins to complete, hence commented
//

// func TestTerraformMultiResource(t *testing.T) {

// 	mpfArgs, err := getTestingMPFArgs()
// 	if err != nil {
// 		t.Skip("required environment variables not set, skipping end to end test")
// 	}
// 	mpfArgs.MPFMode = "terraform"

// 	var tfpath string
// 	if os.Getenv("MPF_TFPATH") == "" {
// 		t.Skip("Terraform Path TF_PATH not set, skipping end to end test")
// 	}
// 	tfpath = os.Getenv("MPF_TFPATH")

// 	_, filename, _, _ := runtime.Caller(0)
// 	curDir := path.Dir(filename)
// 	wrkDir := path.Join(curDir, "../samples/terraform/multi-resource")
// 	varsFile := path.Join(curDir, "../samples/terraform/multi-resource/dev.vars.tfvars")

// 	ctx := context.Background()

// 	mpfConfig := presentation.GetMPFConfig(mpfArgs)

// 	// azAPIClient := azureAPI.NewAzureAPIClients(mpfArgs.SubscriptionID)
// 	var rgManager usecase.ResourceGroupManager
// 	var spRoleAssignmentManager usecase.ServicePrincipalRolemAssignmentManager
// 	rgManager = resourceGroupManager.NewResourceGroupManager(mpfArgs.SubscriptionID)
// 	spRoleAssignmentManager = sproleassignmentmanager.NewSPRoleAssignmentManager(mpfArgs.SubscriptionID)

// 	var deploymentAuthorizationCheckerCleaner usecase.DeploymentAuthorizationCheckerCleaner
// 	var mpfService *usecase.MPFService

// 	initialPermissionsToAdd := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}
// 	permissionsToAddToResult := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}
// 	deploymentAuthorizationCheckerCleaner = terraform.NewTerraformAuthorizationChecker(wrkDir, tfpath, varsFile)
// 	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, false, true, false)

// 	mpfResult, err := mpfService.GetMinimumPermissionsRequired()
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	// Microsoft.ContainerService/managedClusters/delete
// 	// Microsoft.ContainerService/managedClusters/listClusterUserCredential/action
// 	// Microsoft.ContainerService/managedClusters/read
// 	// Microsoft.ContainerService/managedClusters/write
// 	// Microsoft.Network/virtualNetworks/delete
// 	// Microsoft.Network/virtualNetworks/read
// 	// Microsoft.Network/virtualNetworks/subnets/delete
// 	// Microsoft.Network/virtualNetworks/subnets/join/action
// 	// Microsoft.Network/virtualNetworks/subnets/read
// 	// Microsoft.Network/virtualNetworks/subnets/write
// 	// Microsoft.Network/virtualNetworks/write
// 	// Microsoft.Resources/deployments/read
// 	// Microsoft.Resources/deployments/write
// 	// Microsoft.Resources/subscriptions/resourcegroups/delete
// 	// Microsoft.Resources/subscriptions/resourcegroups/read
// 	// Microsoft.Resources/subscriptions/resourcegroups/write
// 	//check if mpfResult.RequiredPermissions is not empty and has 54 permissions for scope ResourceGroupResourceID
// 	assert.NotEmpty(t, mpfResult.RequiredPermissions)
// 	assert.Equal(t, 16, len(mpfResult.RequiredPermissions[mpfConfig.ResourceGroup.ResourceGroupResourceID]))
// }
