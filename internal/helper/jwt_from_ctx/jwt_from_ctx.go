package jwtfromctx

import (
	"errors"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/helper/customerror"
)

type MemberIdentityPair struct {
	MemberID             int
	RegistrationBranchID int
}

var errSubjectClaimIsNotValid = errors.New("subject claim is not valid")

func GetMemberIdentityPair(ectx echo.Context) (MemberIdentityPair, error) {
	user, ok := ectx.Get("user").(string)
	if !ok {
		return MemberIdentityPair{}, customerror.ErrInternal.WithError(errors.New("user is not a string")).WithSource()
	}

	split := strings.Split(user, "-")
	if len(split) != 2 {
		return MemberIdentityPair{}, errSubjectClaimIsNotValid
	}

	memberId, err := strconv.Atoi(split[0])
	if err != nil {
		return MemberIdentityPair{}, errSubjectClaimIsNotValid
	}

	regBranchId, err := strconv.Atoi(split[1])
	if err != nil {
		return MemberIdentityPair{}, errSubjectClaimIsNotValid
	}

	return MemberIdentityPair{
		MemberID:             memberId,
		RegistrationBranchID: regBranchId,
	}, nil
}
