package streams

import (
	"bytes"
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
	"strings"
)

type MetadataItem struct {
	Name string
	Data []byte
}

func (mi *MetadataItem) Clone() *MetadataItem {
	return &MetadataItem{
		Name: mi.Name,
		Data: bytes.Clone(mi.Data),
	}
}

type MetadataCollection map[string]*MetadataItem

func NewMetadataCollection() MetadataCollection {
	return make(map[string]*MetadataItem)
}

func NewMetaDataCollectionFromBytes(collectionBytes []byte) (MetadataCollection, error) {
	var mdc map[string]*MetadataItem
	err := msgpack.Unmarshal(collectionBytes, &mdc)
	return mdc, err
}

func (mc MetadataCollection) ToBytes() ([]byte, error) {
	if mc == nil {
		return nil, fmt.Errorf("metadata Collection is not initialized")
	}

	collectionBytes, err := msgpack.Marshal(mc)
	if err != nil {
		return nil, fmt.Errorf("unable to serialize Metadata Collection: %w", err)
	}

	return collectionBytes, nil
}

func (mc MetadataCollection) AddMetadataItem(mi *MetadataItem) error {
	if mc == nil {
		return fmt.Errorf("metadata Collection is not initialized")
	}

	_, exists := mc[strings.ToUpper(mi.Name)]
	if exists {
		return fmt.Errorf("a metadata item with name \"%s\" already exists", mi.Name)
	}

	mc[strings.ToUpper(mi.Name)] = mi.Clone()

	return nil
}

func (mc MetadataCollection) GetMetadataItem(name string) *MetadataItem {
	if mc == nil {
		return nil
	}

	return mc[strings.ToUpper(name)]
}

func (mc MetadataCollection) Clone() MetadataCollection {
	mcOut := make(map[string]*MetadataItem)

	for name, item := range mc {
		mcOut[name] = item.Clone()
	}

	return mcOut
}
