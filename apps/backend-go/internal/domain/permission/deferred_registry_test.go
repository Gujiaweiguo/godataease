package permission

import (
	"strings"
	"testing"
)

func TestDeferredDimensionRegistry_IsDeferred_KnownDimensions(t *testing.T) {
	r := NewDeferredDimensionRegistry()
	for _, name := range []string{"sysParams", "whiteList", "dept"} {
		if !r.IsDeferred(name) {
			t.Fatalf("expected %q to be deferred", name)
		}
	}
}

func TestDeferredDimensionRegistry_IsDeferred_UnknownDimension(t *testing.T) {
	r := NewDeferredDimensionRegistry()
	if r.IsDeferred("unknown") {
		t.Fatal("expected unknown dimension to not be deferred")
	}
}

func TestDeferredDimensionRegistry_GetRejectionError_KnownDimensions(t *testing.T) {
	r := NewDeferredDimensionRegistry()
	cases := []struct {
		name      string
		errorCode string
	}{
		{"sysParams", "DEFERRED_DIMENSION_SYS_PARAMS"},
		{"whiteList", "DEFERRED_DIMENSION_WHITELIST"},
		{"dept", "DEFERRED_DIMENSION_DEPT"},
	}
	for _, tc := range cases {
		err := r.GetRejectionError(tc.name)
		if err == nil {
			t.Fatalf("expected error for %q", tc.name)
		}
		if !strings.HasPrefix(err.Error(), "["+tc.errorCode+"]") {
			t.Fatalf("expected error prefix [%s], got %q", tc.errorCode, err.Error())
		}
	}
}

func TestDeferredDimensionRegistry_GetRejectionError_UnknownDimension(t *testing.T) {
	r := NewDeferredDimensionRegistry()
	err := r.GetRejectionError("unknown")
	if err == nil {
		t.Fatal("expected error for unknown dimension")
	}
	if !strings.Contains(err.Error(), "not supported") {
		t.Fatalf("expected generic 'not supported' error, got %q", err.Error())
	}
}

func TestDeferredDimensionRegistry_ListDeferred(t *testing.T) {
	r := NewDeferredDimensionRegistry()
	list := r.ListDeferred()
	if len(list) != 3 {
		t.Fatalf("expected 3 deferred dimensions, got %d", len(list))
	}
}
