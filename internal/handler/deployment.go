package handler

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/3lvia/deployvia/internal/config"
	"github.com/3lvia/deployvia/internal/model"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

func PostDeployment(
	ctx context.Context,
	c *gin.Context,
	config *config.Config,
) {
	testingEnableOIDC := os.Getenv("TESTING_ENABLE_OIDC") == "true"
	if testingEnableOIDC {
		log.Errorf("TESTING_ENABLE_OIDC is set to true; THIS SHOULD NEVER BE USED IN PRODUCTION!")
	}

	if !config.Local || testingEnableOIDC {
		gitHubOIDCToken := c.Request.Header.Get("X-GitHub-OIDC-Token")
		if gitHubOIDCToken == "" {
			err := fmt.Errorf("X-GitHub-OIDC-Token header is required")
			log.Error(err)
			c.JSON(400, gin.H{"error": err.Error()})

			return
		}

		const GitHubOIDCURL = "https://token.actions.githubusercontent.com/.well-known/jwks"

		_, err := model.ValidateToken(ctx, gitHubOIDCToken, GitHubOIDCURL)
		if err != nil {
			err := fmt.Errorf("invalid token: %w", err)
			log.Error(err)
			c.JSON(403, gin.H{"error": err.Error()})

			return
		}
	}

	validatedDeployment, err := func() (*model.ValidatedDeployment, error) {
		var deployment model.Deployment
		if err := c.ShouldBindJSON(&deployment); err != nil {
			return nil, err
		}

		return model.ValidateDeployment(&deployment)
	}()
	if err != nil {
		err := fmt.Errorf("invalid deployment: %w", err)
		log.Error(err)
		c.JSON(400, gin.H{"error": err.Error()})

		return
	}

	timeout := func() time.Duration {
		const defaultTimeout = 3 * time.Minute

		timeoutHeader := c.Request.Header.Get("X-Timeout")
		if timeoutHeader == "" {
			return defaultTimeout
		}

		timeout, err := time.ParseDuration(timeoutHeader)
		if err != nil {
			return defaultTimeout
		}

		return timeout
	}()

	gvr := schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "applications",
	}

	err = watchApplicationLifecycle(
		ctx,
		config.KubernetesClient,
		gvr,
		"argocd",
		validatedDeployment,
		timeout,
	)
	if err != nil {
		err := fmt.Errorf("failed to watch application lifecycle: %w", err)
		log.Error(err)
		c.JSON(500, gin.H{"error": err.Error()})

		return
	}

	log.Infof("Deployment %s/%s successfully deployed", validatedDeployment.Deployment.System, validatedDeployment.Deployment.ApplicationName)
	c.JSON(200, gin.H{"message": "Application successfully deployed"})
}

func watchApplicationLifecycle(
	ctx context.Context,
	client dynamic.Interface,
	gvr schema.GroupVersionResource,
	namespace string,
	validatedDeployment *model.ValidatedDeployment,
	timeout time.Duration,
) error {
	application, err := client.Resource(gvr).Namespace(namespace).List(
		ctx,
		metav1.ListOptions{
			LabelSelector: fmt.Sprintf(
				"elvia.no/system=%s,elvia.no/application=%s,elvia.no/cluster-type=%s,kubernetes.io/environment=%s",
				validatedDeployment.Deployment.System,
				validatedDeployment.Deployment.ApplicationName,
				validatedDeployment.Deployment.ClusterType,
				validatedDeployment.Deployment.Environment,
			),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to get application for deployment: %w", err)
	}

	if len(application.Items) == 0 {
		return fmt.Errorf("application not found")
	}

	if len(application.Items) > 1 {
		return fmt.Errorf("multiple applications found")
	}

	applicationName := application.Items[0].GetName()

	w, err := client.Resource(gvr).Namespace(namespace).Watch(
		ctx,
		metav1.ListOptions{
			FieldSelector:  fmt.Sprintf("metadata.name=%s", applicationName),
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

			log.Infof("Event: %s, sync=%s, health=%s\n", evt.Type, syncStatus, healthStatus)

			if syncStatus == "OutOfSync" {
				seenOutOfSync = true
			}

			if seenOutOfSync && syncStatus == "Synced" && healthStatus == "Healthy" {
				log.Infof("Application reached Synced & Healthy after OutOfSync.")

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
