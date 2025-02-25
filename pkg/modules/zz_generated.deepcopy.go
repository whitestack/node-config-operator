//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2023.

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

package modules

import ()

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AptPackages) DeepCopyInto(out *AptPackages) {
	*out = *in
	if in.Packages != nil {
		in, out := &in.Packages, &out.Packages
		*out = make([]AptPackage, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AptPackages.
func (in *AptPackages) DeepCopy() *AptPackages {
	if in == nil {
		return nil
	}
	out := new(AptPackages)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BlockInFiles) DeepCopyInto(out *BlockInFiles) {
	*out = *in
	if in.Blocks != nil {
		in, out := &in.Blocks, &out.Blocks
		*out = make([]BlockInFile, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BlockInFiles.
func (in *BlockInFiles) DeepCopy() *BlockInFiles {
	if in == nil {
		return nil
	}
	out := new(BlockInFiles)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Certificates) DeepCopyInto(out *Certificates) {
	*out = *in
	if in.Certificates != nil {
		in, out := &in.Certificates, &out.Certificates
		*out = make([]Certificate, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Certificates.
func (in *Certificates) DeepCopy() *Certificates {
	if in == nil {
		return nil
	}
	out := new(Certificates)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Crontabs) DeepCopyInto(out *Crontabs) {
	*out = *in
	if in.Entries != nil {
		in, out := &in.Entries, &out.Entries
		*out = make([]Crontab, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Crontabs.
func (in *Crontabs) DeepCopy() *Crontabs {
	if in == nil {
		return nil
	}
	out := new(Crontabs)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GrubKernel) DeepCopyInto(out *GrubKernel) {
	*out = *in
	if in.CmdlineArgs != nil {
		in, out := &in.CmdlineArgs, &out.CmdlineArgs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GrubKernel.
func (in *GrubKernel) DeepCopy() *GrubKernel {
	if in == nil {
		return nil
	}
	out := new(GrubKernel)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Hosts) DeepCopyInto(out *Hosts) {
	*out = *in
	if in.Hosts != nil {
		in, out := &in.Hosts, &out.Hosts
		*out = make([]Host, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Hosts.
func (in *Hosts) DeepCopy() *Hosts {
	if in == nil {
		return nil
	}
	out := new(Hosts)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KernelModules) DeepCopyInto(out *KernelModules) {
	*out = *in
	if in.Modules != nil {
		in, out := &in.Modules, &out.Modules
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KernelModules.
func (in *KernelModules) DeepCopy() *KernelModules {
	if in == nil {
		return nil
	}
	out := new(KernelModules)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KernelParameters) DeepCopyInto(out *KernelParameters) {
	*out = *in
	if in.Parameters != nil {
		in, out := &in.Parameters, &out.Parameters
		*out = make([]KernelParameterKV, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KernelParameters.
func (in *KernelParameters) DeepCopy() *KernelParameters {
	if in == nil {
		return nil
	}
	out := new(KernelParameters)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemdOverrides) DeepCopyInto(out *SystemdOverrides) {
	*out = *in
	if in.Overrides != nil {
		in, out := &in.Overrides, &out.Overrides
		*out = make([]SystemdOverride, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SystemdOverrides.
func (in *SystemdOverrides) DeepCopy() *SystemdOverrides {
	if in == nil {
		return nil
	}
	out := new(SystemdOverrides)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemdUnits) DeepCopyInto(out *SystemdUnits) {
	*out = *in
	if in.Units != nil {
		in, out := &in.Units, &out.Units
		*out = make([]SystemdUnit, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SystemdUnits.
func (in *SystemdUnits) DeepCopy() *SystemdUnits {
	if in == nil {
		return nil
	}
	out := new(SystemdUnits)
	in.DeepCopyInto(out)
	return out
}
