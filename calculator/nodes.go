package calculator

func CalculateNodePrice(hourlyPrice float64) NodePrice {
	return NodePrice{
		Hourly:  hourlyPrice,
		Weekly:  (hourlyPrice * 7),
		Monthly: ((hourlyPrice * 7) * 30),
		Yearly:  (((hourlyPrice * 7) * 30) * 12),
	}
}
