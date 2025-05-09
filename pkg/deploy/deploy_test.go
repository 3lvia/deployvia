package deploy

import (
	"github.com/golang-jwt/jwt/v5"
	"testing"
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
				ImageRepository: "containerregistryelvia.azurecr.io/core-demo-api",
				ImageDigest:     "sha256:6e51725bf9c670677f284412c8d011e1390b143aa1ed6ebb33c5674f3da31373",
			},
			expectError: false,
		},
		{
			name: "valid deployment",
			deployment: &Deployment{
				ApplicationName: "demo-api",
				System:          "core",
				Environment:     "dev",
				ImageDigest:     "sha256:6e51725bf9c670677f284412c8d011e1390b143aa1ed6ebb33c5674f3da31373",
			},
			expectError: false,
		},
		{
			name: "invalid system name",
			deployment: &Deployment{
				ApplicationName: "demo-api",
				System:          "core_1",
				Environment:     "dev",
				ImageDigest:     "sha256:6e51725bf9c670677f284412c8d011e1390b143aa1ed6ebb33c5674f3da31373",
			},
			expectError: true,
		},
		{
			name: "invalid application name",
			deployment: &Deployment{
				ApplicationName: "../demo-api",
				System:          "core",
				Environment:     "dev",
				ImageDigest:     "sha256:6e51725bf9c670677f284412c8d011e1390b143aa1ed6ebb33c5674f3da31373",
			},
			expectError: true,
		},
		{
			name: "missing image digest",
			deployment: &Deployment{
				ApplicationName: "demo-api",
				System:          "core",
				Environment:     "dev",
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
