package geo

type GeometryArea struct {
	ID      string `gorm:"column:id;primaryKey" json:"id"`
	Name    string `gorm:"column:name" json:"name"`
	Code    string `gorm:"column:code" json:"code"`
	GeoJSON string `gorm:"column:geo_json;type:text" json:"geoJson"`
}

func (GeometryArea) TableName() string {
	return "area_geo"
}
