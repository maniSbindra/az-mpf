package terraform

import (
	"errors"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

func GetAddressAndResourceIDFromExistingResourceError(existingResourceErr string) (map[string]string, error) {
	if existingResourceErr != "" && !strings.Contains(existingResourceErr, TFExistingResourceErrorMsg) {
		log.Infoln("Non existing resource error :", existingResourceErr)
		return nil, errors.New("Non existing resource error")
	}

	var resMap map[string]string
	resMap = make(map[string]string)
	// var err error
	re := regexp.MustCompile(`Error: A resource with the ID "([^"]+)" already exists - to be managed via Terraform this resource needs to be imported into the State. Please see the resource documentation for "([^"]+)" for more information.\n\n  with ([^,]+),`)
	// re := regexp.MustCompile(`Error: A resource with the ID "([^']+)" already exists - to be managed via Terraform this resource needs to be imported into the State. Please see the resource documentation for "([^']+)" for more information\.(.*)  with ([^,]+),`)

	matches := re.FindAllStringSubmatch(existingResourceErr, -1)

	if len(matches) == 0 {
		return nil, errors.New("No matches found in 'existingResourceErrorMessage' error message")
	}

	for _, match := range matches {
		if len(match) == 4 {
			// resourceType := match[1]
			resID := match[1]
			addr := match[3]
			resMap[addr] = resID
		}
	}

	return resMap, nil
}
