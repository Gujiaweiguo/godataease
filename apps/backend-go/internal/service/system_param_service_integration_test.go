//go:build integration

package service

import (
	"testing"

	"dataease/backend/internal/domain/audit"
	"dataease/backend/internal/domain/system"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemParamService_SaveOnlineMap_WithAuditService(t *testing.T) {
	// Clean up audit logs
	cleanupTables(&audit.AuditLog{})

	repo := repository.NewSystemParamRepository(testDB)
	auditSvc := NewAuditService(
		repository.NewAuditLogRepository(testDB),
		repository.NewLoginFailureRepository(testDB),
		repository.NewAuditLogDetailRepository(testDB),
	)
	svc := NewSystemParamService(repo, auditSvc)

	editor := &system.OnlineMapEditor{
		MapType:      "gaode",
		Key:          "test-key-with-audit",
		SecurityCode: "test-code",
	}

	err := svc.SaveOnlineMap(editor)
	require.NoError(t, err)

	// Verify audit log was created
	var count int64
	testDB.Model(&audit.AuditLog{}).Where("action_name = ?", "保存在线地图配置").Count(&count)
	assert.GreaterOrEqual(t, count, int64(1))
}

func TestSystemParamService_SaveSQLBot_WithAuditService(t *testing.T) {
	// Clean up audit logs
	cleanupTables(&audit.AuditLog{})

	repo := repository.NewSystemParamRepository(testDB)
	auditSvc := NewAuditService(
		repository.NewAuditLogRepository(testDB),
		repository.NewLoginFailureRepository(testDB),
		repository.NewAuditLogDetailRepository(testDB),
	)
	svc := NewSystemParamService(repo, auditSvc)

	cfg := &system.SQLBotConfig{
		Domain:  "test.domain.with.audit",
		ID:      "test-id-audit",
		Enabled: true,
		Valid:   true,
	}

	err := svc.SaveSQLBot(cfg)
	require.NoError(t, err)

	// Verify audit log was created
	var count int64
	testDB.Model(&audit.AuditLog{}).Where("action_name = ?", "保存SQLBot配置").Count(&count)
	assert.GreaterOrEqual(t, count, int64(1))
}

func TestSystemParamService_SaveBasic_WithAuditService(t *testing.T) {
	// Clean up audit logs
	cleanupTables(&audit.AuditLog{})

	repo := repository.NewSystemParamRepository(testDB)
	auditSvc := NewAuditService(
		repository.NewAuditLogRepository(testDB),
		repository.NewLoginFailureRepository(testDB),
		repository.NewAuditLogDetailRepository(testDB),
	)
	svc := NewSystemParamService(repo, auditSvc)

	items := []system.SettingItem{
		{Pkey: "test.basic.key", Pval: "test-value-audit", Type: "basic", Sort: 1},
	}

	err := svc.SaveBasic(items)
	require.NoError(t, err)

	// Verify audit log was created
	var count int64
	testDB.Model(&audit.AuditLog{}).Where("action_name = ?", "保存基础设置").Count(&count)
	assert.GreaterOrEqual(t, count, int64(1))
}
