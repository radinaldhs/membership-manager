package services

import (
	"context"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/constants"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/helper/customerror"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/models"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/repositories"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/values"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/pkg/errors"
)

type MemberService interface {
	GetMember(ctx context.Context, id, regBranchId int) (models.GetMemberResponse, error)
	UpdateMember(ctx context.Context, memberId, memberRegBranchId int, req models.UpdateMemberRequest) error
}

type memberService struct {
	memberRepo repositories.MemberRepository
}

func NewMemberService(memberRepo repositories.MemberRepository) *memberService {
	return &memberService{
		memberRepo: memberRepo,
	}
}

var genderDisplayMap = map[string]string{
	"L": "Laki-laki",
	"P": "Perempuan",
}

func (svc *memberService) GetMember(ctx context.Context, id, regBranchId int) (models.GetMemberResponse, error) {
	member, err := svc.memberRepo.GetMember(ctx, id, regBranchId)
	if err != nil {
		if err == customerror.ErrNotFound {
			return models.GetMemberResponse{}, customerror.ErrMemberNotFound
		}

		return models.GetMemberResponse{}, err
	}

	cleanedPhoneNum, err := values.ParseDirtyPhoneNumber(member.PhoneNumber)
	if err != nil {
		return models.GetMemberResponse{}, err
	}

	religion := strcase.ToCamel(member.Religion)

	dob := member.DateOfBirth.Format("02 Jan 2006")

	resp := models.GetMemberResponse{
		ID:                   member.ID,
		RegistrationBranchID: member.RegistrationBranchID,
		Name:                 member.Name,
		MemberType:           member.MemberTypeName,
		Points:               member.Points,
		CardNumber:           member.CardNumber,
		PhoneNumber:          cleanedPhoneNum.Standard(),
		Address:              member.Address,
		Province:             member.Province,
		Regency:              member.Regency,
		Religion:             religion,
		Email:                member.Email,
		EmailVerified:        member.EmailVerified,
		DateOfBirth:          dob,
		Gender:               genderDisplayMap[member.Gender],
	}

	return resp, nil
}

func (svc *memberService) UpdateMember(ctx context.Context, memberId, memberRegBranchId int, req models.UpdateMemberRequest) error {
	member, err := svc.memberRepo.GetMember(ctx, memberId, memberRegBranchId)
	if err != nil {
		if err == customerror.ErrNotFound {
			return errors.New(customerror.ErrGroupDataNotFoundErr, "", "Member tidak ditemukan")
		}

		return err
	}

	if req.PhoneNumber != "" {
		member.PhoneNumber = req.PhoneNumber
	}

	if req.DateOfBirth != "" {
		date, err := time.Parse(time.DateOnly, req.DateOfBirth)
		if err != nil {
			return errors.New(customerror.ErrGroupClientErr, "", "Format tanggal lahir salah, format yang benar: yyyy-mm-dd")
		}

		member.DateOfBirth = date
	}

	if req.Gender != "" {
		gender := strings.ToUpper(req.Gender)
		switch gender {
		case constants.GenderMale, constants.GenderFemale:
			member.Gender = gender
		default:
			return errors.New(customerror.ErrGroupClientErr, "", "Jenis kelamin tidak valid")
		}
	}

	if req.Address != "" {
		member.Address = req.Address
	}

	if req.Province != "" {
		member.Province = req.Province
	}

	if req.Regency != "" {
		member.Regency = req.Regency
	}

	if req.Religion != "" {
		religion := strings.ToUpper(req.Religion)
		member.Religion = religion
	}

	err = svc.memberRepo.UpdateMember(ctx, member)

	return err
}
