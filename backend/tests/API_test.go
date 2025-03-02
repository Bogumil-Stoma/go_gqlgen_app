package tests

import (
	"backend/database"
	"backend/graph"
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
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

func setupTestMutation(t *testing.T) (*gorm.DB, graph.MutationResolver) {
	db := setupTestDB() // Assuming setupTestDB is a function that sets up your DB connection.
	r := (&graph.Resolver{DB: db}).Mutation()

	t.Cleanup(func() {
		// Cleanup function that runs at the end of each test.
		err := clearTestTables(db) // Your cleanup function.
		if err != nil {
			t.Fatalf("Failed to clear test tables: %v", err)
		}
	})

	return db, r
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
	db, r := setupTestMutation(t)

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
}

func TestAddTranslation_DuplicateEntry(t *testing.T) {
	_, r := setupTestMutation(t)

	englishWord := "hello"
	polishWord := "cześć"

	_, err := r.AddTranslation(context.Background(), englishWord, polishWord)
	assert.NoError(t, err)

	_, err = r.AddTranslation(context.Background(), englishWord, polishWord)
	assert.Error(t, err, "Expected an error for duplicate translation")
	assert.Contains(t, err.Error(), "translation already exists")
}

func TestAddTranslation_Concurrent(t *testing.T) {
	db, r := setupTestMutation(t)

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
	assert.Equal(t, int64(1), count, "Translation should be stored in the database")
}

func TestAddTranslation_ConcurrencyWithDuplicate(t *testing.T) {
	_, r := setupTestMutation(t)

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
}

func TestAddWord_NewWord(t *testing.T) {
	_, r := setupTestMutation(t)

	word := "hello"
	language := "EN"
	exampleUsage := "A common greeting."

	addedWord, err := r.AddWord(context.Background(), word, language, exampleUsage)

	require.NoError(t, err, "Expected no error while adding word")
	assert.NotNil(t, addedWord, "Added word should not be nil")
	assert.Equal(t, word, addedWord.Word, "The word should be correctly added")
	assert.Equal(t, language, addedWord.Language, "The language should be correctly added")
	assert.Equal(t, exampleUsage, addedWord.ExampleUsage, "The example usage should be correctly added")
}

func TestAddWord_ExistingWord(t *testing.T) {
	_, r := setupTestMutation(t)

	word := "hello"
	language := "EN"
	exampleUsage := "A common greeting."

	_, err := r.AddWord(context.Background(), word, language, exampleUsage)
	require.NoError(t, err, "Expected no error while adding word for the first time")

	addedWord, err := r.AddWord(context.Background(), word, language, exampleUsage)

	require.NoError(t, err, "Expected no error while adding existing word")
	assert.NotNil(t, addedWord, "Added word should not be nil")
	assert.Equal(t, word, addedWord.Word, "The word should be the same")
	assert.Equal(t, language, addedWord.Language, "The language should be the same")
	assert.Equal(t, exampleUsage, addedWord.ExampleUsage, "The example usage should be the same")
}

func TestAddWord_ErrorHandling(t *testing.T) {
	_, r := setupTestMutation(t)

	word := ""
	language := "EN"
	exampleUsage := "A common greeting."

	addedWord, err := r.AddWord(context.Background(), word, language, exampleUsage)

	require.Error(t, err, "Expected an error due to invalid word input")
	assert.Nil(t, addedWord, "Added word should be nil due to the error")
}

func TestAddWord_ConcurrentSameWord(t *testing.T) {
	db, r := setupTestMutation(t)

	word := "hello"
	language := "EN"
	exampleUsage := "A common greeting."

	_, _ = r.AddWord(context.Background(), word, language, exampleUsage)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_, _ = r.AddWord(context.Background(), word, language, exampleUsage)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
	var count int64
	db.Find(&model.Word{}).Count(&count)
	assert.Equal(t, int64(1), count, "Only one word inserted")
}

func TestAddWord_ConcurrentDifferentWords(t *testing.T) {
	db, r := setupTestMutation(t)

	word := "hello"
	language := "EN"
	exampleUsage := "A common greeting."

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			changedWord := word + string(rune(i+64)) //adding some salt
			_, _ = r.AddWord(context.Background(), changedWord, language, exampleUsage)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
	var count int64
	db.Find(&model.Word{}).Count(&count)
	assert.Equal(t, int64(10), count, "Only one word inserted")
}

func TestDeleteWord_ExistingWord(t *testing.T) {
	db, r := setupTestMutation(t)

	word := "hello"
	language := "EN"
	exampleUsage := "A common greeting."

	_, err := r.AddWord(context.Background(), word, language, exampleUsage)
	require.NoError(t, err, "Expected no error while adding word initially")

	deletedWord, err := r.DeleteWord(context.Background(), word, language)

	require.NoError(t, err, "Expected no error while deleting word")
	assert.NotNil(t, deletedWord, "Deleted word should not be nil")
	assert.Equal(t, word, deletedWord.Word, "The deleted word should match")
	assert.Equal(t, language, deletedWord.Language, "The deleted word's language should match")

	var count int64
	err = db.Model(&model.Word{}).Where("word = ? AND language = ?", word, language).Count(&count).Error
	require.NoError(t, err, "Expected no error when counting words in database")
	assert.Equal(t, int64(0), count, "Word should be deleted from the database")
}

func TestDeleteWord_NonExistingWord(t *testing.T) {
	_, r := setupTestMutation(t)

	word := "nonexistent"
	language := "EN"

	deletedWord, err := r.DeleteWord(context.Background(), word, language)

	require.Error(t, err, "Expected error for non-existing word")
	assert.Nil(t, deletedWord, "Deleted word should be nil for non-existing word")
	assert.Equal(t, "word is missing in database", err.Error(), "Error message should match")
}

func TestDeleteWord_TranslationsAlsoDeleted(t *testing.T) {
	db, r := setupTestMutation(t)

	englishWord := "hello"
	polishWord := "cześć"

	_, err := r.AddTranslation(context.Background(), englishWord, polishWord)

	_, err = r.DeleteWord(context.Background(), englishWord, "EN")

	var count int64
	err = db.Model(&model.Word{}).Count(&count).Error
	require.NoError(t, err, "Expected no error when counting words in database")
	assert.Equal(t, int64(1), count, "Only one word should be deleted from the database")

	err = db.Model(&model.Translation{}).Count(&count).Error
	require.NoError(t, err, "Expected no error when counting translations in database")
	assert.Equal(t, int64(0), count, "Translation should be deleted from the database")
}
