package feature

import (
	"testing"
)

func TestNewToggleManager(t *testing.T) {
	tm := NewToggleManager()
	if tm == nil {
		t.Fatal("Expected ToggleManager instance, got nil")
	}
	if tm.flags == nil {
		t.Error("Expected flags map to be initialized")
	}
}

func TestRegister(t *testing.T) {
	tm := NewToggleManager()
	tm.Register("test-feature", true)

	if len(tm.flags) != 1 {
		t.Errorf("Expected 1 flag, got %d", len(tm.flags))
	}

	flag := tm.flags["test-feature"]
	if flag == nil {
		t.Fatal("Expected flag to be registered")
	}
	if flag.Name != "test-feature" {
		t.Errorf("Expected name 'test-feature', got '%s'", flag.Name)
	}
	if !flag.Enabled {
		t.Error("Expected flag to be enabled")
	}
}

func TestIsEnabled(t *testing.T) {
	tm := NewToggleManager()
	tm.Register("enabled-feature", true)
	tm.Register("disabled-feature", false)

	if !tm.IsEnabled("enabled-feature") {
		t.Error("Expected 'enabled-feature' to be enabled")
	}
	if tm.IsEnabled("disabled-feature") {
		t.Error("Expected 'disabled-feature' to be disabled")
	}
	if tm.IsEnabled("non-existent") {
		t.Error("Expected non-existent flag to be disabled")
	}
}

func TestEnable(t *testing.T) {
	tm := NewToggleManager()
	tm.Register("toggle-feature", false)

	if tm.IsEnabled("toggle-feature") {
		t.Error("Expected feature to start disabled")
	}

	tm.Enable("toggle-feature")

	if !tm.IsEnabled("toggle-feature") {
		t.Error("Expected feature to be enabled after Enable()")
	}
}

func TestDisable(t *testing.T) {
	tm := NewToggleManager()
	tm.Register("toggle-feature", true)

	if !tm.IsEnabled("toggle-feature") {
		t.Error("Expected feature to start enabled")
	}

	tm.Disable("toggle-feature")

	if tm.IsEnabled("toggle-feature") {
		t.Error("Expected feature to be disabled after Disable()")
	}
}

func TestEnableDisableNonExistent(t *testing.T) {
	tm := NewToggleManager()

	tm.Enable("non-existent")
	if tm.IsEnabled("non-existent") {
		t.Error("Enabling non-existent flag should not create it")
	}

	tm.Disable("non-existent")
}

func TestSetPercentage(t *testing.T) {
	tm := NewToggleManager()
	tm.Register("percent-feature", true)

	tm.SetPercentage("percent-feature", 50)

	flag := tm.flags["percent-feature"]
	if flag.Percentage != 50 {
		t.Errorf("Expected percentage 50, got %d", flag.Percentage)
	}
}

func TestSetPercentageInvalid(t *testing.T) {
	tm := NewToggleManager()
	tm.Register("percent-feature", true)

	tm.SetPercentage("percent-feature", 150)
	flag := tm.flags["percent-feature"]
	if flag.Percentage == 150 {
		t.Error("Should not accept percentage > 100")
	}

	tm.SetPercentage("percent-feature", -10)
	if flag.Percentage == -10 {
		t.Error("Should not accept negative percentage")
	}
}

func TestAddTenant(t *testing.T) {
	tm := NewToggleManager()
	tm.Register("tenant-feature", false)

	tm.AddTenant("tenant-feature", "tenant-1")
	tm.AddTenant("tenant-feature", "tenant-2")

	flag := tm.flags["tenant-feature"]
	if len(flag.TenantList) != 2 {
		t.Errorf("Expected 2 tenants, got %d", len(flag.TenantList))
	}
}

func TestExcludeTenant(t *testing.T) {
	tm := NewToggleManager()
	tm.Register("tenant-feature", true)

	tm.ExcludeTenant("tenant-feature", "blocked-tenant")

	flag := tm.flags["tenant-feature"]
	if len(flag.ExcludeList) != 1 {
		t.Errorf("Expected 1 excluded tenant, got %d", len(flag.ExcludeList))
	}
}

func TestIsEnabledForTenant(t *testing.T) {
	tm := NewToggleManager()
	tm.Register("tenant-feature", false)
	tm.AddTenant("tenant-feature", "allowed-tenant")

	if !tm.IsEnabledForTenant("tenant-feature", "allowed-tenant") {
		t.Error("Allowed tenant should have access")
	}

	if tm.IsEnabledForTenant("tenant-feature", "other-tenant") {
		t.Error("Non-allowed tenant should not have access")
	}
}

func TestIsEnabledForTenant_Excluded(t *testing.T) {
	tm := NewToggleManager()
	tm.Register("tenant-feature", true)
	tm.ExcludeTenant("tenant-feature", "blocked-tenant")

	if tm.IsEnabledForTenant("tenant-feature", "blocked-tenant") {
		t.Error("Excluded tenant should not have access even when feature is enabled")
	}
}

func TestIsEnabledForTenant_NonExistent(t *testing.T) {
	tm := NewToggleManager()

	if tm.IsEnabledForTenant("non-existent", "tenant-1") {
		t.Error("Non-existent feature should return false")
	}
}

func TestGetAllFlags(t *testing.T) {
	tm := NewToggleManager()
	tm.Register("feature-1", true)
	tm.Register("feature-2", false)

	flags := tm.GetAllFlags()

	if len(flags) != 2 {
		t.Errorf("Expected 2 flags, got %d", len(flags))
	}

	if flags["feature-1"] == nil || flags["feature-2"] == nil {
		t.Error("Expected both flags to be present")
	}
}

func TestGetAllFlags_ReturnsMap(t *testing.T) {
	tm := NewToggleManager()
	tm.Register("feature-1", true)

	flags := tm.GetAllFlags()

	if len(flags) != 1 {
		t.Errorf("Expected 1 flag, got %d", len(flags))
	}

	flags["feature-2"] = &FeatureFlag{Name: "feature-2", Enabled: false}

	if tm.IsEnabled("feature-2") {
		t.Error("Adding to returned map should not affect original")
	}
}

func TestConcurrentAccess(t *testing.T) {
	tm := NewToggleManager()
	tm.Register("concurrent-feature", true)

	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				tm.IsEnabled("concurrent-feature")
				tm.Enable("concurrent-feature")
				tm.Disable("concurrent-feature")
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
