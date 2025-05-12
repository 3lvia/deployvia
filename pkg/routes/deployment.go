package routes

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/3lvia/core/applications/deployvia/pkg/appconfig"
	"github.com/3lvia/core/applications/deployvia/pkg/deploy"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
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

	err = watchApplicationLifecycle(
		ctx,
		config.KubernetesClient,
		gvr,
		"argocd",
		appName,
		5*time.Minute,
	)
	if err != nil {
		err := fmt.Errorf("failed to watch application lifecycle: %w", err)
		log.Error(err)
		c.JSON(500, gin.H{"error": err.Error()})

		return
	}

}

func watchApplicationLifecycle(
	ctx context.Context,
	client dynamic.Interface,
	gvr schema.GroupVersionResource,
	namespace,
	appName string,
	timeout time.Duration,
) error {
	_, err := client.Resource(gvr).Namespace(namespace).Get(ctx, appName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get application %s in namespace %s: %w", appName, namespace, err)
	}

	w, err := client.Resource(gvr).Namespace(namespace).Watch(
		ctx,
		metav1.ListOptions{
			FieldSelector:  fmt.Sprintf("metadata.name=%s", appName),
			TimeoutSeconds: int64Ptr(int64(timeout.Seconds())),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to watch application: %w", err)
	}

	defer w.Stop()

	seenOutOfSync := false
	resultChan := w.ResultChan()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case evt, ok := <-resultChan:
			if !ok {
				return fmt.Errorf("watch closed unexpectedly")
			}

			obj, ok := evt.Object.(*unstructured.Unstructured)
			if !ok {
				continue
			}

			syncStatus, _, _ := unstructured.NestedString(obj.Object, "status", "sync", "status")
			healthStatus, _, _ := unstructured.NestedString(obj.Object, "status", "health", "status")

			fmt.Printf("Event: %s, sync=%s, health=%s\n", evt.Type, syncStatus, healthStatus)

			if syncStatus == "OutOfSync" {
				seenOutOfSync = true
			}

			if seenOutOfSync && syncStatus == "Synced" && healthStatus == "Healthy" {
				fmt.Println("Application reached Synced & Healthy after OutOfSync.")

				return nil
			}
		case <-time.After(timeout):
			return fmt.Errorf("timed out waiting for application lifecycle")
		}
	}
}

func int64Ptr(i int64) *int64 {
	return &i
}
