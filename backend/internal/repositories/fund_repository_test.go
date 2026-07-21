package repositories

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/zpif-analyzer/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
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

	return gormDB, mock
}

func TestFundRepository_GetAll(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewFundRepository(gormDB)
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_start_date", "fund_end_date", "created_at", "updated_at", "deleted_at"}).
		AddRow(1, "Парус ОЗН", "RU000A1022Z1", "", "Парус", "склады", false, true, nil, nil, now, now, nil).
		AddRow(2, "Акцент 5", "RU000A10DQF7", "", "Акцент", "офисы", true, false, nil, nil, now, now, nil)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."deleted_at" IS NULL`)).WillReturnRows(rows)

	emptyRows := sqlmock.NewRows([]string{"id", "fund_id"})
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" IN`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" IN`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" IN`).WillReturnRows(emptyRows)

	funds, err := repo.GetAll()

	assert.NoError(t, err)
	assert.Len(t, funds, 2)
	assert.Equal(t, "Парус ОЗН", funds[0].Name)
	assert.Equal(t, "Акцент 5", funds[1].Name)
}

func TestFundRepository_GetByID(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewFundRepository(gormDB)
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_start_date", "fund_end_date", "created_at", "updated_at", "deleted_at"}).
		AddRow(1, "Парус ОЗН", "RU000A1022Z1", "PARUS", "Парус", "склады", false, true, nil, nil, now, now, nil)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."id" =`)).
		WithArgs(1, 1).
		WillReturnRows(rows)

	emptyRows := sqlmock.NewRows([]string{"id", "fund_id"})
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" =`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" =`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" =`).WillReturnRows(emptyRows)

	fund, err := repo.GetByID(1)

	assert.NoError(t, err)
	assert.NotNil(t, fund)
	assert.Equal(t, "Парус ОЗН", fund.Name)
	assert.Equal(t, "PARUS", fund.Ticker)
}

func TestFundRepository_GetByID_NotFound(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewFundRepository(gormDB)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."id" =`)).
		WithArgs(999, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	fund, err := repo.GetByID(999)

	assert.Error(t, err)
	assert.Nil(t, fund)
}

func TestFundRepository_GetByISIN(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewFundRepository(gormDB)
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_start_date", "fund_end_date", "created_at", "updated_at", "deleted_at"}).
		AddRow(1, "Парус ОЗН", "RU000A1022Z1", "", "Парус", "склады", false, true, nil, nil, now, now, nil)

	mock.ExpectQuery(`isin = .+ AND "funds"\."deleted_at" IS NULL`).
		WithArgs("RU000A1022Z1", 1).
		WillReturnRows(rows)

	fund, err := repo.GetByISIN("RU000A1022Z1")

	assert.NoError(t, err)
	assert.NotNil(t, fund)
	assert.Equal(t, "RU000A1022Z1", fund.ISIN)
}

func TestFundRepository_Create(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewFundRepository(gormDB)

	fund := &models.Fund{
		Name:              "Тестовый фонд",
		ISIN:              "RU000TEST001",
		ManagementCompany: "Тест УК",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "funds"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := repo.Create(fund)

	assert.NoError(t, err)
}

func TestFundRepository_Delete(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewFundRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "funds" SET "deleted_at"=`).
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Delete(1)

	assert.NoError(t, err)
}

func TestFundRepository_Update(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewFundRepository(gormDB)

	now := time.Now()
	fund := &models.Fund{
		ID:                1,
		Name:              "Updated Name",
		ISIN:              "RU000A1022Z1",
		ManagementCompany: "Парус",
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "funds" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Update(fund)

	assert.NoError(t, err)
}

func TestFundRepository_GetAll_Empty(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewFundRepository(gormDB)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."deleted_at" IS NULL`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_start_date", "fund_end_date", "created_at", "updated_at", "deleted_at"}))

	funds, err := repo.GetAll()

	assert.NoError(t, err)
	assert.Empty(t, funds)
}
