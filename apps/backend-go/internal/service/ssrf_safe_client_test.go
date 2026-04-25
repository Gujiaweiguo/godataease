package service

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsBlockedIP(t *testing.T) {
	blocked := []string{
		"127.0.0.1",
		"::1",
		"0.0.0.0",
		"10.0.0.1",
		"172.16.0.1",
		"192.168.1.1",
		"169.254.169.254",
		"fc00::1",
		"fe80::1",
	}
	for _, raw := range blocked {
		ip := net.ParseIP(raw)
		require.NotNil(t, ip)
		assert.True(t, isBlockedIP(ip), raw)
	}

	allowed := []string{"8.8.8.8", "1.1.1.1", "2606:4700:4700::1111"}
	for _, raw := range allowed {
		ip := net.ParseIP(raw)
		require.NotNil(t, ip)
		assert.False(t, isBlockedIP(ip), raw)
	}
}

func TestValidateMarketTemplateURL(t *testing.T) {
	t.Run("rejects invalid scheme", func(t *testing.T) {
		err := validateMarketTemplateURL("file:///etc/passwd")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "scheme")
	})

	t.Run("rejects blocked host", func(t *testing.T) {
		err := validateMarketTemplateURL("http://127.0.0.1:8080/template.json")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "blocked")
	})

	t.Run("rejects missing host", func(t *testing.T) {
		err := validateMarketTemplateURL("https:///template.json")
		require.Error(t, err)
	})
}
