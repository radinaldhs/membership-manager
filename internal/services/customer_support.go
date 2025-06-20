package services

import (
	"context"
	"fmt"
	"net/url"

	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/helper/customerror"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/models"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/repositories"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/values"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/pkg/errors"
)

var errCustomerSupportUnavailable = errors.New(customerror.ErrGroupInternalErr, "-1", "Maaf, saat ini layanan customer support tidak tersedia, silahkan coba beberapa saat lagi")

type CustomerSupportService interface {
	GetAdminWhatsApp(ctx context.Context) (models.GetAdminWhatsAppResponse, error)
}

type customerSupportService struct {
	adminWaRepo repositories.AdminWhatsAppRepository
}

func NewCustomerSupportService(adminWaRepo repositories.AdminWhatsAppRepository) *customerSupportService {
	return &customerSupportService{
		adminWaRepo: adminWaRepo,
	}
}

func (svc *customerSupportService) GetAdminWhatsApp(ctx context.Context) (models.GetAdminWhatsAppResponse, error) {
	adminWa, err := svc.adminWaRepo.Get(ctx)
	if err != nil {
		return models.GetAdminWhatsAppResponse{}, errCustomerSupportUnavailable.WithError(err)
	}

	parsedPhoneNum, err := values.ParseDirtyPhoneNumber(adminWa.PhoneNumber)
	if err != nil {
		return models.GetAdminWhatsAppResponse{}, customerror.ErrInternal.WithError(err).WithSource()
	}

	baseUrl := fmt.Sprintf("https://wa.me/%s", parsedPhoneNum.WithIDCountryCode())
	urlBuild, err := url.Parse(baseUrl)
	if err != nil {
		return models.GetAdminWhatsAppResponse{}, customerror.ErrInternal.WithError(err).WithSource()
	}

	qParam := urlBuild.Query()
	qParam.Set("text", adminWa.TextMessage)

	urlBuild.RawQuery = qParam.Encode()

	return models.GetAdminWhatsAppResponse{
		URL: urlBuild.String(),
	}, nil
}
