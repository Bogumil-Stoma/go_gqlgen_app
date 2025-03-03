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
	var translatedWordIDS []uint

	tx := DB.Begin()
	defer func() {
		tx.Rollback()

	}()

	err := tx.Where("word = ? and language = ?", wordToTranslate, language).First(&word).Error
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
