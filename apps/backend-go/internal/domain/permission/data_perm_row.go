package permission

import "time"

// AuthTargetType 授权目标类型常量
const (
	AuthTargetTypeUser = "user"
	AuthTargetTypeRole = "role"
	AuthTargetTypeDept = "dept"
)

// DataPermRow 行权限实体 - 映射 data_perm_row 表
type DataPermRow struct {
	ID             int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	DatasetID      int64      `gorm:"column:dataset_group_id;not null" json:"datasetGroupId"`
	AuthTargetType string     `gorm:"column:auth_target_type;size:255" json:"authTargetType"`
	AuthTargetID   int64      `gorm:"column:auth_target_id" json:"authTargetId"`
	ExpressionTree string     `gorm:"column:expression_tree;type:text" json:"expressionTree"`
	WhiteListUser  string     `gorm:"column:white_list_user;type:text" json:"whiteListUser"`
	WhiteListRole  string     `gorm:"column:white_list_role;type:text" json:"whiteListRole"`
	WhiteListDept  string     `gorm:"column:white_list_dept;type:text" json:"whiteListDept"`
	Status         int        `gorm:"column:status;default:1" json:"status"`
	CreateBy       *string    `gorm:"column:create_by;size:255" json:"createBy"`
	CreateTime     *time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateBy       *string    `gorm:"column:update_by;size:255" json:"updateBy"`
	UpdateTime     *time.Time `gorm:"column:update_time" json:"updateTime"`
}

func (DataPermRow) TableName() string {
	return "data_perm_row"
}

// DataPermColumn 列权限实体 - 映射 data_perm_column 表
type DataPermColumn struct {
	ID         int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	DatasetID  int64      `gorm:"column:dataset_group_id;not null" json:"datasetGroupId"`
	FieldName  string     `gorm:"column:field_name;size:255" json:"fieldName"`
	PermType   string     `gorm:"column:perm_type;size:255" json:"permType"` // disable, mask
	MaskRule   string     `gorm:"column:mask_rule;type:text" json:"maskRule"`
	Status     int        `gorm:"column:status;default:1" json:"status"`
	CreateBy   *string    `gorm:"column:create_by;size:255" json:"createBy"`
	CreateTime *time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateBy   *string    `gorm:"column:update_by;size:255" json:"updateBy"`
	UpdateTime *time.Time `gorm:"column:update_time" json:"updateTime"`
}

func (DataPermColumn) TableName() string {
	return "data_perm_column"
}

// DatasetRowPermissionsTreeObj 权限树对象结构（对应 Java DatasetRowPermissionsTreeObj）
type DatasetRowPermissionsTreeObj struct {
	Logic string                          `json:"logic"`
	Items []DatasetRowPermissionsTreeItem `json:"items"`
}

// DatasetRowPermissionsTreeItem 权限树节点项（对应 Java DatasetRowPermissionsTreeItem）
type DatasetRowPermissionsTreeItem struct {
	Type       string                        `json:"type"`       // item or tree
	FieldID    int64                         `json:"fieldId"`    // 字段ID
	FieldType  string                        `json:"fieldType"`  // 字段类型
	FilterType string                        `json:"filterType"` // logic or enum
	Term       string                        `json:"term"`       // 条件运算符
	Value      string                        `json:"value"`      // 条件值
	TimeValue  string                        `json:"timeValue"`  // 时间值
	EnumValue  []string                      `json:"enumValue"`  // 枚举值列表
	SubTree    *DatasetRowPermissionsTreeObj `json:"subTree"`    // 子树
}
