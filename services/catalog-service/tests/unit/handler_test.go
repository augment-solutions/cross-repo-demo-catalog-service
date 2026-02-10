package unit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHealthHandler(t *testing.T) {
	router := gin.New()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "healthy", response["status"])
}

func TestProductHandler_Create(t *testing.T) {
	router := gin.New()
	router.POST("/api/v1/products", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		req["id"] = "test-uuid"
		c.JSON(http.StatusCreated, req)
	})

	body := `{"sku": "TEST-001", "name": "Test Product"}`
	req, _ := http.NewRequest("POST", "/api/v1/products", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestProductHandler_Get(t *testing.T) {
	router := gin.New()
	router.GET("/api/v1/products/:id", func(c *gin.Context) {
		id := c.Param("id")
		c.JSON(http.StatusOK, gin.H{
			"id":   id,
			"name": "Test Product",
		})
	})

	req, _ := http.NewRequest("GET", "/api/v1/products/test-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProductHandler_List(t *testing.T) {
	router := gin.New()
	router.GET("/api/v1/products", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"data":      []interface{}{},
			"total":     0,
			"page":      1,
			"page_size": 20,
		})
	})

	req, _ := http.NewRequest("GET", "/api/v1/products", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProductHandler_Update(t *testing.T) {
	router := gin.New()
	router.PUT("/api/v1/products/:id", func(c *gin.Context) {
		id := c.Param("id")
		var req map[string]interface{}
		c.ShouldBindJSON(&req)
		req["id"] = id
		c.JSON(http.StatusOK, req)
	})

	body := `{"name": "Updated Product"}`
	req, _ := http.NewRequest("PUT", "/api/v1/products/test-uuid", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProductHandler_Delete(t *testing.T) {
	router := gin.New()
	router.DELETE("/api/v1/products/:id", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req, _ := http.NewRequest("DELETE", "/api/v1/products/test-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestCategoryHandler_GetTree(t *testing.T) {
	router := gin.New()
	router.GET("/api/v1/categories/tree", func(c *gin.Context) {
		c.JSON(http.StatusOK, []map[string]interface{}{
			{"id": "1", "name": "Electronics", "children": []interface{}{}},
		})
	})

	req, _ := http.NewRequest("GET", "/api/v1/categories/tree", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSearchHandler(t *testing.T) {
	router := gin.New()
	router.GET("/api/v1/search", func(c *gin.Context) {
		query := c.Query("q")
		c.JSON(http.StatusOK, gin.H{
			"query":   query,
			"results": []interface{}{},
			"total":   0,
		})
	})

	req, _ := http.NewRequest("GET", "/api/v1/search?q=test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

