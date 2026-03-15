package service

import (
	"encoding/base64"
	"testing"

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
