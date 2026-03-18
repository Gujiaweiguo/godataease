package handler

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPermissionCompatHandler_TargetPermissionEndpointsReturnExplicitNonSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &PermissionCompatHandler{}

	tests := []struct {
		name string
		path string
		body string
	}{
		{name: "menu target permission", path: "/auth/menuTargetPermission", body: `{"roleId":1}`},
		{name: "business target permission", path: "/auth/busiTargetPermission", body: `{"roleId":1}`},
		{name: "save menu target permission", path: "/auth/saveMenuTargetPer", body: `{"roleId":1,"targetPerms":[]}`},
		{name: "save business target permission", path: "/auth/saveBusiTargetPer", body: `{"roleId":1,"targetPerms":[]}`},
	}

	r := gin.New()
	api := r.Group("/api")
	RegisterPermissionCompatRoutes(api, h)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api"+tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Fatalf("expected status 200, got %d", w.Code)
			}

			var resp map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal response failed: %v", err)
			}
			if resp["code"] != "501000" {
				t.Fatalf("expected code 501000, got %#v", resp["code"])
			}
		})
	}
}
