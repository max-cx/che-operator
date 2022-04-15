//
// Copyright (c) 2019-2021 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation
//
package devfileregistry

import (
	"fmt"

	"github.com/eclipse-che/che-operator/pkg/common/chetypes"
	"github.com/eclipse-che/che-operator/pkg/common/constants"
	defaults "github.com/eclipse-che/che-operator/pkg/common/operator-defaults"
	"github.com/eclipse-che/che-operator/pkg/common/utils"
	"github.com/eclipse-che/che-operator/pkg/deploy"
	"github.com/eclipse-che/che-operator/pkg/deploy/registry"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func (d *DevfileRegistryReconciler) getDevfileRegistryDeploymentSpec(ctx *chetypes.DeployContext) *appsv1.Deployment {
	registryType := "devfile"
	registryImage := defaults.GetDevfileRegistryImage(ctx.CheCluster)
	registryImagePullPolicy := v1.PullPolicy(utils.GetPullPolicyFromDockerImage(registryImage))
	probePath := "/devfiles/"
	devfileImagesEnv := utils.GetEnvByRegExp("^.*devfile_registry_image.*$")

	// If there is a devfile registry deployed by operator
	if !ctx.CheCluster.Spec.ServerComponents.DevfileRegistry.DisableInternalRegistry {
		devfileImagesEnv = append(devfileImagesEnv,
			corev1.EnvVar{
				Name:  "CHE_DEVFILE_REGISTRY_INTERNAL_URL",
				Value: fmt.Sprintf("http://%s.%s.svc:8080", constants.DevfileRegistryName, ctx.CheCluster.Namespace)},
		)
	}

	resources := v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse(constants.DefaultDevfileRegistryMemoryRequest),
			v1.ResourceCPU:    resource.MustParse(constants.DefaultDevfileRegistryCpuRequest),
		},
		Limits: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse(constants.DefaultDevfileRegistryMemoryLimit),
			v1.ResourceCPU:    resource.MustParse(constants.DefaultDevfileRegistryCpuLimit),
		},
	}

	deployment := registry.GetSpecRegistryDeployment(
		ctx,
		registryType,
		registryImage,
		devfileImagesEnv,
		registryImagePullPolicy,
		resources,
		probePath)
	deploy.CustomizeDeployment(deployment, &ctx.CheCluster.Spec.ServerComponents.DevfileRegistry.Deployment, false)
	return deployment
}
