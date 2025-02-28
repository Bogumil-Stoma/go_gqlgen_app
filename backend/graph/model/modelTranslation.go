package model

type Translation struct {
	WordID        string `json:"wordID"`
	TranslationID string `json:"translationID"`
}

func (translation *Translation) SortTranslation() {
	if translation.WordID > translation.TranslationID {
		translation.WordID, translation.TranslationID = translation.TranslationID, translation.WordID
	}
}
