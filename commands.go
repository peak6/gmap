package gmap

type Command struct {
	Action  string
	Entries map[string]interface{}
}

type Entry struct {
	Key   string
	Value interface{}
}

type Sync struct {
	Action string
	Node   string
	Store  *Store
}
