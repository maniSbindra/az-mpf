/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/manisbindra/az-mpf/pkg/domain"
	"github.com/manisbindra/az-mpf/pkg/infrastructure/ARMTemplateShared"
	"github.com/manisbindra/az-mpf/pkg/infrastructure/authorizationCheckers/ARMTemplateWhatIf"
	"github.com/manisbindra/az-mpf/pkg/infrastructure/mpfSharedUtils"
	resourceGroupManager "github.com/manisbindra/az-mpf/pkg/infrastructure/resourceGroupManager"
	sproleassignmentmanager "github.com/manisbindra/az-mpf/pkg/infrastructure/spRoleAssignmentManager"
	"github.com/manisbindra/az-mpf/pkg/presentation"
	"github.com/manisbindra/az-mpf/pkg/usecase"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var flgBicepFilePath string
var flgBicepExecPath string

// armCmd represents the arm command

func NewBicepCommand() *cobra.Command {

	bicepCmd := &cobra.Command{
		Use:   "bicep",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
	and usage of using your command. For example:
	
	Cobra is a CLI library for Go that empowers applications.
	This application is a tool to generate the needed files
	to quickly create a Cobra application.`,
		Run: getMPFBicep,
	}

	bicepCmd.Flags().StringVarP(&flgResourceGroupNamePfx, "resourceGroupNamePfx", "", "testdeployrg", "Resource Group Name Prefix")
	bicepCmd.Flags().StringVarP(&flgDeploymentNamePfx, "deploymentNamePfx", "", "testDeploy", "Deployment Name Prefix")

	bicepCmd.Flags().StringVarP(&flgBicepFilePath, "bicepFilePath", "", "", "Path to bicep File")
	bicepCmd.MarkFlagRequired("bicepFilePath")

	bicepCmd.Flags().StringVarP(&flgParametersFilePath, "parametersFilePath", "", "", "Path to bicep Parameters File")
	bicepCmd.MarkFlagRequired("parametersFilePath")

	bicepCmd.Flags().StringVarP(&flgBicepExecPath, "bicepExecPath", "", "", "Bicep Executable Path")
	bicepCmd.MarkFlagRequired("bicepExecPath")

	bicepCmd.Flags().StringVarP(&flgLocation, "location", "", "eastus", "Location")

	// bicepCmd.Flags().BoolVarP(&flgFullDeployment, "fullDeployment", "", false, "Full Deployment")

	return bicepCmd
}

func getMPFBicep(cmd *cobra.Command, args []string) {
	setLogLevel()

	log.Info("Executing MPF for Bicep")

	log.Debugf("ResourceGroupNamePfx: %s\n", flgResourceGroupNamePfx)
	log.Debugf("DeploymentNamePfx: %s\n", flgDeploymentNamePfx)
	log.Infof("BicepFilePath: %s\n", flgBicepFilePath)
	log.Infof("ParametersFilePath: %s\n", flgParametersFilePath)
	log.Infof("BicepExecPath: %s\n", flgBicepExecPath)

	// validate if template and parameters files exists
	if _, err := os.Stat(flgBicepFilePath); os.IsNotExist(err) {
		log.Fatal("Bicep File does not exist")
	}

	if _, err := os.Stat(flgBicepExecPath); os.IsNotExist(err) {
		log.Fatal("Bicep Executable does not exist")
	}

	if _, err := os.Stat(flgParametersFilePath); os.IsNotExist(err) {
		log.Fatal("Parameters File does not exist")
	}

	flgBicepExecPath, err := getAbsolutePath(flgBicepExecPath)
	if err != nil {
		log.Errorf("Error getting absolute path for bicep executable: %v\n", err)
	}

	flgBicepFilePath, err := getAbsolutePath(flgBicepFilePath)
	if err != nil {
		log.Errorf("Error getting absolute path for bicep file: %v\n", err)
	}

	flgParametersFilePath, err := getAbsolutePath(flgParametersFilePath)
	if err != nil {
		log.Errorf("Error getting absolute path for parameters file: %v\n", err)
	}

	armTemplatePath := strings.TrimSuffix(flgBicepFilePath, ".bicep") + ".json"
	bicepCmd := exec.Command(flgBicepExecPath, "build", flgBicepFilePath, "--outfile", armTemplatePath)
	bicepCmd.Dir = filepath.Dir(flgBicepFilePath)

	_, err = bicepCmd.CombinedOutput()
	if err != nil {
		log.Errorf("error running bicep build: %s", err)
	}
	log.Infoln("Bicep build successful, ARM Template created at:", armTemplatePath)

	ctx := context.Background()

	mpfConfig := getRootMPFConfig()
	mpfRG := domain.ResourceGroup{}
	mpfRG.ResourceGroupName = fmt.Sprintf("%s-%s", flgResourceGroupNamePfx, mpfSharedUtils.GenerateRandomString(7))
	mpfRG.ResourceGroupResourceID = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", flgSubscriptionID, mpfRG.ResourceGroupName)
	mpfRG.Location = flgLocation
	mpfConfig.ResourceGroup = mpfRG
	deploymentName := fmt.Sprintf("%s-%s", flgDeploymentNamePfx, mpfSharedUtils.GenerateRandomString(7))
	armConfig := &ARMTemplateShared.ArmTemplateAdditionalConfig{
		TemplateFilePath:   armTemplatePath,
		ParametersFilePath: flgParametersFilePath,
		DeploymentName:     deploymentName,
	}

	var rgManager usecase.ResourceGroupManager
	var spRoleAssignmentManager usecase.ServicePrincipalRolemAssignmentManager
	rgManager = resourceGroupManager.NewResourceGroupManager(flgSubscriptionID)
	spRoleAssignmentManager = sproleassignmentmanager.NewSPRoleAssignmentManager(flgSubscriptionID)

	var deploymentAuthorizationCheckerCleaner usecase.DeploymentAuthorizationCheckerCleaner
	var mpfService *usecase.MPFService
	var initialPermissionsToAdd []string
	var permissionsToAddToResult []string

	deploymentAuthorizationCheckerCleaner = ARMTemplateWhatIf.NewARMTemplateWhatIfAuthorizationChecker(flgSubscriptionID, *armConfig)
	initialPermissionsToAdd = []string{"Microsoft.Resources/deployments/*", "Microsoft.Resources/subscriptions/operationresults/read"}
	permissionsToAddToResult = []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}

	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, true, false, true)

	mpfResult, err := mpfService.GetMinimumPermissionsRequired()
	if err != nil {
		log.Fatal(err)
	}

	log.Infoln("Deleting Generated ARM Template file...")
	// delete generated ARM template file
	err = os.Remove(armTemplatePath)
	if err != nil {
		log.Errorf("Error deleting Generated ARM template file: %v\n", err)
	}

	displayOptions := presentation.DisplayOptions{
		ShowDetailedOutput:             flgShowDetailedOutput,
		JSONOutput:                     flgJSONOutput,
		DefaultResourceGroupResourceID: mpfConfig.ResourceGroup.ResourceGroupResourceID,
	}

	resultDisplayer := presentation.NewMPFResultDisplayer(mpfResult, displayOptions)
	err = resultDisplayer.DisplayResult(os.Stdout)
	if err != nil {
		log.Fatal(err)
	}

}
