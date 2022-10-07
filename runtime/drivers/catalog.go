package drivers

import (
	"context"
)

// CatalogStore is implemented by drivers capable of storing catalog info for a specific instance
type CatalogStore interface {
	FindObjects(ctx context.Context, instanceID string) []*CatalogObject
	FindObject(ctx context.Context, instanceID string, name string) (*CatalogObject, bool)
	CreateObject(ctx context.Context, instanceID string, object *CatalogObject) error
	UpdateObject(ctx context.Context, instanceID string, object *CatalogObject) error
	DeleteObject(ctx context.Context, instanceID string, name string) error
}

// Constants representing different kinds of catalog objects
const (
	CatalogObjectTypeSource         string = "source"
	CatalogObjectTypeMetricsView    string = "metrics view"
	CatalogObjectTypeUnmanagedTable string = "unmanaged table"
)

// CatalogObject represents one object in the catalog, such as a source
type CatalogObject struct {
	Name string
	Type string
	Blob string
}
