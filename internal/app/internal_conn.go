package app

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/config"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/pkg/logging"
)

type internalConnections struct {
	logger   *logging.Logger
	db       *sqlx.DB
	sqliteDb *sqlx.DB
}

func newInternalConnections(cfg config.Config, logger *logging.Logger) (*internalConnections, error) {
	var conns internalConnections
	conns.logger = logger

	// Connect to database
	mysqlCfg := cfg.Database.MysqlConfig()
	// TODO: I'm not confident on using locale like this, but for now this will works
	mysqlCfg.Loc = time.Local
	connector, err := mysql.NewConnector(mysqlCfg)
	if err != nil {
		return nil, err
	}

	db := sql.OpenDB(connector)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime))
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)

	xdb := sqlx.NewDb(db, "mysql")
	conns.db = xdb

	// SQLite database
	sqliteDb, err := initSqliteDatabase(cfg.SQLite.DBFile)
	if err != nil {
		return nil, err
	}

	conns.sqliteDb = sqliteDb

	return &conns, nil
}

func initSqliteDatabase(dbfile string) (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite", dbfile)
	if err != nil {
		return nil, fmt.Errorf("open SQLite database error: %v", err)
	}

	return db, nil
}

func (ic *internalConnections) close() {
	if err := ic.db.Close(); err != nil {
		ic.logger.Error(fmt.Sprintf("Closing database connection error: %s", err.Error()))
	}
}
