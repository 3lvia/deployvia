package model

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestValidateDeployment(t *testing.T) {
	tests := []struct {
		name        string
		deployment  *Deployment
		claims      jwt.Claims
		expectError bool
	}{
		{
			name: "valid deployment",
			deployment: &Deployment{
				ApplicationName: "demo-api",
				System:          "core",
				Environment:     "dev",
				ClusterType:     "aks",
				Image:           "containerregistryelvia.azurecr.io/core-demo-api@sha256:1234567890abcdef",
			},
			expectError: false,
		},
		{
			name: "valid deployment",
			deployment: &Deployment{
				ApplicationName: "demo-api",
				System:          "core",
				Environment:     "dev",
				ClusterType:     "aks",
				Image:           "containerregistryelvia.azurecr.io/core-demo-api@sha256:1234567890abcdef",
			},
			expectError: false,
		},
		{
			name: "invalid system name",
			deployment: &Deployment{
				ApplicationName: "demo-api",
				System:          "core_1",
				Environment:     "dev",
				ClusterType:     "aks",
				Image:           "containerregistryelvia.azurecr.io/core-demo-api:dev@sha256:1234567890abcdef",
			},
			expectError: true,
		},
		{
			name: "invalid application name",
			deployment: &Deployment{
				ApplicationName: "../demo-api",
				System:          "core",
				Environment:     "dev",
				ClusterType:     "aks",
				Image:           "containerregistryelvia.azurecr.io/core-demo-api:dev@sha256:1234567890abcdef",
			},
			expectError: true,
		},
		{
			name: "missing image",
			deployment: &Deployment{
				ApplicationName: "demo-api",
				System:          "core",
				Environment:     "dev",
				ClusterType:     "gke",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateDeployment(tt.deployment)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateDeployment() error = '%v', expectError %v", err, tt.expectError)
			}
		})
	}
}
