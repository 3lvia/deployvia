package handler_test

import (
    "context"
    "net/http"
    "encoding/json"
    "bytes"
    "net/http/httptest"
	"testing"
    "github.com/gin-gonic/gin"

	"github.com/3lvia/deployvia/internal/config"
	"github.com/3lvia/deployvia/internal/route"
	"github.com/3lvia/deployvia/internal/model"
)

func SetupTestEnvironment(t *testing.T) *gin.Engine {
    ctx := context.Background()

    t.Setenv("LOCAL", "true")

    config, err := config.New(ctx)
    if err != nil {
        t.Fatalf("Failed to load config: %v", err)
    }

    router := route.SetupRouter(config)
    route.RegisterRoutes(ctx, router, config)

    return router
}

func TestPostDeploymentNoBody(t *testing.T) {
    router := SetupTestEnvironment(t)

    // Create a new HTTP request to the /deployment endpoint
    req, err := http.NewRequest("POST", "/deployment", nil)
    if err != nil {
        t.Fatalf("Failed to create request: %v", err)
    }

    rr := httptest.NewRecorder()
    router.ServeHTTP(rr, req)

    expectedStatus := http.StatusBadRequest
    if status := rr.Code; status != expectedStatus {
        t.Errorf("Handler returned wrong status code: got %v want %v", status, expectedStatus)
    }

    expected := `{"error":"invalid deployment: invalid request"}`
    if rr.Body.String() != expected {
        t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
    }

    contentType := rr.Header().Get("Content-Type")
    expectedContentType := "application/json; charset=utf-8"
    if contentType != expectedContentType {
        t.Errorf("Handler returned wrong content type: got %v want %v", contentType, expectedContentType)
    }
}

func TestPostDeployment(t *testing.T) {
    router := SetupTestEnvironment(t)

    deployment := &model.Deployment{
        ApplicationName: "demo-api-go",
        System: "core",
        ClusterType: "aks",
        Environment: "dev",
        Image: "asdf", // not used
    }

    body, err := json.Marshal(deployment)
    if err != nil {
        t.Fatalf("Failed to marshal deployment: %v", err)
    }

    // Create a new HTTP request to the /deployment endpoint
    req, err := http.NewRequest("POST", "/deployment", bytes.NewBuffer(body))
    if err != nil {
        t.Fatalf("Failed to create request: %v", err)
    }

    rr := httptest.NewRecorder()
    router.ServeHTTP(rr, req)

    expectedStatus := http.StatusInternalServerError
    if status := rr.Code; status != expectedStatus {
        t.Errorf("Handler returned wrong status code: got %v want %v", status, expectedStatus)
    }

    expected := `{"error":"failed to watch application lifecycle: application not found"}`
    if rr.Body.String() != expected {
        t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
    }

    contentType := rr.Header().Get("Content-Type")
    expectedContentType := "application/json; charset=utf-8"
    if contentType != expectedContentType {
        t.Errorf("Handler returned wrong content type: got %v want %v", contentType, expectedContentType)
    }
}

func TestPostDeploymentNoToken(t *testing.T) {
    router := SetupTestEnvironment(t)

    t.Setenv("TESTING_ENABLE_OIDC", "true")

    deployment := &model.Deployment{
        ApplicationName: "demo-api-go",
        System: "core",
        ClusterType: "aks",
        Environment: "dev",
        Image: "asdf", // not used
    }

    body, err := json.Marshal(deployment)
    if err != nil {
        t.Fatalf("Failed to marshal deployment: %v", err)
    }

    // Create a new HTTP request to the /deployment endpoint
    req, err := http.NewRequest("POST", "/deployment", bytes.NewBuffer(body))
    if err != nil {
        t.Fatalf("Failed to create request: %v", err)
    }

    rr := httptest.NewRecorder()
    router.ServeHTTP(rr, req)

    expectedStatus := http.StatusBadRequest
    if status := rr.Code; status != expectedStatus {
        t.Errorf("Handler returned wrong status code: got %v want %v", status, expectedStatus)
    }

    expected := `{"error":"X-GitHub-OIDC-Token header is required"}`
    if rr.Body.String() != expected {
        t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
    }

    contentType := rr.Header().Get("Content-Type")
    expectedContentType := "application/json; charset=utf-8"
    if contentType != expectedContentType {
        t.Errorf("Handler returned wrong content type: got %v want %v", contentType, expectedContentType)
    }
}

func TestPostDeploymentInvalidToken(t *testing.T) {
    router := SetupTestEnvironment(t)

    t.Setenv("TESTING_ENABLE_OIDC", "true")

    deployment := &model.Deployment{
        ApplicationName: "demo-api-go",
        System: "core",
        ClusterType: "aks",
        Environment: "dev",
        Image: "asdf", // not used
    }

    body, err := json.Marshal(deployment)
    if err != nil {
        t.Fatalf("Failed to marshal deployment: %v", err)
    }

    // Create a new HTTP request to the /deployment endpoint
    req, err := http.NewRequest("POST", "/deployment", bytes.NewBuffer(body))
    if err != nil {
        t.Fatalf("Failed to create request: %v", err)
    }

    req.Header.Add("X-GitHub-OIDC-Token", "invalid-token")

    rr := httptest.NewRecorder()
    router.ServeHTTP(rr, req)

    expectedStatus := http.StatusForbidden
    if status := rr.Code; status != expectedStatus {
        t.Errorf("Handler returned wrong status code: got %v want %v", status, expectedStatus)
    }

    expected := `{"error":"invalid token: failed to verify token: token is malformed: token contains an invalid number of segments"}`
    if rr.Body.String() != expected {
        t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
    }

    contentType := rr.Header().Get("Content-Type")
    expectedContentType := "application/json; charset=utf-8"
    if contentType != expectedContentType {
        t.Errorf("Handler returned wrong content type: got %v want %v", contentType, expectedContentType)
    }
}
