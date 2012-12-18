package database

type Product struct {
	AccountId  int64
	Pid        string
	Name       string
	ProductUrl string
	ImageUrl   string
	UnitCost   float64
	UnitPrice  float64
	Margin     float64
	MarginRate float64
}

type Person struct {
    MonetateId string
}

