// +build !ignore_autogenerated

/*
Copyright 2020.

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

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AttackSpec) DeepCopyInto(out *AttackSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AttackSpec.
func (in *AttackSpec) DeepCopy() *AttackSpec {
	if in == nil {
		return nil
	}
	out := new(AttackSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReportSpec) DeepCopyInto(out *ReportSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReportSpec.
func (in *ReportSpec) DeepCopy() *ReportSpec {
	if in == nil {
		return nil
	}
	out := new(ReportSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Vegeta) DeepCopyInto(out *Vegeta) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Vegeta.
func (in *Vegeta) DeepCopy() *Vegeta {
	if in == nil {
		return nil
	}
	out := new(Vegeta)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Vegeta) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VegetaList) DeepCopyInto(out *VegetaList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Vegeta, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VegetaList.
func (in *VegetaList) DeepCopy() *VegetaList {
	if in == nil {
		return nil
	}
	out := new(VegetaList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VegetaList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VegetaSpec) DeepCopyInto(out *VegetaSpec) {
	*out = *in
	if in.Attack != nil {
		in, out := &in.Attack, &out.Attack
		*out = new(AttackSpec)
		**out = **in
	}
	if in.Report != nil {
		in, out := &in.Report, &out.Report
		*out = new(ReportSpec)
		**out = **in
	}
	in.Resources.DeepCopyInto(&out.Resources)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VegetaSpec.
func (in *VegetaSpec) DeepCopy() *VegetaSpec {
	if in == nil {
		return nil
	}
	out := new(VegetaSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VegetaStatus) DeepCopyInto(out *VegetaStatus) {
	*out = *in
	if in.Active != nil {
		in, out := &in.Active, &out.Active
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Failed != nil {
		in, out := &in.Failed, &out.Failed
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Succeeded != nil {
		in, out := &in.Succeeded, &out.Succeeded
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VegetaStatus.
func (in *VegetaStatus) DeepCopy() *VegetaStatus {
	if in == nil {
		return nil
	}
	out := new(VegetaStatus)
	in.DeepCopyInto(out)
	return out
}
