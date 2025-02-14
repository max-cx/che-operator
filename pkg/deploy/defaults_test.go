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
package deploy

import (
	"fmt"
	"os"
	"testing"

	"github.com/eclipse-che/che-operator/pkg/util"

	orgv1 "github.com/eclipse-che/che-operator/api/v1"
)

func TestDefaultFromEnv(t *testing.T) {

	cheVersionTest := os.Getenv("CHE_VERSION")

	cheServerImageTest := os.Getenv(util.GetArchitectureDependentEnv("RELATED_IMAGE_che_server"))
	dashboardImageTest := os.Getenv(util.GetArchitectureDependentEnv("RELATED_IMAGE_dashboard"))
	pluginRegistryImageTest := os.Getenv(util.GetArchitectureDependentEnv("RELATED_IMAGE_plugin_registry"))
	devfileRegistryImageTest := os.Getenv(util.GetArchitectureDependentEnv("RELATED_IMAGE_devfile_registry"))
	pvcJobsImageTest := os.Getenv(util.GetArchitectureDependentEnv("RELATED_IMAGE_pvc_jobs"))
	postgresImageTest := os.Getenv(util.GetArchitectureDependentEnv("RELATED_IMAGE_postgres"))
	postgres13ImageTest := os.Getenv(util.GetArchitectureDependentEnv("RELATED_IMAGE_postgres_13_3"))

	if DefaultCheVersion() != cheVersionTest {
		t.Errorf("Expected %s but was %s", cheVersionTest, DefaultCheVersion())
	}

	cheCluster := &orgv1.CheCluster{
		Spec: orgv1.CheClusterSpec{
			Server: orgv1.CheClusterSpecServer{},
		},
	}

	if DefaultCheServerImage(cheCluster) != cheServerImageTest {
		t.Errorf("Expected %s but was %s", cheServerImageTest, DefaultCheServerImage(cheCluster))
	}

	if DefaultDashboardImage(cheCluster) != dashboardImageTest {
		t.Errorf("Expected %s but was %s", dashboardImageTest, DefaultDashboardImage(cheCluster))
	}

	if DefaultPluginRegistryImage(cheCluster) != pluginRegistryImageTest {
		t.Errorf("Expected %s but was %s", pluginRegistryImageTest, DefaultPluginRegistryImage(cheCluster))
	}

	if DefaultDevfileRegistryImage(cheCluster) != devfileRegistryImageTest {
		t.Errorf("Expected %s but was %s", devfileRegistryImageTest, DefaultDevfileRegistryImage(cheCluster))
	}

	if DefaultPvcJobsImage(cheCluster) != pvcJobsImageTest {
		t.Errorf("Expected %s but was %s", pvcJobsImageTest, DefaultPvcJobsImage(cheCluster))
	}

	if DefaultPostgresImage(cheCluster) != postgresImageTest {
		t.Errorf("Expected %s but was %s", postgresImageTest, DefaultPostgresImage(cheCluster))
	}

	if DefaultPostgres13Image(cheCluster) != postgres13ImageTest {
		t.Errorf("Expected %s but was %s", postgres13ImageTest, DefaultPostgres13Image(cheCluster))
	}
}

func TestCorrectImageName(t *testing.T) {
	testCases := map[string]string{
		"docker.io/eclipse/che-operator:latest": "che-operator:latest",
		"eclipse/che-operator:7.1.0":            "che-operator:7.1.0",
		"che-operator:7.2.0":                    "che-operator:7.2.0",
	}
	for k, v := range testCases {
		t.Run(k, func(*testing.T) {
			actual := getImageNameFromFullImage(k)
			if actual != v {
				t.Errorf("Expected %s but was %s", v, actual)
			}
		})
	}
}

func TestCorrectAirGapPatchedImage(t *testing.T) {
	type testcase struct {
		image    string
		expected string
		cr       *orgv1.CheCluster
	}

	var (
		airGapRegistryHostname                                         = "myregistry.org"
		airGapRegistryOrganization                                     = "myorg"
		expectedAirGapPostgresUpstreamImage                            = makeAirGapImagePath(airGapRegistryHostname, airGapRegistryOrganization, getImageNameFromFullImage(defaultPostgresImage))
		expectedAirGapPostgresUpstreamImageOnlyOrgChanged              = makeAirGapImagePath(getHostnameFromImage(defaultPostgresImage), airGapRegistryOrganization, getImageNameFromFullImage(defaultPostgresImage))
		expectedAirGapDevspacesPluginRegistryOnlyOrgChanged            = makeAirGapImagePath(getHostnameFromImage(defaultPluginRegistryImage), airGapRegistryOrganization, getImageNameFromFullImage(defaultPluginRegistryImage))
		expectedAirGapDevspacesPostgresImage                           = makeAirGapImagePath(airGapRegistryHostname, airGapRegistryOrganization, getImageNameFromFullImage(defaultPostgresImage))
		expectedAirGapDevspacesDevfileRegistryImageOnlyHostnameChanged = makeAirGapImagePath(airGapRegistryHostname, getOrganizationFromImage(defaultDevfileRegistryImage), getImageNameFromFullImage(defaultDevfileRegistryImage))
	)

	upstream := &orgv1.CheCluster{
		Spec: orgv1.CheClusterSpec{
			Server: orgv1.CheClusterSpecServer{},
		},
	}
	devspaces := &orgv1.CheCluster{
		Spec: orgv1.CheClusterSpec{
			Server: orgv1.CheClusterSpecServer{
				CheFlavor: "devspaces",
			},
		},
	}
	airGapUpstream := &orgv1.CheCluster{
		Spec: orgv1.CheClusterSpec{
			Server: orgv1.CheClusterSpecServer{
				AirGapContainerRegistryHostname:     airGapRegistryHostname,
				AirGapContainerRegistryOrganization: airGapRegistryOrganization,
			},
		},
	}
	airGapDevspaces := &orgv1.CheCluster{
		Spec: orgv1.CheClusterSpec{
			Server: orgv1.CheClusterSpecServer{
				AirGapContainerRegistryHostname:     airGapRegistryHostname,
				AirGapContainerRegistryOrganization: airGapRegistryOrganization,
				CheFlavor:                           "devspaces",
			},
		},
	}
	upstreamOnlyOrg := &orgv1.CheCluster{
		Spec: orgv1.CheClusterSpec{
			Server: orgv1.CheClusterSpecServer{
				AirGapContainerRegistryOrganization: airGapRegistryOrganization,
			},
		},
	}
	devspacesOnlyOrg := &orgv1.CheCluster{
		Spec: orgv1.CheClusterSpec{
			Server: orgv1.CheClusterSpecServer{
				AirGapContainerRegistryOrganization: airGapRegistryOrganization,
				CheFlavor:                           "devspaces",
			},
		},
	}
	devspacesOnlyHostname := &orgv1.CheCluster{
		Spec: orgv1.CheClusterSpec{
			Server: orgv1.CheClusterSpecServer{
				AirGapContainerRegistryHostname: airGapRegistryHostname,
				CheFlavor:                       "devspaces",
			},
		},
	}

	testCases := map[string]testcase{
		"default postgres":          {image: defaultPostgresImage, expected: defaultPostgresImage, cr: upstream},
		"airgap postgres":           {image: defaultPostgresImage, expected: expectedAirGapPostgresUpstreamImage, cr: airGapUpstream},
		"with only the org changed": {image: defaultPostgresImage, expected: expectedAirGapPostgresUpstreamImageOnlyOrgChanged, cr: upstreamOnlyOrg},
		"devspaces plugin registry with only the org changed": {image: defaultPluginRegistryImage, expected: expectedAirGapDevspacesPluginRegistryOnlyOrgChanged, cr: devspacesOnlyOrg},
		"devspaces postgres":                          {image: defaultPostgresImage, expected: defaultPostgresImage, cr: devspaces},
		"devspaces airgap postgres":                   {image: defaultPostgresImage, expected: expectedAirGapDevspacesPostgresImage, cr: airGapDevspaces},
		"devspaces airgap with only hostname defined": {image: defaultDevfileRegistryImage, expected: expectedAirGapDevspacesDevfileRegistryImageOnlyHostnameChanged, cr: devspacesOnlyHostname},
	}
	for name, tc := range testCases {
		t.Run(name, func(*testing.T) {
			actual := patchDefaultImageName(tc.cr, tc.image)
			if actual != tc.expected {
				t.Errorf("Expected %s but was %s", tc.expected, actual)
			}
		})
	}
}

func makeAirGapImagePath(hostname, org, nameAndTag string) string {
	return fmt.Sprintf("%s/%s/%s", hostname, org, nameAndTag)
}
