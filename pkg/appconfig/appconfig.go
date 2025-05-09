package appconfig

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"os"
	"k8s.io/client-go/dynamic"
)

type Config struct {
	Environment        string
	GitHubOIDCURL      string
    KubernetesClient   *dynamic.DynamicClient
	ApplicationMetrics *ApplicationMetrics
}

func New(ctx context.Context) (*Config, error) {
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		return nil, errors.New("ENVIRONMENT is not set")
	}

	if environment == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

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
		Environment:        environment,
        KubernetesClient:   k8sClient,
		GitHubOIDCURL:      GITHUB_OIDC_URL,
		ApplicationMetrics: applicationMetrics,
	}, nil
}
