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

package v1beta2

import (
	"bitbucket.org/whitestack/node-config-operator/pkg/modules"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var nodeconfiglog = logf.Log.WithName("nodeconfig-resource")

func (r *NodeConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-configuration-whitestack-com-v1beta2-nodeconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=configuration.whitestack.com,resources=nodeconfigs,verbs=create;update,versions=v1beta2,name=mnodeconfig.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &NodeConfig{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *NodeConfig) Default() {
	nodeconfiglog.Info("default", "name", r.Name)
	if len(r.Spec.KernelParameters.Parameters) == 0 {
		r.Spec.KernelParameters.Parameters = []modules.KernelParameterKV{}
		r.Spec.KernelParameters.State = "present"
	}
	if len(r.Spec.KernelModules.Modules) == 0 {
		r.Spec.KernelModules.Modules = []string{}
		r.Spec.KernelModules.State = "present"
	}
	if len(r.Spec.SystemdUnits.Units) == 0 {
		r.Spec.SystemdUnits.Units = []modules.SystemdUnit{}
		r.Spec.SystemdUnits.State = "present"
	}
	if len(r.Spec.SystemdOverrides.Overrides) == 0 {
		r.Spec.SystemdOverrides.Overrides = []modules.SystemdOverride{}
		r.Spec.SystemdOverrides.State = "present"
	}
	if len(r.Spec.Hosts.Hosts) == 0 {
		r.Spec.Hosts.Hosts = []modules.Host{}
		r.Spec.Hosts.State = "present"
	}
	if len(r.Spec.AptPackages.Packages) == 0 {
		r.Spec.AptPackages.Packages = []modules.AptPackage{}
		r.Spec.AptPackages.State = "present"
	}
	if len(r.Spec.BlockInFiles.Blocks) == 0 {
		r.Spec.BlockInFiles.Blocks = []modules.BlockInFile{}
		r.Spec.BlockInFiles.State = "present"
	}
	if len(r.Spec.Certificates.Certificates) == 0 {
		r.Spec.Certificates.Certificates = []modules.Certificate{}
		r.Spec.Certificates.State = "present"
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-configuration-whitestack-com-v1beta2-nodeconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=configuration.whitestack.com,resources=nodeconfigs,verbs=create;update,versions=v1beta2,name=vnodeconfig.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &NodeConfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *NodeConfig) ValidateCreate() error {
	nodeconfiglog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *NodeConfig) ValidateUpdate(old runtime.Object) error {
	nodeconfiglog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *NodeConfig) ValidateDelete() error {
	nodeconfiglog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
