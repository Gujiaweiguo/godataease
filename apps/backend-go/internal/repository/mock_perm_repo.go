package repository

import (
	"errors"

	"dataease/backend/internal/domain/permission"
)

var ErrNotFound = errors.New("not found")

type PermRepositoryInterface interface {
	Create(p *permission.SysPerm) error
	Update(p *permission.SysPerm) error
	Delete(permID int64) error
	GetByID(permID int64) (*permission.SysPerm, error)
	GetByKey(permKey string) (*permission.SysPerm, error)
	List() ([]*permission.SysPerm, error)
	CheckKeyExists(permKey string, excludePermID int64) (int64, error)
	GetByType(permType string) ([]*permission.SysPerm, error)
}

type MockPermRepository struct {
	perms      map[int64]*permission.SysPerm
	permsByKey map[string]*permission.SysPerm
	nextID     int64
}

func NewMockPermRepository() *MockPermRepository {
	return &MockPermRepository{
		perms:      make(map[int64]*permission.SysPerm),
		permsByKey: make(map[string]*permission.SysPerm),
		nextID:     1,
	}
}

func (m *MockPermRepository) Create(p *permission.SysPerm) error {
	p.PermID = m.nextID
	m.nextID++
	m.perms[p.PermID] = p
	m.permsByKey[p.PermKey] = p
	return nil
}

func (m *MockPermRepository) Update(p *permission.SysPerm) error {
	if existing, ok := m.perms[p.PermID]; ok {
		delete(m.permsByKey, existing.PermKey)
		m.perms[p.PermID] = p
		m.permsByKey[p.PermKey] = p
	}
	return nil
}

func (m *MockPermRepository) Delete(permID int64) error {
	if p, ok := m.perms[permID]; ok {
		delete(m.permsByKey, p.PermKey)
		delete(m.perms, permID)
	}
	return nil
}

func (m *MockPermRepository) GetByID(permID int64) (*permission.SysPerm, error) {
	if p, ok := m.perms[permID]; ok {
		return p, nil
	}
	return nil, ErrNotFound
}

func (m *MockPermRepository) GetByKey(permKey string) (*permission.SysPerm, error) {
	if p, ok := m.permsByKey[permKey]; ok {
		return p, nil
	}
	return nil, ErrNotFound
}

func (m *MockPermRepository) List() ([]*permission.SysPerm, error) {
	var perms []*permission.SysPerm
	for _, p := range m.perms {
		perms = append(perms, p)
	}
	return perms, nil
}

func (m *MockPermRepository) CheckKeyExists(permKey string, excludePermID int64) (int64, error) {
	if p, ok := m.permsByKey[permKey]; ok {
		if p.PermID != excludePermID {
			return 1, nil
		}
	}
	return 0, nil
}

func (m *MockPermRepository) GetByType(permType string) ([]*permission.SysPerm, error) {
	var perms []*permission.SysPerm
	for _, p := range m.perms {
		if p.PermType == permType {
			perms = append(perms, p)
		}
	}
	return perms, nil
}
