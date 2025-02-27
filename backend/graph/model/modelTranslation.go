package model

import "strconv"

type Translation struct {
	WordID        string `json:"wordID"`
	TranslationID string `json:"translationID"`
}

func (translation *Translation) SortTranslation() {
	if translation.WordID > translation.TranslationID {
		translation.WordID, translation.TranslationID = translation.TranslationID, translation.WordID
	}
}

func TranslationFromInts(ID1, ID2 uint) Translation {
	return Translation{WordID: strconv.Itoa(int(ID1)), TranslationID: strconv.Itoa(int(ID2))}
}
