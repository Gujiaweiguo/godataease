package feature

import (
	"sync"
)

type FeatureFlag struct {
	Name        string
	Enabled     bool
	Percentage  int
	TenantList  []string
	ExcludeList []string
}

type ToggleManager struct {
	flags map[string]*FeatureFlag
	mu    sync.RWMutex
}

func NewToggleManager() *ToggleManager {
	return &ToggleManager{
		flags: make(map[string]*FeatureFlag),
	}
}

func (m *ToggleManager) Register(name string, enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.flags[name] = &FeatureFlag{
		Name:    name,
		Enabled: enabled,
	}
}

func (m *ToggleManager) IsEnabled(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if flag, ok := m.flags[name]; ok {
		return flag.Enabled
	}
	return false
}

func (m *ToggleManager) Enable(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if flag, ok := m.flags[name]; ok {
		flag.Enabled = true
	}
}

func (m *ToggleManager) Disable(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if flag, ok := m.flags[name]; ok {
		flag.Enabled = false
	}
}

func (m *ToggleManager) SetPercentage(name string, percentage int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if flag, ok := m.flags[name]; ok && percentage >= 0 && percentage <= 100 {
		flag.Percentage = percentage
	}
}

func (m *ToggleManager) AddTenant(name, tenantID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if flag, ok := m.flags[name]; ok {
		flag.TenantList = append(flag.TenantList, tenantID)
	}
}

func (m *ToggleManager) ExcludeTenant(name, tenantID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if flag, ok := m.flags[name]; ok {
		flag.ExcludeList = append(flag.ExcludeList, tenantID)
	}
}

func (m *ToggleManager) IsEnabledForTenant(name, tenantID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	flag, ok := m.flags[name]
	if !ok {
		return false
	}

	for _, excluded := range flag.ExcludeList {
		if excluded == tenantID {
			return false
		}
	}

	if flag.Enabled {
		return true
	}

	for _, allowed := range flag.TenantList {
		if allowed == tenantID {
			return true
		}
	}

	return false
}

func (m *ToggleManager) GetAllFlags() map[string]*FeatureFlag {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*FeatureFlag)
	for k, v := range m.flags {
		result[k] = v
	}
	return result
}
