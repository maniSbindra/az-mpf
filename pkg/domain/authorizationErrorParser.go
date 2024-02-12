package domain

import (
	"errors"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

func GetScopePermissionsFromAuthError(authErrMesg string) (map[string][]string, error) {
	if authErrMesg != "" && !strings.Contains(authErrMesg, "AuthorizationFailed") && !strings.Contains(authErrMesg, "Authorization failed") {
		log.Infoln("Non Authorization Error when creating deployment:", authErrMesg)
		return nil, errors.New("Could not parse deploment error, potentially due to a Non-Authorization error")
	}

	var resMap map[string][]string
	var err error

	if strings.Count(authErrMesg, "AuthorizationFailed") >= 1 {
		resMap, err = parseMultiAuthorizationFailedErrors(authErrMesg)

	}

	// If count of "Authorization failed" in error message is 1 or more than 1, then it is a multi authorization error
	if strings.Count(authErrMesg, "Authorization failed") >= 1 {
		resMap, err = parseMultiAuthorizationErrors(authErrMesg)
	}

	if err != nil {
		return nil, err
	}

	// If map is empty, return error
	if len(resMap) == 0 {
		return nil, errors.New("Could not parse deployment error for scope/permissions")
	}

	// For each /write permission add a /read permission to map
	// traverse resMap for each permission ending with /write add /read permission
	for scope, permissions := range resMap {
		for _, permission := range permissions {
			if strings.HasSuffix(permission, "/write") {
				readPermission := strings.Replace(permission, "/write", "/read", 1)
				resMap[scope] = append(resMap[scope], readPermission)
			}
		}
	}

	return resMap, nil
}

// For 'AuthorizationFailed' errors
func parseMultiAuthorizationFailedErrors(authorizationFailedErrMsg string) (map[string][]string, error) {

	re := regexp.MustCompile(`The client '([^']+)' with object id '([^']+)' does not have authorization to perform action '([^']+)'.* over scope '([^']+)' or the scope is invalid\.`)

	matches := re.FindAllStringSubmatch(authorizationFailedErrMsg, -1)

	if len(matches) == 0 {
		return nil, errors.New("No matches found in 'AuthorizationFailed' error message")
	}

	scopePermissionsMap := make(map[string][]string)

	// Iterate through the matches and populate the map
	for _, match := range matches {
		if len(match) == 5 {
			// resourceType := match[1]
			action := match[3]
			scope := match[4]

			if _, ok := scopePermissionsMap[scope]; !ok {
				scopePermissionsMap[scope] = make([]string, 0)
			}
			scopePermissionsMap[scope] = append(scopePermissionsMap[scope], action)
		}
	}

	// if map is empty, return error
	if len(scopePermissionsMap) == 0 {
		return nil, errors.New("No scope/permissions found in Multi error message")
	}

	return scopePermissionsMap, nil

}

// For 'Authorization failed' errors
func parseMultiAuthorizationErrors(authorizationFailedErrMsg string) (map[string][]string, error) {

	// Regular expression to extract resource information
	re := regexp.MustCompile(`Authorization failed for template resource '([^']+)' of type '([^']+)'\. The client '([^']+)' with object id '([^']+)' does not have permission to perform action '([^']+)' at scope '([^']+)'\.`)

	// Find all matches in the error message
	matches := re.FindAllStringSubmatch(authorizationFailedErrMsg, -1)

	// If No Matches found return error
	if len(matches) == 0 {
		return nil, errors.New("No matches found in 'Authorization failed' error message")
	}

	// Create a map to store scope/permissions
	scopePermissionsMap := make(map[string][]string)

	// Iterate through the matches and populate the map
	for _, match := range matches {
		if len(match) == 7 {
			// resourceType := match[1]
			action := match[5]
			scope := match[6]

			if _, ok := scopePermissionsMap[scope]; !ok {
				scopePermissionsMap[scope] = make([]string, 0)
			}
			scopePermissionsMap[scope] = append(scopePermissionsMap[scope], action)
		}
	}

	// if map is empty, return error
	if len(scopePermissionsMap) == 0 {
		return nil, errors.New("No scope/permissions found in Multi error message")
	}

	return scopePermissionsMap, nil

}
