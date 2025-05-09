package deploy

import (
	"path/filepath"
)

func GetAtlasFileName(validatedDeployment *ValidatedDeployment) string {
	return filepath.Join(
		"/tmp/argocd",
		"manifests/applications/systems",
		validatedDeployment.Deployment.System,
		validatedDeployment.Deployment.ApplicationName,
		"atlas.yml",
	)
}
