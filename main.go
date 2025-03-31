package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gin_test/models"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// .envファイルを読み込む
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	router := gin.Default()
	router.Use(cors.Default())

	// ルーティング設定
	habitRouter := router.Group("/record")
	habitRouter.POST("", AddRecordController)
	habitRouter.GET("", AddRecordController)

	router.Run("localhost:8080")
}

func AddRecordController(c *gin.Context) {
	// クエリパラメータを取得
	category := c.Query("category")

	// バリデーション
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "カテゴリは必須です"})
		return
	}

	// 新しい記録を作成
	newRecord := models.HabitRecord{
		Date:     time.Now(),
		Category: category,
	}

	// リポジトリ関数を呼び出して保存
	err := AddRecordToNotion(&newRecord)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 成功レスポンスを返す
	c.JSON(http.StatusOK, newRecord)
}

func AddRecordToNotion(record *models.HabitRecord) error {
	// 環境変数から設定を読み込み
	apiKey := os.Getenv("NOTION_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("NOTION_API_KEY環境変数が設定されていません")
	}

	databaseID := os.Getenv("NOTION_HABIT_DB_KEY")
	if databaseID == "" {
		return fmt.Errorf("NOTION_HABIT_DB_KEY環境変数が設定されていません")
	}

	// Notionページ作成リクエストを構築 - プロパティタイプを修正
	createPageReq := models.CreatePageRequest{
		Parent: models.Parent{
			DatabaseID: databaseID,
		},
		Properties: models.PageProperties{
			// 日付プロパティ
			Date: models.DateProperty{
				Date: models.DateObject{
					Start: record.Date.Format(time.RFC3339),
				},
			},
			// カテゴリプロパティ - リッチテキスト型に変更
			Category: models.RichTextProperty{
				RichText: []models.RichTextObject{
					{
						Type: "text",
						Text: models.TextObject{
							Content: record.Category,
						},
					},
				},
			},
		},
	}

	// JSONに変換
	reqBody, err := json.Marshal(createPageReq)
	if err != nil {
		return fmt.Errorf("JSON変換エラー: %v", err)
	}

	// デバッグ用：リクエストボディを表示
	log.Printf("Notionリクエスト: %s", string(reqBody))

	// HTTPリクエスト作成
	url := fmt.Sprintf("%s/pages", models.NotionAPIURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("リクエスト作成エラー: %v", err)
	}

	// ヘッダー設定
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", models.NotionVersion)

	// リクエスト送信
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("API呼び出しエラー: %v", err)
	}
	defer resp.Body.Close()

	// レスポンス確認
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("レスポンス読み取りエラー: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("notion APIエラー: %s, コード: %d", string(body), resp.StatusCode)
	}

	// レスポンスをパース
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("レスポンスのJSONデコードエラー: %v", err)
	}

	// 作成されたページIDを設定
	if pageID, ok := response["id"].(string); ok {
		record.ID = pageID
	}

	return nil
}
