// +build !ignore_autogenerated

/*


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

// Code generated by controller-gen. DO NOT EDIT.

package v1beta1

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FieldMapping) DeepCopyInto(out *FieldMapping) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FieldMapping.
func (in *FieldMapping) DeepCopy() *FieldMapping {
	if in == nil {
		return nil
	}
	out := new(FieldMapping)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VaultAuthSpec) DeepCopyInto(out *VaultAuthSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VaultAuthSpec.
func (in *VaultAuthSpec) DeepCopy() *VaultAuthSpec {
	if in == nil {
		return nil
	}
	out := new(VaultAuthSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VaultBinding) DeepCopyInto(out *VaultBinding) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VaultBinding.
func (in *VaultBinding) DeepCopy() *VaultBinding {
	if in == nil {
		return nil
	}
	out := new(VaultBinding)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VaultBinding) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VaultBindingList) DeepCopyInto(out *VaultBindingList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]VaultBinding, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VaultBindingList.
func (in *VaultBindingList) DeepCopy() *VaultBindingList {
	if in == nil {
		return nil
	}
	out := new(VaultBindingList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VaultBindingList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VaultBindingSpec) DeepCopyInto(out *VaultBindingSpec) {
	*out = *in
	if in.VaultSpec != nil {
		in, out := &in.VaultSpec, &out.VaultSpec
		*out = new(VaultSpec)
		**out = **in
	}
	if in.Fields != nil {
		in, out := &in.Fields, &out.Fields
		*out = make([]FieldMapping, len(*in))
		copy(*out, *in)
	}
	if in.Secret != nil {
		in, out := &in.Secret, &out.Secret
		*out = new(v1.SecretReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VaultBindingSpec.
func (in *VaultBindingSpec) DeepCopy() *VaultBindingSpec {
	if in == nil {
		return nil
	}
	out := new(VaultBindingSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VaultBindingStatus) DeepCopyInto(out *VaultBindingStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	out.Vault = in.Vault
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VaultBindingStatus.
func (in *VaultBindingStatus) DeepCopy() *VaultBindingStatus {
	if in == nil {
		return nil
	}
	out := new(VaultBindingStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VaultBindingVaultStatus) DeepCopyInto(out *VaultBindingVaultStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VaultBindingVaultStatus.
func (in *VaultBindingVaultStatus) DeepCopy() *VaultBindingVaultStatus {
	if in == nil {
		return nil
	}
	out := new(VaultBindingVaultStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VaultMirror) DeepCopyInto(out *VaultMirror) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VaultMirror.
func (in *VaultMirror) DeepCopy() *VaultMirror {
	if in == nil {
		return nil
	}
	out := new(VaultMirror)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VaultMirror) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VaultMirrorList) DeepCopyInto(out *VaultMirrorList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]VaultMirror, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VaultMirrorList.
func (in *VaultMirrorList) DeepCopy() *VaultMirrorList {
	if in == nil {
		return nil
	}
	out := new(VaultMirrorList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VaultMirrorList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VaultMirrorSpec) DeepCopyInto(out *VaultMirrorSpec) {
	*out = *in
	if in.Source != nil {
		in, out := &in.Source, &out.Source
		*out = new(VaultSpec)
		**out = **in
	}
	if in.Destination != nil {
		in, out := &in.Destination, &out.Destination
		*out = new(VaultSpec)
		**out = **in
	}
	if in.Interval != nil {
		in, out := &in.Interval, &out.Interval
		*out = new(metav1.Duration)
		**out = **in
	}
	if in.Fields != nil {
		in, out := &in.Fields, &out.Fields
		*out = make([]FieldMapping, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VaultMirrorSpec.
func (in *VaultMirrorSpec) DeepCopy() *VaultMirrorSpec {
	if in == nil {
		return nil
	}
	out := new(VaultMirrorSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VaultMirrorStatus) DeepCopyInto(out *VaultMirrorStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	out.Vault = in.Vault
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VaultMirrorStatus.
func (in *VaultMirrorStatus) DeepCopy() *VaultMirrorStatus {
	if in == nil {
		return nil
	}
	out := new(VaultMirrorStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VaultMirrorVaultStatus) DeepCopyInto(out *VaultMirrorVaultStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VaultMirrorVaultStatus.
func (in *VaultMirrorVaultStatus) DeepCopy() *VaultMirrorVaultStatus {
	if in == nil {
		return nil
	}
	out := new(VaultMirrorVaultStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VaultSpec) DeepCopyInto(out *VaultSpec) {
	*out = *in
	out.TLSConfig = in.TLSConfig
	out.Auth = in.Auth
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VaultSpec.
func (in *VaultSpec) DeepCopy() *VaultSpec {
	if in == nil {
		return nil
	}
	out := new(VaultSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VaultTLSSpec) DeepCopyInto(out *VaultTLSSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VaultTLSSpec.
func (in *VaultTLSSpec) DeepCopy() *VaultTLSSpec {
	if in == nil {
		return nil
	}
	out := new(VaultTLSSpec)
	in.DeepCopyInto(out)
	return out
}