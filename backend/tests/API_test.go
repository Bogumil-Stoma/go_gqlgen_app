package tests

import (
	"backend/database"
	"backend/graph"
	"context"
	"fmt"
	"log"
	"testing"

	"backend/graph/model"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
}

func setupTestDB() *gorm.DB {
	db := database.Connect()
	return db
}

func clearTestTables(db *gorm.DB) error {
	if err := db.Exec("TRUNCATE TABLE words CASCADE").Error; err != nil {
		return fmt.Errorf("error truncating words table: %v", err)
	}

	if err := db.Exec("TRUNCATE TABLE translations CASCADE").Error; err != nil {
		return fmt.Errorf("error truncating translations table: %v", err)
	}
	return nil
}

func TestAddTranslation(t *testing.T) {
	db := setupTestDB()
	r := (&graph.Resolver{DB: db}).Mutation()

	englishWord := "hello"
	polishWord := "cześć"

	translation, err := r.AddTranslation(context.Background(), englishWord, polishWord)
	assert.NoError(t, err, "Expected no error while adding translation")
	assert.NotNil(t, translation, "Translation should not be nil")
	assert.NotZero(t, translation.WordID, "Translation should have a valid WordID")
	assert.NotZero(t, translation.TranslationID, "Translation should have a valid TranslationID")

	var count int64
	db.Model(&model.Translation{}).Where("word_id = ? AND translation_id = ?", translation.WordID, translation.TranslationID).Count(&count)

	assert.Equal(t, int64(1), count, "Translation should be stored in the database")
	err = clearTestTables(db)
	if err != nil {
		return
	}
}

func TestAddTranslation_DuplicateEntry(t *testing.T) {
	db := setupTestDB()
	r := (&graph.Resolver{DB: db}).Mutation()

	englishWord := "hello"
	polishWord := "cześć"

	_, err := r.AddTranslation(context.Background(), englishWord, polishWord)
	assert.NoError(t, err)

	_, err = r.AddTranslation(context.Background(), englishWord, polishWord)
	assert.Error(t, err, "Expected an error for duplicate translation")
	assert.Contains(t, err.Error(), "translation already exists")
	err = clearTestTables(db)
	if err != nil {
		return
	}
}

func TestAddTranslation_Concurrent(t *testing.T) {
	db := setupTestDB()
	r := (&graph.Resolver{DB: db}).Mutation()

	englishWord := "hello"
	polishWord := "cześć"

	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			_, _ = r.AddTranslation(context.Background(), englishWord, polishWord)
			done <- true
		}()
	}
	var count int64
	for i := 0; i < 10; i++ {
		<-done
	}
	db.Find(&model.Translation{}).Count(&count)
	fmt.Println(count)
	assert.Equal(t, int64(1), count, "Translation should be stored in the database")

	err := clearTestTables(db)
	if err != nil {
		return
	}
}

func TestAddTranslation_ConcurrencyWithDuplicate(t *testing.T) {
	db := setupTestDB()
	r := (&graph.Resolver{DB: db}).Mutation()

	englishWord := "hello"
	polishWord := "cześć"

	_, err := r.AddTranslation(context.Background(), englishWord, polishWord)
	assert.NoError(t, err, "Expected no error on first translation")

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := r.AddTranslation(context.Background(), englishWord, polishWord)
			assert.Error(t, err, "Expected error for duplicate translation")
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
	err = clearTestTables(db)
	if err != nil {
		return
	}
}
