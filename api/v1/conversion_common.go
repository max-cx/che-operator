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
	"strings"

	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	logger = ctrl.Log.WithName("conversion")
)

func parseMap(src string) map[string]string {
	m := map[string]string{}

	for _, item := range strings.Split(src, ",") {
		keyValuePair := strings.Split(item, "=")
		if len(keyValuePair) == 1 {
			continue
		}

		key := strings.TrimSpace(keyValuePair[0])
		value := strings.TrimSpace(keyValuePair[1])
		if key != "" && value != "" {
			m[keyValuePair[0]] = keyValuePair[1]
		}
	}

	return m
}

func newMap(m map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		result[k] = v
	}
	return result
}
