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

package v1beta2

import (
	"github.com/whitestack/node-config-operator/internal/modules"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
	// List of Certificates to add to /etc/ssl/certs
	Certificates modules.Certificates `json:"certificates,omitempty"`
	// List of Crontabs to schedule
	Crontabs modules.Crontabs `json:"crontabs,omitempty"`
	// GrubKernelConfig contains kernel version and command line arguments for GRUB configuration
	GrubKernelConfig modules.GrubKernel `json:"grubKernelConfig,omitempty"`

	// Defines the target nodes for this NodeConfig (optional, default is apply to all nodes)
	NodeSelector []metav1.LabelSelectorRequirement `json:"nodeSelector,omitempty"`
}

type NodeStatusType string

const (
	NodeStatusInProgress NodeStatusType = "InProgress"
	NodeStatusAvailable  NodeStatusType = "Available"
	NodeStatusError      NodeStatusType = "Error"
)

type NodeStatus struct {
	LastGeneration int64          `json:"lastGeneration,omitempty"`
	Status         NodeStatusType `json:"status,omitempty"`
	Error          string         `json:"error,omitempty"`
}

type ConditionType string

const (
	NodeConditionInProgress ConditionType = "InProgress"
	NodeConditionAvailable  ConditionType = "Available"
	NodeConditionError      ConditionType = "Error"
)

type Condition struct {
	Type   ConditionType          `json:"type"`
	Status metav1.ConditionStatus `json:"status"`
	Reason string                 `json:"reason,omitempty"`
}

func NewCondition(conditionType ConditionType, status metav1.ConditionStatus, reason string) Condition {
	condition := Condition{
		Type:   conditionType,
		Status: status,
		Reason: reason,
	}
	return condition
}

type ConditionList []Condition

func (c *ConditionList) Set(conditionType ConditionType, status metav1.ConditionStatus, reason string) {
	condition := c.Find(conditionType)

	// If there isn't condition we want to change, add new one
	if condition == nil {
		condition := NewCondition(conditionType, status, reason)
		*c = append(*c, condition)
		return
	}

	// If there is different status, reason or message update it
	if condition.Status != status || condition.Reason != reason {
		condition.Status = status
		condition.Reason = reason
	}
}

func (c ConditionList) Find(conditionType ConditionType) *Condition {
	for i := range c {
		if c[i].Type == conditionType {
			return &c[i]
		}
	}
	return nil
}

func (c *ConditionList) SetInProgress(reason string) {
	c.Set(
		NodeConditionInProgress,
		metav1.ConditionTrue,
		reason,
	)

	c.Set(
		NodeConditionAvailable,
		metav1.ConditionFalse,
		"",
	)

	c.Set(
		NodeConditionError,
		metav1.ConditionFalse,
		"",
	)
}

func (c *ConditionList) SetAvailable(reason string) {
	c.Set(
		NodeConditionInProgress,
		metav1.ConditionFalse,
		"",
	)

	c.Set(
		NodeConditionAvailable,
		metav1.ConditionTrue,
		reason,
	)

	c.Set(
		NodeConditionError,
		metav1.ConditionFalse,
		"",
	)
}

func (c *ConditionList) SetError(reason string) {
	c.Set(
		NodeConditionInProgress,
		metav1.ConditionFalse,
		"",
	)

	c.Set(
		NodeConditionAvailable,
		metav1.ConditionFalse,
		"",
	)

	c.Set(
		NodeConditionError,
		metav1.ConditionTrue,
		reason,
	)
}

// NodeConfigStatus defines the observed state of NodeConfig
type NodeConfigStatus struct {
	// Nodes is the list of the status of all the nodes
	Nodes      map[string]NodeStatus `json:"nodes,omitempty"`
	Conditions ConditionList         `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.status==\"True\")].type",description="Status"
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=".status.conditions[?(@.status==\"True\")].reason",description="Reason"
// +kubebuilder:storageversion

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

func (*NodeConfig) Hub() {}
