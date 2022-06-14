package calculator

import "math"

func RoundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func CalculateNodePrice(hourlyPrice float64) NodePrice {
	hourly := hourlyPrice
	daily := hourlyPrice * 24
	weekly := daily * 7
	monthly := weekly * 30
	yearly := monthly * 12

	return NodePrice{
		Hourly:  RoundFloat(hourly, 4),
		Daily:   RoundFloat(daily, 4),
		Weekly:  RoundFloat(weekly, 4),
		Monthly: RoundFloat(monthly, 4),
		Yearly:  RoundFloat(yearly, 4),
	}
}
