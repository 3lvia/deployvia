package deploy

import "testing"

func TestGetAtlasFileName(t *testing.T) {
	tests := []struct {
		name        string
		deployment  *Deployment
		expected    string
		expectError bool
	}{
		{
			name: "valid input",
			deployment: &Deployment{
				System:          "core",
				ApplicationName: "demo-api",
				Environment:     "dev",
				ImageDigest:     "sha256:1234567890abcdef",
			},
			expected:    "/tmp/argocd/manifests/applications/systems/core/demo-api/atlas.yml",
			expectError: false,
		},
		{
			name: "missing system",
			deployment: &Deployment{
				ApplicationName: "demo-api",
				Environment:     "dev",
				ImageDigest:     "sha256:1234567890abcdef",
			},
			expected:    "",
			expectError: true,
		},
		{
			name: "missing application name",
			deployment: &Deployment{
				System:      "core",
				Environment: "dev",
				ImageDigest: "sha256:1234567890abcdef",
			},
			expected:    "",
			expectError: true,
		},
		{
			name: "missing environment",
			deployment: &Deployment{
				ApplicationName: "demo-api",
				System:          "core",
				ImageDigest:     "sha256:1234567890abcdef",
			},
			expected:    "",
			expectError: true,
		},
		{
			name: "missing image digest",
			deployment: &Deployment{
				ApplicationName: "demo-api",
				System:          "core",
				Environment:     "dev",
			},
			expected:    "",
			expectError: true,
		},
		{
			name: "invalid system name",
			deployment: &Deployment{
				System:          "../test-system",
				ApplicationName: "test-app",
				Environment:     "dev",
				ImageDigest:     "sha256:1234567890abcdef",
			},
			expected:    "",
			expectError: true,
		},
		{
			name: "invalid application name",
			deployment: &Deployment{
				System:          "test-system",
				ApplicationName: "test/app",
				Environment:     "dev",
				ImageDigest:     "sha256:1234567890abcdef",
			},
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := func() (string, error) {
				validatedDeployment, err := ValidateDeployment(tt.deployment)
				if err != nil {
					return "", err
				}

				return GetAtlasFileName(validatedDeployment), nil
			}()
			if (err != nil) != tt.expectError {
				t.Errorf("GetAtlasFileName() error = %v, wantErr %v", err, tt.expectError)
				return
			}
			if result != tt.expected {
				t.Errorf("GetAtlasFileName() = %v, want %v", result, tt.expected)
			}
		})
	}
}
