package domain

type Role struct {
	RoleDefinitionID          string
	RoleDefinitionName        string
	RoleDefinitionDescription string
	RoleDefinitionResourceID  string
}

type ResourceGroup struct {
	ResourceGroupName       string
	ResourceGroupResourceID string
	Location                string
}

type ServicePrincipal struct {
	SPClientID     string
	SPObjectID     string
	SPClientSecret string
}

type MPFConfig struct {
	ResourceGroup  ResourceGroup
	SubscriptionID string
	TenantID       string
	SP             ServicePrincipal
	Role           Role
}

type MPFResult struct {
	// The map from which the minimum permissions will be calculated
	RequiredPermissions map[string][]string
}

func GetMPFResult(requiredPermissions map[string][]string) MPFResult {
	return MPFResult{
		RequiredPermissions: getMapWithUniqueValues(requiredPermissions),
	}
}
