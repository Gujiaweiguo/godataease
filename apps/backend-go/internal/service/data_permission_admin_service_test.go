package service

import (
	"encoding/json"
	"testing"

	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/permission"
)

type fakeRowPermissionStore struct {
	items      []*permission.DataPermRow
	created    *permission.DataPermRow
	updated    *permission.DataPermRow
	deletedID  int64
	lookupByID map[int64]*permission.DataPermRow
}

func (f *fakeRowPermissionStore) PagerByDatasetID(datasetID int64, page, size int) ([]*permission.DataPermRow, int64, error) {
	return f.items, int64(len(f.items)), nil
}

func (f *fakeRowPermissionStore) GetByID(id int64) (*permission.DataPermRow, error) {
	return f.lookupByID[id], nil
}

func (f *fakeRowPermissionStore) Create(perm *permission.DataPermRow) error {
	f.created = perm
	return nil
}

func (f *fakeRowPermissionStore) Update(perm *permission.DataPermRow) error {
	f.updated = perm
	return nil
}

func (f *fakeRowPermissionStore) Delete(id int64) error {
	f.deletedID = id
	return nil
}

type fakeColumnPermissionStore struct {
	items      []*permission.DataPermColumn
	created    *permission.DataPermColumn
	updated    *permission.DataPermColumn
	deletedID  int64
	lookupByID map[int64]*permission.DataPermColumn
}

func (f *fakeColumnPermissionStore) PagerByDatasetID(datasetID int64, page, size int) ([]*permission.DataPermColumn, int64, error) {
	return f.items, int64(len(f.items)), nil
}

func (f *fakeColumnPermissionStore) GetByID(id int64) (*permission.DataPermColumn, error) {
	return f.lookupByID[id], nil
}

func (f *fakeColumnPermissionStore) Create(perm *permission.DataPermColumn) error {
	f.created = perm
	return nil
}

func (f *fakeColumnPermissionStore) Update(perm *permission.DataPermColumn) error {
	f.updated = perm
	return nil
}

func (f *fakeColumnPermissionStore) Delete(id int64) error {
	f.deletedID = id
	return nil
}

type fakeDatasetFieldProvider struct {
	resp *chart.ChartFieldListResponse
}

func (f *fakeDatasetFieldProvider) ListByDQ(datasetGroupID int64, chartID int64) (*chart.ChartFieldListResponse, error) {
	return f.resp, nil
}

func TestDataPermissionAdminService_SaveRowPermission(t *testing.T) {
	rowStore := &fakeRowPermissionStore{}
	fieldProvider := &fakeDatasetFieldProvider{resp: &chart.ChartFieldListResponse{DimensionList: []chart.ChartField{{
		ID:         11,
		OriginName: "region",
		Name:       "region_alias",
	}}}}

	svc := NewDataPermissionAdminService(rowStore, &fakeColumnPermissionStore{}, fieldProvider)
	err := svc.SaveRowPermission(&RowPermissionForm{
		DatasetID:   9,
		FilterType:  permission.AuthTargetTypeRole,
		TargetID:    7,
		FilterField: "region",
		FilterValue: "east",
		WhiteList:   []int64{2, 3},
	})
	if err != nil {
		t.Fatalf("SaveRowPermission failed: %v", err)
	}
	if rowStore.created == nil {
		t.Fatal("expected row permission to be created")
	}
	if rowStore.created.AuthTargetType != permission.AuthTargetTypeRole {
		t.Fatalf("unexpected auth target type: %s", rowStore.created.AuthTargetType)
	}
	if rowStore.created.AuthTargetID != 7 {
		t.Fatalf("unexpected auth target id: %d", rowStore.created.AuthTargetID)
	}

	fieldID, value, err := decodeSimpleExpression(rowStore.created.ExpressionTree)
	if err != nil {
		t.Fatalf("decodeSimpleExpression failed: %v", err)
	}
	if fieldID != 11 || value != "east" {
		t.Fatalf("unexpected expression mapping: fieldID=%d value=%s", fieldID, value)
	}
}

func TestDataPermissionAdminService_RowPermissionPage(t *testing.T) {
	expr, err := encodeSimpleExpression(11, "east")
	if err != nil {
		t.Fatalf("encodeSimpleExpression failed: %v", err)
	}

	rowStore := &fakeRowPermissionStore{items: []*permission.DataPermRow{{
		ID:             5,
		DatasetID:      9,
		AuthTargetType: permission.AuthTargetTypeUser,
		AuthTargetID:   3,
		ExpressionTree: expr,
		Status:         1,
	}}}
	fieldProvider := &fakeDatasetFieldProvider{resp: &chart.ChartFieldListResponse{DimensionList: []chart.ChartField{{
		ID:         11,
		OriginName: "region",
	}}}}

	svc := NewDataPermissionAdminService(rowStore, &fakeColumnPermissionStore{}, fieldProvider)
	page, err := svc.RowPermissionPage(9, 1, 10)
	if err != nil {
		t.Fatalf("RowPermissionPage failed: %v", err)
	}

	list, ok := page.List.([]RowPermissionForm)
	if !ok {
		t.Fatalf("unexpected list type: %T", page.List)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 row, got %d", len(list))
	}
	if list[0].FilterField != "region" || list[0].FilterValue != "east" {
		t.Fatalf("unexpected row mapping: %#v", list[0])
	}
}

func TestDataPermissionAdminService_SaveColumnPermission(t *testing.T) {
	columnStore := &fakeColumnPermissionStore{}
	fieldProvider := &fakeDatasetFieldProvider{resp: &chart.ChartFieldListResponse{DimensionList: []chart.ChartField{{
		ID:         22,
		OriginName: "mobile",
		Type:       "string",
	}}}}

	svc := NewDataPermissionAdminService(&fakeRowPermissionStore{}, columnStore, fieldProvider)
	err := svc.SaveColumnPermission(&ColumnPermissionForm{
		DatasetID: 9,
		FieldName: "mobile",
		RuleType:  permission.PermTypeMask,
		MaskRule:  "custom",
		MaskStart: 2,
		MaskEnd:   3,
	})
	if err != nil {
		t.Fatalf("SaveColumnPermission failed: %v", err)
	}
	if columnStore.created == nil {
		t.Fatal("expected column permission to be created")
	}
	if columnStore.created.PermType != permission.PermTypeMask {
		t.Fatalf("unexpected perm type: %s", columnStore.created.PermType)
	}

	var rule permission.DesensitizationRule
	if err = json.Unmarshal([]byte(columnStore.created.MaskRule), &rule); err != nil {
		t.Fatalf("unmarshal mask rule failed: %v", err)
	}
	if rule.BuiltInRule != permission.BuiltInRuleCustom || rule.M != 2 || rule.N != 3 {
		t.Fatalf("unexpected mask rule: %#v", rule)
	}
}
