package pricer

import (
	"fmt"
	"os"

	ec2instancesinfo "github.com/cristim/ec2-instances-info"
)

type ProviderAWS struct{}

func (a *ProviderAWS) Nodes() ProviderNodesInterface {
	// fmt.Println("price nodes")
	return a
}

func (a *ProviderAWS) Prices() PriceInterface {
	// fmt.Println("price prices")
	return a
}

func (a *ProviderAWS) List() ProviderNodes {
	// fmt.Println("price list")

	data, err := ec2instancesinfo.Data()

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var priced = make(ProviderNodes)

	for _, i := range *data {
		n := Node{}
		nc := NodeCost{}
		nr := NodeResources{}
		nc.Currency = "USD"
		nc.Type = i.InstanceType
		nc.RegionalCost.Value = map[string]float64{
			"sa-east-1":  i.Pricing["sa-east-1"].Linux.OnDemand,
			"us-east-1":  i.Pricing["us-east-1"].Linux.OnDemand,
			"us-east-2":  i.Pricing["us-east-2"].Linux.OnDemand,
			"ap-south-1": i.Pricing["ap-south-1"].Linux.OnDemand,
		}

		nr.CPU = i.PhysicalProcessor
		nr.VCPU = i.VCPU
		nr.MemoryGB = i.Memory

		n.Cost = nc
		n.Resources = nr

		priced[i.InstanceType] = n
	}

	return priced
}
