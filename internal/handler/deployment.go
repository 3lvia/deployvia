package handler

import (
	"context"
	"fmt"
	"os"
	"sync"
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

	err = watchApplicationsLifecycle(
		ctx,
		config.KubernetesClient,
		gvr,
		"argocd",
		validatedDeployment,
		timeout,
	)
	if err != nil {
		log.Error(err)
		c.JSON(500, gin.H{"error": err.Error()})

		return
	}

	c.JSON(200, gin.H{"message": "Application successfully deployed!"})
}

func watchApplicationsLifecycle(
	ctx context.Context,
	client dynamic.Interface,
	gvr schema.GroupVersionResource,
	namespace string,
	validatedDeployment *model.ValidatedDeployment,
	timeout time.Duration,
) error {
	applications, err := client.Resource(gvr).Namespace(namespace).List(
		ctx,
		metav1.ListOptions{
			LabelSelector: getLabelSelector(validatedDeployment),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to get application for deployment: %w", err)
	}

	if len(applications.Items) == 0 {
		return fmt.Errorf("application(s) not found")
	}

	if len(applications.Items) > 1 && !validatedDeployment.Deployment.CheckAllClusters {
		return fmt.Errorf("multiple applications found when only one was expected")
	}

	var applicationNames []string
	for _, application := range applications.Items {
		name, found, err := unstructured.NestedString(application.Object, "metadata", "name")
		if err != nil {
			return fmt.Errorf("failed to get application name: %w", err)
		}

		if found {
			applicationNames = append(applicationNames, name)
		}
	}

	var (
		wg    sync.WaitGroup
		errCh = make(chan error, len(applicationNames)) // buffered to avoid goroutine leaks
	)

	for _, applicationName := range applicationNames {
		wg.Add(1)
		appName := applicationName // avoid loop variable capture issue

		go func() {
			defer wg.Done()
			err := watchApplicationLifecycle(
				ctx,
				client,
				gvr,
				namespace,
				validatedDeployment,
				timeout,
				appName,
			)
			if err != nil {
				errCh <- fmt.Errorf("failed to watch %s: %w", appName, err)
			}
		}()
	}

	wg.Wait()
	close(errCh)

	// Check if any errors occurred
	var combinedErr error
	for err := range errCh {
		if combinedErr == nil {
			combinedErr = err
		} else {
			combinedErr = fmt.Errorf("%v; %w", combinedErr, err)
		}
	}

	if combinedErr != nil {
		return combinedErr
	}

	return nil
}

func watchApplicationLifecycle(
	ctx context.Context,
	client dynamic.Interface,
	gvr schema.GroupVersionResource,
	namespace string,
	validatedDeployment *model.ValidatedDeployment,
	timeout time.Duration,
	applicationName string,
) error {
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

			system, found, err := unstructured.NestedString(obj.Object, "metadata", "labels", "elvia.no/system")
			if err != nil || !found {
				return fmt.Errorf("failed to get system label: %w", err)
			}

			name, found, err := unstructured.NestedString(obj.Object, "metadata", "labels", "elvia.no/application")
			if err != nil || !found {
				return fmt.Errorf("failed to get application label: %w", err)
			}

			environment, found, err := unstructured.NestedString(obj.Object, "metadata", "labels", "kubernetes.io/environment")
			if err != nil || !found {
				return fmt.Errorf("failed to get environment label: %w", err)
			}

			clusterType, found, err := unstructured.NestedString(obj.Object, "metadata", "labels", "elvia.no/cluster-type")
			if err != nil || !found {
				return fmt.Errorf("failed to get cluster-type label: %w", err)
			}

			syncStatus, found, err := unstructured.NestedString(obj.Object, "status", "sync", "status")
			if err != nil || !found {
				return fmt.Errorf("failed to get sync status: %w", err)
			}

			healthStatus, found, err := unstructured.NestedString(obj.Object, "status", "health", "status")
			if err != nil || !found {
				return fmt.Errorf("failed to get health status: %w", err)
			}

			currentImages, found, err := unstructured.NestedStringSlice(
				obj.Object,
				"status",
				"summary",
				"images",
			)
			if err != nil || !found {
				return fmt.Errorf("failed to get current images: %w", err)
			}

			log_ := log.WithFields(log.Fields{
				"system":      system,
				"name":        name,
				"environment": environment,
				"clusterType": clusterType,
			})

			if len(currentImages) != 1 {
				return fmt.Errorf("Expected 1 image, got %d", len(currentImages))
			}

			log_.Infof("Event: %s, sync=%s, health=%s\n", evt.Type, syncStatus, healthStatus)
			currentImage := currentImages[0]
			log_.Infof("Current image: %s", currentImage)

			if syncStatus == "Synced" && healthStatus == "Healthy" && currentImage == validatedDeployment.Deployment.Image {
				log_.Infof("Application is synced and healthy with requested image '%s'", currentImage)

				return nil
			}
		case <-time.After(timeout):
			return fmt.Errorf("timed out waiting for application lifecycle")
		}
	}
}

func getLabelSelector(
	validatedDeployment *model.ValidatedDeployment,
) string {
	baseLabelSelector := fmt.Sprintf(
		"elvia.no/system=%s,elvia.no/application=%s,kubernetes.io/environment=%s",
		validatedDeployment.Deployment.System,
		validatedDeployment.Deployment.ApplicationName,
		validatedDeployment.Deployment.Environment,
	)

	if !validatedDeployment.Deployment.CheckAllClusters {
		return fmt.Sprintf(
			"%s,elvia.no/cluster-type=%s",
			baseLabelSelector,
			validatedDeployment.Deployment.ClusterType,
		)
	}

	return baseLabelSelector
}

func int64Ptr(i int64) *int64 {
	return &i
}
