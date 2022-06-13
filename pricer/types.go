package pricer

type ProviderInterface interface {
	Nodes() ProviderNodesInterface
}

type ProviderNodesInterface interface {
	Prices() PriceInterface
}

type PriceInterface interface {
	List() ProviderNodes
}

// Being the key the instanceType
type ProviderNodes map[string]Node

type Node struct {
	Type      string
	Labels    map[string]string
	Resources NodeResources
	Cost      NodeCost
}

type NodeResources struct {
	CPU      string
	VCPU     int
	Arch     string
	MemoryGB float32
}

type NodeCost struct {
	Type         string
	RegionalCost RegionalCost
	Currency     string
}

type RegionalCost struct {
	Value map[string]float64
}
