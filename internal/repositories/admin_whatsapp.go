package repositories

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/entities"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/helper/customerror"
)

type AdminWhatsAppRepository interface {
	Get(ctx context.Context) (entities.AdminWhatsApp, error)
}

type adminWhatsAppRepository struct {
	db *sqlx.DB
}

func NewAdminWhatsAppRepository(db *sqlx.DB) *adminWhatsAppRepository {
	return &adminWhatsAppRepository{db: db}
}

func (repo *adminWhatsAppRepository) Get(ctx context.Context) (entities.AdminWhatsApp, error) {
	q := `SELECT 
	NomorTelp AS phone_number,
	TeksPesan AS text_message,
	Description AS description
FROM member_mwhatsappadmin mm`

	var adminWa entities.AdminWhatsApp
	row := repo.db.QueryRowxContext(ctx, q)
	if err := row.StructScan(&adminWa); err != nil {
		if err == sql.ErrNoRows {
			return entities.AdminWhatsApp{}, customerror.ErrNotFound
		}

		return entities.AdminWhatsApp{}, customerror.ErrDatabase.WithError(err).WithSource()
	}

	return adminWa, nil
}
