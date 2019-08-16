package burner

const (
	// EventNewBlock indicates the mining of a new burner chain block
	EventNewBlock = "new-block"

	// EventInvalidLocalBlock indicates that the burner chain has or
	// is currently going through a re-org.
	EventInvalidLocalBlock = "burner-chain-reorg"
)
