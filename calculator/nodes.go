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
		Hourly:  RoundFloat(hourly, 2),
		Daily:   RoundFloat(daily, 2),
		Weekly:  RoundFloat(weekly, 2),
		Monthly: RoundFloat(monthly, 2),
		Yearly:  RoundFloat(yearly, 2),
	}
}
