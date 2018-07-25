package logic

import (
	"github.com/ellcrys/elld/database"
)

// ObjectsPut store objects
func (l *Logic) ObjectsPut(objs []*database.KVObject, errCh chan error) error {
	err := l.engine.DB().Put(objs)
	return sendErr(errCh, err)
}

// ObjectsGet store objects
func (l *Logic) ObjectsGet(prefix []byte, result chan []*database.KVObject) error {
	result <- l.engine.DB().GetByPrefix(prefix)
	return nil
}
