package geo

import (
	"testing"
)

func TestGeometryArea_TableName(t *testing.T) {
	g := GeometryArea{}
	if g.TableName() != "area_geo" {
		t.Errorf("Expected table name 'area_geo', got '%s'", g.TableName())
	}
}

func TestGeometryArea_Fields(t *testing.T) {
	g := GeometryArea{
		ID:      "geo-1",
		Name:    "Beijing",
		Code:    "110000",
		GeoJSON: `{"type":"Polygon","coordinates":[[[116,39],[117,39],[117,40],[116,40],[116,39]]]}`,
	}

	if g.ID != "geo-1" {
		t.Errorf("Expected ID 'geo-1', got '%s'", g.ID)
	}
	if g.Name != "Beijing" {
		t.Errorf("Expected Name 'Beijing', got '%s'", g.Name)
	}
	if g.Code != "110000" {
		t.Errorf("Expected Code '110000', got '%s'", g.Code)
	}
	if g.GeoJSON == "" {
		t.Error("Expected GeoJSON to be non-empty")
	}
}
