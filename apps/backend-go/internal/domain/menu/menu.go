package menu

type CoreMenu struct {
	ID        int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Pid       int64  `gorm:"column:pid" json:"pid"`
	Type      int    `gorm:"column:type" json:"type"`
	Name      string `gorm:"column:name;size:100" json:"name"`
	Component string `gorm:"column:component;size:255" json:"component"`
	MenuSort  int    `gorm:"column:menu_sort" json:"menuSort"`
	Icon      string `gorm:"column:icon;size:100" json:"icon"`
	Path      string `gorm:"column:path;size:255" json:"path"`
	Hidden    bool   `gorm:"column:hidden" json:"hidden"`
	InLayout  bool   `gorm:"column:in_layout" json:"inLayout"`
	Auth      bool   `gorm:"column:auth" json:"auth"`
}

func (CoreMenu) TableName() string {
	return "core_menu"
}

type MenuMeta struct {
	Title string `json:"title"`
	Icon  string `json:"icon"`
}

type MenuVO struct {
	ID        int64     `json:"-"`
	Path      string    `json:"path"`
	Component string    `json:"component"`
	Hidden    bool      `json:"hidden"`
	IsPlugin  bool      `json:"isPlugin"`
	Name      string    `json:"name"`
	InLayout  bool      `json:"inLayout"`
	Redirect  string    `json:"redirect,omitempty"`
	Meta      *MenuMeta `json:"meta"`
	Children  []*MenuVO `json:"children,omitempty"`
}
