package gcp

import (
	"context"
	"fmt"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

type GCPConfig struct {
	ProjectID string
	Region    string
	Zone      string
}

func NewGCPComputeService(ctx context.Context, config *GCPConfig) (*compute.Service, error) {
	// Automatically uses GOOGLE_APPLICATION_CREDENTIALS env var

	service, err := compute.NewService(ctx, option.WithScopes(compute.CloudPlatformScope))
	if err != nil {
		return nil, fmt.Errorf("failed to create compute service: %w", err)
	}
	return service, nil
}
