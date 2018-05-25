package database

// DB describes the database access, model and available functionalities
type DB interface {
	Open(namespace string) error
	Close() error
	Address() AddrStore
	WriteBatch([][]byte, [][]byte) error

	// GetByPrefix returns keys matching a prefix. Their key and value are returned
	GetByPrefix([]byte) ([][]byte, [][]byte)

	DeleteByPrefix([]byte) error
}

// AddrStore describes a database interface for accessing and managing addresses
type AddrStore interface {
	GetAll() ([][]byte, error)
	SaveAll([][]byte) error
	ClearAll() error
}
