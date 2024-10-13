package terraform

import (
	"context"
	"os"
	"strings"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/manisbindra/az-mpf/pkg/domain"
	log "github.com/sirupsen/logrus"
)

type terraformDeploymentConfig struct {
	ctx         context.Context
	workingDir  string
	execPath    string
	varFilePath string
}

// This file is created once the destroy phase is entered
const TFDestroyStateEnteredFileName = "azmpfEnteredDestroyPhase.txt"

func NewTerraformAuthorizationChecker(workDir string, execPath string, varFilePath string) *terraformDeploymentConfig {
	err := deleteEnteredDestroyPhaseStateFile(workDir, TFDestroyStateEnteredFileName)
	if err != nil {
		log.Warnf("error deleting enteredDestroyPhaseStateFile: %s", err)
	}

	return &terraformDeploymentConfig{
		workingDir:  workDir,
		execPath:    execPath,
		ctx:         context.Background(),
		varFilePath: varFilePath,
	}
}

func (a *terraformDeploymentConfig) GetDeploymentAuthorizationErrors(mpfConfig domain.MPFConfig) (string, error) {
	return a.deployTerraform(mpfConfig)
}

func (a *terraformDeploymentConfig) CleanDeployment(mpfConfig domain.MPFConfig) error {

	err := deleteEnteredDestroyPhaseStateFile(a.workingDir, TFDestroyStateEnteredFileName)
	if err != nil {
		log.Warnf("error deleting enteredDestroyPhaseStateFile: %s", err)
	}

	tf, err := tfexec.NewTerraform(a.workingDir, a.execPath)
	if err != nil {
		log.Fatalf("error running NewTerraform: %s", err)
	}

	err = tf.Init(context.Background())

	if err != nil {
		log.Fatalf("error running Init: %s", err)
	}

	err = tf.Destroy(a.ctx, tfexec.VarFile(a.varFilePath))
	if err != nil {
		log.Warnf("error running terraform destroy: %s", err)
	}
	return err
}

func (a *terraformDeploymentConfig) deployTerraform(mpfConfig domain.MPFConfig) (string, error) {

	log.Infof("workingDir: %s", a.workingDir)
	log.Infof("varfilePath: %s", a.varFilePath)
	log.Infof("execPath: %s", a.execPath)

	tf, err := tfexec.NewTerraform(a.workingDir, a.execPath)
	if err != nil {
		log.Fatalf("error running NewTerraform: %s", err)
	}

	envVars := map[string]string{
		"ARM_CLIENT_ID":       mpfConfig.SP.SPClientID,
		"ARM_CLIENT_SECRET":   mpfConfig.SP.SPClientSecret,
		"ARM_SUBSCRIPTION_ID": mpfConfig.SubscriptionID,
		"ARM_TENANT_ID":       mpfConfig.TenantID,
	}

	tf.SetEnv(envVars)

	err = tf.Init(context.Background())

	inDestroyPhase := doesEnteredDestroyPhaseStateFileExist(a.workingDir, TFDestroyStateEnteredFileName)

	if err != nil {
		log.Fatalf("error running Init: %s", err)
	}

	if !inDestroyPhase {
		log.Infoln("in apply phase")
		err = tf.Apply(a.ctx, tfexec.VarFile(a.varFilePath))

		if err != nil {
			errorMsg := err.Error()
			log.Debugln(errorMsg)

			if strings.Contains(errorMsg, "Authorization") {
				return errorMsg, nil
			}

			log.Warnf("terraform apply: non authorizaton error occured: %s", errorMsg)

		}
	}

	log.Infoln("in destroy phase")
	if !inDestroyPhase {
		err = createEnteredDestroyPhaseStateFile(a.workingDir, TFDestroyStateEnteredFileName)
		if err != nil {
			log.Warnf("error creating enteredDestroyPhaseStateFile: %s", err)
		}
	}

	err = tf.Destroy(a.ctx, tfexec.VarFile(a.varFilePath))

	if err != nil {
		errorMsg := err.Error()
		log.Debugln(errorMsg)
		if strings.Contains(errorMsg, "Authorization") {
			return errorMsg, nil
		}

		log.Warnf("terraform destroy: non authorizaton error occured: %s", errorMsg)

	}

	return "", nil

}

func doesEnteredDestroyPhaseStateFileExist(workingDir string, fileName string) bool {
	enteredDestroyPhaseStateFileName := workingDir + "/" + fileName

	if _, err := os.Stat(enteredDestroyPhaseStateFileName); err == nil {
		log.Infof("%s file exists \n", enteredDestroyPhaseStateFileName)
		return true
	}
	log.Infof("%s file does not exist \n", enteredDestroyPhaseStateFileName)
	return false
}

func createEnteredDestroyPhaseStateFile(workingDir string, fileName string) error {
	enteredDestroyPhaseStateFileName := workingDir + "/" + fileName

	if _, err := os.Stat(enteredDestroyPhaseStateFileName); err == nil {
		log.Infof("%s file already exists \n", enteredDestroyPhaseStateFileName)
		return nil
	}

	_, err := os.Create(enteredDestroyPhaseStateFileName)
	if err != nil {
		log.Warnf("error creating %s file: %s", enteredDestroyPhaseStateFileName, err)
		return err
	}
	log.Infof("%s created enteredDestroyPhaseStateFile file \n", enteredDestroyPhaseStateFileName)
	return nil
}

func deleteEnteredDestroyPhaseStateFile(workingDir string, fileName string) error {
	enteredDestroyPhaseStateFileName := workingDir + "/" + fileName

	if _, err := os.Stat(enteredDestroyPhaseStateFileName); err != nil {
		log.Infof("%s file does not exist \n", enteredDestroyPhaseStateFileName)
		return nil
	}

	err := os.Remove(enteredDestroyPhaseStateFileName)
	if err != nil {
		log.Warnf("error deleting %s file: %s", enteredDestroyPhaseStateFileName, err)
		return err
	}
	log.Infof("%s deleted enteredDestroyPhaseStateFile file \n", enteredDestroyPhaseStateFileName)
	return nil
}
