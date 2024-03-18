package e2etests

import (
	"context"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/manisbindra/az-mpf/pkg/infrastructure/authorizationCheckers/terraform"
	resourceGroupManager "github.com/manisbindra/az-mpf/pkg/infrastructure/resourceGroupManager"
	sproleassignmentmanager "github.com/manisbindra/az-mpf/pkg/infrastructure/spRoleAssignmentManager"
	"github.com/manisbindra/az-mpf/pkg/usecase"
	"github.com/stretchr/testify/assert"
)

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

	mpfConfig := getMPFConfig(mpfArgs)

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

// 	mpfConfig := presentation.getMPFConfig(mpfArgs)

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
