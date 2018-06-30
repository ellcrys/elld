package logic

import (
	"github.com/ellcrys/elld/database"
)

// ObjectsPut store objects
func (l *Logic) ObjectsPut(addresses []*database.KVObject, errCh chan error) error {
	err := l.engine.DB().WriteBatch(addresses)
	return sendErr(errCh, err)
}

// ObjectsGet store objects
func (l *Logic) ObjectsGet(prefix []byte, result chan []*database.KVObject) error {
	result <- l.engine.DB().GetByPrefix(prefix)
	return nil
}
