package ticket

import (
	"math"

	"github.com/shopspring/decimal"

	"github.com/ellcrys/elld/types"
)

// Manager manages the ticket creation, pricing and expiration.
// It implements types.TicketManager
type Manager struct {
	bChain types.Blockchain
}

// NewManager creates an instance of Manager
func NewManager(bChain types.Blockchain) *Manager {
	return &Manager{bChain: bChain}
}

// DetermineTerm takes a block number returns its term number.
func (m *Manager) DetermineTerm(blockNum uint64) uint {
	if blockNum < BlocksPerTerm {
		return 1
	}
	r := float64(blockNum) / float64(BlocksPerTerm)
	return uint(math.Ceil(r))
}

// DetermineCurrentTerm returns the current ticket term
func (m *Manager) DetermineCurrentTerm() (uint, error) {
	curBlockHeader, err := m.bChain.GetBestChain().Current()
	if err != nil {
		return 0, err
	}
	return m.DetermineTerm(curBlockHeader.GetNumber()), nil
}

// DeterminePrice calculates the price for the current term
// TODO: Devise a proper algorithm
func (m *Manager) DeterminePrice() decimal.Decimal {
	return decimal.NewFromFloat(50)
}
