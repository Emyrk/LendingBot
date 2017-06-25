package balancer

import (
	"fmt"

	"github.com/Emyrk/LendingBot/src/core/poloniex"
)

const (
	PoloniexExchange int = iota
	BitfinexExchange
)

func GetExchangeString(exch int) string {
	switch exch {
	case PoloniexExchange:
		return "Poloniex"
	case BitfinexExchange:
		return "Bitfinex"
	}
	return fmt.Sprintf("Unknown {%d}", exch)
}

type OrderDensity struct {
	Amount float64
	Rate   float64

	Orders []poloniex.PoloniexLoanOrder
}

func GetDensityOfLoans(orders *poloniex.PoloniexLoanOrders) []OrderDensity {
	all := make([]OrderDensity, 2002)
	for _, order := range orders.Offers {
		if int(order.Rate*100000) > 2000 {
			all[2001].AddOrder(order)
		} else {
			all[int(order.Rate*100000)].AddOrder(order)
		}
	}
	return all
}

func (od *OrderDensity) AddOrder(order poloniex.PoloniexLoanOrder) {
	prev := od.Amount
	od.Amount = od.Amount + order.Amount
	od.Rate = od.Rate * prev / od.Amount
	od.Rate = od.Rate + (order.Rate*order.Amount)/od.Amount
	//od.Rate = order.Rate
	od.Orders = append(od.Orders[:], order)
}

func (od *OrderDensity) String() string {
	str := fmt.Sprintf("Loan Density Info of %d loans - Total Coin: %.4f at AVG: %.4f%s", len(od.Orders), od.Amount, od.Rate*100, "%")
	return str
}
