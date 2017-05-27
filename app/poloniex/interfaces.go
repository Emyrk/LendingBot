package poloniex

type IBotExchange interface {
	Setup(exch Exchanges)
	Start()
	SetDefaults()
	GetName() string
	IsEnabled() bool
}
