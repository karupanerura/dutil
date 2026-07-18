package datastore

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"slices"
	"unsafe"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/datastore/apiv1/datastorepb"
)

func NewClient(ctx context.Context, opts Options) (*datastore.Client, error) {
	if opts.Emulator != "" {
		os.Setenv("DATASTORE_EMULATOR_HOST", opts.Emulator)
	}
	return datastore.NewClientWithDatabase(ctx, opts.ProjectID, opts.DatabaseID)
}

type LowLevelClient struct {
	c          *datastore.Client
	lc         datastorepb.DatastoreClient
	dataset    string
	databaseID string
}

// NewLowLevelClient extracts the Datastore SDK's private client, dataset, and
// databaseID fields. Metadata lookup needs those values, which the SDK does not
// expose through its public API. Keep TestNewLowLevelClientSDKCompatibility in
// sync with cloud.google.com/go/datastore upgrades.
func NewLowLevelClient(client *datastore.Client) *LowLevelClient {
	pv := reflect.ValueOf(client)
	sv := pv.Elem()
	lc := extractPrivateField[datastorepb.DatastoreClient](sv, "client")
	dataset := extractPrivateField[string](sv, "dataset")
	databaseID := extractPrivateField[string](sv, "databaseID")
	return &LowLevelClient{c: client, lc: lc, dataset: dataset, databaseID: databaseID}
}

func extractPrivateField[T any](sv reflect.Value, fieldName string) T {
	fv := sv.FieldByName(fieldName)
	cfv := reflect.NewAt(fv.Type(), unsafe.Pointer(fv.UnsafeAddr())).Elem()
	return cfv.Interface().(T)
}

func (c *LowLevelClient) GetMetadata(ctx context.Context, key *datastore.Key) (*EntityMetadata, error) {
	res, err := c.lc.Lookup(ctx, &datastorepb.LookupRequest{
		ProjectId:  c.dataset,
		DatabaseId: c.databaseID,
		Keys:       []*datastorepb.Key{c.toLowLevelKey(key)},
	})
	if err != nil {
		return nil, err
	}

	if len(res.Deferred) != 0 {
		return c.GetMetadata(ctx, key)
	}
	if len(res.Found) == 0 {
		return nil, fmt.Errorf("key=%s is not found", key.String())
	}

	e := res.Found[0]
	return &EntityMetadata{
		CreateTime: e.CreateTime.AsTime(),
		UpdateTime: e.UpdateTime.AsTime(),
		Version:    e.Version,
	}, nil
}

func (c *LowLevelClient) toLowLevelKey(src *datastore.Key) *datastorepb.Key {
	k := src
	var path []*datastorepb.Key_PathElement
	for {
		el := &datastorepb.Key_PathElement{Kind: k.Kind}
		if k.ID != 0 {
			el.IdType = &datastorepb.Key_PathElement_Id{Id: k.ID}
		} else if k.Name != "" {
			el.IdType = &datastorepb.Key_PathElement_Name{Name: k.Name}
		}
		path = append(path, el)
		if k.Parent == nil {
			break
		}
		k = k.Parent
	}
	slices.Reverse(path)

	key := &datastorepb.Key{Path: path}
	if src.Namespace != "" {
		key.PartitionId = &datastorepb.PartitionId{
			NamespaceId: src.Namespace,
		}
	}
	return key
}
