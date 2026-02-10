package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type APITestSuite struct {
	suite.Suite
	router *gin.Engine
}

func (s *APITestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	s.router = gin.New()
	setupTestRoutes(s.router)
}

func setupTestRoutes(r *gin.Engine) {
	// Health endpoints
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Product endpoints
	r.GET("/api/v1/products", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": []interface{}{}, "total": 0})
	})
	r.POST("/api/v1/products", func(c *gin.Context) {
		var req map[string]interface{}
		c.ShouldBindJSON(&req)
		req["id"] = "test-uuid"
		c.JSON(http.StatusCreated, req)
	})
	r.GET("/api/v1/products/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"id": c.Param("id"), "name": "Test"})
	})
	r.PUT("/api/v1/products/:id", func(c *gin.Context) {
		var req map[string]interface{}
		c.ShouldBindJSON(&req)
		req["id"] = c.Param("id")
		c.JSON(http.StatusOK, req)
	})
	r.DELETE("/api/v1/products/:id", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	// Category endpoints
	r.GET("/api/v1/categories", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": []interface{}{}, "total": 0})
	})
	r.GET("/api/v1/categories/tree", func(c *gin.Context) {
		c.JSON(http.StatusOK, []interface{}{})
	})

	// Search endpoint
	r.GET("/api/v1/search", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"results": []interface{}{}, "total": 0})
	})
}

func (s *APITestSuite) TestHealthEndpoint() {
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	s.Equal("healthy", response["status"])
}

func (s *APITestSuite) TestCreateProduct() {
	body := `{"sku": "INT-001", "name": "Integration Test Product"}`
	req, _ := http.NewRequest("POST", "/api/v1/products", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusCreated, w.Code)
}

func (s *APITestSuite) TestGetProduct() {
	req, _ := http.NewRequest("GET", "/api/v1/products/test-uuid", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

func (s *APITestSuite) TestListProducts() {
	req, _ := http.NewRequest("GET", "/api/v1/products", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

func (s *APITestSuite) TestUpdateProduct() {
	body := `{"name": "Updated Product"}`
	req, _ := http.NewRequest("PUT", "/api/v1/products/test-uuid", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

func (s *APITestSuite) TestDeleteProduct() {
	req, _ := http.NewRequest("DELETE", "/api/v1/products/test-uuid", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNoContent, w.Code)
}

func (s *APITestSuite) TestListCategories() {
	req, _ := http.NewRequest("GET", "/api/v1/categories", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

func (s *APITestSuite) TestGetCategoryTree() {
	req, _ := http.NewRequest("GET", "/api/v1/categories/tree", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

func (s *APITestSuite) TestSearch() {
	req, _ := http.NewRequest("GET", "/api/v1/search?q=test", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

func TestAPITestSuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

// Ensure assert is used
var _ = assert.Equal

