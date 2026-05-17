package database

import (
	"fmt"
	"log"
	"time"

	"github.com/haiwo-ci/haiwo/internal/conf"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func OpenPostgres(cfg conf.PostgresConfiguration) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		cfg.Host,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.Port,
		cfg.SSLMode,
		cfg.TimeZone,
	)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger(cfg.LogMode)})
}

func logger(mode string) gormlogger.Interface {
	base := gormlogger.New(log.New(log.Writer(), "\r\n", log.LstdFlags), gormlogger.Config{
		SlowThreshold:             time.Second,
		IgnoreRecordNotFoundError: true,
		Colorful:                  true,
	})
	switch mode {
	case "console":
		return base.LogMode(gormlogger.Info)
	case "slow_query":
		return base.LogMode(gormlogger.Warn)
	default:
		return base.LogMode(gormlogger.Silent)
	}
}
