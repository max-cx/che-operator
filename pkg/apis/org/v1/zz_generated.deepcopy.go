//
// Copyright (c) 2012-2019 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation
//
// +build !ignore_autogenerated

/*
Copyright The Kubernetes Authors.

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

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheCluster) DeepCopyInto(out *CheCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheCluster.
func (in *CheCluster) DeepCopy() *CheCluster {
	if in == nil {
		return nil
	}
	out := new(CheCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CheCluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheClusterList) DeepCopyInto(out *CheClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CheCluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheClusterList.
func (in *CheClusterList) DeepCopy() *CheClusterList {
	if in == nil {
		return nil
	}
	out := new(CheClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CheClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheClusterSpec) DeepCopyInto(out *CheClusterSpec) {
	*out = *in
	out.Server = in.Server
	out.Database = in.Database
	out.Auth = in.Auth
	out.Storage = in.Storage
	out.K8SOnly = in.K8SOnly
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheClusterSpec.
func (in *CheClusterSpec) DeepCopy() *CheClusterSpec {
	if in == nil {
		return nil
	}
	out := new(CheClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheClusterSpecAuth) DeepCopyInto(out *CheClusterSpecAuth) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheClusterSpecAuth.
func (in *CheClusterSpecAuth) DeepCopy() *CheClusterSpecAuth {
	if in == nil {
		return nil
	}
	out := new(CheClusterSpecAuth)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheClusterSpecDB) DeepCopyInto(out *CheClusterSpecDB) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheClusterSpecDB.
func (in *CheClusterSpecDB) DeepCopy() *CheClusterSpecDB {
	if in == nil {
		return nil
	}
	out := new(CheClusterSpecDB)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheClusterSpecK8SOnly) DeepCopyInto(out *CheClusterSpecK8SOnly) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheClusterSpecK8SOnly.
func (in *CheClusterSpecK8SOnly) DeepCopy() *CheClusterSpecK8SOnly {
	if in == nil {
		return nil
	}
	out := new(CheClusterSpecK8SOnly)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheClusterSpecServer) DeepCopyInto(out *CheClusterSpecServer) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheClusterSpecServer.
func (in *CheClusterSpecServer) DeepCopy() *CheClusterSpecServer {
	if in == nil {
		return nil
	}
	out := new(CheClusterSpecServer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheClusterSpecStorage) DeepCopyInto(out *CheClusterSpecStorage) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheClusterSpecStorage.
func (in *CheClusterSpecStorage) DeepCopy() *CheClusterSpecStorage {
	if in == nil {
		return nil
	}
	out := new(CheClusterSpecStorage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheClusterStatus) DeepCopyInto(out *CheClusterStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheClusterStatus.
func (in *CheClusterStatus) DeepCopy() *CheClusterStatus {
	if in == nil {
		return nil
	}
	out := new(CheClusterStatus)
	in.DeepCopyInto(out)
	return out
}
