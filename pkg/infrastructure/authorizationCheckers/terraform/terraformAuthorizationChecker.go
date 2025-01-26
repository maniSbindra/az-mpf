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
	ctx                            context.Context
	workingDir                     string
	execPath                       string
	varFilePath                    string
	importExistingResourcesToState bool
	targetModule                   string
}

var inDestroyPhase bool

// This file is created once the destroy phase is entered
const (
	TFDestroyStateEnteredFileName = ".azmpfEnteredDestroyPhase.txt"
	TFExistingResourceErrorMsg    = "to be managed via Terraform this resource needs to be imported into the State"
	BillingFeaturesPayloadError   = "CurrentBillingFeatures is required in payload"
	// AuthorizationPermissionMismatchErr = "AuthorizationPermissionMismatch"
)

func NewTerraformAuthorizationChecker(workDir string, execPath string, varFilePath string, importExistingResources bool, targetModule string) *terraformDeploymentConfig {
	err := deleteEnteredDestroyPhaseStateFile(workDir, TFDestroyStateEnteredFileName)
	if err != nil {
		log.Warnf("error deleting enteredDestroyPhaseStateFile: %s", err)
	}

	return &terraformDeploymentConfig{
		workingDir:                     workDir,
		execPath:                       execPath,
		ctx:                            context.Background(),
		varFilePath:                    varFilePath,
		importExistingResourcesToState: importExistingResources,
		targetModule:                   targetModule,
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
		log.Warnf("error running Init: %s", err)
		return err
	}

	switch {
	case a.varFilePath == "" && a.targetModule == "":
		err = tf.Destroy(a.ctx)
	case a.varFilePath != "" && a.targetModule == "":
		err = tf.Destroy(a.ctx, tfexec.VarFile(a.varFilePath))
	case a.varFilePath == "" && a.targetModule != "":
		err = tf.Destroy(a.ctx, tfexec.Target(a.targetModule))
	case a.varFilePath != "" && a.targetModule != "":
		err = tf.Destroy(a.ctx, tfexec.VarFile(a.varFilePath), tfexec.Target(a.targetModule))
	}
	if err != nil {
		log.Warnf("error running terraform destroy: %s", err)
	}
	return err
}

func (a *terraformDeploymentConfig) setTFConfig(mpfConfig domain.MPFConfig) (*tfexec.Terraform, error) {
	log.Infof("workingDir: %s", a.workingDir)
	log.Infof("varfilePath: %s", a.varFilePath)
	log.Infof("execPath: %s", a.execPath)

	tf, err := tfexec.NewTerraform(a.workingDir, a.execPath)
	if err != nil {
		log.Fatalf("error running NewTerraform: %s", err)
	}

	pathEnvVal := os.Getenv("PATH")
	var tfLogLevel string

	tfLogPathEnvVal := os.Getenv("TF_LOG_PATH")
	if tfLogPathEnvVal == "" {
		tfLogPathEnvVal = a.workingDir + "/terraform.log"
	}

	tfReattachProviders := os.Getenv("TF_REATTACH_PROVIDERS")

	switch log.GetLevel() {
	case log.InfoLevel:
		tfLogLevel = "INFO"
	case log.WarnLevel:
		tfLogLevel = "WARN"
	case log.DebugLevel:
		tfLogLevel = "DEBUG"
	case log.TraceLevel:
		tfLogLevel = "TRACE"
	default:
		tfLogLevel = "ERROR"
	}

	if tfLogLevel != "ERROR" {
		tf.SetLog(tfLogLevel)
		tf.SetLogPath(tfLogPathEnvVal)
		tf.SetStderr(os.Stderr)
		tf.SetStdout(os.Stdout)
	}

	envVars := map[string]string{
		"ARM_CLIENT_ID":       mpfConfig.SP.SPClientID,
		"ARM_CLIENT_SECRET":   mpfConfig.SP.SPClientSecret,
		"ARM_SUBSCRIPTION_ID": mpfConfig.SubscriptionID,
		"ARM_TENANT_ID":       mpfConfig.TenantID,
		"PATH":                pathEnvVal,
	}

	if tfReattachProviders != "" {
		envVars["TF_REATTACH_PROVIDERS"] = tfReattachProviders
	}

	tf.SetEnv(envVars)
	return tf, nil
}

func (a *terraformDeploymentConfig) deployTerraform(mpfConfig domain.MPFConfig) (string, error) {
	tf, err := a.setTFConfig(mpfConfig)
	if err != nil {
		log.Fatalf("error setting Terraform start config: %s", err)
	}

	inDestroyPhase = doesEnteredDestroyPhaseStateFileExist(a.workingDir, TFDestroyStateEnteredFileName)
	if !inDestroyPhase {
		log.Infof("destroy phase file does not exist, in apply phase")
		msg, err := a.terraformApply(mpfConfig, tf)
		if err != nil || msg != "" {
			return msg, err
		}
	}

	return a.terraformDestroy(mpfConfig, tf)

}

func (a *terraformDeploymentConfig) terraformApply(mpfConfig domain.MPFConfig, tf *tfexec.Terraform) (string, error) {

	err := tf.Init(context.Background())
	if err != nil {
		log.Warnf("error running Init: %s", err)
		return "", err
	}

	log.Infoln("in apply phase")

	switch {
	case a.varFilePath == "" && a.targetModule == "":
		err = tf.Apply(a.ctx)
	case a.varFilePath != "" && a.targetModule == "":
		err = tf.Apply(a.ctx, tfexec.VarFile(a.varFilePath))
	case a.varFilePath == "" && a.targetModule != "":
		err = tf.Apply(a.ctx, tfexec.Target(a.targetModule))
	case a.varFilePath != "" && a.targetModule != "":
		err = tf.Apply(a.ctx, tfexec.VarFile(a.varFilePath), tfexec.Target(a.targetModule))
	}

	if err == nil {
		return "", nil
	}

	errorMsg := err.Error()
	log.Debugln("terraform apply error: ", errorMsg)

	// if strings.Contains(errorMsg, AuthorizationPermissionMismatchErr) {
	// 	return errorMsg, nil
	// }

	if strings.Contains(errorMsg, "Authorization") {
		return errorMsg, nil
	}

	// Temporary fix to workaround issue https://github.com/hashicorp/terraform-provider-azurerm/issues/27961
	// It is observed only once, so retrying works
	if strings.Contains(errorMsg, BillingFeaturesPayloadError) {
		return errorMsg, nil
	}

	// import errors can occur for some resources, when identity does not have all required permissions,
	// as described in https://github.com/hashicorp/terraform-provider-azurerm/issues/27961#issuecomment-2467392936
	if a.importExistingResourcesToState && strings.Contains(errorMsg, TFExistingResourceErrorMsg) {

		msg, err := a.terraformImport(tf, errorMsg)
		if err != nil || msg != "" {
			if strings.Contains(msg, "Authorization") {
				return msg, nil
			}
			return msg, err
		}
		return a.terraformApply(mpfConfig, tf)
	}

	log.Warnf("terraform apply: non authorizaton error occured: %s", errorMsg)
	return errorMsg, err
}

func (a *terraformDeploymentConfig) terraformImport(tf *tfexec.Terraform, existingResErrMesg string) (string, error) {
	log.Warnf("terraform apply: existing resource error occured:|| %s ||\n\n", existingResErrMesg)
	log.Warn("importing existing resources to state")

	exstResAddrAndResIDs, err := GetAddressAndResourceIDFromExistingResourceError(existingResErrMesg)
	if err != nil {
		log.Warnf("error getting existing resource address and resource ID: %s \n", err)
		return existingResErrMesg, err
	}

	for addr, resID := range exstResAddrAndResIDs {
		log.Warnf("importing existing resource: %s, %s ||\n", addr, resID)
		if a.varFilePath != "" {
			err = tf.Import(a.ctx, addr, resID, tfexec.VarFile(a.varFilePath))
		} else {
			err = tf.Import(a.ctx, addr, resID)
		}

		if err != nil {
			log.Warnf("error importing existing resource: %s \n", err)
			return err.Error(), err
		}
		log.Warnf("imported existing resource: %s, %s \n", addr, resID)
	}
	log.Warnf("imported all existing resources to state, triggering deployTerrafom again \n")
	return "", nil
}

func (a *terraformDeploymentConfig) terraformDestroy(mpfConfig domain.MPFConfig, tf *tfexec.Terraform) (string, error) {
	var err error
	log.Infoln("in destroy phase")
	if !inDestroyPhase {
		err = createEnteredDestroyPhaseStateFile(a.workingDir, TFDestroyStateEnteredFileName)
		if err != nil {
			log.Warnf("error creating enteredDestroyPhaseStateFile: %s", err)
		}
	}

	switch {
	case a.varFilePath == "" && a.targetModule == "":
		err = tf.Destroy(a.ctx)
	case a.varFilePath != "" && a.targetModule == "":
		err = tf.Destroy(a.ctx, tfexec.VarFile(a.varFilePath))
	case a.varFilePath == "" && a.targetModule != "":
		err = tf.Destroy(a.ctx, tfexec.Target(a.targetModule))
	case a.varFilePath != "" && a.targetModule != "":
		err = tf.Destroy(a.ctx, tfexec.VarFile(a.varFilePath), tfexec.Target(a.targetModule))
	}

	if err != nil {
		errorMsg := err.Error()
		log.Debugln(errorMsg)
		if strings.Contains(errorMsg, "Authorization") {
			return errorMsg, nil
		}
		log.Warnf("terraform destroy: non authorizaton error occured: %s", errorMsg)
		return errorMsg, err
	}
	return "", nil
}
