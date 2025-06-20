package app

import (
	"context"
	"log/slog"

	firebase "firebase.google.com/go/v4"
	"github.com/jmoiron/sqlx"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/config"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/modules/notification"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/pkg/logging"
	"google.golang.org/api/option"

	_ "modernc.org/sqlite"
)

type modules struct {
	logger        *logging.Logger
	notifProvider notification.NotificationProvider
}

func newModules(cfg config.Config, db *sqlx.DB, logger *logging.Logger) (*modules, error) {
	var mods modules
	ctx := context.Background()

	// Init notification module
	firebaseClient, err := firebase.NewApp(ctx, nil, option.WithCredentialsJSON(cfg.Firebase.ServiceAccountCreds.Decoded))
	if err != nil {
		return nil, err
	}

	fcmClient, err := firebaseClient.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	notif, err := notification.NewNotification(cfg.Notification, notification.NewRepository(db), fcmClient, logger)
	if err != nil {
		return nil, err
	}

	mods.notifProvider = notif

	return &mods, nil
}

func (mods *modules) cleanup() {
	if err := mods.notifProvider.Close(); err != nil {
		mods.logger.Error(err.Error(), slog.String("module", "Notification"))
	}
}
