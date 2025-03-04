package model

type Word struct {
	ID           int     `json:"id" gorm:"primaryKey;autoIncrement"`
	Text         string  `json:"text" gorm:"not null;uniqueIndex:idx_text_language"`
	Translations []*Word `gorm:"many2many:translations;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Language     string  `json:"language" gorm:"not null;uniqueIndex:idx_text_language"`
	ExampleUsage string  `json:"example_usage"`
}
