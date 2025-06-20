package entities

import (
	sqltypes "gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/helper/sql_types"
)

type RegistrationOTP struct {
	MemberID          int                `db:"member_id"`
	MemberRegBranchID int                `db:"member_reg_branch_id"`
	OTP               string             `db:"otp"`
	ExpiredAt         sqltypes.Timestamp `db:"expired_at"`
	NextRegeneration  sqltypes.Timestamp `db:"next_regeneration"`
}
