package gmap

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	. "github.com/peak6/logger"
	"sync"
)

var MyStore = NewSafeStore()

type Store struct {
	Data map[string]*OwnerMap
	lock *sync.RWMutex
}

func (s *Store) String() string {
	return fmt.Sprint(s.Data)
}

type Owner struct {
	Node   *NodeInfo
	Client string
}

type OwnerMap struct {
	Data map[Owner]interface{}
	lock *sync.RWMutex
}

func (om *OwnerMap) String() string {
	return fmt.Sprint(om.Data)
}

type StoreWalker func(path string, owner Owner, value interface{})

func NewStore() *Store {
	return &Store{Data: make(map[string]*OwnerMap)}

}
func NewSafeStore() *Store {
	return &Store{Data: make(map[string]*OwnerMap), lock: &sync.RWMutex{}}
}

func (s *Store) RemoveForNode(node *NodeInfo) {
	if s.lock != nil {
		s.lock.RLock()
		defer s.lock.RUnlock()
	}
	for _, om := range s.Data {
		if om.lock != nil {
			om.lock.Lock()
			defer om.lock.Unlock()
		}
		for owner, _ := range om.Data {
			if owner.Node.Name == node.Name {
				delete(om.Data, owner)
			} else {
				Linfo.Println("NO MATCH: ", owner.Node, node, "\n", spew.Sdump(owner.Node), spew.Sdump(node))
			}
		}
	}
}

func (s *Store) AddAll(store *Store) {
	store.ReadAll(func(path string, owner Owner, value interface{}) {
		s.GetOrCreateOwnerMap(path).Put(owner, value)
	})
}
func (s *Store) Spew() string {
	return spew.Sdump(s)
}

func (s *Store) GetOrCreateOwnerMap(path string) *OwnerMap {
	ret := s.GetOwnerMap(path)
	if ret != nil {
		return ret
	}
	if s.lock != nil {
		s.lock.Lock()
		defer s.lock.Unlock()
	}
	ret, ok := s.Data[path]
	if ok {
		return ret
	}
	ret = &OwnerMap{Data: make(map[Owner]interface{})}
	if s.lock != nil {
		ret.lock = &sync.RWMutex{}
	}
	s.Data[path] = ret
	return ret
}
func (s *Store) ReadAll(proc StoreWalker) {
	if s.lock != nil {
		s.lock.RLock()
		defer s.lock.RUnlock()
	}
	for path, om := range s.Data {
		if om.lock != nil {
			om.lock.RLock()
			defer om.lock.RUnlock()
		}
		for owner, val := range om.Data {
			proc(path, owner, val)
		}
	}
}
func (s *Store) GetMyEntries() *Store {
	ret := NewStore()
	s.ReadAll(func(path string, owner Owner, value interface{}) {
		if owner.Node == &MyNode {
			ret.GetOrCreateOwnerMap(path).Put(owner, value)
		}
	})
	return ret
}

func (s *Store) Put(path string, client string, value interface{}) {
	s.GetOrCreateOwnerMap(path).Put(Owner{Node: &MyNode, Client: client}, value)
}
func (s *Store) PutStatic(path string, value interface{}) {
	s.GetOrCreateOwnerMap(path).Put(Owner{Node: &MyNode, Client: "<static>"}, value)
}

func (s *Store) GetOwnerMap(path string) *OwnerMap {
	if s.lock != nil {
		s.lock.RLock()
		defer s.lock.RUnlock()
	}
	return s.Data[path]
}

func (om *OwnerMap) Put(owner Owner, value interface{}) {
	if om.lock != nil {
		om.lock.Lock()
		defer om.lock.Unlock()
	}
	om.Data[owner] = value
}
func (om *OwnerMap) Get(owner Owner) interface{} {
	if om.lock != nil {
		om.lock.RLock()
		defer om.lock.RUnlock()
	}
	return om.Data[owner]
}

/*type PathEntries map[EntryRef]interface{}

type EntryRef struct {
	Node   string
	Client string
}

type DataStore struct {
	data map[string]PathEntries
	lock sync.Mutex
}

type LockingStore struct {
	DataStore
	lock sync.Mutex
}

func newStore() DataStore {
	return DataStore{data: make(map[string]PathEntries)}
}

var Store = newStore()

func (s *DataStore) Put(path string, ref EntryRef, val interface{}) {
	s.lock.Lock()
	defer s.lock.Unlock()
	ents, ok := s.data[path]
	if !ok {
		ents = make(PathEntries)
		s.data[path] = ents
	}
	ents[ref] = val

}

func (s *DataStore) GetNodeEntries(node string) map[string]PathEntries {
	s.lock.Lock()
	defer s.lock.Unlock()
	ret := make(map[string]PathEntries)
	for path, ents := range s.data {
		for ref, val := range ents {
			if ref.Node == node {
				rp, ok := ret[path]
				if !ok {
					rp = make(PathEntries)
					ret[path] = rp
				}
				rp[ref] = val
			}
		}
	}
	if len(ret) > 0 {
		return ret
	} else {
		return nil
	}
}

func (s *DataStore) Get(path string) PathEntries {
	s.lock.Lock()
	defer s.lock.Unlock()
	e := s.data[path]
	if e == nil {
		return nil
	}
	ret := make(PathEntries, len(e))
	for k, v := range e {
		ret[k] = v
	}
	return ret
}

func (s *DataStore) Spew() string {
	return spew.Sdump(s)
}
*/
