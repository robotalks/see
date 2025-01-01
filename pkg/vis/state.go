package vis

import "sync"

// StateStore saves the current object states
type StateStore interface {
	Objects() (map[string]Object, error)
	DataValues() (map[string]DataValue, error)
	Reset() error
	Update(objs ...Object) error
	UpdateDataValue(id string, val DataValue) error
	Remove(ids ...string) error
}

// MemStateStore is an in-memory implementation of StateStore
type MemStateStore struct {
	lock    sync.RWMutex
	objects map[string]Object
	data    map[string]DataValue
}

// Objects implements StateStore
func (s *MemStateStore) Objects() (objs map[string]Object, err error) {
	objs = make(map[string]Object)
	s.lock.RLock()
	if s.objects != nil {
		for id, obj := range s.objects {
			objs[id] = obj
		}
	}
	s.lock.RUnlock()
	return
}

// DataValues implements StateStore
func (s *MemStateStore) DataValues() (data map[string]DataValue, err error) {
	data = make(map[string]DataValue)
	s.lock.RLock()
	if s.data != nil {
		for id, val := range s.data {
			data[id] = val
		}
	}
	s.lock.RUnlock()
	return
}

// Reset implements StateStore
func (s *MemStateStore) Reset() error {
	s.lock.Lock()
	s.objects = nil
	s.data = nil
	s.lock.Unlock()
	return nil
}

// Update implements StateStore
func (s *MemStateStore) Update(objs ...Object) error {
	s.lock.Lock()
	if s.objects == nil {
		s.objects = make(map[string]Object)
	}
	for _, obj := range objs {
		s.objects[obj.ID()] = obj
	}
	s.lock.Unlock()
	return nil
}

func (s *MemStateStore) UpdateDataValue(id string, val DataValue) error {
	s.lock.Lock()
	if s.data == nil {
		s.data = make(map[string]DataValue)
	}
	s.data[id] = val
	s.lock.Unlock()
	return nil
}

// Remove implements StateStore
func (s *MemStateStore) Remove(ids ...string) error {
	s.lock.Lock()
	if s.objects != nil {
		for _, id := range ids {
			delete(s.objects, id)
		}
	}
	if s.data != nil {
		for _, id := range ids {
			delete(s.data, id)
		}
	}
	s.lock.Unlock()
	return nil
}
