package datastore

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"slices"
	"unsafe"

	"cloud.google.com/go/datastore"
	pb "google.golang.org/genproto/googleapis/datastore/v1"
)

func NewClient(ctx context.Context, opts Options) (*datastore.Client, error) {
	if opts.Emulator != "" {
		os.Setenv("DATASTORE_EMULATOR_HOST", opts.Emulator)
	}
	return datastore.NewClientWithDatabase(ctx, opts.ProjectID, opts.DatabaseID)
}

type LowLevelClient struct {
	c          *datastore.Client
	lc         pb.DatastoreClient
	dataset    string
	databaseID string
}

func NewLowLevelClient(client *datastore.Client) *LowLevelClient {
	pv := reflect.ValueOf(client)
	sv := pv.Elem()
	lc := extractPrivateField[pb.DatastoreClient](sv, "client")
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
	res, err := c.lc.Lookup(ctx, &pb.LookupRequest{
		ProjectId:  c.dataset,
		DatabaseId: c.databaseID,
		Keys:       []*pb.Key{c.toLowLevelKey(key)},
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

func (c *LowLevelClient) toLowLevelKey(src *datastore.Key) *pb.Key {
	k := src
	var path []*pb.Key_PathElement
	for {
		el := &pb.Key_PathElement{Kind: k.Kind}
		if k.ID != 0 {
			el.IdType = &pb.Key_PathElement_Id{Id: k.ID}
		} else if k.Name != "" {
			el.IdType = &pb.Key_PathElement_Name{Name: k.Name}
		}
		path = append(path, el)
		if k.Parent == nil {
			break
		}
		k = k.Parent
	}
	slices.Reverse(path)

	key := &pb.Key{Path: path}
	if src.Namespace != "" {
		key.PartitionId = &pb.PartitionId{
			NamespaceId: src.Namespace,
		}
	}
	return key
}
