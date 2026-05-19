package database

import (
	"titles-mcp/config"
	"titles-mcp/database/models"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"fmt"
	"log"
	"os"
	"time"
)

func NewDb() *gorm.DB {
	var db *gorm.DB
	var err error

	newLogger := logger.New(
		log.New(os.Stderr, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level
			IgnoreRecordNotFoundError: true,           // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      true,           // Don't include params in the SQL log
			Colorful:                  false,          // Disable color
		},
	)

	if os.Getenv("USE_SQLITE") == "true" {
		db, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
			Logger: newLogger,
		})
		if err == nil {
			sqlDB, _ := db.DB()
			sqlDB.SetMaxOpenConns(1)
		}
	} else {
		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			config.AppConfig.DBConfig.Host,
			config.AppConfig.DBConfig.User,
			config.AppConfig.DBConfig.Password,
			config.AppConfig.DBConfig.Name,
			config.AppConfig.DBConfig.Port,
		)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: newLogger,
		})
	}

	if err != nil {
		panic(err)
	}
	db.AutoMigrate(models.Title{})
	return db
}
