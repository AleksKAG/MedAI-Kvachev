package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"medai-assistant/internal/forecast"
	"medai-assistant/internal/lab"
	"medai-assistant/internal/storage"
	"medai-assistant/internal/symptoms"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB
var telegramBot *bot.Bot

func initDB() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=medai port=5432 sslmode=disable"
	}
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect DB")
	}
	storage.Migrate(db)
}

func initBot() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN required")
	}

	webhookURL := os.Getenv("WEBHOOK_URL")
	opts := []bot.Option{
		bot.WithDefaultHandlerFunc(messageHandler),
	}
	if webhookURL != "" {
		opts = append(opts, bot.WithWebhook(bot.Webhook{
			ExternalURL: webhookURL,
			ListenAddr:  "0.0.0.0:8080",
		}))
	}

	var err error
	telegramBot, err = bot.New(token, opts...)
	if err != nil {
		log.Fatal(err)
	}
}

func messageHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	text := strings.TrimSpace(update.Message.Text)

	parts := strings.Fields(text)
	if len(parts) == 0 {
		return
	}

	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	case "/start":
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "🩺 Добро пожаловать в MedAI Assistant!\n\n" +
				"Команды:\n" +
				"/analyze hemoglobin 110 — интерпретация\n" +
				"/trend hemoglobin — динамика\n" +
				"/diagnose fever,weakness — симптомы\n" +
				"/forecast hemoglobin — прогноз",
		})

	case "/analyze":
		if len(args) < 2 {
			b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "Используйте: /analyze <показатель> <значение>"})
			return
		}
		test := args[0]
		value, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "Неверное число"})
			return
		}

		// Сохраняем в БД
		db.Create(&storage.LabResult{
			PatientID: strconv.FormatInt(chatID, 10),
			TestName:  test,
			Value:     value,
			Unit:      "default",
			Date:      time.Now().Unix(),
		})

		interpretation := lab.Interpret(test, value)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "🧪 " + test + ": " + strconv.FormatFloat(value, 'f', 1, 64) + "\n" + interpretation,
		})

	case "/trend":
		if len(args) < 1 {
			b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "Используйте: /trend <показатель>"})
			return
		}
		test := args[0]
		var results []storage.LabResult
		db.Where("patient_id = ? AND test_name = ?", strconv.FormatInt(chatID, 10), test).Order("date").Find(&results)

		if len(results) == 0 {
			b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "Нет данных по " + test})
			return
		}

		values := make([]float64, len(results))
		for i, r := range results {
			values[i] = r.Value
		}

		forecastVal := forecast.LinearTrend(values)
		last := results[len(results)-1].Value

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "📈 Динамика " + test + ":\n" +
				"Последнее: " + strconv.FormatFloat(last, 'f', 1, 64) + "\n" +
				"Прогноз: " + strconv.FormatFloat(forecastVal, 'f', 1, 64) + "\n" +
				"(линейная регрессия на Gonum)",
		})

	case "/diagnose":
		if len(args) < 1 {
			b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "Используйте: /diagnose симптом1,симптом2"})
			return
		}
		symptomList := strings.Split(args[0], ",")
		var results []storage.LabResult
		db.Where("patient_id = ?", strconv.FormatInt(chatID, 10)).Find(&results)

		labMap := make(map[string]float64)
		for _, r := range results {
			labMap[r.TestName] = r.Value
		}

		diagnosis := symptoms.Correlate(symptomList, labMap)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "🔍 Анализ симптомов:\n" + diagnosis,
		})

	default:
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Неизвестная команда. Используйте /start",
		})
	}
}

func main() {
	initDB()
	initBot()

	// Chi router для health-check и будущего Web App API
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Запуск Telegram webhook или polling
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal(err)
		}
	}()

	// Запуск Chi на другом порту (или интеграция через mux)
	go func() {
		log.Println("HTTP сервер запущен на :8081")
		http.ListenAndServe(":8081", r)
	}()

	// Для локального запуска — long polling
	if os.Getenv("WEBHOOK_URL") == "" {
		ctx := context.Background()
		telegramBot.Start(ctx)
	} else {
		select {} // keep alive
	}
}
