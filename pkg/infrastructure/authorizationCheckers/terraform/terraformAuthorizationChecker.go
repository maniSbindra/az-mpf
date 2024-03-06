package terraform

import (
	"context"
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

func NewTerraformAuthorizationChecker(workDir string, execPath string, varFilePath string) *terraformDeploymentConfig {
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

	if err != nil {
		log.Fatalf("error running Init: %s", err)
	}

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

	log.Infoln("in destroy phase")
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
