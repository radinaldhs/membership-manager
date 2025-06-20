package app

import (
	"github.com/jmoiron/sqlx"
	masterdata "gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/modules/master_data"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/repositories"
)

type repositoryCollection struct {
	memberRepo           repositories.MemberRepository
	adminWaRepo          repositories.AdminWhatsAppRepository
	registOtpRepo        repositories.RegistrationOTPRepository
	indonesiaRegionsRepo masterdata.IndonesiaRegionRepository
}

func newRepositoryCollection(db, sqlitedb *sqlx.DB) *repositoryCollection {
	repos := repositoryCollection{
		memberRepo:           repositories.NewMemberRepository(db),
		adminWaRepo:          repositories.NewAdminWhatsAppRepository(db),
		registOtpRepo:        repositories.NewSQLiteRegistrationOTPRepository(sqlitedb),
		indonesiaRegionsRepo: masterdata.NewSQLiteIndonesiaRegionRepository(sqlitedb),
	}

	return &repos
}
