package numberutils

func FloatRound(v float64) float64 {
	decimals := 2
	var pow float64 = 1
	for i := 0; i < decimals; i++ {
		pow *= 10
	}
	if v < 0 {
		return float64(int((v*pow)-0.5)) / pow
	} else {
		return float64(int((v*pow)+0.5)) / pow
	}
}
