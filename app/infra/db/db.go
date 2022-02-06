package db

import (
	"fmt"
	"github.com/HunterXuan/bt/app/domain/model"
	"github.com/HunterXuan/bt/app/infra/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

var DB *gorm.DB

func InitDB() {
	log.Println("DB Initializing...")

	// Generate connection string
	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8mb4&parseTime=True&loc=Local",
		config.Config.GetString("DB_USER"),
		config.Config.GetString("DB_PASS"),
		config.Config.GetString("DB_HOST"),
		config.Config.GetString("DB_PORT"),
		config.Config.GetString("DB_NAME"),
	)

	// Init DB
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Err: %v", err)
	}

	// Migrate DB
	if config.Config.GetString("APP_MODE") == "dev" {
		_ = DB.AutoMigrate(&model.Torrent{}, &model.Peer{})
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Err: %v", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	log.Println("DB Initialized!")
}
