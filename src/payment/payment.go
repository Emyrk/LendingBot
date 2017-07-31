package payment

import (
	"time"
)

type Charge struct {
	LoanAmount float64
	LoanOpen   time.Time
	LoanClose  time.Time
	Charge     float64
	Paid       bool
}
