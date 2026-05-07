package visualization

import (
	"encoding/json"
	"testing"

	"github.com/gin-gonic/gin/binding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlexInt_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("from number", func(t *testing.T) {
		t.Parallel()

		var got FlexInt
		require.NoError(t, json.Unmarshal([]byte(`12`), &got))
		assert.Equal(t, int64(12), got.Int64())
	})

	t.Run("from quoted string", func(t *testing.T) {
		t.Parallel()

		var got FlexInt
		require.NoError(t, json.Unmarshal([]byte(`"34"`), &got))
		assert.Equal(t, int64(34), got.Int64())
	})

	t.Run("returns error for invalid json", func(t *testing.T) {
		t.Parallel()

		var got FlexInt
		err := json.Unmarshal([]byte(`"bad"`), &got)
		require.Error(t, err)
	})
}

func TestSaveRequest_JSONBindingAndValidation(t *testing.T) {
	t.Parallel()

	t.Run("binds json payload", func(t *testing.T) {
		t.Parallel()

		var req SaveRequest
		require.NoError(t, json.Unmarshal([]byte(`{"name":"Dashboard","pid":9,"type":"dashboard"}`), &req))
		require.NotNil(t, req.PID)
		require.NotNil(t, req.Type)
		assert.Equal(t, "Dashboard", req.Name)
		assert.Equal(t, int64(9), *req.PID)
		assert.Equal(t, "dashboard", *req.Type)
	})

	t.Run("validates required name", func(t *testing.T) {
		t.Parallel()

		validator := binding.Validator
		require.NotNil(t, validator)
		err := validator.ValidateStruct(&SaveRequest{})
		require.Error(t, err)
	})
}

func TestDecompressionRequest_JSONFieldMapping(t *testing.T) {
	t.Parallel()

	var req DecompressionRequest
	require.NoError(t, json.Unmarshal([]byte(`{
		"newFrom":"new_outer_template",
		"templateId":8,
		"resourceName":"Template",
		"templateUrl":"https://example.com/template",
		"name":"Imported",
		"type":"dashboard",
		"version":5,
		"canvasStyleData":"{}",
		"componentData":"[]",
		"dynamicData":"{}",
		"appData":"{\"id\":1}",
		"staticResource":"{\"css\":[]}"}`), &req))

	require.NotNil(t, req.TemplateID)
	assert.Equal(t, "new_outer_template", req.NewFrom)
	assert.Equal(t, int64(8), *req.TemplateID)
	assert.Equal(t, "Template", req.ResourceName)
	assert.Equal(t, "https://example.com/template", req.TemplateURL)
	assert.Equal(t, "Imported", req.Name)
	assert.Equal(t, "dashboard", req.Type)
	assert.Equal(t, 5, req.Version)
	assert.Equal(t, "{}", req.CanvasStyleData)
	assert.Equal(t, "[]", req.ComponentData)
	assert.Equal(t, "{}", req.DynamicData)
	assert.Equal(t, `{"id":1}`, req.AppData)
	assert.Equal(t, `{"css":[]}`, req.StaticResource)
}
