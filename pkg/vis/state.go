package vis

import "sync"

// StateStore saves the current object states
type StateStore interface {
	Objects() (map[string]Object, error)
	Reset() error
	Update(objs ...Object) error
	Remove(ids ...string) error
}

// MemStateStore is an in-memory implementation of StateStore
type MemStateStore struct {
	objects map[string]Object
	lock    sync.RWMutex
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

// Reset implements StateStore
func (s *MemStateStore) Reset() error {
	s.lock.Lock()
	s.objects = nil
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

// Remove implements StateStore
func (s *MemStateStore) Remove(ids ...string) error {
	s.lock.Lock()
	if s.objects != nil {
		for _, id := range ids {
			delete(s.objects, id)
		}
	}
	s.lock.Unlock()
	return nil
}
