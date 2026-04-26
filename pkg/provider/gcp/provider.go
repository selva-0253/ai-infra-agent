package gcp

import (
	"context"
	"fmt"
	"google.golang.org/api/compute/v1"
	"ai-infra-agent/pkg/provider"

)

type GCPProvider struct {
	projectID string
	region string
	zone string
	service *compute.Service
}

func NewGCPProvider(ctx context.Context, projectID, region, zone string) (*GCPProvider, error) {
	service, err := NewGCPComputeService(ctx, &GCPConfig{
		ProjectID: projectID,
		Region:    region,
		Zone:      zone,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create GCP provider: %w", err)
	}
	return &GCPProvider{
		projectID: projectID,
		region:    region,
		zone:      zone,
		service:   service,
	}, nil
}

func (g *GCPProvider) CreateInstance(ctx context.Context, req *provider.InstanceRequest) (*provider.InstanceResponse, error) {
	// GCP equivalent of EC2 RunInstances
	instance := &compute.Instance{
		Name: 	  req.Name,
		MachineType: fmt.Sprintf("zones/%s/machineTypes/%s", g.zone, req.MachineType),
		SourceImage: req.ImageName,
		NetworkInterfaces: []*compute.NetworkInterface{
			{
				Network: "global/networks/default",
			}
		},
	}
	op, error := g.service.Instances.Insert(g.projectID, g.zone, instance).Context(ctx).Do()
	if error != nil {
		return nil, fmt.Errorf("failed to create instance: %w", error)
	}
	return &provider.InstanceResponse{
		ID: op.TargetId,
		Name: req.Name,
	}, nil

}


//Implement other methods like ListInstance, DeleteInstance, CreateBucket, CreateVPC similarly using GCP APIs.
