package menu

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSON_ValueAndScan(t *testing.T) {
	t.Run("value nil returns nil", func(t *testing.T) {
		var payload JSON

		value, err := payload.Value()
		require.NoError(t, err)
		assert.Nil(t, value)
	})

	t.Run("value marshals map", func(t *testing.T) {
		payload := JSON{"name": "dashboard", "enabled": true}

		value, err := payload.Value()
		require.NoError(t, err)
		assert.JSONEq(t, `{"enabled":true,"name":"dashboard"}`, string(value.([]byte)))
	})

	t.Run("scan nil resets map", func(t *testing.T) {
		payload := JSON{"legacy": true}

		err := (&payload).Scan(nil)
		require.NoError(t, err)
		assert.Nil(t, payload)
	})

	t.Run("scan ignores unsupported type", func(t *testing.T) {
		payload := JSON{"keep": "value"}

		err := (&payload).Scan("not-bytes")
		require.NoError(t, err)
		assert.Equal(t, JSON{"keep": "value"}, payload)
	})

	t.Run("scan unmarshals bytes", func(t *testing.T) {
		var payload JSON

		err := (&payload).Scan([]byte(`{"action":"view","priority":1}`))
		require.NoError(t, err)
		assert.Equal(t, "view", payload["action"])
		assert.Equal(t, float64(1), payload["priority"])
	})

	t.Run("scan invalid json returns error", func(t *testing.T) {
		var payload JSON

		err := (&payload).Scan([]byte(`{"broken":`))
		require.Error(t, err)
	})
}
