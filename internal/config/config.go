package config

import (
	"fmt"

	"github.com/go-sql-driver/mysql"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/pkg/config"
)

type Database struct {
	Host            string                `env:"DB_HOST"`
	Port            string                `env:"DB_PORT"`
	DatabaseName    string                `env:"DB_NAME"`
	Username        string                `env:"DB_USERNAME"`
	Password        string                `env:"DB_PASSWORD"`
	TLS             string                `env:"DB_TLS" default:"skip-verify"`
	ConnMaxLifetime config.MinuteDuration `env:"DB_CONN_MAX_LIFETIME" default:"3"`
	MaxOpenConns    int                   `env:"DB_MAX_OPEN_CONNS" default:"10"`
	MaxIdleConns    int                   `env:"DB_MAX_IDLE_CONNS" default:"10"`
}

func (d Database) MysqlConfig() *mysql.Config {
	cfg := mysql.Config{
		User:                 d.Username,
		Passwd:               d.Password,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%s", d.Host, d.Port),
		DBName:               d.DatabaseName,
		TLSConfig:            d.TLS,
		CheckConnLiveness:    true,
		AllowNativePasswords: true,
		ParseTime:            true,
	}

	return &cfg
}

type Firebase struct {
	ProjectId           string                  `env:"FIREBASE_PROJECT_ID"`
	ServiceAccountCreds config.RawBase64Encoded `env:"FIREBASE_SERVICE_ACCOUNT_CREDS"`
}

type Notification struct {
	ConcurrentPushLimit         int                   `env:"NOTIFICATION_CONCURRENT_PUSH_LIMIT"`
	WaitWorkerPoolFinishOnClose config.SecondDuration `env:"NOTIFICATION_WAIT_WORKER_POOL_FINISH_ON_CLOSE"`
	SendTimeout                 config.SecondDuration `env:"NOTIFICATION_SEND_TIMEOUT"`
}

type MasterData struct {
	SQLiteDBFilePath string `env:"MASTER_DATA_SQLITE_DB_FILE_PATH"`
}

type Email struct {
	SMTPHost     string `env:"EMAIL_SMTP_HOST"`
	SMTPPort     int    `env:"EMAIL_SMTP_PORT"`
	SMTPUsername string `env:"EMAIL_SMTP_USERNAME"`
	SMTPPassword string `env:"EMAIL_SMTP_PASSWORD"`
	Sender       string `env:"EMAIL_SENDER"`
	TemplateDir  string `env:"EMAIL_TEMPLATE_DIR"`
}

type SQLite struct {
	DBFile string `env:"SQLITE_DB_FILE" default:"toko_mas_jawa.sqlite"`
}

type Config struct {
	HTTPPort     string `env:"HTTP_PORT"`
	LogLevel     string `env:"LOG_LEVEL" default:"DEBUG"`
	Database     Database
	Auth         Auth
	Firebase     Firebase
	Notification Notification
	MasterData   MasterData
	Email        Email
	SQLite       SQLite
}

func Load() (Config, error) {
	var cfg Config
	err := config.LoadEnv(&cfg)

	return cfg, err
}
