package tests

import (
	"backend/database"
	"backend/graph"
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"log"
	"path/filepath"
	"strconv"
	"testing"

	"backend/graph/model"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func init() {
	envPath, err := filepath.Abs("../../.env.test")
	if err != nil {
		log.Fatalf("Error resolving .env.test path: %v", err)
	}
	fmt.Println(envPath)
	if err := godotenv.Load(envPath); err != nil {
		log.Fatal("Error loading .env file")
	}
}

func setupTestMutation(t *testing.T) (*gorm.DB, graph.MutationResolver) {
	db := setupTestDB()
	r := (&graph.Resolver{DB: db}).Mutation()

	t.Cleanup(func() {
		err := clearTestTables(db)
		if err != nil {
			t.Fatalf("Failed to clear test tables: %v", err)
		}
	})

	return db, r
}

func setupTestQuery(t *testing.T) (*gorm.DB, graph.QueryResolver) {
	db := setupTestDB()
	r := (&graph.Resolver{DB: db}).Query()

	t.Cleanup(func() {
		err := clearTestTables(db)
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
	if err := db.Exec("TRUNCATE TABLE words RESTART IDENTITY CASCADE;").Error; err != nil {
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

	translation, err := r.AddTranslation(context.Background(), polishWord, "PL", englishWord, "EN")
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

	_, err := r.AddTranslation(context.Background(), polishWord, "PL", englishWord, "EN")
	assert.NoError(t, err)

	_, err = r.AddTranslation(context.Background(), polishWord, "PL", englishWord, "EN")
	assert.NoError(t, err, "Not expecting an error for duplicate translation")
}

func TestAddTranslation_Concurrent(t *testing.T) {
	db, r := setupTestMutation(t)

	englishWord := "hello"
	polishWord := "cześć"

	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			_, _ = r.AddTranslation(context.Background(), polishWord, "PL", englishWord, "EN")
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
	db, r := setupTestMutation(t)

	englishWord := "hello"
	polishWord := "cześć"

	_, err := r.AddTranslation(context.Background(), polishWord, "PL", englishWord, "EN")
	assert.NoError(t, err, "Expected no error on first translation")

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := r.AddTranslation(context.Background(), polishWord, "PL", englishWord, "EN")
			assert.NoError(t, err, "Not expecting an error for duplicate translation")
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	var count int64
	db.Find(&model.Translation{}).Count(&count)
	assert.Equal(t, int64(1), count, "One ranslation should be stored in the database")
}

func TestAddWord_NewWord(t *testing.T) {
	_, r := setupTestMutation(t)

	word := "hello"
	language := "EN"
	exampleUsage := "A common greeting."

	addedWord, err := r.AddWord(context.Background(), word, language, exampleUsage)

	require.NoError(t, err, "Expected no error while adding word")
	assert.NotNil(t, addedWord, "Added word should not be nil")
	assert.Equal(t, word, addedWord.Text, "The word should be correctly added")
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
	assert.Equal(t, word, addedWord.Text, "The word should be the same")
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
	assert.Equal(t, word, deletedWord.Text, "The deleted word should match")
	assert.Equal(t, language, deletedWord.Language, "The deleted word's language should match")

	var count int64
	err = db.Model(&model.Word{}).Where("text = ? AND language = ?", word, language).Count(&count).Error
	require.NoError(t, err, "Expected no error when counting words in database")
	assert.Equal(t, int64(0), count, "Text should be deleted from the database")
}

func TestDeleteWord_NonExistingWord(t *testing.T) {
	_, r := setupTestMutation(t)

	word := "nonexistent"
	language := "EN"

	deletedWord, err := r.DeleteWord(context.Background(), word, language)

	require.NoError(t, err, "Not expecting error for non-existing word")
	assert.Equal(t, 0, deletedWord.ID, "Deleted word should be empty Word for non-existing word")
}

func TestDeleteWord_TranslationsAlsoDeleted(t *testing.T) {
	db, r := setupTestMutation(t)

	englishWord := "hello"
	polishWord := "cześć"

	_, err := r.AddTranslation(context.Background(), polishWord, "PL", englishWord, "EN")

	_, err = r.DeleteWord(context.Background(), englishWord, "EN")

	var count int64
	err = db.Model(&model.Word{}).Count(&count).Error
	require.NoError(t, err, "Expected no error when counting words in database")
	assert.Equal(t, int64(1), count, "Only one word should be deleted from the database")

	err = db.Model(&model.Translation{}).Count(&count).Error
	require.NoError(t, err, "Expected no error when counting translations in database")
	assert.Equal(t, int64(0), count, "Translation should be deleted from the database")
}

func TestUpdateWord_Success(t *testing.T) {
	_, r := setupTestMutation(t)

	sourceWord := "hello"
	sourceLanguage := "EN"
	updatedWord := "hi"
	updatedExampleUsage := "updated usage"

	_, err := r.AddWord(context.Background(), sourceWord, sourceLanguage, "old usage")
	assert.NoError(t, err)

	word, err := r.UpdateWord(context.Background(), sourceWord, sourceLanguage, updatedWord, updatedExampleUsage)
	assert.NoError(t, err)
	assert.Equal(t, updatedWord, word.Text)
	assert.Equal(t, updatedExampleUsage, word.ExampleUsage)
}

func TestUpdateWord_WordNotFound(t *testing.T) {
	_, r := setupTestMutation(t)

	sourceWord := "nonexistent"
	sourceLanguage := "EN"
	updatedWord := "hi"
	updatedExampleUsage := "updated usage"

	word, err := r.UpdateWord(context.Background(), sourceWord, sourceLanguage, updatedWord, updatedExampleUsage)
	assert.Error(t, err, "Raising error for updating non existing word")
	assert.Nil(t, word)
}

func TestUpdateWord_EmptyWord(t *testing.T) {
	_, r := setupTestMutation(t)

	sourceWord := ""
	sourceLanguage := "EN"
	updatedWord := "hi"
	updatedExampleUsage := "updated usage"

	word, err := r.UpdateWord(context.Background(), sourceWord, sourceLanguage, updatedWord, updatedExampleUsage)
	assert.Error(t, err, "Raising error for updating empty word")
	assert.Nil(t, word)
}

func TestUpdateWord_EmptyLanguage(t *testing.T) {
	_, r := setupTestMutation(t)

	sourceWord := "hello"
	sourceLanguage := ""
	updatedWord := "hi"
	updatedExampleUsage := "updated usage"

	word, err := r.UpdateWord(context.Background(), sourceWord, sourceLanguage, updatedWord, updatedExampleUsage)
	assert.Error(t, err, "Raising error for updating empty word")
	assert.Nil(t, word)
}

func TestTranslations_NoWord(t *testing.T) {
	_, r := setupTestQuery(t)

	words, err := r.GetTranslations(context.Background(), "nonexistent", "PL")
	assert.Error(t, err, "Error for translation with no existing word")
	assert.Nil(t, words)
}

func TestTranslations_TwoTranslations(t *testing.T) {
	_, rq := setupTestQuery(t)
	_, rm := setupTestMutation(t)

	_, _ = rm.AddTranslation(context.Background(), "biegać", "PL", "run", "EN")
	_, _ = rm.AddTranslation(context.Background(), "truchtać", "PL", "run", "EN")

	words, err := rq.GetTranslations(context.Background(), "run", "EN")
	assert.Equal(t, 2, len(words))
	assert.Nil(t, err)

}

func TestTranslations_NoTranslation(t *testing.T) {
	_, r := setupTestQuery(t)
	_, rm := setupTestMutation(t)

	_, _ = rm.AddWord(context.Background(), "run", "EN", "")
	words, err := r.GetTranslations(context.Background(), "run", "EN")
	assert.NoError(t, err, "No error for no translation")
	assert.Equal(t, 0, len(words))
}

func TestDeleteTranslationPLtoEN_NormalCase(t *testing.T) {
	DB, r := setupTestMutation(t)
	englishWord := "hello"
	polishWord := "cześć"

	_, _ = r.AddTranslation(context.Background(), polishWord, "PL", englishWord, "EN")
	translation, err := r.AddTranslation(context.Background(), polishWord, "PL", englishWord, "EN")
	translation, err = r.DeleteTranslation(context.Background(), polishWord, "PL", englishWord, "EN")
	assert.NoError(t, err)
	assert.NotNil(t, translation)
	var count int64
	DB.Find(&model.Translation{}).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestDeleteTranslationPLtoEN_NoWords(t *testing.T) {
	_, r := setupTestMutation(t)
	englishWord := "hello"
	polishWord := "cześć"

	translation, err := r.DeleteTranslation(context.Background(), polishWord, "PL", englishWord, "EN")
	assert.NoError(t, err, "Graceful deletion")
	assert.Equal(t, 0, translation.TranslationID)
	assert.Equal(t, 0, translation.WordID)
}

func TestDeleteTranslationPLtoEN_SecondWordNotFound(t *testing.T) {
	_, r := setupTestMutation(t)
	englishWord := "hello"
	polishWord := "cześć"

	_, _ = r.AddWord(context.Background(), polishWord, "PL", "")
	translation, err := r.DeleteTranslation(context.Background(), polishWord, "PL", englishWord, "EN")
	assert.NoError(t, err, "Graceful deletion")
	assert.Equal(t, 0, translation.TranslationID)
	assert.Equal(t, 0, translation.WordID)
}

func TestDeleteTranslationPLtoEN_NoTranslation(t *testing.T) {
	_, r := setupTestMutation(t)
	englishWord := "hello"
	polishWord := "cześć"

	_, _ = r.AddWord(context.Background(), polishWord, "PL", "")
	_, _ = r.AddWord(context.Background(), englishWord, "EN", "")
	translation, err := r.DeleteTranslation(context.Background(), polishWord, "PL", englishWord, "EN")
	assert.NoError(t, err, "Graceful deletion")
	assert.Equal(t, 0, translation.TranslationID)
	assert.Equal(t, 0, translation.WordID)
}

func TestDeleteWord_Concurrent(t *testing.T) {
	db, r := setupTestMutation(t)

	word := "hello"
	language := "EN"
	exampleUsage := "A common greeting."
	_, _ = r.AddWord(context.Background(), word, language, exampleUsage)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := r.DeleteWord(context.Background(), word, language)
			assert.NoError(t, err, "No error expected when deleting word existing or non existing")
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
	var count int64
	db.Find(&model.Word{}).Count(&count)
	assert.Equal(t, int64(0), count, "No words in db")
}

func TestDeleteTranslation_Concurrent(t *testing.T) {
	db, r := setupTestMutation(t)

	englishWord := "hello"
	polishWord := "cześć"

	_, _ = r.AddTranslation(context.Background(), polishWord, "PL", englishWord, "EN")

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := r.DeleteTranslation(context.Background(), polishWord, "PL", englishWord, "EN")
			assert.NoError(t, err, "Graceful deletion")
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
	var count int64
	db.Find(&model.Translation{}).Count(&count)
	assert.Equal(t, int64(0), count, "No words in db")
}

func TestUpdateWord_Concurrent(t *testing.T) {
	db, r := setupTestMutation(t)

	word := "hello"
	language := "EN"
	exampleUsage := "A common greeting."
	_, _ = r.AddWord(context.Background(), word, language, exampleUsage)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			changedExample := "updated " + strconv.Itoa(i)
			_, err := r.UpdateWord(context.Background(), word, language, word, changedExample)
			assert.NoError(t, err)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
	var count int64
	var updatedWord model.Word
	db.Find(&updatedWord).Count(&count)
	assert.Equal(t, int64(1), count, "One word in db")
	assert.Contains(t, updatedWord.ExampleUsage, "updated ", "Word updated")
}
