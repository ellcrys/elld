package ticket

// Period represents a ticket sale period where
// a new price is calculated and used for the
// duration of the period
type Period interface {
	// GetNumber gets the number of the current period
	GetNumber() uint64
}

// PricePeriod implements a Period
type PricePeriod struct {
	blockNumber uint64
}

// NewPeriodFromNumber creates an instance of PricePeriod
func NewPeriodFromNumber(number uint64) *PricePeriod {
	return &PricePeriod{blockNumber: number}
}

// GetNumber returns block number where the period began
func (pp *PricePeriod) GetNumber() uint64 {
	return pp.blockNumber
}
