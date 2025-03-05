/*
Copyright 2025.

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

package v1beta1

import (
	"github.com/whitestack/node-config-operator/api/v1beta2"
	"github.com/whitestack/node-config-operator/internal/modules"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NodeConfigSpec defines the desired state of NodeConfig
type NodeConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// List of kernel parameters (sysctl). Each parameter should contain name and value
	KernelParameters modules.KernelParameters `json:"kernelParameters,omitempty"`
	// List of kernel modules to load
	KernelModules modules.KernelModules `json:"kernelModules,omitempty"`
	// List of systemd units to install
	SystemdUnits modules.SystemdUnits `json:"systemdUnits,omitempty"`
	// List of systemd overrides to add to existing systemd units
	SystemdOverrides modules.SystemdOverrides `json:"systemdOverrides,omitempty"`
	// List of hosts to install to /etc/hosts
	Hosts modules.Hosts `json:"hosts,omitempty"`
	// List of apt packages to install
	AptPackages modules.AptPackages `json:"aptPackages,omitempty"`
	// List of blocks to add to files
	BlockInFiles modules.BlockInFiles `json:"blockInFiles,omitempty"`

	// Defines the target nodes for this NodeConfig (optional, default is apply to all nodes)
	NodeSelector []metav1.LabelSelectorRequirement `json:"nodeSelector,omitempty"`
}

// NodeConfigStatus defines the observed state of NodeConfig
type NodeConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// NodeConfig is the Schema for the nodeconfigs API
type NodeConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeConfigSpec   `json:"spec,omitempty"`
	Status NodeConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NodeConfigList contains a list of NodeConfig
type NodeConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodeConfig{}, &NodeConfigList{})
}

// ConvertTo converts this v1beta1 to v1beta2. (upgrade)
func (src *NodeConfig) ConvertTo(dstRaw conversion.Hub) error {

	dst := dstRaw.(*v1beta2.NodeConfig)
	dst.ObjectMeta = src.ObjectMeta

	dst.Spec.KernelParameters = src.Spec.KernelParameters
	dst.Spec.KernelModules = src.Spec.KernelModules
	dst.Spec.SystemdUnits = src.Spec.SystemdUnits
	dst.Spec.SystemdOverrides = src.Spec.SystemdOverrides
	dst.Spec.Hosts = src.Spec.Hosts
	dst.Spec.AptPackages = src.Spec.AptPackages
	dst.Spec.BlockInFiles = src.Spec.BlockInFiles
	dst.Spec.NodeSelector = src.Spec.NodeSelector

	return nil
}

// ConvertFrom converts from the Hub version (v1beta2) to (v1beta1). (downgrade)
func (dst *NodeConfig) ConvertFrom(srcRaw conversion.Hub) error {

	src := srcRaw.(*v1beta2.NodeConfig)
	dst.ObjectMeta = src.ObjectMeta

	dst.Spec.KernelParameters = src.Spec.KernelParameters
	dst.Spec.KernelModules = src.Spec.KernelModules
	dst.Spec.SystemdUnits = src.Spec.SystemdUnits
	dst.Spec.SystemdOverrides = src.Spec.SystemdOverrides
	dst.Spec.Hosts = src.Spec.Hosts
	dst.Spec.AptPackages = src.Spec.AptPackages
	dst.Spec.BlockInFiles = src.Spec.BlockInFiles
	dst.Spec.NodeSelector = src.Spec.NodeSelector

	return nil
}
