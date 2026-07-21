package repositories

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/zpif-analyzer/backend/internal/models"
	"gorm.io/gorm"
)

func TestUserRepository_GetByUsername(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewUserRepository(gormDB)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "username", "password_hash", "email", "is_active",
	}).AddRow(1, now, now, nil, "admin", "hashed_password", "admin@test.com", true)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("admin", 1).
		WillReturnRows(rows)

	user, err := repo.GetByUsername("admin")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "admin", user.Username)
	assert.Equal(t, "admin@test.com", user.Email)
	assert.True(t, user.IsActive)
}

func TestUserRepository_GetByUsername_NotFound(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewUserRepository(gormDB)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("nonexistent", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	user, err := repo.GetByUsername("nonexistent")

	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUserRepository_GetByID(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewUserRepository(gormDB)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "username", "password_hash", "email", "is_active",
	}).AddRow(1, now, now, nil, "admin", "hashed_password", "admin@test.com", true)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(rows)

	user, err := repo.GetByID(1)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, uint(1), user.ID)
	assert.Equal(t, "admin", user.Username)
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewUserRepository(gormDB)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
		WithArgs(999, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	user, err := repo.GetByID(999)

	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUserRepository_Create(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewUserRepository(gormDB)
	now := time.Now()

	user := &models.User{
		Username:     "newuser",
		PasswordHash: "hashed",
		Email:        "new@test.com",
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	mock.ExpectCommit()

	err := repo.Create(user)

	assert.NoError(t, err)
}
