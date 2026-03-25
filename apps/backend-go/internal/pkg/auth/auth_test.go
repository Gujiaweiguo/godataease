package auth

import (
	"testing"
	"time"
)

func TestNewJWT(t *testing.T) {
	config := &JWTConfig{
		Secret: "test-secret",
		Expire: 3600,
	}

	jwt := NewJWT(config)

	if jwt == nil {
		t.Fatal("Expected JWT instance, got nil")
	}
	if jwt.config != config {
		t.Error("Config not set correctly")
	}
}

func TestGenerateAndParseToken(t *testing.T) {
	config := &JWTConfig{
		Secret: "test-secret-key-for-testing",
		Expire: 3600,
	}
	jwt := NewJWT(config)

	token, err := jwt.GenerateToken(1, "testuser", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	if token == "" {
		t.Error("Expected non-empty token")
	}

	claims, err := jwt.ParseToken(token)
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	if claims.UserID != 1 {
		t.Errorf("Expected UserID 1, got %d", claims.UserID)
	}
	if claims.Username != "testuser" {
		t.Errorf("Expected Username 'testuser', got '%s'", claims.Username)
	}
	if claims.Role != "admin" {
		t.Errorf("Expected Role 'admin', got '%s'", claims.Role)
	}
	if claims.OrgID != 0 {
		t.Errorf("Expected OrgID 0, got %d", claims.OrgID)
	}
}

func TestGenerateAndParseToken_WithOrgID(t *testing.T) {
	jwt := NewJWT(&JWTConfig{Secret: "org-secret", Expire: 3600})
	token, err := jwt.GenerateTokenWithOrgID(8, "orguser", "user", 5)
	if err != nil {
		t.Fatalf("Failed to generate token with org: %v", err)
	}
	claims, err := jwt.ParseToken(token)
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}
	if claims.OrgID != 5 {
		t.Fatalf("Expected OrgID 5, got %d", claims.OrgID)
	}
}

func TestParseToken_InvalidToken(t *testing.T) {
	config := &JWTConfig{
		Secret: "test-secret",
		Expire: 3600,
	}
	jwt := NewJWT(config)

	_, err := jwt.ParseToken("invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	config1 := &JWTConfig{
		Secret: "secret-1",
		Expire: 3600,
	}
	jwt1 := NewJWT(config1)

	token, _ := jwt1.GenerateToken(1, "user", "role")

	config2 := &JWTConfig{
		Secret: "secret-2",
		Expire: 3600,
	}
	jwt2 := NewJWT(config2)

	_, err := jwt2.ParseToken(token)
	if err == nil {
		t.Error("Expected error when parsing token with wrong secret")
	}
}

func TestRefreshToken(t *testing.T) {
	config := &JWTConfig{
		Secret: "refresh-test-secret",
		Expire: 3600,
	}
	jwt := NewJWT(config)

	originalToken, _ := jwt.GenerateToken(42, "refreshuser", "user")

	newToken, err := jwt.RefreshToken(originalToken)
	if err != nil {
		t.Fatalf("Failed to refresh token: %v", err)
	}
	if newToken == "" {
		t.Error("Expected non-empty refreshed token")
	}

	claims, _ := jwt.ParseToken(newToken)
	if claims.UserID != 42 {
		t.Errorf("Expected UserID 42, got %d", claims.UserID)
	}
	if claims.Username != "refreshuser" {
		t.Errorf("Expected Username 'refreshuser', got '%s'", claims.Username)
	}
	if claims.OrgID != 0 {
		t.Errorf("Expected OrgID 0 after refresh, got %d", claims.OrgID)
	}
}

func TestRefreshToken_PreservesOrgID(t *testing.T) {
	jwt := NewJWT(&JWTConfig{Secret: "refresh-org-secret", Expire: 3600})
	originalToken, _ := jwt.GenerateTokenWithOrgID(42, "refreshuser", "user", 9)
	newToken, err := jwt.RefreshToken(originalToken)
	if err != nil {
		t.Fatalf("Failed to refresh token: %v", err)
	}
	claims, _ := jwt.ParseToken(newToken)
	if claims.OrgID != 9 {
		t.Fatalf("Expected OrgID 9 after refresh, got %d", claims.OrgID)
	}
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	config := &JWTConfig{
		Secret: "test-secret",
		Expire: 3600,
	}
	jwt := NewJWT(config)

	_, err := jwt.RefreshToken("invalid-token")
	if err == nil {
		t.Error("Expected error when refreshing invalid token")
	}
}

func TestHashPassword(t *testing.T) {
	password := "mypassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	if hash == "" {
		t.Error("Expected non-empty hash")
	}
	if hash == password {
		t.Error("Hash should not equal plain password")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "testpassword"

	hash, _ := HashPassword(password)

	if !CheckPassword(password, hash) {
		t.Error("Expected password check to succeed")
	}

	if CheckPassword("wrongpassword", hash) {
		t.Error("Expected password check to fail for wrong password")
	}
}

func TestCheckPassword_EmptyPassword(t *testing.T) {
	if CheckPassword("", "somehash") {
		t.Error("Expected empty password check to fail")
	}
}

func TestClaimsStructure(t *testing.T) {
	config := &JWTConfig{
		Secret: "claims-test-secret",
		Expire: 3600,
	}
	jwt := NewJWT(config)

	token, _ := jwt.GenerateToken(123, "claimsuser", "superadmin")
	claims, _ := jwt.ParseToken(token)

	if claims.UserID != 123 {
		t.Errorf("UserID mismatch: expected 123, got %d", claims.UserID)
	}
	if claims.Username != "claimsuser" {
		t.Errorf("Username mismatch: expected 'claimsuser', got '%s'", claims.Username)
	}
	if claims.Role != "superadmin" {
		t.Errorf("Role mismatch: expected 'superadmin', got '%s'", claims.Role)
	}
	if claims.OrgID != 0 {
		t.Errorf("OrgID mismatch: expected 0, got %d", claims.OrgID)
	}

	if claims.IssuedAt == nil {
		t.Error("IssuedAt should not be nil")
	}
	if claims.ExpiresAt == nil {
		t.Error("ExpiresAt should not be nil")
	}
	if claims.NotBefore == nil {
		t.Error("NotBefore should not be nil")
	}

	if claims.ExpiresAt.Before(time.Now()) {
		t.Error("Token should not be expired immediately after creation")
	}
}

func TestErrorVariables(t *testing.T) {
	if ErrTokenExpired == nil {
		t.Error("ErrTokenExpired should not be nil")
	}
	if ErrTokenInvalid == nil {
		t.Error("ErrTokenInvalid should not be nil")
	}
	if ErrTokenMalformed == nil {
		t.Error("ErrTokenMalformed should not be nil")
	}
	if ErrTokenNotValidYet == nil {
		t.Error("ErrTokenNotValidYet should not be nil")
	}
}
