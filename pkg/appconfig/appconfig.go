package appconfig

import (
	"context"
	"k8s.io/client-go/dynamic"
	"os"
)

type Config struct {
	Environment        string
	GitHubOIDCURL      string
	KubernetesClient   *dynamic.DynamicClient
	ApplicationMetrics *ApplicationMetrics
}

func New(ctx context.Context) (*Config, error) {
	const GITHUB_OIDC_URL = "https://token.actions.githubusercontent.com/.well-known/jwks"

	applicationMetrics, err := ConfigureOpenTelemetry(ctx)
	if err != nil {
		return nil, err
	}

	k8sClient, err := configureKubernetesClient(os.Getenv("LOCAL") == "true")
	if err != nil {
		return nil, err
	}

	return &Config{
		KubernetesClient:   k8sClient,
		GitHubOIDCURL:      GITHUB_OIDC_URL,
		ApplicationMetrics: applicationMetrics,
	}, nil
}
