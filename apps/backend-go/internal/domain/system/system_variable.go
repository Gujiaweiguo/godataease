package system

type SysVariable struct {
	ID        int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Type      string `gorm:"column:type;size:50;index" json:"type"`
	Name      string `gorm:"column:name;size:255;index" json:"name"`
	Min       int64  `gorm:"column:min" json:"min"`
	Max       int64  `gorm:"column:max" json:"max"`
	StartTime string `gorm:"column:start_time;size:64" json:"startTime"`
	EndTime   string `gorm:"column:end_time;size:64" json:"endTime"`
	Root      bool   `gorm:"column:root;default:false" json:"root"`
	Disabled  bool   `gorm:"column:disabled;default:false" json:"disabled"`
}

func (SysVariable) TableName() string {
	return "sys_variable"
}

type SysVariableValue struct {
	ID            int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	SysVariableID int64  `gorm:"column:sys_variable_id;index" json:"sysVariableId"`
	Value         string `gorm:"column:value;size:255" json:"value"`
	ValueDesc     string `gorm:"column:value_desc;size:255" json:"valueDesc"`
	Begin         string `gorm:"column:begin_time;size:64" json:"begin"`
	End           string `gorm:"column:end_time;size:64" json:"end"`
}

func (SysVariableValue) TableName() string {
	return "sys_variable_value"
}

type SysVariableQueryRequest struct {
	ID       int64  `json:"id"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Disabled *bool  `json:"disabled,omitempty"`
}

type SysVariableValueQueryRequest struct {
	ID            int64  `json:"id"`
	SysVariableID int64  `json:"sysVariableId"`
	Value         string `json:"value"`
	ValueDesc     string `json:"valueDesc"`
}

type SysVariableValuePage struct {
	Records []SysVariableValue `json:"records"`
	Total   int64              `json:"total"`
	Current int                `json:"current"`
	Size    int                `json:"size"`
}
