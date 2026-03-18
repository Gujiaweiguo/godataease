package handler

import (
	"testing"

	"dataease/backend/internal/domain/visualization"
)

func TestResolveBusiTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantLen  int
		wantErr  bool
		firstVal string
	}{
		{name: "empty maps to dashboard and dataV", input: "", wantLen: 2, wantErr: false, firstVal: "dashboard"},
		{name: "dashboard-dataV maps to two types", input: "dashboard-dataV", wantLen: 2, wantErr: false, firstVal: "dashboard"},
		{name: "panel maps to dashboard", input: "panel", wantLen: 1, wantErr: false, firstVal: "dashboard"},
		{name: "screen maps to dataV", input: "screen", wantLen: 1, wantErr: false, firstVal: "dataV"},
		{name: "unsupported busiFlag returns error", input: "dataset", wantLen: 0, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveBusiTypes(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tt.wantLen {
				t.Fatalf("unexpected len: got %d want %d", len(got), tt.wantLen)
			}
			if tt.wantLen > 0 && got[0] != tt.firstVal {
				t.Fatalf("unexpected first value: got %s want %s", got[0], tt.firstVal)
			}
		})
	}
}

func TestBuildVisualizationTreeValidation(t *testing.T) {
	validNodeType := "folder"
	invalidNodeType := "unknown"

	tests := []struct {
		name    string
		items   []*visualization.DataVisualizationInfo
		wantErr bool
	}{
		{
			name: "invalid id returns error",
			items: []*visualization.DataVisualizationInfo{{
				ID:       0,
				Name:     "root",
				NodeType: &validNodeType,
			}},
			wantErr: true,
		},
		{
			name: "empty name returns error",
			items: []*visualization.DataVisualizationInfo{{
				ID:       1,
				Name:     "",
				NodeType: &validNodeType,
			}},
			wantErr: true,
		},
		{
			name: "invalid nodeType returns error",
			items: []*visualization.DataVisualizationInfo{{
				ID:       1,
				Name:     "invalid-type",
				NodeType: &invalidNodeType,
			}},
			wantErr: true,
		},
		{
			name: "valid item returns success",
			items: []*visualization.DataVisualizationInfo{{
				ID:       1,
				Name:     "folder-1",
				NodeType: &validNodeType,
			}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes, err := buildVisualizationTree(tt.items, nil)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(nodes) == 0 {
				t.Fatalf("expected non-empty nodes")
			}
		})
	}
}

func TestBuildVisualizationTreeContractShape(t *testing.T) {
	folderType := visualizationNodeTypeFolder
	panelType := visualizationNodeTypePanel
	rootID := int64(10)
	published := 1
	mobileLayout := true

	nodes, err := buildVisualizationTree([]*visualization.DataVisualizationInfo{
		{
			ID:       rootID,
			Name:     "Dashboard Folder",
			NodeType: &folderType,
			Status:   &published,
		},
		{
			ID:           11,
			PID:          &rootID,
			Name:         "Dashboard A",
			NodeType:     &panelType,
			Status:       &published,
			MobileLayout: &mobileLayout,
		},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertVisualizationTreeRootShape(t, nodes)

	leafOnly, err := buildVisualizationTree([]*visualization.DataVisualizationInfo{
		{
			ID:       rootID,
			Name:     "Dashboard Folder",
			NodeType: &folderType,
		},
		{
			ID:       11,
			PID:      &rootID,
			Name:     "Dashboard A",
			NodeType: &panelType,
		},
	}, boolPtr(true))
	if err != nil {
		t.Fatalf("unexpected leaf-filter error: %v", err)
	}
	assertVisualizationLeafOnlyTree(t, leafOnly)
}

func assertVisualizationTreeRootShape(t *testing.T, nodes []treeNode) {
	t.Helper()
	if len(nodes) != 1 {
		t.Fatalf("expected 1 root node, got %d", len(nodes))
	}
	if nodes[0].ID != "10" || nodes[0].Name != "Dashboard Folder" {
		t.Fatalf("unexpected root node: %+v", nodes[0])
	}
	if nodes[0].Leaf {
		t.Fatalf("expected folder root to be non-leaf: %+v", nodes[0])
	}
	if nodes[0].ExtraFlag1 != 1 {
		t.Fatalf("expected published folder extraFlag1=1: %+v", nodes[0])
	}
	if len(nodes[0].Children) != 1 {
		t.Fatalf("expected one child node, got %d", len(nodes[0].Children))
	}
	child := nodes[0].Children[0]
	if child.ID != "11" || child.PID != "10" || child.Name != "Dashboard A" {
		t.Fatalf("unexpected child node: %+v", child)
	}
	if !child.Leaf {
		t.Fatalf("expected panel child to be leaf: %+v", child)
	}
	if child.ExtraFlag != 1 || child.ExtraFlag1 != 1 {
		t.Fatalf("expected mobile published child flags, got %+v", child)
	}
}

func assertVisualizationLeafOnlyTree(t *testing.T, nodes []treeNode) {
	t.Helper()
	if len(nodes) != 0 {
		t.Fatalf("expected no root nodes after leaf-only filter, got %+v", nodes)
	}
}

func boolPtr(v bool) *bool {
	return &v
}
