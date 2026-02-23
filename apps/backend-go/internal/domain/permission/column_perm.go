package permission

const (
	PermTypeDisable = "disable"
	PermTypeMask    = "mask"
)

const (
	BuiltInRuleCompleteDesensitization = "CompleteDesensitization"
	BuiltInRuleKeepFirstAndLastThree   = "KeepFirstAndLastThreeCharacters"
	BuiltInRuleKeepMiddleThree         = "KeepMiddleThreeCharacters"
	BuiltInRuleCustom                  = "custom"
	CustomRuleRetainBeforeMAndAfterN   = "RetainBeforeMAndAfterN"
	CustomRuleRetainMToN               = "RetainMToN"
)

type DesensitizationRule struct {
	BuiltInRule       string `json:"builtInRule"`
	CustomBuiltInRule string `json:"customBuiltInRule,omitempty"`
	M                 int    `json:"m,omitempty"`
	N                 int    `json:"n,omitempty"`
	SpecialCharacter  string `json:"specialCharacter,omitempty"`
}

type ColumnPermissionItem struct {
	ID                  int64                `json:"id"`
	Name                string               `json:"name"`
	DeType              int                  `json:"deType"`
	Selected            bool                 `json:"selected"`
	Opt                 string               `json:"opt"`
	DesensitizationRule *DesensitizationRule `json:"desensitizationRule,omitempty"`
}

type ColumnPermissions struct {
	Enable  bool                    `json:"enable"`
	Columns []*ColumnPermissionItem `json:"columns"`
}

type ColumnPermRule struct {
	FieldName string               `json:"fieldName"`
	PermType  string               `json:"permType"`
	MaskRule  *DesensitizationRule `json:"maskRule,omitempty"`
}
