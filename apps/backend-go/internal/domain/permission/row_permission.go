package permission

// 行权限操作符常量
const (
	OperatorEq      = "eq"
	OperatorNotEq   = "not_eq"
	OperatorLike    = "like"
	OperatorNotLike = "not_like"
	OperatorGt      = "gt"
	OperatorLt      = "lt"
	OperatorGe      = "ge"
	OperatorLe      = "le"
	OperatorIn      = "in"
	OperatorNotIn   = "not_in"
	OperatorNull    = "null"
	OperatorNotNull = "not_null"
)

// 逻辑操作符常量
const (
	LogicAnd = "and"
	LogicOr  = "or"
)

// 节点类型常量
const (
	NodeTypeItem = "item"
	NodeTypeTree = "tree"
)

// RowPermissionTree 行权限树节点结构
// 用于表示条件表达式树，支持嵌套的逻辑组合
type RowPermissionTree struct {
	// ID 节点唯一标识
	ID string `json:"id"`
	// Type 节点类型: item(叶子节点) 或 tree(子树)
	Type string `json:"type"`
	// Field 字段名称
	Field string `json:"field,omitempty"`
	// FieldID 字段ID (用于关联数据集字段)
	FieldID int64 `json:"fieldId,omitempty"`
	// Operator 操作符: eq, not_eq, like, not_like, gt, lt, ge, le, in, not_in, null, not_null
	Operator string `json:"operator,omitempty"`
	// Value 比较值
	Value interface{} `json:"value,omitempty"`
	// Logic 逻辑操作符: and, or (用于子树节点)
	Logic string `json:"logic,omitempty"`
	// Children 子节点列表 (用于树形结构)
	Children []RowPermissionTree `json:"children,omitempty"`
	// SubTree 子树 (兼容Java端的嵌套结构)
	SubTree *RowPermissionTree `json:"subTree,omitempty"`
}

// RowPermissionFilter 行权限过滤器
// 用于存储单个数据集的行权限配置
type RowPermissionFilter struct {
	// DatasetID 数据集ID
	DatasetID int64 `json:"datasetId"`
	// UserID 用户ID
	UserID int64 `json:"userId"`
	// Rules 权限规则树列表
	Rules []RowPermissionTree `json:"rules"`
}

// RowPermissionDTO 行权限数据传输对象
// 映射Java端的DataSetRowPermissionsTreeDTO结构
type RowPermissionDTO struct {
	// ID 权限规则ID
	ID int64 `json:"id"`
	// DatasetID 数据集ID
	DatasetID int64 `json:"datasetId"`
	// AuthTargetType 授权目标类型: user, role, sysParams
	AuthTargetType string `json:"authTargetType"`
	// AuthTargetID 授权目标ID
	AuthTargetID int64 `json:"authTargetId"`
	// Enable 是否启用
	Enable bool `json:"enable"`
	// ExpressionTree 表达式树JSON
	ExpressionTree string `json:"expressionTree"`
	// WhiteListUser 白名单用户ID列表JSON
	WhiteListUser string `json:"whiteListUser"`
	// WhiteListRole 白名单角色ID列表JSON
	WhiteListRole string `json:"whiteListRole"`
	// WhiteListDept 白名单部门ID列表JSON
	WhiteListDept string `json:"whiteListDept"`
	// Tree 解析后的权限树
	Tree *RowPermissionTree `json:"tree,omitempty"`
}

// RowPermissionRequest 行权限查询请求
type RowPermissionRequest struct {
	// DatasetID 数据集ID
	DatasetID int64 `json:"datasetId" binding:"required"`
	// UserID 用户ID (可选，不传则使用当前用户)
	UserID int64 `json:"userId"`
}

// WhereClauseResult WHERE子句构建结果
type WhereClauseResult struct {
	// Clause SQL WHERE子句 (不含WHERE关键字)
	Clause string
	// Args 参数列表
	Args []interface{}
}
