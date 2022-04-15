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

package che

import (
	"fmt"
	"strings"

	"github.com/devfile/devworkspace-operator/pkg/infrastructure"
	"github.com/eclipse-che/che-operator/pkg/common/chetypes"
	"github.com/eclipse-che/che-operator/pkg/deploy"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	NamespaceStrategyErrorMessage  = "Namespace strategies other than 'per user' is not supported anymore. Using the <username> or <userid> placeholder is required in the 'spec.server.workspaceNamespaceDefault' field. The current value is: %s"
	AuthenticationModeErrorMessage = "Single user authentication mode is not supported anymore. To backup your data you can commit workspace configuration to an SCM server and use factories to restore it in multi user mode. To switch to multi user authentication mode set 'spec.serverComponents.extraProperties.CHE_MULTIUSER' to 'true' in %s CheCluster custom resource. Switching to multi user authentication mode without backing up your data will cause data loss."
)

// CheClusterValidator checks CheCluster CR configuration.
// It detect:
// - configurations which miss required field(s) to deploy Che
// - self-contradictory configurations
// - configurations with which it is impossible to deploy Che
type CheClusterValidator struct {
	deploy.Reconcilable
}

func NewCheClusterValidator() *CheClusterValidator {
	return &CheClusterValidator{}
}

func (v *CheClusterValidator) Reconcile(ctx *chetypes.DeployContext) (reconcile.Result, bool, error) {
	if !infrastructure.IsOpenShift() {
		if ctx.CheCluster.Spec.Ingress.Domain == "" {
			return reconcile.Result{}, false, fmt.Errorf("Required field \"spec.auth.gateway.ingress.domain\" is not set")
		}
	}

	workspaceNamespaceDefault := ctx.CheCluster.GetDefaultNamespache()
	if strings.Index(workspaceNamespaceDefault, "<username>") == -1 && strings.Index(workspaceNamespaceDefault, "<userid>") == -1 {
		return reconcile.Result{}, false, fmt.Errorf(NamespaceStrategyErrorMessage, workspaceNamespaceDefault)
	}

	if !ctx.CheCluster.IsMultiUser() {
		return reconcile.Result{}, false, fmt.Errorf(AuthenticationModeErrorMessage, ctx.CheCluster.Name)
	}

	return reconcile.Result{}, true, nil
}

func (v *CheClusterValidator) Finalize(ctx *chetypes.DeployContext) bool {
	return true
}
