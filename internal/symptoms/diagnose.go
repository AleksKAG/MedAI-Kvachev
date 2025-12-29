package symptoms

func Correlate(symptoms []string, labResults map[string]float64) string {
	// Простой rule-based движок (в будущем — ML)
	if contains(symptoms, "fever") && labResults["leukocytes"] > 9.0 {
		return "⚠️ Вероятна бактериальная инфекция"
	}
	if labResults["hemoglobin"] < 120 && contains(symptoms, "weakness") {
		return "⚠️ Вероятна железодефицитная анемия"
	}
	return "ℹ️ Нет явных корреляций"
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
