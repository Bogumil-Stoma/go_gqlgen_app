package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Word struct {
	ID           uint    `gorm:"primaryKey;autoIncrement"`
	Word         string  `gorm:"unique"`
	Translations []*Word `gorm:"many2many:translations;joinForeignKey:word_id;joinReferences:translation_id"`
	Language     string
}

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

	err = DB.AutoMigrate(&Word{})
	if err != nil {
		log.Fatal(err)
	}

	return DB
}
