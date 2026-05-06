package threshold

import (
	"encoding/json"
	"testing"
)

func TestFilterTreeObj_JSONRoundTrip(t *testing.T) {
	original := FilterTreeObj{
		Logic: "and",
		Items: []FilterTreeItem{
			{
				Type:       "item",
				FieldID:    json.Number("100"),
				FilterType: "logic",
				Term:       "gt",
				Value:      "90",
				ValueType:  "fixed",
			},
			{
				Type:           "item",
				FieldID:        json.Number("200"),
				FilterType:     "enum",
				EnumValue:      []string{"a", "b", "c"},
				FilterTypeTime: "year",
				TimeType:       "exact",
			},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal FilterTreeObj: %v", err)
	}

	var decoded FilterTreeObj
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal FilterTreeObj: %v", err)
	}

	if decoded.Logic != "and" {
		t.Errorf("Logic: got %q, want 'and'", decoded.Logic)
	}
	if len(decoded.Items) != 2 {
		t.Fatalf("Items length: got %d, want 2", len(decoded.Items))
	}
	if decoded.Items[0].Term != "gt" {
		t.Errorf("Items[0].Term: got %q, want 'gt'", decoded.Items[0].Term)
	}
	if decoded.Items[0].Value != "90" {
		t.Errorf("Items[0].Value: got %q, want '90'", decoded.Items[0].Value)
	}
	assertStringSliceEqual(t, "Items[1].EnumValue", []string{"a", "b", "c"}, decoded.Items[1].EnumValue)
}

func TestFilterTreeObj_NestedSubTree(t *testing.T) {
	inner := FilterTreeObj{
		Logic: "or",
		Items: []FilterTreeItem{
			{Type: "item", FieldID: json.Number("1"), Term: "eq", Value: "10"},
		},
	}

	outer := FilterTreeObj{
		Logic: "and",
		Items: []FilterTreeItem{
			{Type: "item", FieldID: json.Number("2"), Term: "lt", Value: "5"},
			{Type: "tree", SubTree: &inner},
		},
	}

	data, err := json.Marshal(outer)
	if err != nil {
		t.Fatalf("marshal nested FilterTreeObj: %v", err)
	}

	var decoded FilterTreeObj
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal nested FilterTreeObj: %v", err)
	}

	if len(decoded.Items) != 2 {
		t.Fatalf("Items length: got %d, want 2", len(decoded.Items))
	}

	treeItem := decoded.Items[1]
	if treeItem.Type != "tree" {
		t.Errorf("nested item Type: got %q, want 'tree'", treeItem.Type)
	}
	if treeItem.SubTree == nil {
		t.Fatal("SubTree should not be nil")
	}
	if treeItem.SubTree.Logic != "or" {
		t.Errorf("SubTree.Logic: got %q, want 'or'", treeItem.SubTree.Logic)
	}
	if len(treeItem.SubTree.Items) != 1 {
		t.Fatalf("SubTree.Items length: got %d, want 1", len(treeItem.SubTree.Items))
	}
	if treeItem.SubTree.Items[0].Term != "eq" {
		t.Errorf("SubTree nested item Term: got %q, want 'eq'", treeItem.SubTree.Items[0].Term)
	}
}

func TestFilterTreeObj_DeeplyNested(t *testing.T) {
	level3 := FilterTreeObj{
		Logic: "or",
		Items: []FilterTreeItem{
			{Type: "item", FieldID: json.Number("3"), Term: "like", Value: "%test%"},
		},
	}
	level2 := FilterTreeObj{
		Logic: "and",
		Items: []FilterTreeItem{
			{Type: "tree", SubTree: &level3},
		},
	}
	level1 := FilterTreeObj{
		Logic: "and",
		Items: []FilterTreeItem{
			{Type: "tree", SubTree: &level2},
		},
	}

	data, err := json.Marshal(level1)
	if err != nil {
		t.Fatalf("marshal deeply nested: %v", err)
	}

	var decoded FilterTreeObj
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal deeply nested: %v", err)
	}

	inner := decoded.Items[0].SubTree.Items[0].SubTree
	if inner == nil {
		t.Fatal("level 3 SubTree should not be nil")
	}
	if inner.Items[0].Value != "%test%" {
		t.Errorf("deep nested value: got %q, want '%%test%%'", inner.Items[0].Value)
	}
}

func TestFilterTreeObj_Empty(t *testing.T) {
	obj := FilterTreeObj{}
	data, err := json.Marshal(obj)
	if err != nil {
		t.Fatalf("marshal empty: %v", err)
	}

	var decoded FilterTreeObj
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal empty: %v", err)
	}

	if decoded.Logic != "" {
		t.Errorf("Logic: got %q, want empty", decoded.Logic)
	}
	if len(decoded.Items) != 0 {
		t.Errorf("Items: got %d items, want 0", len(decoded.Items))
	}
}

func TestFilterTreeItem_FieldIDAsJSONNumber(t *testing.T) {
	input := `{"type":"item","fieldId":"42","filterType":"logic","term":"gt","value":"100"}`
	var item FilterTreeItem
	if err := json.Unmarshal([]byte(input), &item); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if item.FieldID.String() != "42" {
		t.Errorf("FieldID: got %q, want '42'", item.FieldID.String())
	}
}

func TestFilterTreeItem_FieldIDNumeric(t *testing.T) {
	input := `{"type":"item","fieldId":999,"filterType":"logic","term":"lt","value":"50"}`
	var item FilterTreeItem
	if err := json.Unmarshal([]byte(input), &item); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if item.FieldID.String() != "999" {
		t.Errorf("FieldID: got %q, want '999'", item.FieldID.String())
	}
}

func TestFilterTreeItem_FieldAsAny(t *testing.T) {
	tests := []struct {
		name  string
		field string
		want  string
	}{
		{"string field", `{"field":"name"}`, "name"},
		{"object field", `{"field":{"id":1,"name":"col"}}`, `{"id":1,"name":"col"}`},
		{"null field", `{"field":null}`, ""},
		{"number field", `{"field":42}`, "42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := tt.field
			var item FilterTreeItem
			if err := json.Unmarshal([]byte(input), &item); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if item.Field == nil && tt.want != "" {
				t.Errorf("Field: got nil, want non-nil")
			}
		})
	}
}

func TestFilterTreeItem_NilSubTree(t *testing.T) {
	item := FilterTreeItem{
		Type:    "item",
		FieldID: json.Number("1"),
		Term:    "eq",
		Value:   "100",
	}

	if item.SubTree != nil {
		t.Error("SubTree should be nil for non-tree items")
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded FilterTreeItem
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.SubTree != nil {
		t.Error("SubTree should remain nil after round-trip for item type")
	}
}

func TestFilterTreeObj_DeserializeFromRealWorldJSON(t *testing.T) {
	input := `{
		"logic": "and",
		"items": [
			{
				"type": "item",
				"fieldId": "1001",
				"field": {"id": 1001, "name": "sales", "type": "int"},
				"filterType": "logic",
				"term": "gt",
				"value": "1000",
				"enumValue": [],
				"valueType": "fixed"
			},
			{
				"type": "tree",
				"subTree": {
					"logic": "or",
					"items": [
						{
							"type": "item",
							"fieldId": "2002",
							"field": {"id": 2002, "name": "region", "type": "text"},
							"filterType": "enum",
							"term": "in",
							"value": "",
							"enumValue": ["US", "EU"],
							"valueType": "fixed"
						}
					]
				}
			}
		]
	}`

	var obj FilterTreeObj
	if err := json.Unmarshal([]byte(input), &obj); err != nil {
		t.Fatalf("unmarshal real-world JSON: %v", err)
	}

	if obj.Logic != "and" {
		t.Errorf("Logic: got %q, want 'and'", obj.Logic)
	}
	if len(obj.Items) != 2 {
		t.Fatalf("Items length: got %d, want 2", len(obj.Items))
	}

	item0 := obj.Items[0]
	if item0.Type != "item" {
		t.Errorf("Items[0].Type: got %q", item0.Type)
	}
	if item0.Term != "gt" {
		t.Errorf("Items[0].Term: got %q", item0.Term)
	}
	if item0.Value != "1000" {
		t.Errorf("Items[0].Value: got %q", item0.Value)
	}

	treeItem := obj.Items[1]
	if treeItem.Type != "tree" {
		t.Errorf("Items[1].Type: got %q", treeItem.Type)
	}
	if treeItem.SubTree == nil {
		t.Fatal("SubTree should not be nil")
	}
	if len(treeItem.SubTree.Items) != 1 {
		t.Fatalf("SubTree items: got %d, want 1", len(treeItem.SubTree.Items))
	}
	assertStringSliceEqual(t, "EnumValue", []string{"US", "EU"}, treeItem.SubTree.Items[0].EnumValue)
}

func TestFilterTreeItem_AllTerms(t *testing.T) {
	terms := []string{"eq", "not_eq", "lt", "gt", "le", "ge", "like", "not_like", "in", "not_in", "between", "is_null", "is_not_null", "empty", "not_empty"}

	for _, term := range terms {
		t.Run(term, func(t *testing.T) {
			item := FilterTreeItem{
				Type:       "item",
				FieldID:    json.Number("1"),
				FilterType: "logic",
				Term:       term,
				Value:      "test",
			}

			data, err := json.Marshal(item)
			if err != nil {
				t.Fatalf("marshal term %q: %v", term, err)
			}

			var decoded FilterTreeItem
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("unmarshal term %q: %v", term, err)
			}

			if decoded.Term != term {
				t.Errorf("Term: got %q, want %q", decoded.Term, term)
			}
		})
	}
}

func TestFilterTreeItem_TimeRelatedFields(t *testing.T) {
	item := FilterTreeItem{
		Type:           "item",
		FieldID:        json.Number("1"),
		FilterTypeTime: "year",
		TimeType:       "dynamic",
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded FilterTreeItem
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.FilterTypeTime != "year" {
		t.Errorf("FilterTypeTime: got %q, want 'year'", decoded.FilterTypeTime)
	}
	if decoded.TimeType != "dynamic" {
		t.Errorf("TimeType: got %q, want 'dynamic'", decoded.TimeType)
	}
}

func TestFilterTreeObj_LogicValues(t *testing.T) {
	for _, logic := range []string{"and", "or"} {
		t.Run(logic, func(t *testing.T) {
			obj := FilterTreeObj{Logic: logic}
			data, err := json.Marshal(obj)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var decoded FilterTreeObj
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if decoded.Logic != logic {
				t.Errorf("Logic: got %q, want %q", decoded.Logic, logic)
			}
		})
	}
}
