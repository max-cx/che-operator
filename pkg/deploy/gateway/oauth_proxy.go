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

package gateway

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/devfile/devworkspace-operator/pkg/infrastructure"
	chev2 "github.com/eclipse-che/che-operator/api/v2"
	"github.com/eclipse-che/che-operator/pkg/common/chetypes"
	"github.com/eclipse-che/che-operator/pkg/common/constants"
	defaults "github.com/eclipse-che/che-operator/pkg/common/operator-defaults"
	"github.com/eclipse-che/che-operator/pkg/deploy"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getGatewayOauthProxyConfigSpec(ctx *chetypes.DeployContext, cookieSecret string) corev1.ConfigMap {
	instance := ctx.CheCluster

	var config string
	if infrastructure.IsOpenShift() {
		config = openshiftOauthProxyConfig(ctx, cookieSecret)
	} else {
		config = kubernetesOauthProxyconfig(ctx, cookieSecret)
	}
	return corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "che-gateway-config-oauth-proxy",
			Namespace: instance.Namespace,
			Labels:    deploy.GetLabels(GatewayServiceName),
		},
		Data: map[string]string{
			"oauth-proxy.cfg": config,
		},
	}
}

func openshiftOauthProxyConfig(ctx *chetypes.DeployContext, cookieSecret string) string {
	return fmt.Sprintf(`
http_address = ":%d"
https_address = ""
provider = "openshift"
redirect_url = "https://%s/oauth/callback"
upstreams = [
	"http://127.0.0.1:8081/"
]
client_id = "%s"
client_secret = "%s"
scope = "user:full"
openshift_service_account = "%s"
cookie_secret = "%s"
cookie_expire = "24h0m0s"
email_domains = "*"
cookie_httponly = false
pass_access_token = true
skip_provider_button = true
%s
`, GatewayServicePort,
		ctx.CheCluster.GetCheHost(),
		ctx.CheCluster.Spec.Ingress.Auth.OAuthClientName,
		ctx.CheCluster.Spec.Ingress.Auth.OAuthSecret,
		GatewayServiceName,
		cookieSecret,
		skipAuthConfig(ctx.CheCluster))
}

func kubernetesOauthProxyconfig(ctx *chetypes.DeployContext, cookieSecret string) string {
	return fmt.Sprintf(`
proxy_prefix = "/oauth"
http_address = ":%d"
https_address = ""
provider = "oidc"
redirect_url = "https://%s/oauth/callback"
oidc_issuer_url = "%s"
insecure_oidc_skip_issuer_verification = true
ssl_insecure_skip_verify = true
upstreams = [
	"http://127.0.0.1:8081/"
]
client_id = "%s"
client_secret = "%s"
cookie_secret = "%s"
cookie_expire = "24h0m0s"
email_domains = "*"
cookie_httponly = false
pass_authorization_header = true
skip_provider_button = true
%s
`, GatewayServicePort,
		ctx.CheCluster.GetCheHost(),
		ctx.CheCluster.Spec.Ingress.Auth.IdentityProviderURL,
		ctx.CheCluster.Spec.Ingress.Auth.OAuthClientName,
		ctx.CheCluster.Spec.Ingress.Auth.OAuthSecret,
		cookieSecret,
		skipAuthConfig(ctx.CheCluster))
}

func skipAuthConfig(instance *chev2.CheCluster) string {
	var skipAuthPaths []string
	if !instance.Spec.ServerComponents.PluginRegistry.DisableInternalRegistry {
		skipAuthPaths = append(skipAuthPaths, "^/"+constants.PluginRegistryName)
	}
	if !instance.Spec.ServerComponents.DevfileRegistry.DisableInternalRegistry {
		skipAuthPaths = append(skipAuthPaths, "^/"+constants.DevfileRegistryName)
	}
	skipAuthPaths = append(skipAuthPaths, "/healthz$")
	if len(skipAuthPaths) > 0 {
		propName := "skip_auth_routes"
		if infrastructure.IsOpenShift() {
			propName = "skip_auth_regex"
		}
		return fmt.Sprintf("%s = \"%s\"", propName, strings.Join(skipAuthPaths, "|"))
	}
	return ""
}

func getOauthProxyContainerSpec(ctx *chetypes.DeployContext) corev1.Container {
	return corev1.Container{
		Name:            "oauth-proxy",
		Image:           defaults.GetGatewayAuthenticationSidecarImage(ctx.CheCluster),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Args: []string{
			"--config=/etc/oauth-proxy/oauth-proxy.cfg",
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "oauth-proxy-config",
				MountPath: "/etc/oauth-proxy",
			},
		},
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("512Mi"),
				corev1.ResourceCPU:    resource.MustParse("0.5"),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("64Mi"),
				corev1.ResourceCPU:    resource.MustParse("0.1"),
			},
		},
		Ports: []corev1.ContainerPort{
			{ContainerPort: GatewayServicePort, Protocol: "TCP"},
		},
		Env: []corev1.EnvVar{
			{
				Name:  "http_proxy",
				Value: ctx.Proxy.HttpProxy,
			},
			{
				Name:  "https_proxy",
				Value: ctx.Proxy.HttpsProxy,
			},
			{
				Name:  "no_proxy",
				Value: ctx.Proxy.NoProxy,
			},
		},
	}
}

func getOauthProxyConfigVolume() corev1.Volume {
	return corev1.Volume{
		Name: "oauth-proxy-config",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "che-gateway-config-oauth-proxy",
				},
			},
		},
	}
}
