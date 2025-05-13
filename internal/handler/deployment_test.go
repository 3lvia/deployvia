package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPostDeployment(t *testing.T) {
	router := setupRouter()

	// Define the route for the POST request
	router.POST("/deployment", func(c *gin.Context) {
		// Simulate the deployment logic here
		c.JSON(200, gin.H{"status": "success"})
	})

	// Create a new HTTP request
	req, err := http.NewRequest("POST", "/deployment", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new HTTP recorder to capture the response
	w := httptest.NewRecorder()

	// Serve the HTTP request using the router
	router.ServeHTTP(w, req)

	// Check the response status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}

	// Check the response body
	expectedBody := `{"status":"success"}`
	if w.Body.String() != expectedBody {
		t.Errorf("Expected body %s, got %s", expectedBody, w.Body.String())
	}
}
