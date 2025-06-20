package masterdata

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/helper/customerror"
)

type RegionLevel int

type Region struct {
	RegionCode string `db:"region_code" json:"region_code"`
	Name       string `db:"name" json:"name"`
}

type IndonesiaRegionRepository interface {
	GetAllProvince(ctx context.Context) ([]Region, error)
	GetAllRegencyByProvinceRegionCode(ctx context.Context, provinceRegionCode string) ([]Region, error)
}

type SQLiteIndonesiaRegionRepository struct {
	db *sqlx.DB
}

func NewSQLiteIndonesiaRegionRepository(db *sqlx.DB) *SQLiteIndonesiaRegionRepository {
	return &SQLiteIndonesiaRegionRepository{db: db}
}

func (repo *SQLiteIndonesiaRegionRepository) GetAllProvince(ctx context.Context) ([]Region, error) {
	q := `SELECT region_code, name FROM master_data_indonesia_provinces;`

	rows, err := repo.db.QueryxContext(ctx, q)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, customerror.ErrDatabase.WithError(err).WithSource()
	}

	var regions []Region
	for rows.Next() {
		var region Region
		if err := rows.StructScan(&region); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, customerror.ErrDatabase.WithError(err).WithSource()
		}

		regions = append(regions, region)
	}

	return regions, nil
}

func (repo *SQLiteIndonesiaRegionRepository) GetAllRegencyByProvinceRegionCode(ctx context.Context, provinceRegionCode string) ([]Region, error) {
	q := `SELECT region_code, name FROM master_data_indonesia_regencies WHERE province_region_code = ?`
	rows, err := repo.db.QueryxContext(ctx, q, provinceRegionCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, customerror.ErrDatabase.WithError(err).WithSource()
	}

	var regions []Region
	for rows.Next() {
		var region Region
		if err := rows.StructScan(&region); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, customerror.ErrDatabase.WithError(err).WithSource()
		}

		regions = append(regions, region)
	}

	return regions, nil
}
