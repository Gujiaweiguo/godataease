package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	pkgauth "dataease/backend/internal/pkg/auth"

	"github.com/gin-gonic/gin"
)

func TestAuthMiddleware_AcceptsXDETokenHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	SetRoleIDsResolver(nil)
	jwtInstance := pkgauth.NewJWT(&pkgauth.JWTConfig{Secret: "test-secret", Expire: 3600})
	token, err := jwtInstance.GenerateTokenWithOrgID(42, "tester", "", 7)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	r := gin.New()
	r.GET("/protected", Auth(jwtInstance), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"user_id": GetUserID(c), "org_id": GetOrgID(c)})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("X-DE-TOKEN", token)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200 with X-DE-TOKEN, got %d body=%s", resp.Code, resp.Body.String())
	}
	if resp.Body.String() != "{\"org_id\":7,\"user_id\":42}" {
		t.Fatalf("expected org-aware auth context, got %s", resp.Body.String())
	}
}

func TestAuthMiddleware_PopulatesRoleIDsWhenResolverConfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)
	SetRoleIDsResolver(func(userID uint64) ([]int64, error) {
		if userID != 42 {
			t.Fatalf("expected user id 42, got %d", userID)
		}
		return []int64{2, 3}, nil
	})
	t.Cleanup(func() {
		SetRoleIDsResolver(nil)
	})

	jwtInstance := pkgauth.NewJWT(&pkgauth.JWTConfig{Secret: "test-secret", Expire: 3600})
	token, err := jwtInstance.GenerateTokenWithOrgID(42, "tester", "", 7)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	r := gin.New()
	r.GET("/protected", Auth(jwtInstance), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"role_ids": c.MustGet("role_ids")})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("X-DE-TOKEN", token)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	var body struct {
		RoleIDs []int64 `json:"role_ids"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if len(body.RoleIDs) != 2 || body.RoleIDs[0] != 2 || body.RoleIDs[1] != 3 {
		t.Fatalf("expected role ids [2 3], got %#v", body.RoleIDs)
	}
}
