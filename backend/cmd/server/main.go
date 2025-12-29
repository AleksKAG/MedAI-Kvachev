package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"medai-backend/internal/parser"
	"medai-backend/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB
var telegramBot *bot.Bot
var ocrServiceURL = "http://ocr-service:5000/ocr"

func initDB() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=db user=postgres password=postgres dbname=medai port=5432 sslmode=disable"
	}
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("DB error")
	}
	storage.Migrate(db)
}

func callOCRService(fileData io.Reader, filename string) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filename)
	io.Copy(part, fileData)
	writer.Close()

	resp, err := http.Post(ocrServiceURL, writer.FormDataContentType(), body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct{ Text string }
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Text, nil
}

func telegramHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID

	// Обработка Web App данных
	if update.Message.WebAppData != nil {
		var payload struct {
			Action   string        `json:"action"`
			Results  []parser.LabResult `json:"results"`
		}
		json.Unmarshal([]byte(update.Message.WebAppData.Data), &payload)

		patientID := strconv.FormatInt(chatID, 10)
		for _, r := range payload.Results {
			db.Create(&storage.LabResult{
				PatientID: patientID,
				TestName:  r.Name,
				Value:     r.Value,
				Unit:      r.Unit,
				Date:      time.Now().Unix(),
			})
		}

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "✅ Результаты сохранены! Используйте /history для просмотра.",
		})
		return
	}

	// Кнопка открытия Web App
	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{
				Text: "🖥️ Открыть MedAI",
				WebApp: &models.WebAppInfo{
					URL: os.Getenv("WEBAPP_URL"),
				},
			}},
		},
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        "🩺 Откройте интерфейс для загрузки анализов:",
		ReplyMarkup: keyboard,
	})
}

func main() {
	initDB()

	// Telegram bot
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN required")
	}

	var err error
	telegramBot, err = bot.New(token, bot.WithDefaultHandlerFunc(telegramHandler))
	if err != nil {
		log.Fatal(err)
	}

	// Web API для Web App
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*.telegram.org", "http://localhost:*"},
		AllowedMethods: []string{"GET", "POST"},
	}))

	r.Post("/api/ocr-parse", func(w http.ResponseWriter, r *http.Request) {
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "no file", 400)
			return
		}
		defer file.Close()

		// Вызываем OCR
		text, err := callOCRService(file, "upload")
		if err != nil {
			http.Error(w, "ocr failed", 500)
			return
		}

		// Парсим
		results := parser.ParseLabResults(text)

		// Добавляем интерпретацию
		var enriched []map[string]interface{}
		for _, r := range results {
			enriched = append(enriched, map[string]interface{}{
				"name":           r.Name,
				"value":          r.Value,
				"unit":           r.Unit,
				"interpretation": parser.Interpret(r.Name, r.Value),
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"results": enriched,
		})
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Запуск
	go func() {
		log.Println("Telegram bot starting...")
		telegramBot.Start(context.Background())
	}()

	log.Println("Web API server starting on :8081")
	http.ListenAndServe(":8081", r)
}
