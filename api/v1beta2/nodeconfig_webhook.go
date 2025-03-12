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
	"context"
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *NodeConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	mPresent := os.Getenv("VALIDATION_MODULE_PRESENT_ENABLED")
	modulePresent := false
	if mPresent == "true" {
		modulePresent = true
	}

	s := validatorSettings{
		modulePresent: modulePresent,
	}

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		WithDefaulter(&NodeConfigDefaulter{c: mgr.GetClient()}).
		WithValidator(&NodeConfigValidator{c: mgr.GetClient(), s: s}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-configuration-whitestack-com-v1beta2-nodeconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=configuration.whitestack.com,resources=nodeconfigs,verbs=create;update,versions=v1beta2,name=mnodeconfig.kb.io,admissionReviewVersions=v1
// +kubebuilder:object:generate=false

type NodeConfigDefaulter struct {
	c client.Client
}

func (nd *NodeConfigDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	logger := log.FromContext(ctx)
	nc := obj.(*NodeConfig)
	logger.Info("default", "name", nc.Name)

	return nil
}

type validatorSettings struct {
	modulePresent bool
}

// +kubebuilder:webhook:path=/validate-configuration-whitestack-com-v1beta2-nodeconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=configuration.whitestack.com,resources=nodeconfigs,verbs=create;update,versions=v1beta2,name=vnodeconfig.kb.io,admissionReviewVersions=v1
// +kubebuilder:object:generate=false

type NodeConfigValidator struct {
	c client.Client
	s validatorSettings
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (nv *NodeConfigValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	logger := log.FromContext(ctx)

	nc := obj.(*NodeConfig)
	logger.Info("validate create", "name", nc.Name)

	return nil, nv.validate(ctx, nc)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (nv *NodeConfigValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	logger := log.FromContext(ctx)

	ncNew := newObj.(*NodeConfig)
	logger.Info("validate update", "name", ncNew.Name)

	return nil, nv.validate(ctx, ncNew)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (nv *NodeConfigValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (nv *NodeConfigValidator) validate(ctx context.Context, nc *NodeConfig) error {
	if nv.s.modulePresent {
		err := nv.validateModulePresent(ctx, nc)
		if err != nil {
			return err
		}
	}

	return nil
}

// validateModulePresent checks that there isn't another NodeConfig in the
// cluster that has configured the same modules for the same node selector
func (nv *NodeConfigValidator) validateModulePresent(ctx context.Context, nc *NodeConfig) error {
	ncList := &NodeConfigList{}
	err := nv.c.List(ctx, ncList, &client.ListOptions{})
	if err != nil {
		return err
	}

	ncNamespacedName := getNamespacedNameFromObject(nc)

	for _, nodeConfig := range ncList.Items {
		nodeConfigNamespacedName := getNamespacedNameFromObject(&nodeConfig)
		if ncNamespacedName == nodeConfigNamespacedName {
			// same object
			continue
		}

		sameNodeSelector, err := compareSelectors(nc.Spec.NodeSelector, nodeConfig.Spec.NodeSelector)
		if err != nil {
			return err
		}
		if !sameNodeSelector {
			continue
		}

		getError := func(moduleName string) error {
			return fmt.Errorf("%s module already defined in %s", moduleName, nodeConfigNamespacedName)
		}

		// Validate all modules
		if nc.Spec.AptPackages.IsPresent() && nodeConfig.Spec.AptPackages.IsPresent() {
			return getError("apt")
		}
		if nc.Spec.BlockInFiles.IsPresent() && nodeConfig.Spec.BlockInFiles.IsPresent() {
			return getError("blockInFiles")
		}
		if nc.Spec.Certificates.IsPresent() && nodeConfig.Spec.Certificates.IsPresent() {
			return getError("certificates")
		}
		if nc.Spec.Crontabs.IsPresent() && nodeConfig.Spec.Crontabs.IsPresent() {
			return getError("crontabs")
		}
		if nc.Spec.GrubKernelConfig.IsPresent() && nodeConfig.Spec.GrubKernelConfig.IsPresent() {
			return getError("grubKernelConfig")
		}
		if nc.Spec.Hosts.IsPresent() && nodeConfig.Spec.Hosts.IsPresent() {
			return getError("hosts")
		}
		if nc.Spec.KernelModules.IsPresent() && nodeConfig.Spec.KernelModules.IsPresent() {
			return getError("kernelModules")
		}
		if nc.Spec.KernelParameters.IsPresent() && nodeConfig.Spec.KernelParameters.IsPresent() {
			return getError("kernelParameters")
		}
		if nc.Spec.SystemdUnits.IsPresent() && nodeConfig.Spec.SystemdUnits.IsPresent() {
			return getError("systemdUnits")
		}
		if nc.Spec.SystemdOverrides.IsPresent() && nodeConfig.Spec.SystemdOverrides.IsPresent() {
			return getError("systemdOverrides")
		}
	}
	return nil
}
