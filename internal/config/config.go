package config

import (
	"context"
	"os"

	"k8s.io/client-go/dynamic"
)

type Config struct {
	Environment        string
	GitHubOIDCURL      string
	KubernetesClient   *dynamic.DynamicClient
	ApplicationMetrics *ApplicationMetrics
	Local              bool
	Port               string
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

	local := os.Getenv("LOCAL") == "true"

	port := func() string {
		port_ := os.Getenv("PORT")
		if port_ == "" {
			return "8080"
		}

		return port_
	}()

	return &Config{
		KubernetesClient:   k8sClient,
		GitHubOIDCURL:      GITHUB_OIDC_URL,
		ApplicationMetrics: applicationMetrics,
		Local:              local,
		Port:               port,
	}, nil
}
