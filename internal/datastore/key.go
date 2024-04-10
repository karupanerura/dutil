package datastore

import "cloud.google.com/go/datastore"

type Key struct {
	Kind      string `json:"kind"`
	ID        int64  `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	Parent    *Key   `json:"parent,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

func (k *Key) ToDatastore() *datastore.Key {
	key := &datastore.Key{
		Kind:      k.Kind,
		ID:        k.ID,
		Name:      k.Name,
		Namespace: k.Namespace,
	}
	if k.Parent != nil {
		key.Parent = k.Parent.ToDatastore()
	}
	return key
}

func DecodeKey(encoded string) (*Key, error) {
	key, err := datastore.DecodeKey(encoded)
	if err != nil {
		return nil, err
	}
	return FromDatastoreKey(key), nil
}

func FromDatastoreKey(src *datastore.Key) *Key {
	dest := &Key{
		Kind:      src.Kind,
		ID:        src.ID,
		Name:      src.Name,
		Namespace: src.Namespace,
	}
	if src.Parent != nil {
		dest.Parent = FromDatastoreKey(src.Parent)
	}
	return dest
}

type Keys []*Key

func (k Keys) ToDatastore() []*datastore.Key {
	keys := make([]*datastore.Key, len(k))
	for i, k := range k {
		keys[i] = k.ToDatastore()
	}
	return keys
}
