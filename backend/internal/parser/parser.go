package parser

import (
	"regexp"
	"strconv"
	"strings"
)

type LabResult struct {
	Name  string
	Value float64
	Unit  string
}

var Norms = map[string]struct{ Min, Max float64 }{
	"hemoglobin": {120, 160},
	"glucose":    {3.3, 5.5},
	"leukocytes": {4.0, 9.0},
}

func Interpret(testName string, value float64) string {
	if norm, ok := Norms[testName]; ok {
		if value < norm.Min {
			return "🔴 Понижен"
		} else if value > norm.Max {
			return "🔺 Повышен"
		} else {
			return "✅ В норме"
		}
	}
	return "❓"
}

func ParseLabResults(text string) []LabResult {
	var results []LabResult
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "гемоглобин") || strings.Contains(lower, "hemoglobin") {
			if v := extractNumber(lower); v > 0 {
				results = append(results, LabResult{"hemoglobin", v, "г/л"})
			}
		} else if strings.Contains(lower, "глюкоза") || strings.Contains(lower, "glucose") {
			if v := extractNumber(lower); v > 0 {
				results = append(results, LabResult{"glucose", v, "ммоль/л"})
			}
		} else if strings.Contains(lower, "лейкоциты") || strings.Contains(lower, "leukocytes") {
			if v := extractNumber(lower); v > 0 {
				results = append(results, LabResult{"leukocytes", v, "x10^9/л"})
			}
		}
	}
	return results
}

func extractNumber(s string) float64 {
	re := regexp.MustCompile(`[\d,]+\.?\d*`)
	matches := re.FindAllString(s, -1)
	if len(matches) == 0 {
		return 0
	}
	val, _ := strconv.ParseFloat(strings.Replace(matches[0], ",", ".", 1), 64)
	return val
}
