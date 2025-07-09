package datastore

import (
	"net/url"
	"slices"
	"strconv"
	"strings"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/datastore/apiv1/datastorepb"
	"google.golang.org/protobuf/encoding/prototext"
)

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

func (k *Key) ToProto() *datastorepb.Key {
	root := k
	keys := []*Key{root}
	for root.Parent != nil {
		root = root.Parent
		keys = append(keys, root)
	}
	slices.Reverse(keys)

	path := make([]*datastorepb.Key_PathElement, len(keys))
	for i, k := range keys {
		if k.Name != "" {
			path[i] = &datastorepb.Key_PathElement{
				Kind:   k.Kind,
				IdType: &datastorepb.Key_PathElement_Name{Name: k.Name},
			}
		} else {
			path[i] = &datastorepb.Key_PathElement{
				Kind:   k.Kind,
				IdType: &datastorepb.Key_PathElement_Id{Id: k.ID},
			}
		}
	}

	var partitionId *datastorepb.PartitionId
	if k.Namespace != "" {
		partitionId = &datastorepb.PartitionId{
			NamespaceId: k.Namespace,
		}
	}

	return &datastorepb.Key{
		PartitionId: partitionId,
		Path:        path,
	}
}

func (k *Key) ToProtoEncoded() string {
	return url.QueryEscape(prototext.Format(k.ToProto()))
}

func (k *Key) String() string {
	var s strings.Builder
	s.WriteString("KEY(")
	if k.Namespace != "" {
		s.WriteString("NAMESPACE(")
		s.WriteString(strconv.Quote(k.Namespace))
		s.WriteString(")")
	}
	k.string(&s)
	s.WriteString(")")
	return s.String()
}

func (k *Key) string(s *strings.Builder) {
	if k.Parent != nil {
		k.Parent.string(s)
		s.WriteByte(',')
	}
	s.WriteString(k.Kind)
	s.WriteByte(',')
	if k.ID != 0 {
		s.WriteString(strconv.FormatInt(k.ID, 10))
	} else {
		s.WriteString(strconv.Quote(k.Name))
	}
}

func DecodeKey(encoded string) (*Key, error) {
	key, err := datastore.DecodeKey(encoded)
	if err != nil {
		return nil, err
	}
	return FromDatastoreKey(key), nil
}

func ParseEncodedProtoKey(s string) (*Key, error) {
	decoded, err := url.QueryUnescape(s)
	if err != nil {
		return nil, err
	}
	return ParseTextProtoKey([]byte(decoded))
}

func ParseTextProtoKey(b []byte) (*Key, error) {
	var key datastorepb.Key
	if err := prototext.Unmarshal(b, &key); err != nil {
		return nil, err
	}
	return FromProtoKey(&key), nil
}

func FromProtoKey(src *datastorepb.Key) *Key {
	var namespace string
	if src.PartitionId != nil {
		namespace = src.PartitionId.NamespaceId
	}

	var target *Key
	for _, elem := range src.Path {
		key := &Key{
			Kind:      elem.Kind,
			Namespace: namespace,
			Parent:    target,
		}
		if id, ok := elem.IdType.(*datastorepb.Key_PathElement_Id); ok {
			key.ID = id.Id
		} else if name, ok := elem.IdType.(*datastorepb.Key_PathElement_Name); ok {
			key.Name = name.Name
		}
		target = key
	}
	return target
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

type KeyFormatter struct {
	Format string
}

func (k *KeyFormatter) FormatKey(key *Key) any {
	switch k.Format {
	case "json":
		return key
	case "gql":
		return key.String()
	case "encoded":
		return key.ToDatastore().Encode()
	case "proto":
		return key.ToProtoEncoded()
	default:
		return key
	}
}
