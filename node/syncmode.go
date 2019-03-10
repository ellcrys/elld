package node

import "sync"

// SyncMode describes how the state
// of the node should be synchronized
// with external nodes.
type SyncMode struct {
	sync.RWMutex
	disabled bool
}

// NewDefaultSyncMode creates a SyncMode object
// that describes the default sync behaviour
func NewDefaultSyncMode(disabled bool) *SyncMode {
	return &SyncMode{
		disabled: disabled,
	}
}

// IsDisabled checks whether the syncing
// has been disabled.
func (s *SyncMode) IsDisabled() bool {
	s.RLock()
	defer s.RUnlock()
	return s.disabled
}

// Enable enables the sync mode
func (s *SyncMode) Enable() {
	s.Lock()
	defer s.Unlock()
	s.disabled = false
}

// Disable disables the sync mode
func (s *SyncMode) Disable() {
	s.Lock()
	defer s.Unlock()
	s.disabled = true
}
