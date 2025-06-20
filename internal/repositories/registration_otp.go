package repositories

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/entities"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/helper/customerror"
)

type RegistrationOTPRepository interface {
	Save(ctx context.Context, regOtp entities.RegistrationOTP) error
	Get(ctx context.Context, memberId, memberRegBranchId int) (entities.RegistrationOTP, error)
	Delete(ctx context.Context, memberId, memberRegBranchId int) error
}

type SQLiteRegistrationOTPRepository struct {
	db *sqlx.DB
}

func NewSQLiteRegistrationOTPRepository(db *sqlx.DB) *SQLiteRegistrationOTPRepository {
	return &SQLiteRegistrationOTPRepository{db: db}
}

func (repo *SQLiteRegistrationOTPRepository) Save(ctx context.Context, regOtp entities.RegistrationOTP) error {
	q := `INSERT INTO registration_otps(
		member_id,
		member_reg_branch_id,
		otp, 
		expired_at,
		next_regeneration
	) VALUES (:member_id, :member_reg_branch_id, :otp, :expired_at, :next_regeneration)
	ON CONFLICT (member_id, member_reg_branch_id) DO UPDATE SET otp = :otp, expired_at = :expired_at, next_regeneration = :next_regeneration`
	_, err := repo.db.NamedExecContext(ctx, q, regOtp)
	if err != nil {
		return customerror.ErrDatabase.WithError(err).WithSource()
	}

	return nil
}

func (repo *SQLiteRegistrationOTPRepository) Get(ctx context.Context, memberId, memberRegBranchId int) (entities.RegistrationOTP, error) {
	var regOtp entities.RegistrationOTP
	q := `SELECT 
		member_id, 
		member_reg_branch_id,
		otp, 
		expired_at,
		next_regeneration
	FROM registration_otps 
	WHERE member_id = ? AND member_reg_branch_id = ?`
	row := repo.db.QueryRowxContext(ctx, q, memberId, memberRegBranchId)
	if err := row.StructScan(&regOtp); err != nil {
		if err == sql.ErrNoRows {
			return regOtp, customerror.ErrNotFound
		}

		return regOtp, customerror.ErrDatabase.WithError(err).WithSource()
	}

	return regOtp, nil
}

func (repo *SQLiteRegistrationOTPRepository) Delete(ctx context.Context, memberId, memberRegBranchId int) error {
	q := `DELETE FROM registration_otps WHERE member_id = ? AND member_reg_branch_id = ?`
	_, err := repo.db.ExecContext(ctx, q, memberId, memberRegBranchId)
	if err != nil {
		return customerror.ErrDatabase.WithError(err).WithSource()
	}

	return nil
}
