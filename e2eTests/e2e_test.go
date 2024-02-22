package e2etests

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/manisbindra/az-mpf/pkg/infrastructure/ARMTemplateShared"
	"github.com/manisbindra/az-mpf/pkg/infrastructure/authorizationCheckers/ARMTemplateDeployment"
	"github.com/manisbindra/az-mpf/pkg/infrastructure/authorizationCheckers/ARMTemplateWhatIf"
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
	mpfArgs.TemplateFilePath = "../templates/samples/multi-resource-template.json"
	mpfArgs.ParametersFilePath = "../templates/samples/multi-resource-parameters.json"

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
	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult)

	mpfResult, err := mpfService.GetMinimumPermissionsRequired()
	if err != nil {
		t.Error(err)
	}

	//check if mpfResult.RequiredPermissions is not empty and has 54 permissions for scope ResourceGroupResourceID
	assert.NotEmpty(t, mpfResult.RequiredPermissions)
	assert.Equal(t, 54, len(mpfResult.RequiredPermissions[mpfConfig.ResourceGroup.ResourceGroupResourceID]))
}

func TestARMTemplateFullDeploymentMultiResourceTemplate(t *testing.T) {

	mpfArgs, err := getTestingMPFArgs()
	if err != nil {
		t.Skip("required environment variables not set, skipping end to end test")
	}
	mpfArgs.TemplateFilePath = "../templates/samples/multi-resource-template.json"
	mpfArgs.ParametersFilePath = "../templates/samples/multi-resource-parameters.json"

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
	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, []string{}, []string{})

	mpfResult, err := mpfService.GetMinimumPermissionsRequired()
	if err != nil {
		t.Error(err)
	}

	//check if mpfResult.RequiredPermissions is not empty and has 54 permissions for scope ResourceGroupResourceID
	assert.NotEmpty(t, mpfResult.RequiredPermissions)
	assert.Equal(t, 54, len(mpfResult.RequiredPermissions[mpfConfig.ResourceGroup.ResourceGroupResourceID]))
}
