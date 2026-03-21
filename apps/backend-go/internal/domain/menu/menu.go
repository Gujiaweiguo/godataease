package menu

import (
	"database/sql/driver"
	"encoding/json"
)

// JSON is a custom type for handling JSON fields in GORM
type JSON map[string]interface{}

func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

type CoreMenu struct {
	ID           int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Pid          int64  `gorm:"column:pid" json:"pid"`
	Type         int    `gorm:"column:type" json:"type"`
	Name         string `gorm:"column:name;size:100" json:"name"`
	Component    string `gorm:"column:component;size:255" json:"component"`
	MenuSort     int    `gorm:"column:menu_sort" json:"menuSort"`
	Icon         string `gorm:"column:icon;size:100" json:"icon"`
	Path         string `gorm:"column:path;size:255" json:"path"`
	Hidden       bool   `gorm:"column:hidden" json:"hidden"`
	InLayout     bool   `gorm:"column:in_layout" json:"inLayout"`
	Auth         bool   `gorm:"column:auth" json:"auth"`
	MenuLocation string `gorm:"column:menu_location;size:20" json:"menuLocation"`
	MenuType     string `gorm:"column:menu_type;size:20" json:"menuType"`
	ActionConfig JSON   `gorm:"column:action_config;type:json" json:"actionConfig"`
}

func (CoreMenu) TableName() string {
	return "core_menu"
}

type MenuMeta struct {
	Title string `json:"title"`
	Icon  string `json:"icon"`
}

type MenuVO struct {
	ID           int64                  `json:"id"`
	Pid          int64                  `json:"pid"`
	Type         int                    `json:"type"`
	MenuSort     int                    `json:"menuSort"`
	Icon         string                 `json:"icon"`
	Auth         bool                   `json:"auth"`
	Path         string                 `json:"path"`
	Component    string                 `json:"component"`
	Hidden       bool                   `json:"hidden"`
	IsPlugin     bool                   `json:"isPlugin"`
	Name         string                 `json:"name"`
	InLayout     bool                   `json:"inLayout"`
	Redirect     string                 `json:"redirect,omitempty"`
	MenuLocation string                 `json:"menuLocation,omitempty"`
	MenuType     string                 `json:"menuType,omitempty"`
	ActionConfig map[string]interface{} `json:"actionConfig,omitempty"`
	Meta         *MenuMeta              `json:"meta"`
	Children     []*MenuVO              `json:"children,omitempty"`
}
