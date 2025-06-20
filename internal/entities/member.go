package entities

import (
	"time"

	sqltypes "gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/helper/sql_types"
)

// TODO: Maybe use this in Member struct
type MemberCompositeID struct {
	ID                   int `db:"id"`
	RegistrationBranchID int `db:"registration_branch_id"`
}

// Member holds the member profile information.
// We use custom column name because this data is composed
// from multiple tables, so to avoid confusion, using
// custom column names will be more readable
type Member struct {
	ID                   int       `db:"id"`
	RegistrationBranchID int       `db:"registration_branch_id"`
	CardNumber           string    `db:"card_number"`
	Name                 string    `db:"name"`
	PhoneNumber          string    `db:"phone_number"`
	MemberTypeName       string    `db:"membership_type"`
	Points               float64   `db:"membership_points"`
	Province             string    `db:"province"`
	Regency              string    `db:"regency"`
	Address              string    `db:"address"`
	Religion             string    `db:"religion"`
	Email                string    `db:"email"`
	EmailVerified        bool      `db:"email_verified"`
	DateOfBirth          time.Time `db:"date_of_birth"`
	Gender               string    `db:"gender"`
}

type MembershipStatus struct {
	Status                     int    `db:"status"`
	MemberID                   int    `db:"member_id"`
	MemberRegistrationBranchID int    `db:"member_registration_branch_id"`
	MemberName                 string `db:"member_name"`
	MemberPhoneNumber          string `db:"member_phone_number"`
}

type MemberFailedLoginAttempts struct {
	Counter     int                           `db:"Counter"`
	LastAttempt sqltypes.DateTimeWithTimezone `db:"LastAttemptAt"`
}

type MemberEmailCreds struct {
	MemberID                   int    `db:"member_id"`
	MemberRegistrationBranchID int    `db:"member_registration_branch_id"`
	MutationStatus             int    `db:"mutation_status"`
	CardNumber                 string `db:"card_number"`
	Email                      string `db:"email"`
	EmailVerified              bool   `db:"email_verified"`
	Password                   []byte `db:"password"`
	GoogleAccountID            string `db:"google_account_id"`
}
