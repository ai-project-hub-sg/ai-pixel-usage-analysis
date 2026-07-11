package analytics

func PercentChange(current, baseline float64) *float64 {
	if baseline == 0 {
		return nil
	}
	value := (current - baseline) / baseline * 100
	return &value
}
