package usecase

import (
	"context"
	"strings"

	"github.com/manisbindra/az-mpf/pkg/domain"
	log "github.com/sirupsen/logrus"
)

type MPFService struct {
	ctx                                 context.Context
	rgManager                           ResourceGroupManager
	spRoleAssignmentManager             ServicePrincipalRolemAssignmentManager
	deploymentAuthCheckerCleaner        DeploymentAuthorizationCheckerCleaner
	mpfConfig                           domain.MPFConfig
	initialPermissionsToAdd             []string
	permissionsToAddToResult            []string
	requiredPermissions                 map[string][]string
	autoAddReadPermissionForEachWrite   bool
	autoAddDeletePermissionForEachWrite bool
	autoCreateResourceGroup             bool
}

func NewMPFService(ctx context.Context, rgMgr ResourceGroupManager, spRoleAssgnMgr ServicePrincipalRolemAssignmentManager, deploymentAuthChkCln DeploymentAuthorizationCheckerCleaner, mpfConfig domain.MPFConfig, initialPermissionsToAdd []string, permissionsToAddToResult []string, autoAddReadPermissionForEachWrite bool, autoAddDeletePermissionForEachWrite bool, autoCreateResourceGroup bool) *MPFService {
	return &MPFService{
		ctx:                                 ctx,
		rgManager:                           rgMgr,
		spRoleAssignmentManager:             spRoleAssgnMgr,
		deploymentAuthCheckerCleaner:        deploymentAuthChkCln,
		mpfConfig:                           mpfConfig,
		initialPermissionsToAdd:             initialPermissionsToAdd,
		permissionsToAddToResult:            permissionsToAddToResult,
		requiredPermissions:                 make(map[string][]string),
		autoAddReadPermissionForEachWrite:   autoAddReadPermissionForEachWrite,
		autoAddDeletePermissionForEachWrite: autoAddDeletePermissionForEachWrite,
		autoCreateResourceGroup:             autoCreateResourceGroup,
	}
}

func (s *MPFService) GetMinimumPermissionsRequired() (domain.MPFResult, error) {

	if s.autoCreateResourceGroup {
		// Create Resource Group
		log.Infof("Creating Resource Group: %s \n", s.mpfConfig.ResourceGroup.ResourceGroupName)
		err := s.rgManager.CreateResourceGroup(s.ctx, s.mpfConfig.ResourceGroup.ResourceGroupName, s.mpfConfig.ResourceGroup.Location)
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("Resource Group: %s created successfully \n", s.mpfConfig.ResourceGroup.ResourceGroupName)
		// defer s.deploymentAuthCheckerCleaner.CleanDeployment(s.mpfConfig)
	}

	defer s.CleanUpResources()

	// Delete all existing role assignments for the service principal
	err := s.spRoleAssignmentManager.DetachRolesFromSP(s.ctx, s.mpfConfig.SubscriptionID, s.mpfConfig.SP.SPObjectID, s.mpfConfig.Role)
	if err != nil {
		log.Warnf("Unable to delete Role Assignments: %v\n", err)
		return domain.MPFResult{}, err
	}
	log.Info("Deleted all existing role assignments for service principal \n")

	// Initialize new custom role
	log.Infoln("Initializing Custom Role")
	// err = mpf.CreateUpdateCustomRole([]string{})

	err = s.spRoleAssignmentManager.CreateUpdateCustomRole(s.mpfConfig.SubscriptionID, s.mpfConfig.Role, s.initialPermissionsToAdd)
	if err != nil {
		log.Warn(err)
		return domain.MPFResult{}, err
	}
	log.Infoln("Custom role initialized successfully")

	// Assign new custom role to service principal
	log.Infoln("Assigning new custom role to service principal")
	// err = mpf.AssignRoleToSP()
	err = s.spRoleAssignmentManager.AssignRoleToSP(s.mpfConfig.SubscriptionID, s.mpfConfig.SP.SPObjectID, s.mpfConfig.Role)
	if err != nil {
		log.Warn(err)
		return domain.MPFResult{}, err
	}
	log.Infoln("New Custom Role assigned to service principal successfully")

	// Add initial permissions to requiredPermissions map
	log.Infoln("Adding initial permissions to requiredPermissions map")
	for _, permission := range s.permissionsToAddToResult {
		s.requiredPermissions[s.mpfConfig.ResourceGroup.ResourceGroupResourceID] = append(s.requiredPermissions[s.mpfConfig.ResourceGroup.ResourceGroupResourceID], permission)
	}
	// s.requiredPermissions[s.mpfConfig.ResourceGroup.ResourceGroupResourceID] = append(s.requiredPermissions[s.mpfConfig.ResourceGroup.ResourceGroupResourceID], s.permissionsToAddToResult...)

	maxIterations := 15
	iterCount := 0
	for {
		authErrMesg, err := s.deploymentAuthCheckerCleaner.GetDeploymentAuthorizationErrors(s.mpfConfig)

		if authErrMesg == "" && err == nil {
			log.Infoln("Authorization Successful")
			break
		}

		if err != nil {
			log.Errorf("Non Authorization error received: %v \n", err)
			return domain.MPFResult{}, err
		}

		log.Debugln("Deployment Authorization Error:", authErrMesg)

		scpMp, err := domain.GetScopePermissionsFromAuthError(authErrMesg)
		if err != nil {
			log.Warnf("Could Not Parse Deployment Authorization Error: %v \n", err)
			return domain.MPFResult{}, err
		}

		log.Infoln("Successfully Parsed Deployment Authorization Error")
		log.Debugln("scope permissions found from deployment error:", scpMp)

		// auto add read and delete permissions as per configuration
		for scope, permissions := range scpMp {
			for _, permission := range permissions {
				if s.autoAddReadPermissionForEachWrite && strings.HasSuffix(permission, "/write") {
					readPermission := strings.Replace(permission, "/write", "/read", 1)
					scpMp[scope] = append(scpMp[scope], readPermission)
				}
				if s.autoAddDeletePermissionForEachWrite && strings.HasSuffix(permission, "/write") {
					deletePermission := strings.Replace(permission, "/write", "/delete", 1)
					scpMp[scope] = append(scpMp[scope], deletePermission)
				}
			}
		}

		log.Infoln("Adding mising scopes/permissions to final result map...")
		for k, v := range scpMp {
			s.requiredPermissions[k] = append(s.requiredPermissions[k], v...)
			s.requiredPermissions[s.mpfConfig.ResourceGroup.ResourceGroupResourceID] = append(s.requiredPermissions[s.mpfConfig.ResourceGroup.ResourceGroupResourceID], v...)
		}

		// assign permission to role
		log.Infoln("Adding permission/scope to role...........")
		log.Debugln("Number of Permissions added to role:", len(s.requiredPermissions[s.mpfConfig.ResourceGroup.ResourceGroupResourceID]))

		permissionsIncludingInitialPermissions := append(s.initialPermissionsToAdd, s.requiredPermissions[s.mpfConfig.ResourceGroup.ResourceGroupResourceID]...)
		err = s.spRoleAssignmentManager.CreateUpdateCustomRole(s.mpfConfig.SubscriptionID, s.mpfConfig.Role, permissionsIncludingInitialPermissions)

		// err = s.spRoleAssignmentManager.CreateUpdateCustomRole(s.mpfConfig.SubscriptionID, s.mpfConfig.ResourceGroup.ResourceGroupName, s.mpfConfig.Role, s.requiredPermissions[s.mpfConfig.ResourceGroup.ResourceGroupResourceID])

		if err != nil {
			log.Infoln("Error when adding permission/scope to role: \n", err)
			log.Warn(err)
			return domain.MPFResult{}, err
		}
		log.Infoln("Permission/scope added to role successfully")

		iterCount++
		if iterCount == maxIterations {
			log.Warnln("max iterations for fetching authorization errors reached, exiting...")
			return domain.MPFResult{}, err
		}
	}

	return domain.GetMPFResult(s.requiredPermissions), nil

}

func (s *MPFService) CleanUpResources() {
	log.Infoln("Cleaning up resources...")
	log.Infoln("*************************")

	// Cancel deployment. Even if cancelling deployment fails attempt to delete other resources
	// _ = m.CancelDeployment(deploymentName)

	err := s.deploymentAuthCheckerCleaner.CleanDeployment(s.mpfConfig)
	if err != nil {
		log.Warnln("Cleaning up deployment returned an error, attempting to clean rest of the resources")
	}

	// Detach Roles from SP
	err = s.spRoleAssignmentManager.DetachRolesFromSP(s.ctx, s.mpfConfig.SubscriptionID, s.mpfConfig.SP.SPObjectID, s.mpfConfig.Role)
	if err != nil {
		log.Warnf("Could not detach roles from SP: %s\n", err)
	}

	// Delete Custom Role
	err = s.spRoleAssignmentManager.DeleteCustomRole(s.mpfConfig.SubscriptionID, s.mpfConfig.Role)
	if err != nil {
		log.Warnf("Could not delete custom role: %s\n", err)
	}

	// Delete Resource Group
	if s.autoCreateResourceGroup {
		err = s.rgManager.DeleteResourceGroup(s.ctx, s.mpfConfig.ResourceGroup.ResourceGroupName)
		if err != nil {
			log.Warnf("Error when deleting resource group: %s \n", err)
		}
		log.Infoln("Resource group deletion initiated successfully...")
	}

}
