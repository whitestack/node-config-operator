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

package controllers

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktypes "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	configurationv1beta2 "bitbucket.org/whitestack/node-config-operator/api/v1beta2"
	"bitbucket.org/whitestack/node-config-operator/pkg/modules"
)

var nodeName = os.Getenv("NODE_NAME")
var nodeConfigFinalizer = fmt.Sprintf("nodeconfig.whitestack.com/finalizer-%s", nodeName)
var logging = log.Log.WithName("nodeconfig_controller")

// NodeConfigReconciler reconciles a NodeConfig object
type NodeConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=configuration.whitestack.com,resources=nodeconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=configuration.whitestack.com,resources=nodeconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=configuration.whitestack.com,resources=nodeconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=nodes,verbs=list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NodeConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *NodeConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logging.WithValues("objectName", req.Name, "node", os.Getenv("NODE_NAME"))
	configs := []modules.Config{}

	nodeConfig := &configurationv1beta2.NodeConfig{}
	err := r.Get(ctx, req.NamespacedName, nodeConfig)
	if err != nil {
		if kerrors.IsNotFound(err) {
			logger.Info("NodeConfig resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
	}

	// Check if selector matches
	if len(nodeConfig.Spec.NodeSelector) > 0 {
		logger.Info("node selector found", "selector", nodeConfig.Spec.NodeSelector)
		matches, err := checkNodeBySelector(r.Client, nodeConfig, logger)
		if err != nil {
			logger.Error(err, "error while checking if node matches")
			return ctrl.Result{}, err
		}

		if !matches {
			logger.Info("selector doesn't match this node, ignoring...")
			return ctrl.Result{}, nil
		}
	}

	// START of config types handling
	if len(nodeConfig.Spec.AptPackages.Packages) != 0 {
		configs = append(
			configs,
			modules.AptModuleConfig{
				AptPackages: nodeConfig.Spec.AptPackages,
				Logger:      logger.WithName("apt-packages"),
			},
		)
	}
	if len(nodeConfig.Spec.KernelModules.Modules) != 0 {
		configs = append(
			configs,
			modules.NewKernelModuleConfig(
				nodeConfig.Spec.KernelModules,
				logger.WithName("kernel-modules"),
			),
		)
	}

	if len(nodeConfig.Spec.KernelParameters.Parameters) != 0 {
		configs = append(
			configs,
			modules.NewKernelParameterConfig(
				nodeConfig.Spec.KernelParameters,
				logger.WithName("kernel-parameter"),
			),
		)
	}

	if len(nodeConfig.Spec.Hosts.Hosts) != 0 {
		configs = append(
			configs,
			modules.NewHostModuleConfig(
				nodeConfig.Spec.Hosts,
				logger.WithName("hosts"),
			),
		)
	}

	if len(nodeConfig.Spec.BlockInFiles.Blocks) != 0 {
		configs = append(
			configs,
			modules.BlockInFileConfig{
				BlockInFiles: nodeConfig.Spec.BlockInFiles,
				Log:          logger.WithName("block-in-files"),
			},
		)
	}

	if len(nodeConfig.Spec.SystemdUnits.Units) != 0 {
		configs = append(
			configs,
			modules.NewSystemdUnitConfig(
				nodeConfig.Spec.SystemdUnits,
				logger.WithName("systemd-units"),
			),
		)
	}

	if len(nodeConfig.Spec.Certificates.Certificates) != 0 {
		configs = append(
			configs,
			modules.CertificateConfig{
				Certificates: nodeConfig.Spec.Certificates,
				Log:          logger.WithName("certificates"),
			},
		)
	}

	if len(nodeConfig.Spec.SystemdOverrides.Overrides) != 0 {
		configs = append(
			configs,
			modules.NewSystemdOverrideConfig(
				nodeConfig.Spec.SystemdOverrides,
				logger.WithName("systemd-overrides"),
			),
		)
	}

	if len(nodeConfig.Spec.Crontabs.Entries) != 0 {
		configs = append(
			configs,
			modules.CrontabsConfig{
				Crontabs: nodeConfig.Spec.Crontabs,
				Log:      logger.WithName("crontabs"),
			},
		)
	}
	// END of config types handling

	if !nodeConfig.ObjectMeta.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(nodeConfig, nodeConfigFinalizer) {
			// TODO: our finalizer is present, so lets handle any external dependency

			// update object before deleting finalizer to avoid the error
			// "the object has been modified..."
			if err := r.Get(ctx, req.NamespacedName, nodeConfig); err != nil {
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(nodeConfig, nodeConfigFinalizer)
			if err := r.Update(ctx, nodeConfig); err != nil {
				return ctrl.Result{}, err
			}
		}
		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	// Reconciliation logic
	// Loop over all configs and call Reconcile
	logger.Info("reconciling node")
	for _, config := range configs {
		if err := config.Reconcile(); err != nil {
			return ctrl.Result{}, err
		}
		// TODO: add status and (optionally) create an event
	}

	// Add our finalizer to the object
	if !controllerutil.ContainsFinalizer(nodeConfig, nodeConfigFinalizer) {
		// update object before adding finalizer to avoid the error
		// "the object has been modified..."
		if err := r.Get(ctx, req.NamespacedName, nodeConfig); err != nil {
			return ctrl.Result{}, err
		}

		controllerutil.AddFinalizer(nodeConfig, nodeConfigFinalizer)
		if err := r.Update(ctx, nodeConfig); err != nil {
			return ctrl.Result{}, err
		}
	}

	logger.Info("node reconciled")
	return ctrl.Result{RequeueAfter: 2 * time.Minute}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	err := ctrl.NewControllerManagedBy(mgr).
		For(&configurationv1beta2.NodeConfig{}).
		Complete(r)
	if err != nil {
		return err
	}

	err = ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: nodeName}}).
		Watches(
			&source.Kind{
				Type: &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: nodeName,
					},
				},
			},
			handler.EnqueueRequestsFromMapFunc(
				func(a client.Object) []reconcile.Request {
					routes := &configurationv1beta2.NodeConfigList{}
					if err := r.List(context.Background(), routes); err != nil {
						logging.Error(err, "Failed to list NodeConfigs")
						return nil
					}

					var result []reconcile.Request
					for _, route := range routes.Items {
						result = append(result, reconcile.Request{
							NamespacedName: ktypes.NamespacedName{
								Name:      route.GetName(),
								Namespace: route.GetNamespace(),
							},
						})
					}
					return result
				},
			)).
		WithEventFilter(
			&predicate.Funcs{
				CreateFunc: func(e event.CreateEvent) bool {
					return false
				},
				UpdateFunc: func(e event.UpdateEvent) bool {
					if len(e.ObjectNew.GetLabels()) != len(e.ObjectOld.GetLabels()) {
						logging.Info("Node label amount changed. Submitting all NodeConfig CRs for reconciliation.")
						return true
					}
					for k, v := range e.ObjectOld.GetLabels() {
						if e.ObjectNew.GetLabels()[k] != v {
							logging.Info("Node labels changed. Submitting all NodeConfig CRs for reconciliation.")
							return true
						}
					}
					return false
				},
				DeleteFunc: func(e event.DeleteEvent) bool {
					return false
				},
			}).
		Complete(r)

	return err
}

// checkNodeBySelector returns true if the current node matches the
// nodeSelector in the nodeConfig object, returning an error if any
// function call fails
func checkNodeBySelector(c client.Client, nodeConfig *configurationv1beta2.NodeConfig, logger logr.Logger) (bool, error) {
	nodes := &corev1.NodeList{}
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchExpressions: append(nodeConfig.Spec.NodeSelector,
			metav1.LabelSelectorRequirement{
				Key:      "kubernetes.io/hostname",
				Operator: metav1.LabelSelectorOpIn,
				Values:   []string{nodeName},
			}),
	})
	if err != nil {
		return false, err
	}

	listOptions := &client.ListOptions{LabelSelector: selector}
	if err := c.List(context.Background(), nodes, listOptions); err != nil {
		logger.Error(err, "Failed to fetch nodes")
		return false, err
	} else if len(nodes.Items) == 0 {
		logger.Info("Node not found with the given selectors", "Value", selector)
		return false, nil
	}
	return true, nil
}
