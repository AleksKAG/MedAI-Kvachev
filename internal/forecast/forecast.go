package forecast

import (
	"gonum.org/v1/gonum/stat"
	"math"
)

// Простая линейная регрессия: прогноз следующего значения
func LinearTrend(values []float64) float64 {
	n := len(values)
	if n < 2 {
		return values[0]
	}

	x := make([]float64, n)
	for i := 0; i < n; i++ {
		x[i] = float64(i)
	}

	slope, _ := stat.LinearRegression(x, values, nil, false)
	next := values[n-1] + slope
	return math.Round(next*100) / 100
}
