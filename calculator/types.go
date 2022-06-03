package calculator

type Calculator interface {
	Nodes() NodePrice
}

type NodePrice struct {
	Hourly  float64
	Daily   float64
	Weekly  float64
	Monthly float64
	Yearly  float64
}
