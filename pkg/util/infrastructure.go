package util

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	configv1 "github.com/openshift/api/config/v1"
)

const (
	// InfrastructureName is the name of the Infrastructure,
	// as Infrastructure is a singleton within the cluster.
	InfrastructureName = "cluster"
)

// GetInfrastructure returns the infrastructure matching the infrastructureName.
func GetInfrastructure(ctx context.Context, cl client.Client) (*configv1.Infrastructure, error) {
	infrastructure := &configv1.Infrastructure{}
	if err := cl.Get(ctx, client.ObjectKey{Name: InfrastructureName}, infrastructure); err != nil {
		return nil, fmt.Errorf("unable to get infrastructure object: %w", err)
	}

	return infrastructure, nil
}
