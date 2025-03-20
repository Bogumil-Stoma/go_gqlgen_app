package database

import (
	"backend/graph/model"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"time"
)

var DB *gorm.DB

func Connect() *gorm.DB {
	dbString := "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC"
	host := os.Getenv("POSTGRES_HOST")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_NAME")
	dbPort := os.Getenv("POSTGRES_PORT")

	dsn := fmt.Sprintf(dbString, host, user, password, dbName, dbPort)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to database")

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("failed to get SQL DB", err)
	}

	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(2 * time.Hour)

	err = DB.AutoMigrate(&model.Word{})
	if err != nil {
		log.Fatal(err)
	}

	return DB
}
