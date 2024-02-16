package usecase

import (
	"context"

	"github.com/manisbindra/az-mpf/pkg/domain"
)

type ServicePrincipalAssignmentModifier interface {
	DetachRolesFromSP(ctx context.Context, subscription string, resourceGroupName string, SPOBjectID string, role domain.Role) error
	AssignRoleToSP(subscription string, resourceGroupName string, SPOBjectID string, role domain.Role) error
}

type CustomRoleCreatorModifier interface {
	CreateUpdateCustomRole(subscription string, resourceGroupName string, role domain.Role, permissions []string) error
	DeleteCustomRole(subscription string, resourceGroupName string, role domain.Role) error
}

type ServicePrincipalRolemAssignmentManager interface {
	ServicePrincipalAssignmentModifier
	CustomRoleCreatorModifier
}
