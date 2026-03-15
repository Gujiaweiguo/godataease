package http

import "testing"

func TestRegisterRoutes_RegistersVisualizationCompatibilityRoutes(t *testing.T) {
	router := NewRouter(nil, nil)
	router.RegisterRoutes()

	wantRoutes := map[string]bool{
		"GET /dataVisualization/findDvType/:id":             true,
		"POST /dataVisualization/tree":                      true,
		"POST /dataVisualization/nameCheck":                 true,
		"POST /dataVisualization/checkCanvasChange":         true,
		"POST /dataVisualization/findById":                  true,
		"POST /dataVisualization/list":                      true,
		"POST /dataVisualization/updateBase":                true,
		"POST /dataVisualization/move":                      true,
		"POST /dataVisualization/updatePublishStatus":       true,
		"POST /dataVisualization/recoverToPublished":        true,
		"POST /dataVisualization/saveCanvas":                true,
		"POST /dataVisualization/updateCanvas":              true,
		"POST /dataVisualization/deleteLogic/:id":           true,
		"POST /dataVisualization/deleteLogic/:id/:busiFlag": true,
	}

	for _, route := range router.Engine().Routes() {
		key := route.Method + " " + route.Path
		delete(wantRoutes, key)
	}

	if len(wantRoutes) > 0 {
		t.Fatalf("missing visualization compatibility routes: %+v", wantRoutes)
	}
}
