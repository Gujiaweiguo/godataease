package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	pkgauth "dataease/backend/internal/pkg/auth"

	"github.com/gin-gonic/gin"
)

func TestAuthMiddleware_AcceptsXDETokenHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
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
