package model

type Word struct {
	ID           string  `json:"id" gorm:"primaryKey;autoIncrement"`
	Word         string  `json:"word" gorm:"not null;uniqueIndex:idx_word_language"`
	Translations []*Word `gorm:"many2many:translations;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Language     string  `json:"language" gorm:"not null;uniqueIndex:idx_word_language"`
	ExampleUsage string  `json:"example_usage"`
}
