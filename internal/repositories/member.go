package repositories

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/entities"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/helper/customerror"
)

type MemberRepository interface {
	GetMembershipStatus(ctx context.Context, cardNum string) (entities.MembershipStatus, error)
	GetFailedLoginAttempt(ctx context.Context, id int, branchRegId int) (entities.MemberFailedLoginAttempts, bool, error)
	IncrFailedLoginAttempt(ctx context.Context, id, branchRegId int) error
	ResetLoginAttemptCounter(ctx context.Context, id, branchRegId int) error
	GetMember(ctx context.Context, id, regBranchId int) (entities.Member, error)
	GetMemberCompositeIDByCardNumber(ctx context.Context, cardNum string) (entities.MemberCompositeID, error)
	GetMemberEmail(ctx context.Context, memberId, memberBranchRegistId int) (string, error)
	IsMemberEmailVerified(ctx context.Context, memberId, memberRegBranchId int) (bool, error)
	IsEmailRegistered(ctx context.Context, email string) (bool, error)
	SaveMemberEmailAndPassword(ctx context.Context, memberId, memberRegBranchId int, email string, password []byte) error
	SetEmailAsVerified(ctx context.Context, memberId, memberRegBranchId int) error
	GetMemberEmailCreds(ctx context.Context, memberId, memberRegBranchId int) (entities.MemberEmailCreds, error)
	SetMemberEmailAndGoogleUserID(ctx context.Context, memberId, memberRegBranchId int, email, googleUserId string) error
	UpdateMember(ctx context.Context, member entities.Member) error
}

type memberRepository struct {
	db *sqlx.DB
}

func NewMemberRepository(db *sqlx.DB) *memberRepository {
	return &memberRepository{
		db: db,
	}
}

func (repo *memberRepository) GetMembershipStatus(ctx context.Context, cardNum string) (entities.MembershipStatus, error) {
	q := `SELECT
	(CASE WHEN mt.TipeMutasi IS NULL THEN 0 ELSE mt.TipeMutasi END) AS status,
	mm.IdMMember AS member_id,
	mm.IdMCabangDaftar AS member_registration_branch_id,
	mm.Nama AS member_name,
	mm.Telp AS member_phone_number
FROM member_mkartumember mk
	LEFT JOIN member_mutasimember mt ON mt.IdMKartuMemberLama = mk.IdMKartuMember
	LEFT JOIN member_mmember mm ON 
		mm.IdMKartuMember = mk.IdMKartuMember
		OR mm.IdMKartuMember = mt.IdMKartuMemberBaru
WHERE mk.NomorKartu = ?`

	var membershipStatus entities.MembershipStatus
	row := repo.db.QueryRowxContext(ctx, q, cardNum)
	if err := row.StructScan(&membershipStatus); err != nil {
		if err == sql.ErrNoRows {
			return entities.MembershipStatus{}, customerror.ErrNotFound
		}

		return entities.MembershipStatus{}, customerror.ErrDatabase.WithError(err).WithSource()
	}

	return membershipStatus, nil
}

// TODO: This might can be merged with GetMemberByCardNumber
func (repo *memberRepository) GetFailedLoginAttempt(ctx context.Context, id int, branchRegId int) (entities.MemberFailedLoginAttempts, bool, error) {
	q := `SELECT 
	Counter, 
	(
		DATE_FORMAT(
			TimeUpdate,
			CONCAT('%Y-%m-%d %H:%i:%s ', (SELECT @@global.time_zone))
		)
	) AS LastAttemptAt	
FROM member_percobaanLoginMember
WHERE 
	IdMMember = ?
	AND IdMCabangDaftar = ?`

	var loginAttempts entities.MemberFailedLoginAttempts
	row := repo.db.QueryRowxContext(ctx, q, id, branchRegId)
	if err := row.StructScan(&loginAttempts); err != nil {
		if err == sql.ErrNoRows {
			return loginAttempts, false, nil
		}

		return entities.MemberFailedLoginAttempts{}, false, customerror.ErrDatabase.WithError(err).WithSource()
	}

	return loginAttempts, true, nil
}

func (repo *memberRepository) IncrFailedLoginAttempt(ctx context.Context, id, branchRegId int) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return customerror.ErrDatabase.WithError(err).WithSource()
	}

	q := `CALL member_IncrementPercobaanLoginMember(?, ?)`

	_, err = tx.ExecContext(ctx, q, id, branchRegId)
	if err != nil {
		tx.Rollback()
		return customerror.ErrDatabase.WithError(err).WithSource()
	}

	if err := tx.Commit(); err != nil {
		return customerror.ErrDatabase.WithError(err).WithSource()
	}

	return nil
}

func (repo *memberRepository) ResetLoginAttemptCounter(ctx context.Context, id, branchRegId int) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return customerror.ErrDatabase.WithError(err).WithSource()
	}

	branchId, err := repo.tx_getBranchId(ctx, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	q := `CALL member_resetCounterPercobaanLoginMember(0, ?, NOW(), ?, ?, 0)`
	_, err = tx.ExecContext(ctx, q, branchId, id, branchRegId)
	if err != nil {
		tx.Rollback()
		return customerror.ErrDatabase.WithError(err).WithSource()
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return customerror.ErrDatabase.WithError(err).WithSource()
	}

	return nil
}

func (repo *memberRepository) tx_getBranchId(ctx context.Context, tx *sql.Tx) (int, error) {
	q := `SELECT getIdMCabang()`
	var branchId int

	row := tx.QueryRowContext(ctx, q)
	if err := row.Scan(&branchId); err != nil {
		return 0, customerror.ErrDatabase.WithError(err).WithSource()
	}

	return branchId, nil
}

func (repo *memberRepository) GetMember(ctx context.Context, id, regBranchId int) (entities.Member, error) {
	q := `SELECT
	mm.IdMMember AS id,
	mm.IdMCabangDaftar AS registration_branch_id,
	mm.Nama AS name,
	mtk.NamaTipe AS membership_type,
	mm.JmlPoint AS membership_points,
	mk.NomorKartu AS card_number,
	mm.Telp AS phone_number,
	mm.Alamat AS address,
	mm.Agama AS religion,
	mm.Propinsi AS province,
	mm.Kota AS regency,
	mv.Email AS email,
	mv.EmailVerified AS email_verified,
	mm.TglLahir AS date_of_birth,
	mm.Kelamin AS gender
FROM member_mmember mm 
	LEFT JOIN member_mkartumember mk ON mk.IdMKartuMember = mm.IdMKartuMember
	LEFT JOIN member_mtipekartumember mtk ON mtk.IdMTipeKartuMember = mk.IdMTipeKartuMember
	LEFT JOIN member_mmemberverification mv ON mv.IdMMember = mm.IdMMember AND mv.IdMCabangDaftar = mm.IdMCabangDaftar
WHERE 
	mm.IdMMember = ?
	AND mm.IdMCabangDaftar = ?`

	var member entities.Member
	row := repo.db.QueryRowxContext(ctx, q, id, regBranchId)
	if err := row.StructScan(&member); err != nil {
		if err == sql.ErrNoRows {
			return entities.Member{}, customerror.ErrNotFound
		}

		return entities.Member{}, customerror.ErrDatabase.WithError(err).WithSource()
	}

	return member, nil
}

func (repo *memberRepository) GetMemberCompositeIDByCardNumber(ctx context.Context, cardNum string) (entities.MemberCompositeID, error) {
	q := `SELECT	
	mm.IdMMember AS id,
	mm.IdMCabangDaftar AS registration_branch_id
FROM member_mkartumember mk	
	LEFT JOIN member_mmember mm ON 
		mm.IdMKartuMember = mk.IdMKartuMember	
WHERE mk.NomorKartu = ?;`

	var compositeId entities.MemberCompositeID
	row := repo.db.QueryRowxContext(ctx, q, cardNum)
	if err := row.StructScan(&compositeId); err != nil {
		if err == sql.ErrNoRows {
			return entities.MemberCompositeID{}, customerror.ErrNotFound
		}

		return entities.MemberCompositeID{}, customerror.ErrDatabase.WithError(err).WithSource()
	}

	return compositeId, nil
}

func (repo *memberRepository) GetMemberEmail(ctx context.Context, memberId, memberBranchRegistId int) (string, error) {
	q := `SELECT email FROM member_mmemberverification WHERE IdMMember = ? AND IdMCabangDaftar = ?`
	var email string
	row := repo.db.QueryRowContext(ctx, q, memberId, memberBranchRegistId)
	if err := row.Scan(&email); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}

		return "", customerror.ErrDatabase.WithError(err).WithSource()
	}

	return email, nil
}

func (repo *memberRepository) IsMemberEmailVerified(ctx context.Context, memberId, memberRegBranchId int) (bool, error) {
	q := `SELECT EmailVerified FROM member_mmemberverification
WHERE IdMMember = ? AND IdMCabangDaftar = ?;`

	var emailVerified bool
	row := repo.db.QueryRowContext(ctx, q, memberId, memberRegBranchId)
	if err := row.Scan(&emailVerified); err != nil {
		if err == sql.ErrNoRows {
			return false, customerror.ErrNotFound
		}

		return false, customerror.ErrDatabase.WithError(err).WithSource()
	}

	return emailVerified, nil
}

func (repo *memberRepository) IsEmailRegistered(ctx context.Context, email string) (bool, error) {
	q := `SELECT email FROM member_mmemberverification WHERE email = ?`
	var emailFromDb string
	row := repo.db.QueryRowContext(ctx, q, email)
	if err := row.Scan(&emailFromDb); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}

		return false, customerror.ErrDatabase.WithError(err).WithSource()
	}

	return true, nil
}

func (repo *memberRepository) SaveMemberEmailAndPassword(ctx context.Context, memberId, memberRegBranchId int, email string, password []byte) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return customerror.ErrDatabase.WithError(err).WithSource()
	}

	q := `CALL member_UpdateEmailDanPassword(?, ?, ?, ?)`

	_, err = tx.ExecContext(ctx, q, memberId, memberRegBranchId, email, password)
	if err != nil {
		tx.Rollback()
		return customerror.ErrDatabase.WithError(err).WithSource()
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return customerror.ErrDatabase.WithError(err).WithSource()
	}

	return nil
}

func (repo *memberRepository) SetEmailAsVerified(ctx context.Context, memberId, memberRegBranchId int) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return customerror.ErrDatabase.WithError(err).WithSource()
	}

	q := `CALL member_ValidateEmailViaOTP(?, ?)`

	_, err = tx.ExecContext(ctx, q, memberId, memberRegBranchId)
	if err != nil {
		tx.Rollback()
		return customerror.ErrDatabase.WithError(err).WithSource()
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return customerror.ErrDatabase.WithError(err).WithSource()
	}

	return nil
}

func (repo *memberRepository) GetMemberEmailCreds(ctx context.Context, memberId, memberRegBranchId int) (entities.MemberEmailCreds, error) {
	q := `SELECT
	mm.IdMMember AS member_id,
	mm.IdMCabangDaftar AS member_registration_branch_id,
	mk.NomorKartu AS card_number,
	(CASE WHEN mt.TipeMutasi IS NULL THEN 0 ELSE mt.TipeMutasi END) AS mutation_status,
	mv.Email AS email,
	mv.EmailVerified AS email_verified,
	mv.AppPassword AS password,
	mv.GoogleAccountID AS google_account_id
FROM member_mmemberverification mv
LEFT JOIN member_mmember mm ON 
	mm.IdMMember = mv.IdMMember
	AND mm.IdMCabangDaftar = mv.IdMCabangDaftar
LEFT JOIN member_mkartumember mk ON
	mk.IdMKartuMember = mm.IdMKartuMember	
LEFT JOIN member_mutasimember mt ON 
	mt.IdMKartuMemberLama = mk.IdMKartuMember
WHERE 
	mm.IdMMember = ?
	AND mm.IdMCabangDaftar = ?`

	var emailCreds entities.MemberEmailCreds
	row := repo.db.QueryRowxContext(ctx, q, memberId, memberRegBranchId)
	if err := row.StructScan(&emailCreds); err != nil {
		if err == sql.ErrNoRows {
			return entities.MemberEmailCreds{}, customerror.ErrNotFound
		}

		return entities.MemberEmailCreds{}, customerror.ErrInternal.WithError(err).WithSource()
	}

	return emailCreds, nil
}

func (repo *memberRepository) SetMemberEmailAndGoogleUserID(ctx context.Context, memberId, memberRegBranchId int, email, googleUserId string) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return customerror.ErrDatabase.WithError(err).WithSource()
	}

	q := `CALL member_UpdateGoogleAccountID(?, ?, ?, ?)`
	_, err = tx.ExecContext(ctx, q, memberId, memberRegBranchId, email, googleUserId)
	if err != nil {
		tx.Rollback()
		return customerror.ErrDatabase.WithError(err).WithSource()
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return customerror.ErrDatabase.WithError(err).WithSource()
	}

	return nil
}

func (repo *memberRepository) UpdateMember(ctx context.Context, member entities.Member) error {
	q := `CALL Member_UpdateUserInfo(
		:id,
		:registration_branch_id,
		:address,
		:regency,
		:province,
		:phone_number,
		:date_of_birth,
		:gender,
		:religion
	)`

	tx, err := repo.db.BeginTxx(ctx, nil)
	if err != nil {
		return customerror.ErrDatabase.WithError(err).WithSource()
	}

	_, err = tx.NamedExecContext(ctx, q, member)
	if err != nil {
		tx.Rollback()
		return customerror.ErrDatabase.WithError(err).WithSource()
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return customerror.ErrDatabase.WithError(err).WithSource()
	}

	return nil
}
