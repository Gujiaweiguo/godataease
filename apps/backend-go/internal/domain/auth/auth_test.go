package auth

import (
	"testing"
)

func TestPwdLoginDTO_Fields(t *testing.T) {
	dto := PwdLoginDTO{
		Name:   "admin",
		Pwd:    "password123",
		Origin: 0,
	}

	if dto.Name != "admin" {
		t.Errorf("Expected Name 'admin', got '%s'", dto.Name)
	}
	if dto.Pwd != "password123" {
		t.Errorf("Expected Pwd 'password123', got '%s'", dto.Pwd)
	}
	if dto.Origin != 0 {
		t.Errorf("Expected Origin 0, got %d", dto.Origin)
	}
}

func TestTokenVO_Fields(t *testing.T) {
	vo := TokenVO{
		Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		Exp:   1700000000,
	}

	if vo.Token == "" {
		t.Error("Expected Token to be non-empty")
	}
	if vo.Exp != 1700000000 {
		t.Errorf("Expected Exp 1700000000, got %d", vo.Exp)
	}
}

func TestTokenClaims_Fields(t *testing.T) {
	claims := TokenClaims{
		Uid: 1,
		Oid: 1,
	}

	if claims.Uid != 1 {
		t.Errorf("Expected Uid 1, got %d", claims.Uid)
	}
	if claims.Oid != 1 {
		t.Errorf("Expected Oid 1, got %d", claims.Oid)
	}
}

func TestDefaultLoginConfig(t *testing.T) {
	cfg := DefaultLoginConfig()

	if cfg == nil {
		t.Fatal("Expected LoginConfig, got nil")
	}
	if cfg.AdminUsername != "admin" {
		t.Errorf("Expected AdminUsername 'admin', got '%s'", cfg.AdminUsername)
	}
	if cfg.AdminPasswordEnv != "ADMIN_PASSWORD" {
		t.Errorf("Expected AdminPasswordEnv 'ADMIN_PASSWORD', got '%s'", cfg.AdminPasswordEnv)
	}
	if cfg.DefaultAdminPassword != "dataease" {
		t.Errorf("Expected DefaultAdminPassword 'dataease', got '%s'", cfg.DefaultAdminPassword)
	}
}

func TestLoginConfig_Fields(t *testing.T) {
	cfg := &LoginConfig{
		AdminUsername:        "superadmin",
		AdminPasswordEnv:     "SUPER_ADMIN_PASSWORD",
		DefaultAdminPassword: "supersecret",
	}

	if cfg.AdminUsername != "superadmin" {
		t.Errorf("Expected AdminUsername 'superadmin', got '%s'", cfg.AdminUsername)
	}
}
