# MedAI Kvachev
MedAI Kvachev — AI-система на Go для анализа медицинской истории пациента. Создает "медицинскую биографию" на основе анализов. Pet-project с фокусом на приватность данных. Основной язык: Go.

 Возможности
- Загрузка любых анализов: Поддержка PDF, изображения, текст (анализы крови, УЗИ и т.д.).
- Построение динамики: Визуализация изменений показателей во времени (графики).
- Нормальные значения: Сравнение с нормами по возрасту/полу.
- Выявление отклонений: Автоматическое выделение аномалий.
- Рекомендации врачу: Предложения на основе правил и простых ML.
- Медицинская биография: Генерация отчета с историей здоровья.

 Архитектура
- Сервис-ориентированная: Модули для парсинга, анализа, рекомендаций.
- Backend: Go с Fiber для API, Ent для ORM (MySQL).
- AI-компоненты: Gonum для статистики, интеграция с NLP (via Hugging Face API) для парсинга текста.
- Хранение: MySQL для данных, encrypted storage для приватности (crypto package).
- Обработка: Goroutines для параллельного анализа.
- Развертывание: Docker, с GDPR-compliant.

 Инструменты и технологии
- Язык: Go 1.21+
- Frontend: Telegram-бот + Web App
- Фреймворки: Fiber (HTTP), Ent (ORM), Gonum (stats).
- Базы данных: MySQL.
- AI/ML: Gonum, внешние API для NLP (gRPC).
- Парсинг: pdfcpu для PDF, tesseract-go для OCR.
- CI/CD: CircleCI.
- Тестирование: Go test, fuzzing для парсинга.
- Другие: Docker, crypto/std для шифрования.

 Установка
1. Клонируйте: `git clone https://github.com/yourusername/healthsummary-ai.git`
2. `go mod tidy`
3. .env с DB creds.
4. `go run main.go`
5. Docker: `docker-compose up`

 Использование
- Загрузка: POST `/api/upload/analysis`
- Анализ: GET `/api/summary/{patient_id}`
- Демо: Примеры анализов в repo.

## Автор
Квачёв Александр — Go-разработчик  
GitHub: [AleksKAG](https://github.com/AleksKAG)  
Telegram: [@Kurtalex27](https://t.me/Kurtalex27)

