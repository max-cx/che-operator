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
package server

import (
	chev2 "github.com/eclipse-che/che-operator/api/v2"
	"github.com/eclipse-che/che-operator/pkg/common/chetypes"
	defaults "github.com/eclipse-che/che-operator/pkg/common/operator-defaults"
	"github.com/eclipse-che/che-operator/pkg/deploy"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type CheServerReconciler struct {
	deploy.Reconcilable
}

func NewCheServerReconciler() *CheServerReconciler {
	return &CheServerReconciler{}
}

func (s *CheServerReconciler) Reconcile(ctx *chetypes.DeployContext) (reconcile.Result, bool, error) {
	done, err := s.syncCheConfigMap(ctx)
	if !done {
		return reconcile.Result{}, false, err
	}

	// ensure configmap is created
	// the version of the object is used in the deployment
	exists, err := deploy.GetNamespacedObject(ctx, CheConfigMapName, &corev1.ConfigMap{})
	if !exists {
		return reconcile.Result{}, false, err
	}

	done, err = s.syncDeployment(ctx)
	if !done {
		return reconcile.Result{}, false, err
	}

	done, err = s.syncChePhase(ctx)
	if !done {
		return reconcile.Result{}, false, err
	}

	done, err = s.updateCheVersion(ctx)
	if !done {
		return reconcile.Result{}, false, err
	}

	return reconcile.Result{}, true, nil
}

func (s *CheServerReconciler) Finalize(ctx *chetypes.DeployContext) bool {
	return true
}

func (s *CheServerReconciler) syncCheConfigMap(ctx *chetypes.DeployContext) (bool, error) {
	data, err := s.getCheConfigMapData(ctx)
	if err != nil {
		return false, err
	}

	return deploy.SyncConfigMapDataToCluster(ctx, CheConfigMapName, data, getComponentName(ctx))
}

func (s *CheServerReconciler) syncChePhase(ctx *chetypes.DeployContext) (bool, error) {
	cheDeployment := &appsv1.Deployment{}
	exists, err := deploy.GetNamespacedObject(ctx, getComponentName(ctx), cheDeployment)
	if err != nil {
		return false, err
	}

	if exists {
		if cheDeployment.Status.AvailableReplicas < 1 {
			if ctx.CheCluster.Status.ChePhase != chev2.ClusterPhaseInactive {
				ctx.CheCluster.Status.ChePhase = chev2.ClusterPhaseInactive
				err := deploy.UpdateCheCRStatus(ctx, "Phase", chev2.ClusterPhaseInactive)
				return err == nil, err
			}
		} else if cheDeployment.Status.Replicas != 1 {
			if ctx.CheCluster.Status.ChePhase != chev2.RollingUpdate {
				ctx.CheCluster.Status.ChePhase = chev2.RollingUpdate
				err := deploy.UpdateCheCRStatus(ctx, "Phase", chev2.RollingUpdate)
				return err == nil, err
			}
		} else {
			if ctx.CheCluster.Status.ChePhase != chev2.ClusterPhaseActive {
				cheFlavor := defaults.GetCheFlavor()
				name := "Eclipse Che"
				if cheFlavor == "devspaces" {
					name = "Red Hat OpenShift Dev Spaces"
				}

				logrus.Infof(name+" is now available at: %s", ctx.CheCluster.Status.CheURL)
				ctx.CheCluster.Status.ChePhase = chev2.ClusterPhaseActive
				err := deploy.UpdateCheCRStatus(ctx, "Phase", chev2.ClusterPhaseActive)
				return err == nil, err
			}
		}
	} else {
		ctx.CheCluster.Status.ChePhase = chev2.ClusterPhaseInactive
		err := deploy.UpdateCheCRStatus(ctx, "Phase", chev2.ClusterPhaseInactive)
		return err == nil, err
	}

	return true, nil
}

func (s *CheServerReconciler) syncDeployment(ctx *chetypes.DeployContext) (bool, error) {
	spec, err := s.getDeploymentSpec(ctx)
	if err != nil {
		return false, err
	}

	return deploy.SyncDeploymentSpecToCluster(ctx, spec, deploy.DefaultDeploymentDiffOpts)
}

func (s CheServerReconciler) updateCheVersion(ctx *chetypes.DeployContext) (bool, error) {
	cheVersion := defaults.GetCheVersion()
	if ctx.CheCluster.Status.CheVersion != cheVersion {
		ctx.CheCluster.Status.CheVersion = cheVersion
		err := deploy.UpdateCheCRStatus(ctx, "version", cheVersion)
		return err == nil, err
	}
	return true, nil
}
