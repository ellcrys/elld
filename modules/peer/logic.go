package peer

// Logic represents the peers local logic.
// This protocol determines messaging type and format,
// initialization, storage and shutdown behaviour etc
type Logic interface {
	SendHandShake(*Peer) error
}
