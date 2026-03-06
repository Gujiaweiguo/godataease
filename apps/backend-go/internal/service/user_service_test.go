package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserService_ResolveDefaultPassword(t *testing.T) {
	svc := &UserService{}

	t.Setenv(DefaultPasswordEnvName, "")
	assert.Equal(t, FallbackDefaultPwd, svc.ResolveDefaultPassword())

	t.Setenv(DefaultPasswordEnvName, "custom-password")
	assert.Equal(t, "custom-password", svc.ResolveDefaultPassword())
}
