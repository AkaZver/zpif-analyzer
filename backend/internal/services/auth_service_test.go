package services

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/zpif-analyzer/backend/internal/repositories"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"golang.org/x/crypto/bcrypt"
)

func setupTestAuthService(t *testing.T) (*AuthService, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 db,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm: %v", err)
	}

	userRepo := repositories.NewUserRepository(gormDB)
	service := NewAuthService(userRepo)

	cleanup := func() {
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
	}

	return service, mock, cleanup
}

func TestAuthService_Authenticate_Success(t *testing.T) {
	service, mock, cleanup := setupTestAuthService(t)
	defer cleanup()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "username", "password_hash", "email", "is_active",
	}).AddRow(1, now, now, nil, "admin", string(hashedPassword), "admin@test.com", true)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("admin", 1).
		WillReturnRows(rows)

	user, err := service.Authenticate("admin", "password123")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "admin", user.Username)
}

func TestAuthService_Authenticate_WrongPassword(t *testing.T) {
	service, mock, cleanup := setupTestAuthService(t)
	defer cleanup()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct_password"), bcrypt.DefaultCost)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "username", "password_hash", "email", "is_active",
	}).AddRow(1, now, now, nil, "admin", string(hashedPassword), "admin@test.com", true)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("admin", 1).
		WillReturnRows(rows)

	user, err := service.Authenticate("admin", "wrong_password")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "invalid credentials")
}

func TestAuthService_Authenticate_UserNotFound(t *testing.T) {
	service, mock, cleanup := setupTestAuthService(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("nonexistent", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	user, err := service.Authenticate("nonexistent", "password")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "invalid credentials")
}

func TestAuthService_Authenticate_DisabledAccount(t *testing.T) {
	service, mock, cleanup := setupTestAuthService(t)
	defer cleanup()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "username", "password_hash", "email", "is_active",
	}).AddRow(1, now, now, nil, "admin", string(hashedPassword), "admin@test.com", false)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("admin", 1).
		WillReturnRows(rows)

	user, err := service.Authenticate("admin", "password123")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "disabled")
}

func TestAuthService_CreateUser_Success(t *testing.T) {
	service, mock, cleanup := setupTestAuthService(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("newuser", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	mock.ExpectCommit()

	user, err := service.CreateUser("newuser", "password123", "new@test.com")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "newuser", user.Username)
	assert.Equal(t, "new@test.com", user.Email)
}

func TestAuthService_CreateUser_DuplicateUsername(t *testing.T) {
	service, mock, cleanup := setupTestAuthService(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "username", "password_hash", "email", "is_active",
	}).AddRow(1, now, now, nil, "existing", "hash", "existing@test.com", true)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("existing", 1).
		WillReturnRows(rows)

	user, err := service.CreateUser("existing", "password123", "new@test.com")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "already exists")
}

func TestAuthService_GetUserByID(t *testing.T) {
	service, mock, cleanup := setupTestAuthService(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "username", "password_hash", "email", "is_active",
	}).AddRow(1, now, now, nil, "admin", "hash", "admin@test.com", true)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
		WithArgs(uint(1), 1).
		WillReturnRows(rows)

	user, err := service.GetUserByID(1)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "admin", user.Username)
}
