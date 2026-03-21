package service

import (
	"encoding/json"
	"strings"
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
		MaskRule:  "keep_ends",
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

func TestDataPermissionAdminService_ColumnPermissionPage(t *testing.T) {
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
	if list[0].MaskRule != "custom" || list[0].MaskStart != 2 || list[0].MaskEnd != 4 {
		t.Fatalf("unexpected custom mask mapping: %#v", list[0])
	}
	if list[1].MaskRule != "all" {
		t.Fatalf("expected empty mask rule to map to all, got %#v", list[1])
	}
}

func TestDataPermissionAdminService_HelperBranches(t *testing.T) {
	if _, _, err := decodeSimpleExpression("not-json"); err == nil {
		t.Fatal("expected decodeSimpleExpression to fail for invalid json")
	}
	if fieldID, value, err := decodeSimpleExpression(""); err != nil || fieldID != 0 || value != "" {
		t.Fatalf("expected empty decode result, got fieldID=%d value=%q err=%v", fieldID, value, err)
	}

	maskAll, err := encodeMaskRule(&ColumnPermissionForm{RuleType: permission.PermTypeMask, MaskRule: "all"})
	if err != nil {
		t.Fatalf("encodeMaskRule all failed: %v", err)
	}
	if !strings.Contains(maskAll, string(permission.BuiltInRuleCompleteDesensitization)) {
		t.Fatalf("expected full mask rule, got %s", maskAll)
	}

	maskKeepEnds, err := encodeMaskRule(&ColumnPermissionForm{RuleType: permission.PermTypeMask, MaskRule: "keep_ends"})
	if err != nil {
		t.Fatalf("encodeMaskRule keep_ends failed: %v", err)
	}
	if !strings.Contains(maskKeepEnds, string(permission.BuiltInRuleKeepFirstAndLastThree)) {
		t.Fatalf("expected keep_ends rule, got %s", maskKeepEnds)
	}

	keepEndsForm := &ColumnPermissionForm{}
	applyMaskRuleToForm(keepEndsForm, maskKeepEnds)
	if keepEndsForm.MaskRule != "keep_ends" {
		t.Fatalf("expected keep_ends mapping, got %#v", keepEndsForm)
	}

	invalidForm := &ColumnPermissionForm{}
	applyMaskRuleToForm(invalidForm, "not-json")
	if invalidForm.MaskRule != "all" {
		t.Fatalf("expected invalid mask rule to fall back to all, got %#v", invalidForm)
	}

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
