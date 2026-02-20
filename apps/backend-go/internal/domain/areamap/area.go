package areamap

type Area struct {
	ID    string `gorm:"column:id;primaryKey" json:"id"`
	Level string `gorm:"column:level" json:"level"`
	Name  string `gorm:"column:name" json:"name"`
	Pid   string `gorm:"column:pid" json:"pid"`
}

func (Area) TableName() string {
	return "area"
}

type CoreAreaCustom struct {
	ID    string `gorm:"column:id;primaryKey" json:"id"`
	Level string `gorm:"column:level" json:"level"`
	Name  string `gorm:"column:name" json:"name"`
	Pid   string `gorm:"column:pid" json:"pid"`
}

func (CoreAreaCustom) TableName() string {
	return "core_area_custom"
}

type AreaNode struct {
	ID       string      `json:"id"`
	Level    string      `json:"level"`
	Name     string      `json:"name"`
	Pid      string      `json:"pid,omitempty"`
	Custom   bool        `json:"custom"`
	Country  string      `json:"country,omitempty"`
	Children []*AreaNode `json:"children,omitempty"`
}
