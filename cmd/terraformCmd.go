/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/manisbindra/az-mpf/pkg/infrastructure/authorizationCheckers/terraform"
	resourceGroupManager "github.com/manisbindra/az-mpf/pkg/infrastructure/resourceGroupManager"
	sproleassignmentmanager "github.com/manisbindra/az-mpf/pkg/infrastructure/spRoleAssignmentManager"
	"github.com/manisbindra/az-mpf/pkg/usecase"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	flgTFPath                         string
	flgWorkingDir                     string
	flgVarFilePath                    string
	flgImportExistingResourcesToState bool
	flgTargetModule                   string
)

const FoundPermissionsFromFailedRunFilename = ".permissionsFromFailedRun.json"

// terraformCmd represents the terraform command
// var

func NewTerraformCommand() *cobra.Command {
	terraformCmd := &cobra.Command{
		Use:   "terraform",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
	and usage of using your command. For example:
	
	Cobra is a CLI library for Go that empowers applications.
	This application is a tool to generate the needed files
	to quickly create a Cobra application.`,
		Run: getMPFTerraform,
	}

	terraformCmd.Flags().StringVarP(&flgTFPath, "tfPath", "", "", "Path to Terraform Executable")
	terraformCmd.MarkFlagRequired("tfPath")

	terraformCmd.Flags().StringVarP(&flgWorkingDir, "workingDir", "", "", "Path to Terraform Working Directory")
	terraformCmd.MarkFlagRequired("workingDir")

	terraformCmd.Flags().StringVarP(&flgVarFilePath, "varFilePath", "", "", "Path to Terraform Variable File")
	terraformCmd.Flags().BoolVarP(&flgImportExistingResourcesToState, "importExistingResourcesToState", "", true, "On existing resource error, import existing resources into to Terraform State. This will also destroy the imported resources before MPF execution completes.")

	terraformCmd.Flags().StringVarP(&flgTargetModule, "targetModule", "", "", "The Terraform module to Target Module to run MPF on")

	return terraformCmd
}

func getMPFTerraform(cmd *cobra.Command, args []string) {
	setLogLevel()

	log.Info("Executin MPF for Terraform")
	log.Infof("TFPath: %s\n", flgTFPath)
	log.Infof("WorkingDir: %s\n", flgWorkingDir)
	log.Infof("VarFilePath: %s\n", flgVarFilePath)
	log.Infof("ImportExistingResourcesToState: %t\n", flgImportExistingResourcesToState)

	// validate if working directory exists
	if _, err := os.Stat(flgWorkingDir); os.IsNotExist(err) {
		log.Fatalf("Working Directory does not exist: %s\n", flgWorkingDir)
	}

	flgWorkingDir, err := getAbsolutePath(flgWorkingDir)
	if err != nil {
		log.Errorf("Error getting absolute path for terraform working directory: %v\n", err)
	}

	// validate if tfPath exists
	if _, err := os.Stat(flgTFPath); os.IsNotExist(err) {
		log.Fatalf("Terraform Executable does not exist: %s\n", flgTFPath)
	}

	flgTFPath, err := getAbsolutePath(flgTFPath)
	if err != nil {
		log.Errorf("Error getting absolute path for terraform executable: %v\n", err)
	}

	if flgVarFilePath != "" {

		if _, err := os.Stat(flgVarFilePath); os.IsNotExist(err) {
			log.Fatalf("Terraform Variable File does not exist: %s\n", flgVarFilePath)
		}

		flgVarFilePath, err = getAbsolutePath(flgVarFilePath)
		if err != nil {
			log.Errorf("Error getting absolute path for terraform variable file: %v\n", err)
		}

	}

	ctx := context.Background()

	mpfConfig := getRootMPFConfig()

	var rgManager usecase.ResourceGroupManager
	var spRoleAssignmentManager usecase.ServicePrincipalRolemAssignmentManager
	rgManager = resourceGroupManager.NewResourceGroupManager(flgSubscriptionID)
	spRoleAssignmentManager = sproleassignmentmanager.NewSPRoleAssignmentManager(flgSubscriptionID)

	var deploymentAuthorizationCheckerCleaner usecase.DeploymentAuthorizationCheckerCleaner
	var mpfService *usecase.MPFService

	initialPermissionsToAdd := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}
	permissionsToAddToResult := []string{"Microsoft.Resources/deployments/read", "Microsoft.Resources/deployments/write"}

	// Check if permissions file from previous failed run exists
	if terraform.DoesTFFileExist(flgWorkingDir, FoundPermissionsFromFailedRunFilename) {
		prevResult, err := terraform.LoadMPFResultFromFile(flgWorkingDir, FoundPermissionsFromFailedRunFilename)
		if err != nil {
			log.Warnf("Error loading permissions from previous failed run: %v\n, continuing....", err)
		}
		prevRunFoundPermissions := prevResult.RequiredPermissions[""]
		if len(prevRunFoundPermissions) > 0 {
			log.Warnf("Found permissions from previous failed run: %v\n Adding the Permissions....", prevRunFoundPermissions)
			initialPermissionsToAdd = append(initialPermissionsToAdd, prevRunFoundPermissions...)
			permissionsToAddToResult = append(permissionsToAddToResult, prevRunFoundPermissions...)
		}
	}

	deploymentAuthorizationCheckerCleaner = terraform.NewTerraformAuthorizationChecker(flgWorkingDir, flgTFPath, flgVarFilePath, flgImportExistingResourcesToState, flgTargetModule)
	mpfService = usecase.NewMPFService(ctx, rgManager, spRoleAssignmentManager, deploymentAuthorizationCheckerCleaner, mpfConfig, initialPermissionsToAdd, permissionsToAddToResult, false, true, false)

	displayOptions := getDislayOptions(flgShowDetailedOutput, flgJSONOutput, mpfConfig.ResourceGroup.ResourceGroupResourceID)

	mpfResult, err := mpfService.GetMinimumPermissionsRequired()
	if err != nil {
		if len(mpfResult.RequiredPermissions) > 0 {
			fmt.Println("Error occurred while getting minimum permissions required. However, some permissions were identified prior to the error.")
			_ = terraform.SaveMPFResultsToFile(flgWorkingDir, FoundPermissionsFromFailedRunFilename, mpfResult)

			displayResult(mpfResult, displayOptions)
		}
		log.Fatal(err)
	}

	if terraform.DoesTFFileExist(flgWorkingDir, FoundPermissionsFromFailedRunFilename) {
		_ = terraform.DeleteTFFile(flgWorkingDir, FoundPermissionsFromFailedRunFilename)
	}

	displayResult(mpfResult, displayOptions)

}
