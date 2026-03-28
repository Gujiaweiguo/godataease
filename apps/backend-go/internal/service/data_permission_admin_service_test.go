package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/permission"
)

type fakeRowPermissionStore struct {
	items       []*permission.DataPermRow
	targetItems map[string][]*permission.DataPermRow
	created     *permission.DataPermRow
	updated     *permission.DataPermRow
	deletedID   int64
	lookupByID  map[int64]*permission.DataPermRow
	pagerErr    error
	targetErr   error
	getErr      error
	deleteErr   error
}

func (f *fakeRowPermissionStore) PagerByDatasetID(datasetID int64, page, size int) ([]*permission.DataPermRow, int64, error) {
	if f.pagerErr != nil {
		return nil, 0, f.pagerErr
	}
	return f.items, int64(len(f.items)), nil
}

func (f *fakeRowPermissionStore) PagerByDatasetIDAndTarget(datasetID int64, targetType string, targetID int64, page, size int) ([]*permission.DataPermRow, int64, error) {
	if f.targetErr != nil {
		return nil, 0, f.targetErr
	}
	if f.targetItems == nil {
		return []*permission.DataPermRow{}, 0, nil
	}
	key := fmt.Sprintf("%d:%s:%d", datasetID, targetType, targetID)
	items := f.targetItems[key]
	return items, int64(len(items)), nil
}

func (f *fakeRowPermissionStore) GetByID(id int64) (*permission.DataPermRow, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
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
	if f.deleteErr != nil {
		return f.deleteErr
	}
	f.deletedID = id
	return nil
}

type fakeColumnPermissionStore struct {
	items      []*permission.DataPermColumn
	created    *permission.DataPermColumn
	updated    *permission.DataPermColumn
	deletedID  int64
	lookupByID map[int64]*permission.DataPermColumn
	pagerErr   error
	getErr     error
	deleteErr  error
}

func (f *fakeColumnPermissionStore) PagerByDatasetID(datasetID int64, page, size int) ([]*permission.DataPermColumn, int64, error) {
	if f.pagerErr != nil {
		return nil, 0, f.pagerErr
	}
	return f.items, int64(len(f.items)), nil
}

func (f *fakeColumnPermissionStore) GetByID(id int64) (*permission.DataPermColumn, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
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
	if f.deleteErr != nil {
		return f.deleteErr
	}
	f.deletedID = id
	return nil
}

type fakeDatasetFieldProvider struct {
	resp *chart.ChartFieldListResponse
	err  error
}

func (f *fakeDatasetFieldProvider) ListByDQ(datasetGroupID int64, chartID int64) (*chart.ChartFieldListResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
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

func TestDataPermissionAdminService_RowPermissionPageByTarget(t *testing.T) {
	expr, err := encodeSimpleExpression(11, "east")
	if err != nil {
		t.Fatalf("encodeSimpleExpression failed: %v", err)
	}

	rowStore := &fakeRowPermissionStore{targetItems: map[string][]*permission.DataPermRow{
		"9:role:7": {{
			ID:             5,
			DatasetID:      9,
			AuthTargetType: permission.AuthTargetTypeRole,
			AuthTargetID:   7,
			ExpressionTree: expr,
			Status:         1,
		}},
	}}
	fieldProvider := &fakeDatasetFieldProvider{resp: &chart.ChartFieldListResponse{DimensionList: []chart.ChartField{{
		ID:         11,
		OriginName: "region",
	}}}}

	svc := NewDataPermissionAdminService(rowStore, &fakeColumnPermissionStore{}, fieldProvider)
	page, err := svc.RowPermissionPageByTarget(9, permission.AuthTargetTypeRole, 7, 1, 10)
	if err != nil {
		t.Fatalf("RowPermissionPageByTarget failed: %v", err)
	}

	list, ok := page.List.([]RowPermissionForm)
	if !ok {
		t.Fatalf("unexpected list type: %T", page.List)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 row, got %d", len(list))
	}
	if list[0].FilterType != permission.AuthTargetTypeRole || list[0].TargetID != 7 {
		t.Fatalf("unexpected target row mapping: %#v", list[0])
	}
}

func TestDataPermissionAdminService_RowPermissionPageByTarget_RejectsUnsupportedType(t *testing.T) {
	svc := NewDataPermissionAdminService(&fakeRowPermissionStore{}, &fakeColumnPermissionStore{}, &fakeDatasetFieldProvider{})
	if _, err := svc.RowPermissionPageByTarget(9, permission.AuthTargetTypeDept, 7, 1, 10); err == nil {
		t.Fatal("expected unsupported targetType to fail")
	}
}

func TestDataPermissionAdminService_RowPermissionPageByTarget_RejectsMissingTargetID(t *testing.T) {
	svc := NewDataPermissionAdminService(&fakeRowPermissionStore{}, &fakeColumnPermissionStore{}, &fakeDatasetFieldProvider{})
	if _, err := svc.RowPermissionPageByTarget(9, permission.AuthTargetTypeRole, 0, 1, 10); err == nil || err.Error() != "targetId is required" {
		t.Fatalf("unexpected error: %v", err)
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
		MaskRule:  maskRuleCustom,
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

func TestDataPermissionAdminService_SaveRowPermission_Validation(t *testing.T) {
	fieldProvider := &fakeDatasetFieldProvider{resp: &chart.ChartFieldListResponse{DimensionList: []chart.ChartField{{ID: 11, OriginName: "region", Name: "region_alias"}}}}
	svc := NewDataPermissionAdminService(&fakeRowPermissionStore{}, &fakeColumnPermissionStore{}, fieldProvider)

	if err := svc.SaveRowPermission(&RowPermissionForm{TargetID: 1, FilterType: permission.AuthTargetTypeUser, FilterField: "region"}); err == nil || err.Error() != "datasetId is required" {
		t.Fatalf("unexpected datasetId validation error: %v", err)
	}
	if err := svc.SaveRowPermission(&RowPermissionForm{DatasetID: 9, FilterType: permission.AuthTargetTypeUser, FilterField: "region"}); err == nil || err.Error() != "targetId is required" {
		t.Fatalf("unexpected targetId validation error: %v", err)
	}
	if err := svc.SaveRowPermission(&RowPermissionForm{DatasetID: 9, TargetID: 1, FilterType: permission.AuthTargetTypeDept, FilterField: "region"}); err == nil || !strings.Contains(err.Error(), "filterType dept is not supported") {
		t.Fatalf("unexpected filterType validation error: %v", err)
	}
	if err := svc.SaveRowPermission(&RowPermissionForm{DatasetID: 9, TargetID: 1, FilterType: permission.AuthTargetTypeUser, FilterField: "   "}); err == nil || err.Error() != "filterField is required" {
		t.Fatalf("unexpected filterField validation error: %v", err)
	}
	if err := svc.SaveRowPermission(&RowPermissionForm{DatasetID: 9, TargetID: 1, FilterType: permission.AuthTargetTypeUser, FilterField: "missing"}); err == nil || err.Error() != "dataset field missing not found" {
		t.Fatalf("unexpected dataset field error: %v", err)
	}
}

func TestDataPermissionAdminService_SaveColumnPermission_Validation(t *testing.T) {
	t.Run("validation errors", testSaveColumnPermissionValidationErrors)
	t.Run("persists mask and disable rules", testSaveColumnPermissionValidationPersistsRules)
}

func testSaveColumnPermissionValidationErrors(t *testing.T) {
	fieldProvider := &fakeDatasetFieldProvider{resp: &chart.ChartFieldListResponse{DimensionList: []chart.ChartField{{ID: 22, OriginName: "mobile", Type: "string"}}}}
	svc := NewDataPermissionAdminService(&fakeRowPermissionStore{}, &fakeColumnPermissionStore{}, fieldProvider)

	if err := svc.SaveColumnPermission(&ColumnPermissionForm{FieldName: "mobile", RuleType: permission.PermTypeMask}); err == nil || err.Error() != "datasetId is required" {
		t.Fatalf("unexpected datasetId validation error: %v", err)
	}
	if err := svc.SaveColumnPermission(&ColumnPermissionForm{DatasetID: 9, RuleType: permission.PermTypeMask}); err == nil || err.Error() != "fieldName is required" {
		t.Fatalf("unexpected fieldName validation error: %v", err)
	}
	if err := svc.SaveColumnPermission(&ColumnPermissionForm{DatasetID: 9, FieldName: "mobile", RuleType: "unknown"}); err == nil || !strings.Contains(err.Error(), "ruleType unknown is not supported") {
		t.Fatalf("unexpected ruleType validation error: %v", err)
	}
	if err := svc.SaveColumnPermission(&ColumnPermissionForm{DatasetID: 9, FieldName: "missing", RuleType: permission.PermTypeMask}); err == nil || err.Error() != "dataset field missing not found" {
		t.Fatalf("unexpected dataset field error: %v", err)
	}
}

func testSaveColumnPermissionValidationPersistsRules(t *testing.T) {
	fieldProvider := &fakeDatasetFieldProvider{resp: &chart.ChartFieldListResponse{DimensionList: []chart.ChartField{{ID: 22, OriginName: "mobile", Type: "string"}}}}
	columnStore := &fakeColumnPermissionStore{}
	svc := NewDataPermissionAdminService(&fakeRowPermissionStore{}, columnStore, fieldProvider)
	err := svc.SaveColumnPermission(&ColumnPermissionForm{DatasetID: 9, FieldName: "mobile", RuleType: permission.PermTypeMask, MaskRule: maskRuleCustom, MaskStart: 1, MaskEnd: 5})
	if err != nil {
		t.Fatalf("expected custom mask permission to succeed: %v", err)
	}
	var rule permission.DesensitizationRule
	if err = json.Unmarshal([]byte(columnStore.created.MaskRule), &rule); err != nil {
		t.Fatalf("unmarshal custom mask rule failed: %v", err)
	}
	if rule.M != 1 || rule.N != 5 {
		t.Fatalf("expected custom mask bounds to be preserved, got %#v", rule)
	}

	columnStore = &fakeColumnPermissionStore{}
	svc = NewDataPermissionAdminService(&fakeRowPermissionStore{}, columnStore, fieldProvider)
	err = svc.SaveColumnPermission(&ColumnPermissionForm{DatasetID: 9, FieldName: "mobile", RuleType: permission.PermTypeDisable})
	if err != nil {
		t.Fatalf("expected disable permission to succeed: %v", err)
	}
	if columnStore.created == nil || columnStore.created.MaskRule != "" {
		t.Fatalf("expected disable permission to store empty mask rule, got %#v", columnStore.created)
	}
}

func TestDataPermissionAdminService_SaveRowPermission_UpdateExisting(t *testing.T) {
	expr, err := encodeSimpleExpression(11, "old")
	if err != nil {
		t.Fatalf("encodeSimpleExpression failed: %v", err)
	}

	row := &permission.DataPermRow{ID: 8, DatasetID: 1, DatasetGroupID: 1, ExpressionTree: expr}
	rowStore := &fakeRowPermissionStore{lookupByID: map[int64]*permission.DataPermRow{8: row}}
	fieldProvider := &fakeDatasetFieldProvider{resp: &chart.ChartFieldListResponse{DimensionList: []chart.ChartField{{
		ID:         11,
		OriginName: "region",
		Name:       "region_alias",
	}}}}

	svc := NewDataPermissionAdminService(rowStore, &fakeColumnPermissionStore{}, fieldProvider)
	err = svc.SaveRowPermission(&RowPermissionForm{
		ID:          8,
		DatasetID:   9,
		FilterType:  permission.AuthTargetTypeUser,
		TargetID:    3,
		FilterField: "region",
		FilterValue: "west",
	})
	if err != nil {
		t.Fatalf("SaveRowPermission update failed: %v", err)
	}
	if rowStore.updated == nil {
		t.Fatal("expected row permission to be updated")
	}
	if rowStore.updated.DatasetGroupID != 9 {
		t.Fatalf("expected dataset group id 9, got %d", rowStore.updated.DatasetGroupID)
	}
	fieldID, value, err := decodeSimpleExpression(rowStore.updated.ExpressionTree)
	if err != nil {
		t.Fatalf("decodeSimpleExpression failed: %v", err)
	}
	if fieldID != 11 || value != "west" {
		t.Fatalf("unexpected updated expression mapping: fieldID=%d value=%s", fieldID, value)
	}
}

func TestDataPermissionAdminService_SaveColumnPermission_UpdateExisting(t *testing.T) {
	column := &permission.DataPermColumn{ID: 6, DatasetID: 1, DatasetGroupID: 1, FieldName: "mobile"}
	columnStore := &fakeColumnPermissionStore{lookupByID: map[int64]*permission.DataPermColumn{6: column}}
	fieldProvider := &fakeDatasetFieldProvider{resp: &chart.ChartFieldListResponse{DimensionList: []chart.ChartField{{
		ID:         22,
		OriginName: "mobile",
		Type:       "string",
	}}}}

	svc := NewDataPermissionAdminService(&fakeRowPermissionStore{}, columnStore, fieldProvider)
	err := svc.SaveColumnPermission(&ColumnPermissionForm{
		ID:        6,
		DatasetID: 9,
		FieldName: "mobile",
		RuleType:  permission.PermTypeMask,
		MaskRule:  maskRuleKeepEnds,
	})
	if err != nil {
		t.Fatalf("SaveColumnPermission update failed: %v", err)
	}
	if columnStore.updated == nil {
		t.Fatal("expected column permission to be updated")
	}
	if columnStore.updated.DatasetGroupID != 9 {
		t.Fatalf("expected dataset group id 9, got %d", columnStore.updated.DatasetGroupID)
	}
	var rule permission.DesensitizationRule
	if err = json.Unmarshal([]byte(columnStore.updated.MaskRule), &rule); err != nil {
		t.Fatalf("unmarshal updated mask rule failed: %v", err)
	}
	if rule.BuiltInRule != permission.BuiltInRuleKeepFirstAndLastThree {
		t.Fatalf("unexpected updated mask rule: %#v", rule)
	}
}

func TestDataPermissionAdminService_DeleteRowPermission(t *testing.T) {
	rowStore := &fakeRowPermissionStore{}
	svc := NewDataPermissionAdminService(rowStore, &fakeColumnPermissionStore{}, &fakeDatasetFieldProvider{})

	if err := svc.DeleteRowPermission(0); err == nil {
		t.Fatal("expected error when row permission id is missing")
	}

	if err := svc.DeleteRowPermission(9); err != nil {
		t.Fatalf("DeleteRowPermission failed: %v", err)
	}
	if rowStore.deletedID != 9 {
		t.Fatalf("expected deleted row permission id 9, got %d", rowStore.deletedID)
	}
}

func TestDataPermissionAdminService_DeleteColumnPermission(t *testing.T) {
	columnStore := &fakeColumnPermissionStore{}
	svc := NewDataPermissionAdminService(&fakeRowPermissionStore{}, columnStore, &fakeDatasetFieldProvider{})

	if err := svc.DeleteColumnPermission(0); err == nil {
		t.Fatal("expected error when column permission id is missing")
	}

	if err := svc.DeleteColumnPermission(12); err != nil {
		t.Fatalf("DeleteColumnPermission failed: %v", err)
	}
	if columnStore.deletedID != 12 {
		t.Fatalf("expected deleted column permission id 12, got %d", columnStore.deletedID)
	}
}

func TestDataPermissionAdminService_RowPermissionStoreErrorPaths(t *testing.T) {
	t.Run("target pager error propagates", func(t *testing.T) {
		svc := NewDataPermissionAdminService(&fakeRowPermissionStore{targetErr: errors.New("target pager failed")}, &fakeColumnPermissionStore{}, &fakeDatasetFieldProvider{})
		page, err := svc.RowPermissionPageByTarget(9, permission.AuthTargetTypeUser, 7, 1, 10)
		if !errors.Is(err, svc.rowStore.(*fakeRowPermissionStore).targetErr) {
			t.Fatalf("expected target pager error, got %v", err)
		}
		if page != nil {
			t.Fatalf("expected nil page, got %#v", page)
		}
	})

	t.Run("row page store error", func(t *testing.T) {
		svc := NewDataPermissionAdminService(&fakeRowPermissionStore{pagerErr: errors.New("row pager failed")}, &fakeColumnPermissionStore{}, &fakeDatasetFieldProvider{})
		page, err := svc.RowPermissionPage(9, 1, 10)
		if !errors.Is(err, svc.rowStore.(*fakeRowPermissionStore).pagerErr) {
			t.Fatalf("expected row pager error, got %v", err)
		}
		if page != nil {
			t.Fatalf("expected nil page, got %#v", page)
		}
	})

	t.Run("delete error propagates", func(t *testing.T) {
		rowErr := errors.New("row delete failed")
		svc := NewDataPermissionAdminService(&fakeRowPermissionStore{deleteErr: rowErr}, &fakeColumnPermissionStore{}, &fakeDatasetFieldProvider{})
		err := svc.DeleteRowPermission(9)
		if !errors.Is(err, rowErr) {
			t.Fatalf("expected row delete error, got %v", err)
		}
	})

	t.Run("update get by id error", func(t *testing.T) {
		fieldProvider := &fakeDatasetFieldProvider{resp: &chart.ChartFieldListResponse{DimensionList: []chart.ChartField{{ID: 11, OriginName: "region"}}}}
		lookupErr := errors.New("row lookup failed")
		svc := NewDataPermissionAdminService(&fakeRowPermissionStore{getErr: lookupErr}, &fakeColumnPermissionStore{}, fieldProvider)
		err := svc.SaveRowPermission(&RowPermissionForm{ID: 8, DatasetID: 9, FilterType: permission.AuthTargetTypeUser, TargetID: 3, FilterField: "region", FilterValue: "west"})
		if !errors.Is(err, lookupErr) {
			t.Fatalf("expected row lookup error, got %v", err)
		}
	})

	t.Run("decode expression error still builds fallback row", func(t *testing.T) {
		rowStore := &fakeRowPermissionStore{items: []*permission.DataPermRow{{ID: 9, DatasetID: 9, AuthTargetType: permission.AuthTargetTypeUser, AuthTargetID: 3, ExpressionTree: "not-json", Status: 1}}}
		fieldProvider := &fakeDatasetFieldProvider{resp: &chart.ChartFieldListResponse{DimensionList: []chart.ChartField{{ID: 11, OriginName: "region"}}}}
		svc := NewDataPermissionAdminService(rowStore, &fakeColumnPermissionStore{}, fieldProvider)
		page, err := svc.RowPermissionPage(9, 1, 10)
		if err != nil {
			t.Fatalf("RowPermissionPage failed: %v", err)
		}
		list, ok := page.List.([]RowPermissionForm)
		if !ok || len(list) != 1 {
			t.Fatalf("unexpected row list: %#v", page.List)
		}
		if list[0].FilterField != "" || list[0].FilterValue != "" || list[0].Name != "规则-9" {
			t.Fatalf("expected fallback row mapping, got %#v", list[0])
		}
	})
}

func TestDataPermissionAdminService_ColumnPermissionStoreErrorPaths(t *testing.T) {
	t.Run("column page store error", func(t *testing.T) {
		svc := NewDataPermissionAdminService(&fakeRowPermissionStore{}, &fakeColumnPermissionStore{pagerErr: errors.New("column pager failed")}, &fakeDatasetFieldProvider{})
		page, err := svc.ColumnPermissionPage(9, 1, 10)
		if !errors.Is(err, svc.columnStore.(*fakeColumnPermissionStore).pagerErr) {
			t.Fatalf("expected column pager error, got %v", err)
		}
		if page != nil {
			t.Fatalf("expected nil page, got %#v", page)
		}
	})

	t.Run("delete error propagates", func(t *testing.T) {
		columnErr := errors.New("column delete failed")
		svc := NewDataPermissionAdminService(&fakeRowPermissionStore{}, &fakeColumnPermissionStore{deleteErr: columnErr}, &fakeDatasetFieldProvider{})
		err := svc.DeleteColumnPermission(12)
		if !errors.Is(err, columnErr) {
			t.Fatalf("expected column delete error, got %v", err)
		}
	})

	t.Run("update get by id error", func(t *testing.T) {
		fieldProvider := &fakeDatasetFieldProvider{resp: &chart.ChartFieldListResponse{DimensionList: []chart.ChartField{{ID: 22, OriginName: "mobile", Type: "string"}}}}
		lookupErr := errors.New("column lookup failed")
		svc := NewDataPermissionAdminService(&fakeRowPermissionStore{}, &fakeColumnPermissionStore{getErr: lookupErr}, fieldProvider)
		err := svc.SaveColumnPermission(&ColumnPermissionForm{ID: 6, DatasetID: 9, FieldName: "mobile", RuleType: permission.PermTypeMask, MaskRule: maskRuleKeepEnds})
		if !errors.Is(err, lookupErr) {
			t.Fatalf("expected column lookup error, got %v", err)
		}
	})

	t.Run("unknown field leaves type empty", func(t *testing.T) {
		columnStore := &fakeColumnPermissionStore{items: []*permission.DataPermColumn{{ID: 1, DatasetID: 9, FieldName: "missing", PermType: permission.PermTypeDisable, MaskRule: ""}}}
		fieldProvider := &fakeDatasetFieldProvider{resp: &chart.ChartFieldListResponse{DimensionList: []chart.ChartField{{ID: 21, OriginName: "known", Type: "string"}}}}
		svc := NewDataPermissionAdminService(&fakeRowPermissionStore{}, columnStore, fieldProvider)
		page, err := svc.ColumnPermissionPage(9, 1, 10)
		if err != nil {
			t.Fatalf("ColumnPermissionPage failed: %v", err)
		}
		list, ok := page.List.([]ColumnPermissionForm)
		if !ok || len(list) != 1 {
			t.Fatalf("unexpected column list: %#v", page.List)
		}
		if list[0].FieldType != "" {
			t.Fatalf("expected empty field type for unknown field, got %#v", list[0])
		}
		if list[0].MaskRule != maskRuleAll {
			t.Fatalf("expected empty mask rule to map to all, got %#v", list[0])
		}
	})
}

func TestDataPermissionAdminService_ColumnPermissionPage(t *testing.T) {
	t.Run("maps custom and disable rules", testColumnPermissionPageBase)
	t.Run("maps keep ends mask rule", testColumnPermissionPageKeepEnds)
}

func testColumnPermissionPageBase(t *testing.T) {
	customRuleBytes, err := json.Marshal(permission.DesensitizationRule{
		BuiltInRule:       permission.BuiltInRuleCustom,
		CustomBuiltInRule: permission.CustomRuleRetainBeforeMAndAfterN,
		M:                 2,
		N:                 4,
	})
	if err != nil {
		t.Fatalf("marshal custom rule failed: %v", err)
	}

	columnStore := &fakeColumnPermissionStore{items: []*permission.DataPermColumn{
		{
			ID:        1,
			DatasetID: 9,
			FieldName: "mobile",
			PermType:  permission.PermTypeMask,
			MaskRule:  string(customRuleBytes),
		},
		{
			ID:        2,
			DatasetID: 9,
			FieldName: "email",
			PermType:  permission.PermTypeDisable,
			MaskRule:  "",
		},
	}}
	fieldProvider := &fakeDatasetFieldProvider{resp: &chart.ChartFieldListResponse{DimensionList: []chart.ChartField{
		{ID: 21, OriginName: "mobile", Type: "string"},
		{ID: 22, OriginName: "email", Type: "string"},
	}}}

	svc := NewDataPermissionAdminService(&fakeRowPermissionStore{}, columnStore, fieldProvider)
	page, err := svc.ColumnPermissionPage(9, 0, 0)
	if err != nil {
		t.Fatalf("ColumnPermissionPage failed: %v", err)
	}

	list, ok := page.List.([]ColumnPermissionForm)
	if !ok {
		t.Fatalf("unexpected list type: %T", page.List)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 column permissions, got %d", len(list))
	}
	if page.Current != 1 || page.Size != 10 {
		t.Fatalf("expected normalized paging (1,10), got (%d,%d)", page.Current, page.Size)
	}
	if list[0].MaskRule != maskRuleCustom || list[0].MaskStart != 2 || list[0].MaskEnd != 4 {
		t.Fatalf("unexpected custom mask mapping: %#v", list[0])
	}
	if list[1].MaskRule != maskRuleAll {
		t.Fatalf("expected empty mask rule to map to all, got %#v", list[1])
	}
}

func testColumnPermissionPageKeepEnds(t *testing.T) {
	keepEndsBytes, err := json.Marshal(permission.DesensitizationRule{BuiltInRule: permission.BuiltInRuleKeepFirstAndLastThree})
	if err != nil {
		t.Fatalf("marshal keep ends rule failed: %v", err)
	}

	columnStore := &fakeColumnPermissionStore{items: []*permission.DataPermColumn{{ID: 3, DatasetID: 10, FieldName: "mobile", PermType: permission.PermTypeMask, MaskRule: string(keepEndsBytes)}}}
	fieldProvider := &fakeDatasetFieldProvider{resp: &chart.ChartFieldListResponse{DimensionList: []chart.ChartField{{ID: 21, OriginName: "mobile", Type: "string"}}}}
	svc := NewDataPermissionAdminService(&fakeRowPermissionStore{}, columnStore, fieldProvider)

	page, err := svc.ColumnPermissionPage(10, 1, 10)
	if err != nil {
		t.Fatalf("ColumnPermissionPage failed: %v", err)
	}
	list, ok := page.List.([]ColumnPermissionForm)
	if !ok || len(list) != 1 {
		t.Fatalf("unexpected column list: %#v", page.List)
	}
	if list[0].MaskRule != maskRuleKeepEnds {
		t.Fatalf("expected keep_ends mask mapping, got %#v", list[0])
	}
}

func TestDataPermissionAdminService_HelperBranches(t *testing.T) {
	t.Run("decode and encode mask helpers", testDataPermissionMaskHelpers)
	t.Run("apply mask helpers", testDataPermissionApplyMaskHelpers)
	t.Run("rule name helpers", testDataPermissionRuleNameHelpers)
}

func testDataPermissionMaskHelpers(t *testing.T) {
	if _, _, err := decodeSimpleExpression("not-json"); err == nil {
		t.Fatal("expected decodeSimpleExpression to fail for invalid json")
	}
	if fieldID, value, err := decodeSimpleExpression(""); err != nil || fieldID != 0 || value != "" {
		t.Fatalf("expected empty decode result, got fieldID=%d value=%q err=%v", fieldID, value, err)
	}

	maskAll, err := encodeMaskRule(&ColumnPermissionForm{RuleType: permission.PermTypeMask, MaskRule: maskRuleAll})
	if err != nil {
		t.Fatalf("encodeMaskRule all failed: %v", err)
	}
	if !strings.Contains(maskAll, string(permission.BuiltInRuleCompleteDesensitization)) {
		t.Fatalf("expected full mask rule, got %s", maskAll)
	}

	maskKeepEnds, err := encodeMaskRule(&ColumnPermissionForm{RuleType: permission.PermTypeMask, MaskRule: maskRuleKeepEnds})
	if err != nil {
		t.Fatalf("encodeMaskRule keep_ends failed: %v", err)
	}
	if !strings.Contains(maskKeepEnds, string(permission.BuiltInRuleKeepFirstAndLastThree)) {
		t.Fatalf("expected keep_ends rule, got %s", maskKeepEnds)
	}

	maskEmpty, err := encodeMaskRule(&ColumnPermissionForm{RuleType: permission.PermTypeDisable})
	if err != nil {
		t.Fatalf("encodeMaskRule non-mask failed: %v", err)
	}
	if maskEmpty != "" {
		t.Fatalf("expected non-mask rule to encode empty string, got %q", maskEmpty)
	}
}

func testDataPermissionApplyMaskHelpers(t *testing.T) {
	maskKeepEnds, err := encodeMaskRule(&ColumnPermissionForm{RuleType: permission.PermTypeMask, MaskRule: maskRuleKeepEnds})
	if err != nil {
		t.Fatalf("encodeMaskRule keep_ends failed: %v", err)
	}

	keepEndsForm := &ColumnPermissionForm{}
	applyMaskRuleToForm(keepEndsForm, maskKeepEnds)
	if keepEndsForm.MaskRule != maskRuleKeepEnds {
		t.Fatalf("expected keep_ends mapping, got %#v", keepEndsForm)
	}

	invalidForm := &ColumnPermissionForm{}
	applyMaskRuleToForm(invalidForm, "not-json")
	if invalidForm.MaskRule != maskRuleAll {
		t.Fatalf("expected invalid mask rule to fall back to all, got %#v", invalidForm)
	}

	customBytes, err := json.Marshal(permission.DesensitizationRule{BuiltInRule: permission.BuiltInRuleCustom, CustomBuiltInRule: permission.CustomRuleRetainBeforeMAndAfterN, M: 2, N: 5})
	if err != nil {
		t.Fatalf("marshal custom rule failed: %v", err)
	}
	customForm := &ColumnPermissionForm{}
	applyMaskRuleToForm(customForm, string(customBytes))
	if customForm.MaskRule != maskRuleCustom || customForm.MaskStart != 2 || customForm.MaskEnd != 5 {
		t.Fatalf("expected custom mask mapping, got %#v", customForm)
	}

	fieldID, value, err := decodeSimpleExpression(`{"logic":"OR","items":[]}`)
	if err != nil || fieldID != 0 || value != "" {
		t.Fatalf("expected empty items decode result, got fieldID=%d value=%q err=%v", fieldID, value, err)
	}
}

func testDataPermissionRuleNameHelpers(t *testing.T) {
	if buildRuleName("field", "value", 1) != "field = value" {
		t.Fatal("expected buildRuleName to use field/value pair")
	}
	if buildRuleName("field", "", 1) != "field" {
		t.Fatal("expected buildRuleName to fall back to field")
	}
	if buildRuleName("", "", 9) != "规则-9" {
		t.Fatal("expected buildRuleName to fall back to rule id")
	}
}

func TestDataPermissionAdminService_FieldProviderErrorsAndMaps(t *testing.T) {
	t.Run("field provider errors", testDataPermissionFieldProviderErrors)
	t.Run("dataset field maps", testDataPermissionDatasetFieldMaps)
	t.Run("display field helpers", testDataPermissionDisplayFieldHelpers)
}

func testDataPermissionFieldProviderErrors(t *testing.T) {
	fieldErr := errors.New("field provider failed")
	svc := NewDataPermissionAdminService(&fakeRowPermissionStore{}, &fakeColumnPermissionStore{}, &fakeDatasetFieldProvider{err: fieldErr})

	if _, err := svc.RowPermissionPage(9, 1, 10); !errors.Is(err, fieldErr) {
		t.Fatalf("expected row page field provider error, got %v", err)
	}
	if _, err := svc.ColumnPermissionPage(9, 1, 10); !errors.Is(err, fieldErr) {
		t.Fatalf("expected column page field provider error, got %v", err)
	}
	if _, err := svc.RowPermissionPageByTarget(9, permission.AuthTargetTypeUser, 1, 1, 10); !errors.Is(err, fieldErr) {
		t.Fatalf("expected target row page field provider error, got %v", err)
	}
}

func testDataPermissionDatasetFieldMaps(t *testing.T) {
	provider := &fakeDatasetFieldProvider{resp: &chart.ChartFieldListResponse{DimensionList: []chart.ChartField{{ID: 31, Name: "raw_name", OriginName: "origin_name", DataeaseName: "dataease_name", FieldShortName: "short_name"}}}}
	svc := NewDataPermissionAdminService(&fakeRowPermissionStore{}, &fakeColumnPermissionStore{}, provider)
	byID, byName, err := svc.datasetFieldMaps(9)
	if err != nil {
		t.Fatalf("datasetFieldMaps failed: %v", err)
	}
	if len(byID) != 1 || len(byName) != 4 {
		t.Fatalf("unexpected field maps: byID=%d byName=%d", len(byID), len(byName))
	}
	for _, key := range []string{"raw_name", "origin_name", "dataease_name", "short_name"} {
		if _, ok := byName[key]; !ok {
			t.Fatalf("expected alias %s to be indexed", key)
		}
	}

	provider = &fakeDatasetFieldProvider{resp: &chart.ChartFieldListResponse{DimensionList: []chart.ChartField{{ID: 31, Name: "dim_name", OriginName: "dim_origin"}}, QuotaList: []chart.ChartField{{ID: 32, Name: "quota_name", OriginName: "quota_origin", DataeaseName: "quota_dataease", FieldShortName: "quota_short"}}}}
	svc = NewDataPermissionAdminService(&fakeRowPermissionStore{}, &fakeColumnPermissionStore{}, provider)
	_, byName, err = svc.datasetFieldMaps(9)
	if err != nil {
		t.Fatalf("datasetFieldMaps with quota failed: %v", err)
	}
	for _, key := range []string{"quota_name", "quota_origin", "quota_dataease", "quota_short"} {
		if _, ok := byName[key]; !ok {
			t.Fatalf("expected quota alias %s to be indexed", key)
		}
	}
}

func testDataPermissionDisplayFieldHelpers(t *testing.T) {
	if displayFieldName(chart.ChartField{ID: 1, OriginName: "", Name: "", DataeaseName: "dataease", FieldShortName: "short"}) != "dataease" {
		t.Fatal("expected displayFieldName to fall back to dataease name")
	}
	if displayFieldName(chart.ChartField{ID: 2, Name: "field_name"}) != "field_name" {
		t.Fatal("expected displayFieldName to fall back to name")
	}
	if displayFieldName(chart.ChartField{ID: 3, Name: "", OriginName: "", DataeaseName: "", FieldShortName: "short_name"}) != "short_name" {
		t.Fatal("expected displayFieldName to fall back to field short name")
	}
	if displayFieldName(chart.ChartField{ID: 2}) != "field_2" {
		t.Fatal("expected displayFieldName to fall back to generated field name")
	}
	if normalizePage(0) != 1 || normalizePage(2) != 2 || normalizeSize(0) != 10 || normalizeSize(5) != 5 {
		t.Fatal("expected page/size normalization helpers to return defaults only when invalid")
	}
}
