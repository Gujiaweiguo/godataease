package audit

import "time"

type ActionType string

const (
	ActionTypeUserAction       ActionType = "USER_ACTION"
	ActionTypePermissionChange ActionType = "PERMISSION_CHANGE"
	ActionTypeDataAccess       ActionType = "DATA_ACCESS"
	ActionTypeSystemConfig     ActionType = "SYSTEM_CONFIG"
)

type Operation string

const (
	OperationCreate Operation = "CREATE"
	OperationUpdate Operation = "UPDATE"
	OperationDelete Operation = "DELETE"
	OperationExport Operation = "EXPORT"
	OperationLogin  Operation = "LOGIN"
	OperationLogout Operation = "LOGOUT"
)

type ResourceType string

const (
	ResourceTypeUser         ResourceType = "USER"
	ResourceTypeOrganization ResourceType = "ORGANIZATION"
	ResourceTypeRole         ResourceType = "ROLE"
	ResourceTypePermission   ResourceType = "PERMISSION"
	ResourceTypeDataset      ResourceType = "DATASET"
	ResourceTypeDashboard    ResourceType = "DASHBOARD"
)

type Status string

const (
	StatusSuccess Status = "SUCCESS"
	StatusFailed  Status = "FAILED"
)

type AuditLog struct {
	ID             int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID         *int64     `gorm:"column:user_id" json:"userId"`
	Username       *string    `gorm:"column:username;size:100" json:"username"`
	ActionType     ActionType `gorm:"column:action_type;size:50;not null" json:"actionType"`
	ActionName     string     `gorm:"column:action_name;size:100;not null" json:"actionName"`
	ResourceType   *string    `gorm:"column:resource_type;size:50" json:"resourceType"`
	ResourceID     *int64     `gorm:"column:resource_id" json:"resourceId"`
	ResourceName   *string    `gorm:"column:resource_name;size:200" json:"resourceName"`
	Operation      Operation  `gorm:"column:operation;size:20;not null" json:"operation"`
	Status         Status     `gorm:"column:status;size:20;not null;default:SUCCESS" json:"status"`
	FailureReason  *string    `gorm:"column:failure_reason;size:500" json:"failureReason"`
	IPAddress      *string    `gorm:"column:ip_address;size:50" json:"ipAddress"`
	UserAgent      *string    `gorm:"column:user_agent;size:500" json:"userAgent"`
	BeforeValue    *string    `gorm:"column:before_value;type:text" json:"beforeValue"`
	AfterValue     *string    `gorm:"column:after_value;type:text" json:"afterValue"`
	OrganizationID *int64     `gorm:"column:organization_id" json:"organizationId"`
	CreateTime     time.Time  `gorm:"column:create_time;not null;autoCreateTime" json:"createTime"`
}

func (AuditLog) TableName() string {
	return "de_audit_log"
}

type AuditLogDetail struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AuditLogID  int64     `gorm:"column:audit_log_id;not null" json:"auditLogId"`
	DetailType  *string   `gorm:"column:detail_type;size:50" json:"detailType"`
	DetailKey   *string   `gorm:"column:detail_key;size:100" json:"detailKey"`
	DetailValue *string   `gorm:"column:detail_value;type:text" json:"detailValue"`
	CreateTime  time.Time `gorm:"column:create_time;not null;autoCreateTime" json:"createTime"`
}

func (AuditLogDetail) TableName() string {
	return "de_audit_log_detail"
}

type LoginFailure struct {
	ID            int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Username      string    `gorm:"column:username;size:100;not null" json:"username"`
	IPAddress     *string   `gorm:"column:ip_address;size:50" json:"ipAddress"`
	FailureReason *string   `gorm:"column:failure_reason;size:200" json:"failureReason"`
	UserAgent     *string   `gorm:"column:user_agent;size:500" json:"userAgent"`
	CreateTime    time.Time `gorm:"column:create_time;not null;autoCreateTime" json:"createTime"`
}

func (LoginFailure) TableName() string {
	return "de_login_failure"
}

type AuditLogQuery struct {
	UserID         *int64
	Username       *string
	ActionType     *ActionType
	ResourceType   *ResourceType
	OrganizationID *int64
	StartTime      *time.Time
	EndTime        *time.Time
	Status         *Status
	Page           int
	PageSize       int
}

type AuditLogCreateRequest struct {
	UserID         *int64     `json:"userId"`
	Username       *string    `json:"username"`
	ActionType     ActionType `json:"actionType" binding:"required"`
	ActionName     string     `json:"actionName" binding:"required"`
	ResourceType   *string    `json:"resourceType"`
	ResourceID     *int64     `json:"resourceId"`
	ResourceName   *string    `json:"resourceName"`
	Operation      Operation  `json:"operation" binding:"required"`
	Status         *Status    `json:"status"`
	FailureReason  *string    `json:"failureReason"`
	IPAddress      *string    `json:"ipAddress"`
	UserAgent      *string    `json:"userAgent"`
	BeforeValue    *string    `json:"beforeValue"`
	AfterValue     *string    `json:"afterValue"`
	OrganizationID *int64     `json:"organizationId"`
}

type LoginFailureRequest struct {
	Username      string  `json:"username" binding:"required"`
	IPAddress     *string `json:"ipAddress"`
	FailureReason *string `json:"failureReason"`
	UserAgent     *string `json:"userAgent"`
}
