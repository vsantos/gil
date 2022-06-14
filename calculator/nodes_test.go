package calculator

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestRoundFloat(t *testing.T) {
	assert.Equal(t, RoundFloat(0.123, 1), 0.1)
	assert.Equal(t, RoundFloat(0.123, 2), 0.12)
	assert.Equal(t, RoundFloat(0.123, 3), 0.123)
	assert.Equal(t, RoundFloat(0.1, 2), 0.1)
	assert.Equal(t, RoundFloat(0.1, 3), 0.1)
}

func TestCalculateNodePrice(t *testing.T) {
	type pricesCase struct {
		price    float64
		expected NodePrice
	}

	pCases := []pricesCase{
		{
			price: 312,
			expected: NodePrice{
				Hourly:  312,
				Daily:   7488,
				Weekly:  52416,
				Monthly: 1572480,
				Yearly:  18869760,
			},
		},
		{
			price: 0.06,
			expected: NodePrice{
				Hourly:  0.06,
				Daily:   1.44,
				Weekly:  10.08,
				Monthly: 302.4,
				Yearly:  3628.8,
			},
		},
		{
			price: 0.00,
			expected: NodePrice{
				Hourly:  0,
				Daily:   0,
				Weekly:  0,
				Monthly: 0,
				Yearly:  0,
			},
		},
	}

	for _, tc := range pCases {
		cPrices := CalculateNodePrice(tc.price)
		assert.Equal(t, cPrices.Daily, tc.expected.Daily)
		assert.Equal(t, cPrices.Hourly, tc.expected.Hourly)
		assert.Equal(t, cPrices.Weekly, tc.expected.Weekly)
		assert.Equal(t, cPrices.Monthly, tc.expected.Monthly)
		assert.Equal(t, cPrices.Yearly, tc.expected.Yearly)
	}
}
