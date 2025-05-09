package deploy

import (
	"fmt"
	"regexp"
)

// We use a 'Validated(MyStruct)' pattern to wrap struct types that need to be validated, e.g. fields are checked for zero values or regex patterns.
// Once the struct is validated, we wrap it in a 'Validated' struct, certifying that it has been validated.

type ValidatedDeployment struct {
	Deployment *Deployment
}

type Deployment struct {
	ApplicationName string `json:"application_name"`
	System          string `json:"system"`
    ClusterType     string `json:"cluster_type"`
	Environment     string `json:"environment"`
    Image          string `json:"image"`
}

func ValidateDeployment(deployment *Deployment) (*ValidatedDeployment, error) {
	if deployment == nil {
		return nil, fmt.Errorf("deployment is nil")
	}

	if deployment.System == "" {
		return nil, fmt.Errorf("system is required")
	}

	if deployment.ApplicationName == "" {
		return nil, fmt.Errorf("application name is required")
	}

	re := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

	if !re.MatchString(deployment.System) {
		return nil, fmt.Errorf("system name must only contain alphanumeric characters and hyphens")
	}

	if !re.MatchString(deployment.ApplicationName) {
		return nil, fmt.Errorf("application name must only contain alphanumeric characters and hyphens")
	}

    if deployment.ClusterType == "" {
        return nil, fmt.Errorf("cluster type is required")
    }

	if deployment.Environment == "" {
		return nil, fmt.Errorf("environment is required")
	}

	if deployment.Image == "" {
		return nil, fmt.Errorf("image is required")
	}

	return &ValidatedDeployment{Deployment: deployment}, nil
}
