package lab

var Norms = map[string]struct{ Min, Max float64 }{
	"hemoglobin": {120, 160},
	"glucose":    {3.3, 5.5},
	"leukocytes": {4.0, 9.0},
}

func Interpret(testName string, value float64) string {
	if norm, ok := Norms[testName]; ok {
		if value < norm.Min {
			return "🔴 Понижен: возможна анемия, недоедание"
		} else if value > norm.Max {
			return "🔺 Повышен: возможна инфекция, обезвоживание"
		} else {
			return "✅ В норме"
		}
	}
	return "❓ Неизвестный показатель"
}
