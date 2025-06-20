package app

import (
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/config"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/modules/email"
	masterdata "gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/modules/master_data"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/services"
)

type serviceCollection struct {
	authSvc            services.AuthService
	memberSvc          services.MemberService
	customerSupportSvc services.CustomerSupportService
	masterDataSvc      masterdata.MasterDataProvider
}

func newServiceCollection(cfg config.Config, repos *repositoryCollection) (*serviceCollection, error) {
	mailSvc, err := email.NewEmailService(cfg.Email)
	if err != nil {
		return nil, err
	}

	return &serviceCollection{
		authSvc:            services.NewAuthService(cfg.Auth, repos.memberRepo, repos.registOtpRepo, mailSvc),
		memberSvc:          services.NewMemberService(repos.memberRepo),
		customerSupportSvc: services.NewCustomerSupportService(repos.adminWaRepo),
		masterDataSvc:      masterdata.NewMasterData(repos.indonesiaRegionsRepo),
	}, nil
}
