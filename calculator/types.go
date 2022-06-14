package calculator

type Calculator interface {
	Nodes() NodePrice
}

type NodePrice struct {
	Hourly  float64 `json:"hourly"`
	Daily   float64 `json:"daily"`
	Weekly  float64 `json:"weekly"`
	Monthly float64 `json:"monthly"`
	Yearly  float64 `json:"yearly"`
}
