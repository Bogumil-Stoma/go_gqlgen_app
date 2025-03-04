package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.66

import (
	"backend/graph/model"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AddTranslation is the resolver for the addTranslation field.
func (r *mutationResolver) AddTranslation(ctx context.Context, sourceText string, sourceTextLanguage string, translatedText string, translatedTextLanguage string) (*model.Translation, error) {
	var sourceWord model.Word
	var translatedWord model.Word
	var existingTranslation model.Translation

	tx := r.DB.Begin()
	defer func() {
		tx.Rollback()
	}()

	err := tx.FirstOrCreate(&sourceWord, model.Word{Text: sourceText, Language: sourceTextLanguage}).Error
	if err != nil {
		return nil, fmt.Errorf("an error has occured while inserting source word")
	}
	err = tx.FirstOrCreate(&translatedWord, model.Word{Text: translatedText, Language: translatedTextLanguage}).Error
	if err != nil {
		return nil, fmt.Errorf("an error has occured while inserting translated word")
	}

	sortedTranslation := model.Translation{WordID: translatedWord.ID, TranslationID: sourceWord.ID}
	sortedTranslation.SortTranslation()

	err = tx.Where("word_id = ? AND translation_id = ?", sortedTranslation.WordID, sortedTranslation.TranslationID).First(&existingTranslation).Error
	if err == nil {
		return nil, fmt.Errorf("translation already exists")
	}

	err = tx.Create(&sortedTranslation).Error
	if err != nil {
		return nil, fmt.Errorf("database error while inserting translation: %w", err)
	}
	fmt.Println(sortedTranslation)
	tx.Commit()

	return &sortedTranslation, nil
}

// AddWord is the resolver for the addWord field.
func (r *mutationResolver) AddWord(ctx context.Context, text string, language string, exampleUsage string) (*model.Word, error) {
	var addedWord model.Word
	tx := r.DB.Begin()
	defer func() {
		tx.Rollback()
	}()
	if text == "" || language == "" {
		return nil, fmt.Errorf("word and language must not be empty")
	}
	err := tx.FirstOrCreate(&addedWord, model.Word{Text: text, Language: language, ExampleUsage: exampleUsage}).Error
	if err != nil {
		return nil, fmt.Errorf("database error while inserting word: %w", err)
	}

	tx.Commit()
	return &addedWord, nil
}

// DeleteWord is the resolver for the deleteWord field.
func (r *mutationResolver) DeleteWord(ctx context.Context, text string, language string) (*model.Word, error) {
	var deletedWord model.Word
	tx := r.DB.Begin()
	defer func() {
		tx.Rollback()
	}()
	err := tx.Where("text = ? and language = ?", text, language).First(&deletedWord).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("database error while finding word: %w", err)
	}
	err = tx.Select(clause.Associations).Delete(&deletedWord).Error
	if err != nil {
		return nil, fmt.Errorf("database error while removing word: %w", err)
	}

	tx.Commit()
	return &deletedWord, nil
}

// UpdateWord is the resolver for the updateWord field.
func (r *mutationResolver) UpdateWord(ctx context.Context, sourceText string, sourceLanguage string, updatedText string, updatedExampleUsage string) (*model.Word, error) {
	if sourceText == "" || sourceLanguage == "" {
		return nil, fmt.Errorf("word and language must not be empty")
	}
	var word model.Word
	tx := r.DB.Begin()
	defer func() {
		tx.Rollback()
	}()

	err := tx.Where("text = ? and language = ?", sourceText, sourceLanguage).First(&word).Error
	if err != nil {
		return nil, fmt.Errorf("word is missing in database")
	}

	word.Text = updatedText
	word.ExampleUsage = updatedExampleUsage
	err = tx.Save(&word).Error
	if err != nil {
		return nil, fmt.Errorf("database error while updating word: %w", err)
	}

	tx.Commit()
	return &word, nil
}

// DeleteTranslation is the resolver for the deleteTranslation field.
func (r *mutationResolver) DeleteTranslation(ctx context.Context, sourceText string, sourceTextLanguage string, translatedText string, translatedTextLanguage string) (*model.Translation, error) {
	var sourceWord, translatedWord model.Word

	tx := r.DB.Begin()
	defer func() { tx.Rollback() }()

	err := tx.First(&sourceWord, model.Word{Text: sourceText, Language: sourceTextLanguage}).Error
	if err != nil {
		return nil, fmt.Errorf("source word not found in database")
	}

	err = tx.First(&translatedWord, model.Word{Text: translatedText, Language: translatedTextLanguage}).Error
	if err != nil {
		return nil, fmt.Errorf("translated word not found in database")
	}

	sortedTranslation := model.Translation{WordID: sourceWord.ID, TranslationID: translatedWord.ID}
	sortedTranslation.SortTranslation()

	result := tx.Where("word_id = ? AND translation_id = ?", sortedTranslation.WordID, sortedTranslation.TranslationID).Delete(&sortedTranslation)
	if result.Error != nil {
		return nil, fmt.Errorf("database error while deleting translation: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}

	tx.Commit()
	return &sortedTranslation, nil
}

// GetTranslations is the resolver for the getTranslations field.
func (r *queryResolver) GetTranslations(ctx context.Context, textToTranslate string, language string) ([]*model.Word, error) {
	var word model.Word
	var translatedWords []*model.Word
	var translations []*model.Translation
	var translatedWordIDS []int

	tx := r.DB.Begin()
	defer func() {
		tx.Rollback()
	}()

	err := tx.Where("text = ? and language = ?", textToTranslate, language).First(&word).Error
	if err != nil {
		return nil, fmt.Errorf("give word is not in database")
	}

	err = tx.Where("translation_id = ? or word_id = ?", word.ID, word.ID).Find(&translations).Error
	if err != nil || len(translations) == 0 {
		return nil, fmt.Errorf("no translations of given word were found")
	}

	for _, t := range translations {
		if word.ID == t.WordID {
			translatedWordIDS = append(translatedWordIDS, t.TranslationID)
		} else {
			translatedWordIDS = append(translatedWordIDS, t.WordID)
		}
	}

	err = tx.Where("id in (?)", translatedWordIDS).Find(&translatedWords).Error
	if err != nil {
		return nil, fmt.Errorf("database error while searching translation: %w", err)
	}

	tx.Commit()
	return translatedWords, nil
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
