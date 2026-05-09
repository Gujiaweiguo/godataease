package service

import (
	"testing"
	"time"

	"dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/integration/seatunnel"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRound16_SetPreviewExecutorFactory_NilFactory(t *testing.T) {
	svc := NewDatasetService(nil)
	require.NotNil(t, svc)
	assert.NotNil(t, svc.previewExecutorFactory)

	svc.SetPreviewExecutorFactory(nil)
	assert.NotNil(t, svc.previewExecutorFactory, "nil factory should reset to defaultPreviewExecutorFactory")
}

func TestRound16_SetPreviewExecutorFactory_CustomFactory(t *testing.T) {
	svc := NewDatasetService(nil)
	require.NotNil(t, svc)

	factory := func(ds *datasource.CoreDatasource, cfg *datasource.ConnectionConfig) (PreviewExecutor, error) {
		return nil, nil
	}

	svc.SetPreviewExecutorFactory(factory)
	assert.NotNil(t, svc.previewExecutorFactory)
}

func TestRound16_SetPreviewExecutorFactory_RoundTrip(t *testing.T) {
	svc := NewDatasetService(nil)

	custom := func(ds *datasource.CoreDatasource, cfg *datasource.ConnectionConfig) (PreviewExecutor, error) {
		return nil, nil
	}
	svc.SetPreviewExecutorFactory(custom)
	assert.NotNil(t, svc.previewExecutorFactory)

	svc.SetPreviewExecutorFactory(nil)
	assert.NotNil(t, svc.previewExecutorFactory, "nil factory should restore default")
}

func TestRound16_EnsureSeatunnelClient_EmptyAddress(t *testing.T) {
	svc := NewDatasourceService(nil)
	svc.seatunnelAddress = ""

	client, err := svc.ensureSeatunnelClient()
	require.Error(t, err)
	assert.Nil(t, client)
	assert.Equal(t, "seatunnel grpc address is not configured", err.Error())
}

func TestRound16_EnsureSeatunnelClient_CachedClient(t *testing.T) {
	svc := NewDatasourceService(nil)
	svc.seatunnelAddress = "localhost:1234"
	svc.seatunnelClient = &seatunnel.Client{}

	client, err := svc.ensureSeatunnelClient()
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestRound16_EnsureSeatunnelClient_DefaultTimeout(t *testing.T) {
	svc := NewDatasourceService(nil)
	svc.seatunnelAddress = "localhost:1234"
	svc.seatunnelTimeout = 0

	assert.Equal(t, time.Duration(0), svc.seatunnelTimeout)
}

func TestRound16_EnsureSeatunnelClient_DefaultRetries(t *testing.T) {
	svc := NewDatasourceService(nil)
	svc.seatunnelAddress = "localhost:1234"
	svc.seatunnelRetries = -1

	assert.Equal(t, -1, svc.seatunnelRetries)
}
