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
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestTerraformWithImport(t *testing.T) {

	// import errors can occur for some resources, when identity does not have all required permissions,
	// as described in https://github.com/hashicorp/terraform-provider-azurerm/issues/27961#issuecomment-2467392936
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
	log.Infof("curDir: %s", curDir)
	wrkDir := path.Join(curDir, "../samples/terraform/existing-resource-import")
	log.Infof("wrkDir: %s", wrkDir)
	ctx := context.Background()

	mpfConfig := getMPFConfig(mpfArgs)

	var rgManager usecase.ResourceGroupManager
	var spRoleAssignmentManager usecase.ServicePrincipalRolemAssignmentManager
	rgManager = resourceGroupManager.NewResourceGroupManager(mpfArgs.SubscriptionID)
	spRoleAssignmentManager = sproleassignmentmanager.NewSPRoleAssignmentManager(mpfArgs.SubscriptionID)

	var deploymentAuthorizationCheckerCleaner usecase.DeploymentAuthorizationCheckerCleaner
	var mpfService *usecase.MPFService

	initialPermissionsToAdd := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}
	permissionsToAddToResult := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}
	deploymentAuthorizationCheckerCleaner = terraform.NewTerraformAuthorizationChecker(wrkDir, tfpath, "", true, "")
	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, false, true, false)

	mpfResult, err := mpfService.GetMinimumPermissionsRequired()
	if err != nil {
		t.Error(err)
	}

	assert.NotEmpty(t, mpfResult.RequiredPermissions)
	assert.Equal(t, 13, len(mpfResult.RequiredPermissions[mpfConfig.ResourceGroup.ResourceGroupResourceID]))
}

func TestTerraformWithTargetting(t *testing.T) {

	// import errors can occur for some resources, when identity does not have all required permissions,
	// as described in https://github.com/hashicorp/terraform-provider-azurerm/issues/27961#issuecomment-2467392936
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
	log.Infof("curDir: %s", curDir)
	wrkDir := path.Join(curDir, "../samples/terraform/module-test-with-targetting")
	log.Infof("wrkDir: %s", wrkDir)
	ctx := context.Background()

	mpfConfig := getMPFConfig(mpfArgs)

	var rgManager usecase.ResourceGroupManager
	var spRoleAssignmentManager usecase.ServicePrincipalRolemAssignmentManager
	rgManager = resourceGroupManager.NewResourceGroupManager(mpfArgs.SubscriptionID)
	spRoleAssignmentManager = sproleassignmentmanager.NewSPRoleAssignmentManager(mpfArgs.SubscriptionID)

	var deploymentAuthorizationCheckerCleaner usecase.DeploymentAuthorizationCheckerCleaner
	var mpfService *usecase.MPFService

	initialPermissionsToAdd := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}
	permissionsToAddToResult := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}
	deploymentAuthorizationCheckerCleaner = terraform.NewTerraformAuthorizationChecker(wrkDir, tfpath, "", true, "module.law")
	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, false, true, false)

	mpfResult, err := mpfService.GetMinimumPermissionsRequired()
	if err != nil {
		t.Error(err)
	}

	assert.NotEmpty(t, mpfResult.RequiredPermissions)
	assert.Equal(t, 8, len(mpfResult.RequiredPermissions[mpfConfig.ResourceGroup.ResourceGroupResourceID]))
}
