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

package helpers

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/cluster-control-plane-machine-set-operator/test/e2e/framework"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

// CheckInfrastructureHasFailureDomains checks infrastructure to see if failure domains are configured.
func CheckInfrastructureHasFailureDomains(testFramework framework.Framework, gomegaArgs ...interface{}) bool {
	infra := testFramework.NewEmptyInfrastructure()

	Eventually(komega.Get(infra), gomegaArgs...).Should(Succeed(), "cluster infrastructure should exist")

	platformType := testFramework.GetPlatformType()

	if platformType == configv1.VSpherePlatformType {
		if len(infra.Spec.PlatformSpec.VSphere.FailureDomains) > 0 {
			return true
		}
	}

	return false
}

// AddFailureDomain adds a failure domain to existing infrastructure.
func AddFailureDomain(testFramework framework.Framework, fd *configv1.VSpherePlatformFailureDomainSpec, index int) {
	k8sClient := testFramework.GetClient()
	ctx := testFramework.GetContext()

	platformType := testFramework.GetPlatformType()
	infra := testFramework.NewEmptyInfrastructure()

	if err := k8sClient.Get(ctx, testFramework.InfrastructureKey(), infra); err != nil && !apierrors.IsNotFound(err) {
		Fail(fmt.Sprintf("error when checking if an infrastructure exists: %v", err))
	} else if err == nil {
		switch platformType {
		case configv1.VSpherePlatformType:
			infra.Spec.PlatformSpec.VSphere.FailureDomains = insertFailureDomain(infra.Spec.PlatformSpec.VSphere.FailureDomains, index, fd)
			UpdateInfrastructure(testFramework, ctx, infra)
		default:
			Fail(fmt.Sprintf("adding failure domain unsupported for platform type %s", platformType))
		}
	}
}

// insertFailureDomain a utility method to insert a failure domain into a failure domain array.
func insertFailureDomain(a []configv1.VSpherePlatformFailureDomainSpec, index int, value *configv1.VSpherePlatformFailureDomainSpec) []configv1.VSpherePlatformFailureDomainSpec {
	if len(a) == index {
		return append(a, *value)
	}

	a = append(a[:index+1], a[index:]...)
	a[index] = *value

	return a
}

// RemoveFailureDomain removes the failure domain found at the specified index from the current infrastructure.
func RemoveFailureDomain(testFramework framework.Framework, index int) *configv1.VSpherePlatformFailureDomainSpec {
	var origFailureDomain *configv1.VSpherePlatformFailureDomainSpec

	k8sClient := testFramework.GetClient()
	ctx := testFramework.GetContext()

	platformType := testFramework.GetPlatformType()

	infra := testFramework.NewEmptyInfrastructure()
	if err := k8sClient.Get(ctx, testFramework.InfrastructureKey(), infra); err != nil && !apierrors.IsNotFound(err) {
		Fail(fmt.Sprintf("error when checking if an infrastructure exists: %v", err))
	} else if err == nil {
		// The infrastructure exists, so now remove the failure domain.
		switch platformType {
		case configv1.VSpherePlatformType:
			failureDomains := infra.Spec.PlatformSpec.VSphere.FailureDomains
			origFailureDomain = failureDomains[index].DeepCopy()
			failureDomains = append(failureDomains[:index], failureDomains[index+1:]...)
			infra.Spec.PlatformSpec.VSphere.FailureDomains = failureDomains
			UpdateInfrastructure(testFramework, ctx, infra)

			return origFailureDomain
		default:
			return nil
		}
	}

	Fail("manual support for the control plane machine set not yet implemented")

	return nil
}

// UpdateInfrastructure updates the infrastructure object.
func UpdateInfrastructure(testFramework framework.Framework, ctx context.Context, infra *configv1.Infrastructure) {
	By("Updating Infrastructure")

	k8sClient := testFramework.GetClient()
	Expect(k8sClient.Update(ctx, infra)).To(Succeed(), "infrastructure should have been updated")
}
