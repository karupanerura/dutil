package datastore

import (
	"encoding/json"
	"fmt"

	"cloud.google.com/go/datastore"
	proto "google.golang.org/genproto/googleapis/datastore/v1"
)

type Property struct {
	Value
	Name    string `json:"name"`
	NoIndex bool   `json:"noIndex,omitempty"`
}

func NewPropertiesByProtoValueMap(m map[string]any) []Property {
	props := make([]Property, 0, len(m))
	for name, value := range m {
		prop := Property{Name: name}

		v, ok := value.(*proto.Value)
		if !ok {
			panic(fmt.Sprintf("unexpected value type: %T", value))
		}
		prop.fromDatastoreProtoValue(v)

		props = append(props, prop)
	}
	return props
}

func (p *Property) fromDatastoreProperty(prop datastore.Property) {
	p.Name = prop.Name
	p.Value.fromDatastoreValue(prop.Value)
	p.NoIndex = prop.NoIndex
}

func (p *Property) toDatastoreProperty() (prop datastore.Property) {
	prop.Name = p.Name
	prop.Value = p.Value.toDatastoreValue()
	prop.NoIndex = p.NoIndex
	return
}

func (p *Property) UnmarshalJSON(b []byte) error {
	var box struct {
		Name    string `json:"name"`
		NoIndex bool   `json:"noIndex,omitempty"`
	}
	if err := json.Unmarshal(b, &box); err != nil {
		return err
	}
	p.Name = box.Name
	p.NoIndex = box.NoIndex

	return p.Value.UnmarshalJSON(b)
}
