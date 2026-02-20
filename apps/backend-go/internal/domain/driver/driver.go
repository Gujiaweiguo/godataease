package driver

type Driver struct {
	ID       int64   `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name     string  `gorm:"column:name" json:"name"`
	Type     string  `gorm:"column:type" json:"type"`
	TypeDesc *string `gorm:"column:type_desc" json:"typeDesc"`
	Desc     *string `gorm:"column:desc" json:"desc"`
}

func (Driver) TableName() string {
	return "de_driver"
}

type DriverDTO struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	TypeDesc *string `json:"typeDesc"`
	Desc     *string `json:"desc"`
}

type DriverJar struct {
	ID         int64   `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	DriverID   int64   `gorm:"column:driver_id" json:"driverId"`
	FileName   string  `gorm:"column:file_name" json:"fileName"`
	FilePath   string  `gorm:"column:file_path" json:"filePath"`
	Version    *string `gorm:"column:version" json:"version"`
	CreateBy   *string `gorm:"column:create_by" json:"createBy"`
	CreateTime *int64  `gorm:"column:create_time" json:"createTime"`
}

func (DriverJar) TableName() string {
	return "de_driver_jar"
}

type DriverJarDTO struct {
	ID       int64   `json:"id"`
	DriverID int64   `json:"driverId"`
	FileName string  `json:"fileName"`
	FilePath string  `json:"filePath"`
	Version  *string `json:"version"`
}
