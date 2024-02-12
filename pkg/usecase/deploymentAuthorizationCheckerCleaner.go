package usecase

import "github.com/manisbindra/az-mpf/pkg/domain"

type DeploymentAuthorizationChecker interface {

	// Check if the user has the required permissions to deploy the template
	// If Authorization Error is received the authorization error message string is returned, and error is nil
	// If string is empty and error is not nill, then non authorization error is received
	// If string is empty and error is nil, then authorization is successful
	GetDeploymentAuthorizationErrors(mpfCoreConfig domain.MPFConfig) (string, error)
}

type DeploymentCleaner interface {
	CleanDeployment(mpfCoreConfig domain.MPFConfig) error
}

type DeploymentAuthorizationCheckerCleaner interface {
	DeploymentAuthorizationChecker
	DeploymentCleaner
}
