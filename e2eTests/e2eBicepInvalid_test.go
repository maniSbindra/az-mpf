package e2etests

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/manisbindra/az-mpf/pkg/infrastructure/ARMTemplateShared"
	"github.com/manisbindra/az-mpf/pkg/infrastructure/authorizationCheckers/ARMTemplateWhatIf"
	mpfSharedUtils "github.com/manisbindra/az-mpf/pkg/infrastructure/mpfSharedUtils"
	resourceGroupManager "github.com/manisbindra/az-mpf/pkg/infrastructure/resourceGroupManager"
	sproleassignmentmanager "github.com/manisbindra/az-mpf/pkg/infrastructure/spRoleAssignmentManager"
	"github.com/manisbindra/az-mpf/pkg/usecase"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestBicepInvalidParams(t *testing.T) {

	mpfArgs, err := getTestingMPFArgs()
	if err != nil {
		t.Skip("required environment variables not set, skipping end to end test")
	}

	if checkBicepTestEnvVars() {
		t.Skip("required environment variables not set, skipping end to end test")
	}

	bicepExecPath := os.Getenv("MPF_BICEPEXECPATH")
	bicepFilePath := "../samples/bicep/aks-private-subnet.bicep"
	parametersFilePath := "../samples/bicep/aks-invalid-params.json"

	bicepFilePath, _ = getAbsolutePath(bicepFilePath)
	parametersFilePath, _ = getAbsolutePath(parametersFilePath)

	armTemplatePath := strings.TrimSuffix(bicepFilePath, ".bicep") + ".json"
	bicepCmd := exec.Command(bicepExecPath, "build", bicepFilePath, "--outfile", armTemplatePath)
	bicepCmd.Dir = filepath.Dir(bicepFilePath)

	_, err = bicepCmd.CombinedOutput()
	if err != nil {
		log.Error(err)
		t.Error(err)
	}
	// defer os.Remove(armTemplatePath)

	ctx := context.Background()

	mpfConfig := getMPFConfig(mpfArgs)

	deploymentName := fmt.Sprintf("%s-%s", mpfArgs.DeploymentNamePfx, mpfSharedUtils.GenerateRandomString(7))
	armConfig := &ARMTemplateShared.ArmTemplateAdditionalConfig{
		TemplateFilePath:   armTemplatePath,
		ParametersFilePath: parametersFilePath,
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

	_, err = mpfService.GetMinimumPermissionsRequired()
	assert.Error(t, err)

	if !errors.Is(err, ARMTemplateShared.ErrInvalidTemplate) {
		t.Errorf("Error is not of type InvalidTemplate")
	}
}

func TestBicepInvalidResourceFile(t *testing.T) {

	_, err := getTestingMPFArgs()
	if err != nil {
		t.Skip("required environment variables not set, skipping end to end test")
	}

	if checkBicepTestEnvVars() {
		t.Skip("required environment variables not set, skipping end to end test")
	}

	bicepExecPath := os.Getenv("MPF_BICEPEXECPATH")
	bicepFilePath := "../samples/bicep/invalid-bicep.bicep"

	bicepFilePath, err = getAbsolutePath(bicepFilePath)
	if err != nil {
		t.Error(err)
	}

	armTemplatePath := strings.TrimSuffix(bicepFilePath, ".bicep") + ".json"
	bicepCmd := exec.Command(bicepExecPath, "build", bicepFilePath, "--outfile", armTemplatePath)
	bicepCmd.Dir = filepath.Dir(bicepFilePath)

	_, err = bicepCmd.CombinedOutput()
	if err == nil {
		// log.Error(err)
		t.Error("expected error, got nil")
	}
	// defer os.Remove(armTemplatePath)

}
