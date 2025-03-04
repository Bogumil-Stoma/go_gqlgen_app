package utils

import (
	"backend/graph/model"
	"fmt"
	"gorm.io/gorm"
)

func GetTranslations(wordToTranslate string, language string, DB *gorm.DB) ([]*model.Word, error) {
	var word model.Word
	var translatedWords []*model.Word
	var translations []*model.Translation
	var translatedWordIDS []int

	tx := DB.Begin()
	defer func() {
		tx.Rollback()

	}()

	err := tx.Where("text = ? and language = ?", wordToTranslate, language).First(&word).Error
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
		return nil, fmt.Errorf("an error has occured while selecting translations")
	}

	tx.Commit()
	return translatedWords, nil
}

func DeleteTranslation(DB *gorm.DB, sourceText string, sourceTextLanguage string, translatedText string, translatedTextLanguage string) (*model.Translation, error) {
	var sourceWord, translatedWord model.Word

	tx := DB.Begin()
	defer func() {
		tx.Rollback()

	}()

	err := DB.First(&sourceWord, model.Word{Text: sourceText, Language: sourceTextLanguage}).Error
	if err != nil {
		return nil, fmt.Errorf("source word not found in database")
	}

	err = DB.First(&translatedWord, model.Word{Text: translatedText, Language: translatedTextLanguage}).Error
	if err != nil {
		return nil, fmt.Errorf("translated word not found in database")
	}

	sortedTranslation := model.Translation{WordID: sourceWord.ID, TranslationID: translatedWord.ID}
	sortedTranslation.SortTranslation()

	err = DB.Where("word_id = ? AND translation_id = ?", sortedTranslation.WordID, sortedTranslation.TranslationID).Delete(&sortedTranslation).Error
	if err != nil {
		return nil, fmt.Errorf("translated word not found in database")
	}

	tx.Commit()
	return &sortedTranslation, nil
}
