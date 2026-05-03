package jobs

import (
	"context"
	"testing"

	"dataease/backend/internal/domain/audit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubAuditSettingsReader struct {
	settings *audit.AuditAlertSettings
	err      error
}

func (s *stubAuditSettingsReader) QueryAuditAlertSettings() (*audit.AuditAlertSettings, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.settings, nil
}

type stubAuditLogCleaner struct {
	days int
	call int
	err  error
}

func (s *stubAuditLogCleaner) DeleteAuditLogsBeforeDate(days int) (int64, error) {
	s.call++
	s.days = days
	return 0, s.err
}

type stubAuditAlertDetector struct {
	called bool
	ctx    context.Context
	err    error
}

type testContextKey struct{}

func (s *stubAuditAlertDetector) DetectAndAlert(ctx context.Context) error {
	s.called = true
	s.ctx = ctx
	return s.err
}

func TestAuditJobDefinitionsMetadata(t *testing.T) {
	cleanup := NewAuditCleanupDefinition(&stubAuditSettingsReader{settings: audit.DefaultAuditAlertSettings()}, &stubAuditLogCleaner{})
	alertCheck := NewAuditAlertCheckDefinition(&stubAuditAlertDetector{})

	assert.Equal(t, AuditCleanupJobName, cleanup.Metadata.Key)
	assert.Equal(t, AuditCleanupJobSpec, cleanup.Metadata.Spec)
	assert.Equal(t, auditCleanupDescription, cleanup.Metadata.Description)
	assert.True(t, cleanup.Metadata.Enabled)
	assert.True(t, cleanup.Metadata.Distributed)

	assert.Equal(t, AuditAlertCheckJobName, alertCheck.Metadata.Key)
	assert.Equal(t, AuditAlertCheckJobSpec, alertCheck.Metadata.Spec)
	assert.Equal(t, auditAlertCheckDescription, alertCheck.Metadata.Description)
	assert.True(t, alertCheck.Metadata.Enabled)
	assert.True(t, alertCheck.Metadata.Distributed)
}

func TestAuditCleanupDefinitionRunUsesSettingsRetention(t *testing.T) {
	settings := audit.DefaultAuditAlertSettings()
	settings.RetentionDays = 30
	reader := &stubAuditSettingsReader{settings: settings}
	cleaner := &stubAuditLogCleaner{}

	err := NewAuditCleanupDefinition(reader, cleaner).Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, cleaner.call)
	assert.Equal(t, 30, cleaner.days)
}

func TestAuditAlertCheckDefinitionRunCallsDetectAndAlert(t *testing.T) {
	detector := &stubAuditAlertDetector{}
	ctx := context.WithValue(context.Background(), testContextKey{}, "audit")

	err := NewAuditAlertCheckDefinition(detector).Run(ctx)
	require.NoError(t, err)
	assert.True(t, detector.called)
	assert.Equal(t, "audit", detector.ctx.Value("request"))
}
