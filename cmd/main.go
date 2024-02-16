// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	ARMTemplateShared "github.com/manisbindra/az-mpf/pkg/infrastructure/ARMTemplateShared"
	"github.com/manisbindra/az-mpf/pkg/infrastructure/authorizationCheckers/ARMTemplateDeployment"
	"github.com/manisbindra/az-mpf/pkg/infrastructure/authorizationCheckers/ARMTemplateWhatIf"
	mpfSharedUtils "github.com/manisbindra/az-mpf/pkg/infrastructure/mpfSharedUtils"
	resourceGroupManager "github.com/manisbindra/az-mpf/pkg/infrastructure/resourceGroupManager"
	sproleassignmentmanager "github.com/manisbindra/az-mpf/pkg/infrastructure/spRoleAssignmentManager"
	"github.com/manisbindra/az-mpf/pkg/presentation"
	"github.com/manisbindra/az-mpf/pkg/usecase"
)

func main() {

	logLevel, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		logLevel = log.ErrorLevel
	}
	log.SetLevel(logLevel)

	ctx := context.Background()

	mpfArgs, err := presentation.GetCLIArgs()
	if err != nil {
		log.Fatal(err)
	}

	mpfConfig := presentation.GetMPFConfig(mpfArgs)

	deploymentName := fmt.Sprintf("%s-%s", mpfArgs.DeploymentNamePfx, mpfSharedUtils.GenerateRandomString(7))
	armConfig := &ARMTemplateShared.ArmTemplateAdditionalConfig{
		TemplateFilePath:   mpfArgs.TemplateFilePath,
		ParametersFilePath: mpfArgs.ParametersFilePath,
		DeplomentName:      deploymentName,
	}

	// Create Azure API Clients
	// azAPIClient := azureAPI.NewAzureAPIClients(mpfArgs.SubscriptionID)

	var rgManager usecase.ResourceGroupManager
	var spRoleAssignmentManager usecase.ServicePrincipalRolemAssignmentManager
	rgManager = resourceGroupManager.NewResourceGroupManager(mpfArgs.SubscriptionID)
	spRoleAssignmentManager = sproleassignmentmanager.NewSPRoleAssignmentManager(mpfArgs.SubscriptionID)

	var deploymentAuthorizationCheckerCleaner usecase.DeploymentAuthorizationCheckerCleaner
	var mpfService *usecase.MPFService

	switch mpfArgs.MPFMode {
	case "whatif":
		deploymentAuthorizationCheckerCleaner = ARMTemplateWhatIf.NewARMTemplateWhatIfAuthorizationChecker(mpfArgs.SubscriptionID, *armConfig)
		initialPermissionsToAdd := []string{"Microsoft.Resources/deployments/*", "Microsoft.Resources/subscriptions/operationresults/read"}
		permissionsToAddToResult := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}
		mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult)
	case "fullDeployment":
		deploymentAuthorizationCheckerCleaner = ARMTemplateDeployment.NewARMTemplateDeploymentAuthorizationChecker(mpfArgs.SubscriptionID, *armConfig)
		mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, []string{}, []string{})
	default:
		deploymentAuthorizationCheckerCleaner = ARMTemplateDeployment.NewARMTemplateDeploymentAuthorizationChecker(mpfArgs.SubscriptionID, *armConfig)
		mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, []string{}, []string{})
	}

	mpfResult, err := mpfService.GetMinimumPermissionsRequired()
	if err != nil {
		log.Fatal(err)
	}

	displayOptions := presentation.DisplayOptions{
		ShowDetailedOutput:             mpfArgs.ShowDetailedOutput,
		JSONOutput:                     mpfArgs.JSONOutput,
		DefaultResourceGroupResourceID: mpfConfig.ResourceGroup.ResourceGroupResourceID,
	}

	resultDisplayer := presentation.NewMPFResultDisplayer(mpfResult, displayOptions)
	err = resultDisplayer.DisplayResult(os.Stdout)
	if err != nil {
		log.Fatal(err)
	}

}

// Function to generate a random string of a given length
