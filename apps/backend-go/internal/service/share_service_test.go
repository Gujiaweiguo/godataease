package service

import (
	"testing"
	"time"

	"dataease/backend/internal/domain/share"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestShareServiceHelpers_PasswordAndExpiry(t *testing.T) {
	pwd, err := generatePassword(8)
	assert.NoError(t, err)
	assert.Len(t, pwd, 8)

	assert.False(t, isShareExpired(0))
	assert.True(t, isShareExpired(time.Now().Add(-time.Hour).UnixMilli()))
	assert.False(t, isShareExpired(time.Now().Add(time.Hour).UnixMilli()))
	assert.True(t, isShareExpired(time.Now().Add(-time.Hour).Unix()))

	assert.True(t, isValidSharePassword("Ab1!"))
	assert.False(t, isValidSharePassword("abc"))
	assert.False(t, isValidSharePassword("abcdef"))
	assert.False(t, isValidSharePassword("123456"))
	assert.False(t, isValidSharePassword("Abcdef"))
	assert.False(t, isValidSharePassword("Ab1中文"))
}

func TestShareServiceHelpers_EnsureShareOwner(t *testing.T) {
	svc := &ShareService{}
	assert.Equal(t, gorm.ErrRecordNotFound, svc.ensureShareOwner(nil, 1))
	assert.EqualError(t, svc.ensureShareOwner(&share.Share{Creator: 1}, 2), "forbidden")
	assert.NoError(t, svc.ensureShareOwner(&share.Share{Creator: 1}, 1))
}
