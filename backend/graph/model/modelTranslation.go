package model

type Translation struct {
	WordID        uint `json:"wordID"`
	TranslationID uint `json:"translationID"`
}

func (translation *Translation) SortTranslation() {
	if translation.WordID > translation.TranslationID {
		translation.WordID, translation.TranslationID = translation.TranslationID, translation.WordID
	}
}
