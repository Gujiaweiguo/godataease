package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := map[string]string{"key": "value"}
	Success(c, data)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Code != "000000" {
		t.Errorf("Expected code '000000', got '%s'", resp.Code)
	}
	if resp.Message != "success" {
		t.Errorf("Expected message 'success', got '%s'", resp.Message)
	}
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Error(c, "10001", "test error")

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Code != "10001" {
		t.Errorf("Expected code '10001', got '%s'", resp.Code)
	}
	if resp.Message != "test error" {
		t.Errorf("Expected message 'test error', got '%s'", resp.Message)
	}
}

func TestErrorWithData(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ErrorWithData(c, "10002", "error with data", map[string]int{"count": 5})

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Code != "10002" {
		t.Errorf("Expected code '10002', got '%s'", resp.Code)
	}
	if resp.Data == nil {
		t.Error("Expected data to be present")
	}
}

func TestBadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	BadRequest(c, "invalid request")

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Code != "10001" {
		t.Errorf("Expected code '10001', got '%s'", resp.Code)
	}
}

func TestUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Unauthorized(c, "unauthorized access")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Code != "20001" {
		t.Errorf("Expected code '20001', got '%s'", resp.Code)
	}
}

func TestForbidden(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Forbidden(c, "forbidden access")

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Code != "70001" {
		t.Errorf("Expected code '70001', got '%s'", resp.Code)
	}
}

func TestNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	NotFound(c, "resource not found")

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Code != "50001" {
		t.Errorf("Expected code '50001', got '%s'", resp.Code)
	}
}

func TestForbiddenExport(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ForbiddenExport(c, "export forbidden")

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Code != "403001" {
		t.Errorf("Expected code '403001', got '%s'", resp.Code)
	}
}

func TestNotFoundExport(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	NotFoundExport(c, "export not found")

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Code != "404001" {
		t.Errorf("Expected code '404001', got '%s'", resp.Code)
	}
}

func TestInternalError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	InternalError(c, "internal server error")

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Code != "40001" {
		t.Errorf("Expected code '40001', got '%s'", resp.Code)
	}
}
