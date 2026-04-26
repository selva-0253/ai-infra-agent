package provider

import "context"

type CloudProvider interface {
	CreateInstance(ctx context.Context, req *InstanceRequest) (*InstanceResponse, error)
	ListInstance(ctx context.Context) ([]Instance, error)
	DeleteInstance(ctx context.Context, id string) error
	//Storage Operations
	CreateBucket(ctx context.Context, req *BucketRequest) (*BucketResponse, error)

	// Network Operations
	CreateVPC(ctx context.Context, req *VPCRequest) (*VPCResponse, error)
}

type InstanceRequest struct {
	Name        string
	MachineType string
	ImageName   string
	Count       int32
}

type InstanceResponse struct {
	ID   string
	Name string
}

type Instance struct {
	ID    string
	Name  string
	State string
}

type BucketRequest struct {
	Name     string
	Location string
}

type BucketResponse struct {
	Name string
}

type VPCRequest struct {
	Name string
	CIDR string
}

type VPCResponse struct {
	ID   string
	Name string
}
