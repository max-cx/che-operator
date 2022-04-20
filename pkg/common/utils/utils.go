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
package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/yaml"
)

func Contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func Remove(list []string, s string) []string {
	for i, v := range list {
		if v == s {
			list = append(list[:i], list[i+1:]...)
		}
	}
	return list
}

func GeneratePassword(stringLength int) (passwd string) {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")
	length := stringLength
	buf := make([]rune, length)
	for i := range buf {
		buf[i] = chars[rand.Intn(len(chars))]
	}
	passwd = string(buf)
	return passwd
}

func IsK8SResourceServed(discoveryClient discovery.DiscoveryInterface, resourceName string) bool {
	_, resourceList, err := discoveryClient.ServerGroupsAndResources()
	if err != nil {
		return false
	}

	return hasAPIResourceNameInList(resourceName, resourceList)
}

func hasAPIResourceNameInList(name string, resources []*metav1.APIResourceList) bool {
	for _, l := range resources {
		for _, r := range l.APIResources {
			if r.Name == name {
				return true
			}
		}
	}

	return false
}

func GetValue(value string, defaultValue string) string {
	if value == "" {
		value = defaultValue
	}
	return value
}

func GetEnv(envs []corev1.EnvVar, name string) string {
	for _, env := range envs {
		if env.Name == name {
			return env.Value
		}
	}

	return ""
}

func GetEnvByRegExp(regExp string) []corev1.EnvVar {
	var env []corev1.EnvVar
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		envName := pair[0]
		rxp := regexp.MustCompile(regExp)
		if rxp.MatchString(envName) {
			envName = GetArchitectureDependentEnv(envName)
			env = append(env, corev1.EnvVar{Name: envName, Value: pair[1]})
		}
	}
	return env
}

// GetArchitectureDependentEnv returns environment variable dependending on architecture
// by adding "_<ARCHITECTURE>" suffix. If variable is not set then the default will be return.
func GetArchitectureDependentEnv(env string) string {
	archEnv := env + "_" + runtime.GOARCH
	if _, ok := os.LookupEnv(archEnv); ok {
		return archEnv
	}

	return env
}

// GetImageNameAndTag returns the image repository and tag name from the provided image
// Referenced from https://github.com/che-incubator/chectl/blob/main/src/util.ts
func GetImageNameAndTag(image string) (string, string) {
	var imageName, imageTag string
	if strings.Contains(image, "@") {
		// Image is referenced via a digest
		index := strings.Index(image, "@")
		imageName = image[:index]
		imageTag = image[index+1:]
	} else {
		// Image is referenced via a tag
		lastColonIndex := strings.LastIndex(image, ":")
		if lastColonIndex == -1 {
			imageName = image
			imageTag = "latest"
		} else {
			beforeLastColon := image[:lastColonIndex]
			afterLastColon := image[lastColonIndex+1:]
			if strings.Contains(afterLastColon, "/") {
				// The colon is for registry port and not for a tag
				imageName = image
				imageTag = "latest"
			} else {
				// The colon separates image name from the tag
				imageName = beforeLastColon
				imageTag = afterLastColon
			}
		}
	}
	return imageName, imageTag
}

func ReadObjectInto(yamlFile string, obj interface{}) error {
	data, err := ioutil.ReadFile(yamlFile)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, obj)
	if err != nil {
		return err
	}

	return nil
}

func ComputeHash256(data []byte) string {
	hasher := sha256.New()
	hasher.Write(data)
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

func GetPullPolicyFromDockerImage(dockerImage string) string {
	tag := ""
	parts := strings.Split(dockerImage, ":")
	if len(parts) > 1 {
		tag = parts[1]
	}
	if tag == "latest" || tag == "nightly" || tag == "next" {
		return "Always"
	}
	return "IfNotPresent"
}

func GetMap(value map[string]string, defaultValue map[string]string) map[string]string {
	ret := value
	if len(value) < 1 {
		ret = defaultValue
	}

	return ret
}

func InNamespaceEventFilter(namespace string) predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(ce event.CreateEvent) bool {
			return namespace == ce.Object.GetNamespace()
		},
		DeleteFunc: func(de event.DeleteEvent) bool {
			return namespace == de.Object.GetNamespace()
		},
		UpdateFunc: func(ue event.UpdateEvent) bool {
			return namespace == ue.ObjectOld.GetNamespace()
		},
		GenericFunc: func(ge event.GenericEvent) bool {
			return namespace == ge.Object.GetNamespace()
		},
	}
}
