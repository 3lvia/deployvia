package routes

import (
	"context"
	"fmt"
    "os"

	"github.com/3lvia/core/applications/deployvia/pkg/appconfig"
	"github.com/3lvia/core/applications/deployvia/pkg/deploy"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func PostDeployment(ctx context.Context, c *gin.Context, config *appconfig.Config) {
    local := os.Getenv("LOCAL") == "true"

    if !local {
	gitHubOIDCToken := c.Request.Header.Get("X-GitHub-OIDC-Token")
	if gitHubOIDCToken == "" {
		err := fmt.Errorf("X-GitHub-OIDC-Token header is required")
		log.Error(err)
		c.JSON(400, gin.H{"error": err.Error()})

		return
	}

	const GitHubOIDCURL = "https://token.actions.githubusercontent.com/.well-known/jwks"

	_, err := deploy.ValidateToken(ctx, gitHubOIDCToken, GitHubOIDCURL)
	if err != nil {
		err := fmt.Errorf("invalid token: %w", err)
		log.Error(err)
		c.JSON(403, gin.H{"error": err.Error()})

		return
	}
    }

	validatedDeployment, err := func() (*deploy.ValidatedDeployment, error) {
		var deployment deploy.Deployment
		if err := c.ShouldBindJSON(&deployment); err != nil {
			return nil, err
		}

		return deploy.ValidateDeployment(&deployment)
	}()
	if err != nil {
		err := fmt.Errorf("invalid deployment: %w", err)
		log.Error(err)
		c.JSON(400, gin.H{"error": err.Error()})

		return
	}

	gvr := schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "applications",
	}

    appName := fmt.Sprintf(
        "%s-%s-%s-%s",
        validatedDeployment.Deployment.System,
        validatedDeployment.Deployment.ApplicationName,
        validatedDeployment.Deployment.ClusterType,
        validatedDeployment.Deployment.Environment,
    )

    app, err := config.KubernetesClient.Resource(gvr).Namespace("argocd").Get(ctx, appName, v1.GetOptions{})
    if err != nil {
        log.Warnf("failed to get ArgoCD application: %v", err)
        c.JSON(404, gin.H{"error": "ArgoCD application not found"})

        return
    }

    c.JSON(200, gin.H{
        "message": "Deployment validated successfully",
        "application": app,
    })

	return
}
