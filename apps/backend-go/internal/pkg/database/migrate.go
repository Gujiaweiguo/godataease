package database

import (
	"dataease/backend/internal/domain/areamap"
	"dataease/backend/internal/domain/audit"
	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/domain/chart"
	datafillingdomain "dataease/backend/internal/domain/datafilling"
	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/domain/driver"
	"dataease/backend/internal/domain/embedded"
	"dataease/backend/internal/domain/engine"
	"dataease/backend/internal/domain/geo"
	"dataease/backend/internal/domain/menu"
	"dataease/backend/internal/domain/org"
	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/domain/static"
	"dataease/backend/internal/domain/system"
	"dataease/backend/internal/domain/user"
	"dataease/backend/internal/domain/visualization"
	applogger "dataease/backend/internal/pkg/logger"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	models := []interface{}{
		&user.SysUser{},
		&user.SysUserRole{},
		&user.SysUserPerm{},
		&role.SysRole{},
		&role.RoleMenu{},
		&org.SysOrg{},
		&permission.SysPerm{},
		&permission.SysRolePerm{},
		&permission.SysResource{},
		&permission.SysResourcePerm{},
		&permission.DataPermRow{},
		&permission.DataPermColumn{},
		&static.StaticResource{},
		&static.Store{},
		&static.Typeface{},
		// areamap
		&areamap.Area{},
		&areamap.CoreAreaCustom{},
		// audit
		&audit.AuditLog{},
		&audit.AuditLogDetail{},
		&audit.LoginFailure{},
		// chart
		&chart.CoreChartView{},
		&datafillingdomain.DataFillingForm{},
		&datafillingdomain.DfCommitLog{},
		&datafillingdomain.DataFillingTask{},
		&datafillingdomain.DataFillingSubTask{},
		&datafillingdomain.DataFillingSubInstance{},
		// dataset
		&dataset.CoreDatasetGroup{},
		&dataset.CoreDatasetTable{},
		&dataset.CoreDatasetTableField{},
		// datasource
		&datasource.CoreDatasource{},
		// driver
		&driver.Driver{},
		&driver.DriverJar{},
		// embedded
		&embedded.CoreEmbedded{},
		// engine
		&engine.Engine{},
		// geo
		&geo.GeometryArea{},
		// menu
		&menu.CoreMenu{},
		// system
		&system.SysVariable{},
		&system.SysVariableValue{},
		// visualization
		&visualization.DataVisualizationInfo{},
		&visualization.SnapshotCanvasChartView{},
		&visualization.Watermark{},
		// auto-generated domain models
		&auto.CoreAPITraffic{},
		&auto.CoreCopilotConfig{},
		&auto.CoreCopilotToken{},
		&auto.CoreCustomGeoArea{},
		&auto.CoreCustomGeoSubArea{},
		&auto.CoreDatasetTableSqlLog{},
		&auto.CoreDatasourceTask{},
		&auto.CoreDatasourceTaskLog{},
		&auto.CoreDeEngine{},
		&auto.CoreDriver{},
		&auto.CoreDriverJar{},
		&auto.CoreDsFinishPage{},
		&auto.CoreExportDownloadTask{},
		&auto.CoreExportTask{},
		&auto.CoreFont{},
		&auto.CoreOptRecent{},
		&auto.CoreRsa{},
		&auto.CoreShareTicket{},
		&auto.CoreStore{},
		&auto.CoreSysStartupJob{},
		&auto.SnapshotDataVisualizationInfo{},
		&auto.SnapshotVisualizationLinkJump{},
		&auto.SnapshotVisualizationLinkJumpInfo{},
		&auto.SnapshotVisualizationLinkJumpTargetViewInfo{},
		&auto.SnapshotVisualizationLinkage{},
		&auto.SnapshotVisualizationLinkageField{},
		&auto.SnapshotVisualizationOuterParam{},
		&auto.SnapshotVisualizationOuterParamsInfo{},
		&auto.SnapshotVisualizationOuterParamsTargetViewInfo{},
		&auto.VisualizationBackground{},
		&auto.VisualizationBackgroundImage{},
		&auto.VisualizationLinkJump{},
		&auto.VisualizationLinkJumpInfo{},
		&auto.VisualizationLinkJumpTargetViewInfo{},
		&auto.VisualizationLinkage{},
		&auto.VisualizationLinkageField{},
		&auto.VisualizationOuterParam{},
		&auto.VisualizationOuterParamsInfo{},
		&auto.VisualizationOuterParamsTargetViewInfo{},
		&auto.VisualizationReportFilter{},
		&auto.VisualizationSubject{},
		&auto.VisualizationTemplate{},
		&auto.VisualizationTemplateCategory{},
		&auto.VisualizationTemplateCategoryMap{},
		&auto.VisualizationTemplateExtendDatum{},
		&auto.XpackPlatformToken{},
		&auto.XpackSettingAuthentication{},
		&auto.XpackShare{},
		&auto.XpackThresholdInfo{},
		&auto.XpackThresholdInstance{},
		&auto.XpackWebhook{},
	}

	if err := db.AutoMigrate(models...); err != nil {
		return err
	}

	// core_sys_setting lives in repository package (circular dep), migrate inline
	if err := db.AutoMigrate(&coreSysSettingMigrate{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&coreVisualizationTemplateMigrate{}); err != nil {
		return err
	}

	applogger.Info("Database migration completed",
		zap.Int("tables", len(models)+2),
	)

	return nil
}

type coreSysSettingMigrate struct {
	ID   int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Pkey string `gorm:"column:pkey"`
	Pval string `gorm:"column:pval"`
	Type string `gorm:"column:type"`
	Sort int    `gorm:"column:sort"`
}

func (coreSysSettingMigrate) TableName() string { return "core_sys_setting" }

type coreVisualizationTemplateMigrate struct {
	ID            int64      `gorm:"column:id;primaryKey;autoIncrement"`
	Name          string     `gorm:"column:name;size:255"`
	Pid           int64      `gorm:"column:pid;index"`
	Level         int        `gorm:"column:level"`
	DvType        string     `gorm:"column:dv_type;size:50"`
	NodeType      string     `gorm:"column:node_type;size:50"`
	CreateBy      string     `gorm:"column:create_by;size:255"`
	CreateTime    *time.Time `gorm:"column:create_time"`
	Snapshot      string     `gorm:"column:snapshot;type:longtext"`
	TemplateType  string     `gorm:"column:template_type;size:50"`
	TemplateStyle string     `gorm:"column:template_style;type:longtext"`
	TemplateData  string     `gorm:"column:template_data;type:longtext"`
	DynamicData   string     `gorm:"column:dynamic_data;type:longtext"`
	AppData       string     `gorm:"column:app_data;type:longtext"`
	UseCount      int        `gorm:"column:use_count;default:0"`
	Version       int        `gorm:"column:version;default:3"`
}

func (coreVisualizationTemplateMigrate) TableName() string { return "core_visualization_template" }
