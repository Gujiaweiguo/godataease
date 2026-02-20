package static

type StaticResource struct {
	ID   string `gorm:"column:id;primaryKey" json:"id"`
	Name string `gorm:"column:name" json:"name"`
	Path string `gorm:"column:path" json:"path"`
	Type string `gorm:"column:type" json:"type"`
}

func (StaticResource) TableName() string {
	return "static_resource"
}

type Store struct {
	ID   string `gorm:"column:id;primaryKey" json:"id"`
	Name string `gorm:"column:name" json:"name"`
	URL  string `gorm:"column:url" json:"url"`
}

func (Store) TableName() string {
	return "store"
}

type Typeface struct {
	ID   string `gorm:"column:id;primaryKey" json:"id"`
	Name string `gorm:"column:name" json:"name"`
	File string `gorm:"column:file" json:"file"`
}

func (Typeface) TableName() string {
	return "typeface"
}
