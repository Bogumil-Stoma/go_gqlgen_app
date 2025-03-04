package model

type Translation struct {
	WordID        int `json:"wordID" gorm:"column:word_id;primaryKey"`
	TranslationID int `json:"translationID" gorm:"column:translation_id;primaryKey"`
}

func (translation *Translation) SortTranslation() {
	if translation.WordID > translation.TranslationID {
		translation.WordID, translation.TranslationID = translation.TranslationID, translation.WordID
	}
}
