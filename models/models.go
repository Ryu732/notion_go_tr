package models

import "time"

// 習慣記録の構造体
type HabitRecord struct {
	ID       string    `json:"id,omitempty"`
	Date     time.Time `json:"date"`
	Category string    `json:"category"`
}

// NotionAPIの設定
const (
	NotionAPIURL  = "https://api.notion.com/v1"
	NotionVersion = "2022-06-28"
)

// Notionページ作成のリクエスト構造体
type CreatePageRequest struct {
	Parent     Parent         `json:"parent"`
	Properties PageProperties `json:"properties"`
}

type Parent struct {
	DatabaseID string `json:"database_id"`
}

// 実際のNotionデータベースプロパティに合わせて修正
type PageProperties struct {
	// Dateプロパティ - 日付型
	Date DateProperty `json:"Date"`

	// Categoryプロパティ - リッチテキスト型に修正
	Category RichTextProperty `json:"Category"`
}

type DateProperty struct {
	Date DateObject `json:"date"`
}

type DateObject struct {
	Start string `json:"start"`
}

type RichTextProperty struct {
	RichText []RichTextObject `json:"rich_text"`
}

type RichTextObject struct {
	Type string     `json:"type"`
	Text TextObject `json:"text"`
}

type TextObject struct {
	Content string `json:"content"`
}
