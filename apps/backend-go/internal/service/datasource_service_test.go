package service

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"dataease/backend/internal/domain/datasource"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatasourceServiceHelpers_DecodeMaybeBase64JSONMap(t *testing.T) {
	raw := base64.StdEncoding.EncodeToString([]byte(`{"a":1}`))
	result, err := decodeMaybeBase64JSONMap(raw)
	require.NoError(t, err)
	assert.Equal(t, float64(1), result["a"])

	result, err = decodeMaybeBase64JSONMap(`{"b":2}`)
	require.NoError(t, err)
	assert.Equal(t, float64(2), result["b"])

	_, err = decodeMaybeBase64JSONMap("")
	assert.Error(t, err)
}

func TestDatasourceServiceHelpers_ParseIDs(t *testing.T) {
	id, err := parseDatasourceID(map[string]string{"datasourceId": "12"})
	require.NoError(t, err)
	assert.Equal(t, int64(12), id)

	_, err = parseDatasourceID(map[string]string{})
	assert.Error(t, err)

	_, err = parseDatasourceID(map[string]string{"id": "bad"})
	assert.Error(t, err)

	assert.Equal(t, int64(99), parseTaskID("99"))
	assert.Equal(t, int64(0), parseTaskID("bad"))
}

func TestDatasourceService_Validate(t *testing.T) {
	svc := NewDatasourceService(nil)

	t.Run("missing host and port returns error status", func(t *testing.T) {
		dsType := "mysql"
		cfgJSON, err := json.Marshal(&datasource.ConnectionConfig{})
		require.NoError(t, err)
		cfgStr := base64.StdEncoding.EncodeToString(cfgJSON)

		resp, err := svc.Validate(&datasource.ValidateRequest{
			Type:          &dsType,
			Configuration: &cfgStr,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, datasource.StatusError, resp.Status)
		assert.Contains(t, resp.Message, "missing host/port")
	})

	t.Run("unreachable host returns connectivity error status", func(t *testing.T) {
		dsType := "mysql"
		cfgJSON, err := json.Marshal(&datasource.ConnectionConfig{Host: "198.51.100.1", Port: 81})
		require.NoError(t, err)
		cfgStr := base64.StdEncoding.EncodeToString(cfgJSON)

		resp, err := svc.Validate(&datasource.ValidateRequest{
			Type:          &dsType,
			Configuration: &cfgStr,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, datasource.StatusError, resp.Status)
		assert.Contains(t, resp.Message, "failed to connect")
	})
}

func TestDatasourceServiceHelpers_PingTCPTimeout(t *testing.T) {
	err := pingTCP("198.51.100.1", 81, time.Millisecond)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect")
}
