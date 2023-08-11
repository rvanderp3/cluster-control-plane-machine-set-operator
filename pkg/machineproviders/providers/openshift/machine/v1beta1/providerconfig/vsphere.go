/*
Copyright 2023 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package providerconfig

import (
	"fmt"

	"github.com/go-logr/logr"
	v1 "github.com/openshift/api/config/v1"
	machinev1 "github.com/openshift/api/machine/v1"
	machinev1beta1 "github.com/openshift/api/machine/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

// VSphereProviderConfig holds the provider spec of a VSphere Machine.
// It allows external code to extract and inject failure domain information,
// as well as gathering the stored config.
type VSphereProviderConfig struct {
	providerConfig machinev1beta1.VSphereMachineProviderSpec
	infrastructure *v1.Infrastructure
}

// InjectFailureDomain returns a new VSphereProviderConfig configured with the failure domain.
//
//nolint:nestif
func (v VSphereProviderConfig) InjectFailureDomain(fd machinev1.VSphereFailureDomain) VSphereProviderConfig {
	newVSphereProviderConfig := v

	for _, failureDomain := range v.infrastructure.Spec.PlatformSpec.VSphere.FailureDomains {
		if failureDomain.Name == fd.Name {
			workspace := newVSphereProviderConfig.providerConfig.Workspace
			topology := failureDomain.Topology

			if len(topology.ComputeCluster) > 0 {
				workspace.ResourcePool = fmt.Sprintf("%s/resources", topology.ComputeCluster)
			}

			if len(topology.ResourcePool) > 0 {
				workspace.ResourcePool = topology.ResourcePool
			}

			if len(topology.Datacenter) > 0 {
				workspace.Datacenter = topology.Datacenter
			}

			if len(topology.Datastore) > 0 {
				workspace.Datastore = topology.Datastore
			}

			if len(topology.Networks) > 0 {
				newVSphereProviderConfig.providerConfig.Network = machinev1beta1.NetworkSpec{
					Devices: []machinev1beta1.NetworkDeviceSpec{
						{
							NetworkName: topology.Networks[0],
						},
					},
				}
			}

			// TO-DO: fix in SPLAT-1141/1140
			if len(topology.Template) > 0 {
				newVSphereProviderConfig.providerConfig.Template = topology.Template
			} else if len(v.infrastructure.Spec.PlatformSpec.VSphere.FailureDomains) > 1 {
				newVSphereProviderConfig.providerConfig.Template = fmt.Sprintf("%s-rhcos-%s-%s", v.infrastructure.Status.InfrastructureName, failureDomain.Region, failureDomain.Zone)
			}

			if len(topology.Folder) > 0 {
				workspace.Folder = topology.Folder
			} else {
				workspace.Folder = fmt.Sprintf("/%s/vm/", workspace.Datacenter)
			}
		}
	}

	return newVSphereProviderConfig
}

// ExtractFailureDomain is used to extract a failure domain from the ProviderConfig.
func (v VSphereProviderConfig) ExtractFailureDomain() machinev1.VSphereFailureDomain {
	workspace := v.providerConfig.Workspace
	failureDomains := v.infrastructure.Spec.PlatformSpec.VSphere.FailureDomains

	for _, failureDomain := range failureDomains {
		topology := failureDomain.Topology
		if workspace.Datacenter == topology.Datacenter &&
			workspace.Datastore == topology.Datastore &&
			workspace.Server == failureDomain.Server &&
			workspace.ResourcePool == topology.ResourcePool {
			return machinev1.VSphereFailureDomain{
				Name: failureDomain.Name,
			}
		}
	}

	return machinev1.VSphereFailureDomain{}
}

// Config returns the stored VSphereMachineProviderSpec.
func (v VSphereProviderConfig) Config() machinev1beta1.VSphereMachineProviderSpec {
	return v.providerConfig
}

// newVSphereProviderConfig creates a VSphere type ProviderConfig from the raw extension.
// It should return an error if the provided RawExtension does not represent a VSphereProviderConfig.
func newVSphereProviderConfig(logger logr.Logger, raw *runtime.RawExtension, infrastructure *v1.Infrastructure) (ProviderConfig, error) {
	var vsphereMachineProviderSpec machinev1beta1.VSphereMachineProviderSpec

	if err := checkForUnknownFieldsInProviderSpecAndUnmarshal(logger, raw, &vsphereMachineProviderSpec); err != nil {
		return nil, fmt.Errorf("failed to check for unknown fields in the provider spec: %w", err)
	}

	// if multiple failure domains are defined, remove fields that are ambiguous in a multiple when
	// multiple failure domains are present.
	if len(infrastructure.Spec.PlatformSpec.VSphere.FailureDomains) > 1 {
		vsphereMachineProviderSpec.Template = ""
		vsphereMachineProviderSpec.Workspace = &machinev1beta1.Workspace{}
		vsphereMachineProviderSpec.Network = machinev1beta1.NetworkSpec{}
	}

	VSphereProviderConfig := VSphereProviderConfig{
		providerConfig: vsphereMachineProviderSpec,
		infrastructure: infrastructure,
	}

	config := providerConfig{
		platformType: v1.VSpherePlatformType,
		vsphere:      VSphereProviderConfig,
	}

	return config, nil
}
