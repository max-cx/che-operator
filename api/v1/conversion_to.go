//
// Copyright (c) 2019-2022 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation
//

package v1

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/devfile/devworkspace-operator/pkg/infrastructure"
	chev2 "github.com/eclipse-che/che-operator/api/v2"
	"github.com/eclipse-che/che-operator/pkg/common/constants"
	k8shelper "github.com/eclipse-che/che-operator/pkg/common/k8s-helper"
	defaults "github.com/eclipse-che/che-operator/pkg/common/operator-defaults"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *CheCluster) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*chev2.CheCluster)
	dst.ObjectMeta = src.ObjectMeta

	if err := src.convertTo_ServerComponents(dst); err != nil {
		return err
	}

	if err := src.convertTo_Ingress(dst); err != nil {
		return err
	}

	if err := src.convertTo_Workspaces(dst); err != nil {
		return err
	}

	if err := src.convertTo_ContainerRegistry(dst); err != nil {
		return err
	}

	if err := src.convertTo_DevWorkspace(dst); err != nil {
		return err
	}

	if err := src.convertTo_Status(dst); err != nil {
		return err
	}

	return nil
}

func (src *CheCluster) convertTo_Status(dst *chev2.CheCluster) error {
	dst.Status.PostgresVersion = src.Spec.Database.PostgresVersion
	dst.Status.CheURL = src.Status.CheURL
	dst.Status.CheVersion = src.Status.CheVersion
	dst.Status.DevfileRegistryURL = src.Status.DevfileRegistryURL
	dst.Status.PluginRegistryURL = src.Status.PluginRegistryURL
	dst.Status.Message = src.Status.Message
	dst.Status.Reason = src.Status.Reason
	dst.Status.GatewayPhase = chev2.GatewayPhase(src.Status.DevworkspaceStatus.GatewayPhase)
	dst.Status.WorkspaceBaseDomain = src.Status.DevworkspaceStatus.WorkspaceBaseDomain

	switch src.Status.CheClusterRunning {
	case "Available":
		dst.Status.ChePhase = chev2.ClusterPhaseActive
	case "Unavailable":
		dst.Status.ChePhase = chev2.ClusterPhaseInactive
	case "Available, Rolling Update in Progress":
		dst.Status.ChePhase = chev2.RollingUpdate
	}

	return nil
}

func (src *CheCluster) convertTo_DevWorkspace(dst *chev2.CheCluster) error {
	dst.Spec.DevWorkspace.Deployment = chev2.Deployment{
		Containers: []chev2.Container{
			{
				Name:  constants.DevWorkspaceController,
				Image: src.Spec.DevWorkspace.ControllerImage,
			},
		},
	}
	dst.Spec.DevWorkspace.RunningLimit = src.Spec.DevWorkspace.RunningLimit

	return nil
}

func (src *CheCluster) convertTo_ContainerRegistry(dst *chev2.CheCluster) error {
	dst.Spec.ContainerRegistry.Hostname = src.Spec.Server.AirGapContainerRegistryHostname
	dst.Spec.ContainerRegistry.Organization = src.Spec.Server.AirGapContainerRegistryOrganization

	return nil
}

func (src *CheCluster) convertTo_Workspaces(dst *chev2.CheCluster) error {
	if src.Spec.Server.GitSelfSignedCert {
		if src.Status.GitServerTLSCertificateConfigMapName != "" {
			dst.Spec.Workspaces.TrustedCerts.GitTrustedCertsConfigMapName = src.Status.GitServerTLSCertificateConfigMapName
		} else {
			dst.Spec.Workspaces.TrustedCerts.GitTrustedCertsConfigMapName = constants.DefaultGitSelfSignedCertsConfigMapName
		}
	}

	dst.Spec.Workspaces.DefaultNamespace = src.Spec.Server.WorkspaceNamespaceDefault
	dst.Spec.Workspaces.NodeSelector = newMap(src.Spec.Server.WorkspacePodNodeSelector)

	dst.Spec.Workspaces.Tolerations = []corev1.Toleration{}
	for _, v := range src.Spec.Server.WorkspacePodTolerations {
		dst.Spec.Workspaces.Tolerations = append(dst.Spec.Workspaces.Tolerations, v)
	}

	dst.Spec.Workspaces.DefaultPlugins = make([]chev2.WorkspaceDefaultPlugins, 0)
	for _, p := range src.Spec.Server.WorkspacesDefaultPlugins {
		dst.Spec.Workspaces.DefaultPlugins = append(dst.Spec.Workspaces.DefaultPlugins,
			chev2.WorkspaceDefaultPlugins{
				Editor:  p.Editor,
				Plugins: p.Plugins,
			})
	}

	if err := src.convertTo_Workspaces_Storage(dst); err != nil {
		return err
	}

	return nil
}

func (src *CheCluster) convertTo_Workspaces_Storage(dst *chev2.CheCluster) error {
	dst.Spec.Workspaces.Storage.PreCreateSubPaths = src.Spec.Storage.PreCreateSubPaths
	dst.Spec.Workspaces.Storage.Pvc = chev2.PVC{
		ClaimSize:    src.Spec.Storage.PvcClaimSize,
		StorageClass: src.Spec.Storage.WorkspacePVCStorageClassName,
	}
	dst.Spec.Workspaces.Storage.PvcJobsImage = src.Spec.Storage.PvcJobsImage
	dst.Spec.Workspaces.Storage.PvcStrategy = src.Spec.Storage.PvcStrategy

	return nil
}

func (src *CheCluster) convertTo_Ingress(dst *chev2.CheCluster) error {
	if infrastructure.IsOpenShift() {
		dst.Spec.Ingress = chev2.CheClusterSpecIngress{
			Labels:        parseMap(src.Spec.Server.CheServerRoute.Labels),
			Annotations:   newMap(src.Spec.Server.CheServerRoute.Annotations),
			Hostname:      src.Spec.Server.CheHost,
			Domain:        src.Spec.Server.CheServerRoute.Domain,
			TlsSecretName: src.Spec.Server.CheHostTLSSecret,
		}
	} else {
		dst.Spec.Ingress = chev2.CheClusterSpecIngress{
			Labels:   parseMap(src.Spec.Server.CheServerIngress.Labels),
			Domain:   src.Spec.K8s.IngressDomain,
			Hostname: src.Spec.Server.CheHost,
		}

		if src.Spec.Server.CheHostTLSSecret != "" {
			dst.Spec.Ingress.TlsSecretName = src.Spec.Server.CheHostTLSSecret
		} else {
			dst.Spec.Ingress.TlsSecretName = src.Spec.K8s.TlsSecretName
		}

		dst.Spec.Ingress.Annotations = make(map[string]string)
		if src.Spec.K8s.IngressClass != "" {
			dst.Spec.Ingress.Annotations["kubernetes.io/ingress.class"] = src.Spec.K8s.IngressClass
		}
		dst.Spec.Ingress.Annotations = labels.Merge(dst.Spec.Ingress.Annotations, src.Spec.Server.CheServerIngress.Annotations)
	}

	if err := src.convertTo_Ingress_Auth(dst); err != nil {
		return err
	}

	return nil
}

func (src *CheCluster) convertTo_Ingress_Auth(dst *chev2.CheCluster) error {
	dst.Spec.Ingress.Auth.IdentityProviderURL = src.Spec.Auth.IdentityProviderURL
	dst.Spec.Ingress.Auth.OAuthClientName = src.Spec.Auth.OAuthClientName
	dst.Spec.Ingress.Auth.OAuthSecret = src.Spec.Auth.OAuthSecret

	if err := src.convertTo_Ingress_Auth_Gateway(dst); err != nil {
		return err
	}

	return nil
}

func (src *CheCluster) convertTo_Ingress_Auth_Gateway(dst *chev2.CheCluster) error {
	dst.Spec.Ingress.Auth.Gateway.ConfigLabels = newMap(src.Spec.Server.SingleHostGatewayConfigMapLabels)

	dst.Spec.Ingress.Auth.Gateway.Deployment.Containers = []chev2.Container{}
	if src.Spec.Server.SingleHostGatewayImage != "" {
		dst.Spec.Ingress.Auth.Gateway.Deployment.Containers = append(
			dst.Spec.Ingress.Auth.Gateway.Deployment.Containers,
			chev2.Container{
				Name:  constants.GatewayContainerName,
				Image: src.Spec.Server.SingleHostGatewayImage,
			},
		)
	}

	if src.Spec.Server.SingleHostGatewayConfigSidecarImage != "" {
		dst.Spec.Ingress.Auth.Gateway.Deployment.Containers = append(
			dst.Spec.Ingress.Auth.Gateway.Deployment.Containers,
			chev2.Container{
				Name:  constants.GatewayConfigSideCarContainerName,
				Image: src.Spec.Server.SingleHostGatewayConfigSidecarImage,
			},
		)
	}

	if src.Spec.Auth.GatewayAuthenticationSidecarImage != "" {
		dst.Spec.Ingress.Auth.Gateway.Deployment.Containers = append(
			dst.Spec.Ingress.Auth.Gateway.Deployment.Containers,
			chev2.Container{
				Name:  constants.GatewayAuthenticationContainerName,
				Image: src.Spec.Auth.GatewayAuthenticationSidecarImage,
			},
		)
	}

	if src.Spec.Auth.GatewayAuthorizationSidecarImage != "" {
		dst.Spec.Ingress.Auth.Gateway.Deployment.Containers = append(
			dst.Spec.Ingress.Auth.Gateway.Deployment.Containers,
			chev2.Container{
				Name:  constants.GatewayAuthorizationContainerName,
				Image: src.Spec.Auth.GatewayAuthorizationSidecarImage,
			},
		)
	}

	return nil
}

func (src *CheCluster) convertTo_ServerComponents(dst *chev2.CheCluster) error {
	dst.Spec.ServerComponents.ExtraProperties = newMap(src.Spec.Server.CustomCheProperties)

	if err := src.convertTo_ServerComponents_Dashboard(dst); err != nil {
		return err
	}

	if err := src.convertTo_ServerComponents_DevfileRegistry(dst); err != nil {
		return err
	}

	if err := src.convertTo_ServerComponents_PluginRegistry(dst); err != nil {
		return err
	}

	if err := src.convertTo_ServerComponents_CheServer(dst); err != nil {
		return err
	}

	if err := src.convertTo_ServerComponents_Database(dst); err != nil {
		return err
	}

	if err := src.convertTo_ServerComponents_Metrics(dst); err != nil {
		return err
	}

	if err := src.convertTo_ServerComponents_ImagePuller(dst); err != nil {
		return err
	}

	return nil
}

func (src *CheCluster) convertTo_ServerComponents_ImagePuller(dst *chev2.CheCluster) error {
	dst.Spec.ServerComponents.ImagePuller.Enable = src.Spec.ImagePuller.Enable
	dst.Spec.ServerComponents.ImagePuller.Spec = src.Spec.ImagePuller.Spec
	return nil
}

func (src *CheCluster) convertTo_ServerComponents_Metrics(dst *chev2.CheCluster) error {
	dst.Spec.ServerComponents.Metrics.Enable = src.Spec.Metrics.Enable
	return nil
}

func (src *CheCluster) convertTo_ServerComponents_CheServer(dst *chev2.CheCluster) error {
	dst.Spec.ServerComponents.CheServer.LogLevel = src.Spec.Server.CheLogLevel
	dst.Spec.ServerComponents.CheServer.ClusterRoles = strings.Split(src.Spec.Server.CheClusterRoles, ",")

	if src.Spec.Server.CheDebug != "" {
		debug, err := strconv.ParseBool(src.Spec.Server.CheDebug)
		if err != nil {
			return err
		} else {
			dst.Spec.ServerComponents.CheServer.Debug = debug
		}
	}

	dst.Spec.ServerComponents.CheServer.Proxy = chev2.Proxy{
		Url:                   src.Spec.Server.ProxyURL,
		Port:                  src.Spec.Server.ProxyPort,
		NonProxyHosts:         strings.Split(src.Spec.Server.NonProxyHosts, "|"),
		CredentialsSecretName: src.Spec.Server.ProxySecret,
	}

	if src.Spec.Server.ProxySecret == "" && src.Spec.Server.ProxyUser != "" && src.Spec.Server.ProxyPassword != "" {
		if err := createCredentialsSecret(
			src.Spec.Server.ProxyUser,
			src.Spec.Server.ProxyPassword,
			constants.DefaultProxyCredentialsSecret,
			src.ObjectMeta.Namespace); err != nil {
			return err
		}

		dst.Spec.ServerComponents.CheServer.Proxy.CredentialsSecretName = constants.DefaultProxyCredentialsSecret
	}

	runAsUser, fsGroup, err := parseSecurityContext(src)
	if err != nil {
		return err
	}

	dst.Spec.ServerComponents.CheServer.Deployment = chev2.Deployment{
		Containers: []chev2.Container{
			{
				Name:            defaults.GetCheFlavor(),
				Image:           map[bool]string{true: src.Spec.Server.CheImage + ":" + src.Spec.Server.CheImageTag, false: ""}[src.Spec.Server.CheImage != ""],
				ImagePullPolicy: src.Spec.Server.CheImagePullPolicy,
				Resources: chev2.ResourceRequirements{
					Requests: chev2.ResourceList{
						Memory: src.Spec.Server.ServerMemoryRequest,
						Cpu:    src.Spec.Server.ServerCpuRequest,
					},
					Limits: chev2.ResourceList{
						Memory: src.Spec.Server.ServerMemoryLimit,
						Cpu:    src.Spec.Server.ServerCpuLimit,
					},
				},
			},
		},
		SecurityContext: chev2.PodSecurityContext{
			RunAsUser: runAsUser,
			FsGroup:   fsGroup,
		},
	}

	if src.Spec.Server.ServerTrustStoreConfigMapName != "" {
		if err := renameTrustStoreConfigMapToDefault(src.Spec.Server.ServerTrustStoreConfigMapName, src.Namespace); err != nil {
			return err
		}
	}

	return nil
}

func (src *CheCluster) convertTo_ServerComponents_PluginRegistry(dst *chev2.CheCluster) error {
	dst.Spec.ServerComponents.PluginRegistry.DisableInternalRegistry = src.Spec.Server.ExternalPluginRegistry

	if dst.Spec.ServerComponents.PluginRegistry.DisableInternalRegistry {
		dst.Spec.ServerComponents.PluginRegistry.ExternalPluginRegistries = []chev2.ExternalPluginRegistry{
			{
				Url: src.Spec.Server.PluginRegistryUrl,
			},
		}
	}

	dst.Spec.ServerComponents.PluginRegistry.Deployment = chev2.Deployment{
		Containers: []chev2.Container{
			{
				Name:            constants.PluginRegistryName,
				Image:           src.Spec.Server.PluginRegistryImage,
				ImagePullPolicy: corev1.PullPolicy(src.Spec.Server.PluginRegistryPullPolicy),
				Resources: chev2.ResourceRequirements{
					Requests: chev2.ResourceList{
						Memory: src.Spec.Server.PluginRegistryMemoryRequest,
						Cpu:    src.Spec.Server.PluginRegistryCpuRequest,
					},
					Limits: chev2.ResourceList{
						Memory: src.Spec.Server.PluginRegistryMemoryLimit,
						Cpu:    src.Spec.Server.PluginRegistryCpuLimit,
					},
				},
			},
		},
	}

	return nil
}

func (src *CheCluster) convertTo_ServerComponents_DevfileRegistry(dst *chev2.CheCluster) error {
	dst.Spec.ServerComponents.DevfileRegistry.DisableInternalRegistry = src.Spec.Server.ExternalDevfileRegistry

	dst.Spec.ServerComponents.DevfileRegistry.ExternalDevfileRegistries = []chev2.ExternalDevfileRegistry{}
	for _, r := range src.Spec.Server.ExternalDevfileRegistries {
		dst.Spec.ServerComponents.DevfileRegistry.ExternalDevfileRegistries = append(dst.Spec.ServerComponents.DevfileRegistry.ExternalDevfileRegistries,
			chev2.ExternalDevfileRegistry{
				Url: r.Url,
			})
	}

	dst.Spec.ServerComponents.DevfileRegistry.Deployment = chev2.Deployment{
		Containers: []chev2.Container{
			{
				Name:            constants.DevfileRegistryName,
				Image:           src.Spec.Server.DevfileRegistryImage,
				ImagePullPolicy: corev1.PullPolicy(src.Spec.Server.DevfileRegistryPullPolicy),
				Resources: chev2.ResourceRequirements{
					Requests: chev2.ResourceList{
						Memory: src.Spec.Server.DevfileRegistryMemoryRequest,
						Cpu:    src.Spec.Server.DevfileRegistryCpuRequest,
					},
					Limits: chev2.ResourceList{
						Memory: src.Spec.Server.DevfileRegistryMemoryLimit,
						Cpu:    src.Spec.Server.DevfileRegistryCpuLimit,
					},
				},
			},
		},
	}

	return nil
}

func (src *CheCluster) convertTo_ServerComponents_Database(dst *chev2.CheCluster) error {
	dst.Spec.ServerComponents.Database.CredentialsSecretName = src.Spec.Database.ChePostgresSecret

	if src.Spec.Database.ChePostgresSecret == "" && src.Spec.Database.ChePostgresUser != "" && src.Spec.Database.ChePostgresPassword != "" {
		if err := createCredentialsSecret(
			src.Spec.Database.ChePostgresUser,
			src.Spec.Database.ChePostgresPassword,
			constants.DefaultPostgresCredentialsSecret,
			src.ObjectMeta.Namespace); err != nil {
			return err
		}
		dst.Spec.ServerComponents.Database.CredentialsSecretName = constants.DefaultPostgresCredentialsSecret
	}

	dst.Spec.ServerComponents.Database.Deployment = chev2.Deployment{
		Containers: []chev2.Container{
			{
				Name:            constants.PostgresName,
				Image:           src.Spec.Database.PostgresImage,
				ImagePullPolicy: corev1.PullPolicy(src.Spec.Database.PostgresImagePullPolicy),
				Resources: chev2.ResourceRequirements{
					Requests: chev2.ResourceList{
						Memory: src.Spec.Database.ChePostgresContainerResources.Requests.Memory,
						Cpu:    src.Spec.Database.ChePostgresContainerResources.Requests.Cpu,
					},
					Limits: chev2.ResourceList{
						Memory: src.Spec.Database.ChePostgresContainerResources.Limits.Memory,
						Cpu:    src.Spec.Database.ChePostgresContainerResources.Limits.Cpu,
					},
				},
			},
		},
	}

	dst.Spec.ServerComponents.Database.ExternalDb = src.Spec.Database.ExternalDb
	dst.Spec.ServerComponents.Database.PostgresDb = src.Spec.Database.ChePostgresDb
	dst.Spec.ServerComponents.Database.PostgresHostName = src.Spec.Database.ChePostgresHostName
	dst.Spec.ServerComponents.Database.PostgresPort = src.Spec.Database.ChePostgresPort
	dst.Spec.ServerComponents.Database.Pvc = chev2.PVC{
		ClaimSize:    src.Spec.Database.PvcClaimSize,
		StorageClass: src.Spec.Storage.PostgresPVCStorageClassName,
	}

	return nil
}

func (src *CheCluster) convertTo_ServerComponents_Dashboard(dst *chev2.CheCluster) error {
	runAsUser, fsGroup, err := parseSecurityContext(src)
	if err != nil {
		return err
	}

	dst.Spec.ServerComponents.Dashboard.Deployment = chev2.Deployment{
		Containers: []chev2.Container{
			{
				Name:            defaults.GetCheFlavor() + "-dashboard",
				Image:           src.Spec.Server.DashboardImage,
				ImagePullPolicy: corev1.PullPolicy(src.Spec.Server.DashboardImagePullPolicy),
				Resources: chev2.ResourceRequirements{
					Requests: chev2.ResourceList{
						Memory: src.Spec.Server.DashboardMemoryRequest,
						Cpu:    src.Spec.Server.DashboardCpuRequest,
					},
					Limits: chev2.ResourceList{
						Memory: src.Spec.Server.DashboardMemoryLimit,
						Cpu:    src.Spec.Server.DashboardCpuLimit,
					},
				},
			},
		},
		SecurityContext: chev2.PodSecurityContext{
			RunAsUser: runAsUser,
			FsGroup:   fsGroup,
		},
	}
	dst.Spec.ServerComponents.Dashboard.Warning = src.Spec.Dashboard.Warning

	return nil
}

func parseSecurityContext(checluster *CheCluster) (*int64, *int64, error) {
	var runAsUser *int64 = nil
	if checluster.Spec.K8s.SecurityContextRunAsUser != "" {
		intValue, err := strconv.ParseInt(checluster.Spec.K8s.SecurityContextRunAsUser, 10, 64)
		if err != nil {
			return nil, nil, err
		}

		runAsUser = pointer.Int64Ptr(intValue)
	}

	var fsGroup *int64 = nil
	if checluster.Spec.K8s.SecurityContextFsGroup != "" {
		intValue, err := strconv.ParseInt(checluster.Spec.K8s.SecurityContextFsGroup, 10, 64)
		if err != nil {
			return nil, nil, err
		}
		fsGroup = pointer.Int64Ptr(intValue)
	}

	return runAsUser, fsGroup, nil
}

// Create a secret with a user's credentials
// Username and password are stored in `user` and `password` fields correspondingly.
func createCredentialsSecret(username string, password string, secretName string, namespace string) error {
	k8sHelper := k8shelper.New()

	_, err := k8sHelper.GetClientset().CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err == nil {
		// Credentials secret already exists, we can't proceed
		return fmt.Errorf("secret %s already exists", secretName)
	}

	secret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/part-of": "che.eclipse.org",
			},
		},
		Data: map[string][]byte{
			"user":     []byte(username),
			"password": []byte(password),
		},
	}

	if _, err := k8sHelper.GetClientset().CoreV1().Secrets(namespace).Create(context.TODO(), secret, metav1.CreateOptions{}); err != nil {
		return err
	}

	logger.Info("Credentials secret '" + secretName + "' with created.")
	return nil
}

// Convert `server.ServerTrustStoreConfigMapName` field from API V1 to API V2
// Since we API V2 does not have `server.ServerTrustStoreConfigMapName` field, we need to create
// the same ConfigMap but with a default name to be correctly handled by a controller.
func renameTrustStoreConfigMapToDefault(trustStoreConfigMapName string, namespace string) error {
	if trustStoreConfigMapName == constants.DefaultServerTrustStoreConfigMapName {
		// Already in default name
		return nil
	}

	k8sHelper := k8shelper.New()

	_, err := k8sHelper.GetClientset().CoreV1().ConfigMaps(namespace).Get(context.TODO(), constants.DefaultServerTrustStoreConfigMapName, metav1.GetOptions{})
	if err == nil {
		// ConfigMap with a default name already exists, we can't proceed
		return fmt.Errorf("TrustStore ConfigMap %s already exists", constants.DefaultServerTrustStoreConfigMapName)
	}

	existedTrustStoreConfigMap, err := k8sHelper.GetClientset().CoreV1().ConfigMaps(namespace).Get(context.TODO(), trustStoreConfigMapName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// ConfigMap not found, nothing to rename
			return nil
		}
		return err
	}

	// must have labels
	newTrustStoreConfigMapLabels := map[string]string{
		"app.kubernetes.io/part-of":   "che.eclipse.org",
		"app.kubernetes.io/component": "ca-bundle",
	}

	newTrustStoreConfigMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.DefaultServerTrustStoreConfigMapName,
			Namespace: namespace,
			Labels:    labels.Merge(newTrustStoreConfigMapLabels, existedTrustStoreConfigMap.Labels),
		},
		Data: existedTrustStoreConfigMap.Data,
	}

	// Create TrustStore ConfigMap with a default name
	if _, err = k8sHelper.GetClientset().CoreV1().ConfigMaps(namespace).Create(context.TODO(), newTrustStoreConfigMap, metav1.CreateOptions{}); err != nil {
		return err
	}

	// Delete legacy TrustStore ConfigMap
	if err = k8sHelper.GetClientset().CoreV1().ConfigMaps(namespace).Delete(context.TODO(), trustStoreConfigMapName, metav1.DeleteOptions{}); err != nil {
		return err
	}

	logger.Info("TrustStore ConfigMap '" + constants.DefaultServerTrustStoreConfigMapName + "' created.")
	return nil
}
