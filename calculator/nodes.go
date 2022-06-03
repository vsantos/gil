package calculator

func CalculateNodePrice(hourlyPrice float64) NodePrice {
	return NodePrice{
		Hourly:  hourlyPrice,
		Daily:   (hourlyPrice * 24),
		Weekly:  ((hourlyPrice * 24) * 7),
		Monthly: (((hourlyPrice * 24) * 7) * 30),
		Yearly:  ((((hourlyPrice * 24) * 7) * 30) * 12),
	}
}
