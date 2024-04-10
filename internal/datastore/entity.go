package datastore

import (
	"cloud.google.com/go/datastore"
)

type Entity struct {
	Key        *Key       `json:"key"`
	Properties []Property `json:"properties"`
}

func (e *Entity) LoadKey(key *datastore.Key) error {
	e.Key = FromDatastoreKey(key)
	return nil
}

func (e *Entity) Load(props []datastore.Property) error {
	e.Properties = make([]Property, len(props))
	for i, prop := range props {
		e.Properties[i].fromDatastoreProperty(prop)
	}
	return nil
}

func (e *Entity) Save() ([]datastore.Property, error) {
	props := make([]datastore.Property, len(e.Properties))
	for i, prop := range e.Properties {
		props[i] = prop.toDatastoreProperty()
	}
	return props, nil
}
