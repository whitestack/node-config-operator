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

package controller

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	configurationv1beta2 "github.com/whitestack/node-config-operator/api/v1beta2"
	"github.com/whitestack/node-config-operator/internal/modules"
)

var _ = Describe("NodeConfig Controller", Ordered, func() {
	const (
		nodeName1 = "test-node-1"
		nodeName2 = "test-node-2"
		nodeName3 = "test-node-not-ready"
	)

	var (
		controllerReconciler1 *NodeConfigReconciler
		controllerReconciler2 *NodeConfigReconciler
		controllerReconciler3 *NodeConfigReconciler
	)

	BeforeAll(func() {
		By("setting the reconcilers")
		controllerReconciler1 = &NodeConfigReconciler{
			Client:   k8sClient,
			Scheme:   k8sClient.Scheme(),
			NodeName: nodeName1,
		}
		controllerReconciler2 = &NodeConfigReconciler{
			Client:   k8sClient,
			Scheme:   k8sClient.Scheme(),
			NodeName: nodeName2,
		}
		controllerReconciler3 = &NodeConfigReconciler{
			Client:   k8sClient,
			Scheme:   k8sClient.Scheme(),
			NodeName: nodeName3,
		}

		By("creating nodes in K8s")
		node1 := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: nodeName1,
				Labels: map[string]string{
					"ready":                  "true",
					"kubernetes.io/hostname": nodeName1,
				},
			},
			Status: corev1.NodeStatus{
				Conditions: []corev1.NodeCondition{
					{
						Type:   corev1.NodeReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
		}
		node2 := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: nodeName2,
				Labels: map[string]string{
					"ready":                  "true",
					"kubernetes.io/hostname": nodeName2,
				},
			},
			Status: corev1.NodeStatus{
				Conditions: []corev1.NodeCondition{
					{
						Type:   corev1.NodeReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
		}

		node3 := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: nodeName3,
				Labels: map[string]string{
					"not-ready":              "true",
					"kubernetes.io/hostname": nodeName3,
				},
			},
			Status: corev1.NodeStatus{
				Conditions: []corev1.NodeCondition{
					{
						Type:   corev1.NodeReady,
						Status: corev1.ConditionFalse,
					},
				},
			},
		}

		Expect(k8sClient.Create(ctx, node1)).To(Succeed())
		Expect(k8sClient.Create(ctx, node2)).To(Succeed())
		Expect(k8sClient.Create(ctx, node3)).To(Succeed())
	})

	Context("When reconciling an empty resource", func() {
		const resourceName = "test-resource"

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		nodeconfig := &configurationv1beta2.NodeConfig{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind NodeConfig")
			err := k8sClient.Get(ctx, typeNamespacedName, nodeconfig)
			if err != nil && errors.IsNotFound(err) {
				resource := &configurationv1beta2.NodeConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: configurationv1beta2.NodeConfigSpec{
						NodeSelector: []metav1.LabelSelectorRequirement{
							{
								Key:      "ready",
								Operator: metav1.LabelSelectorOpIn,
								Values: []string{
									"true",
								},
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}

		})

		AfterEach(func() {
			resource := &configurationv1beta2.NodeConfig{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())

			By("Cleanup the specific resource instance NodeConfig")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			_, err := controllerReconciler1.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			_, err = controllerReconciler2.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking that the correct node status is set")
			resource := &configurationv1beta2.NodeConfig{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())

			nodeStatus, ok := resource.Status.Nodes[nodeName1]
			Expect(ok).To(BeTrue())
			Expect(nodeStatus.Status).To(Equal(configurationv1beta2.NodeStatusAvailable))

			By("Checking that the correct nodeConfig condition is set")
			conditionAvailable := resource.Status.Conditions.Find(configurationv1beta2.NodeConditionAvailable)
			Expect(conditionAvailable.Status).To(Equal(metav1.ConditionTrue))
			Expect(conditionAvailable.Reason).To(Equal("all nodes configured"))
			conditionInProgress := resource.Status.Conditions.Find(configurationv1beta2.NodeConditionInProgress)
			Expect(conditionInProgress.Status).To(Equal(metav1.ConditionFalse))
			conditionError := resource.Status.Conditions.Find(configurationv1beta2.NodeConditionError)
			Expect(conditionError.Status).To(Equal(metav1.ConditionFalse))
		})
	})

	Context("When reconciling a failing resource", func() {
		const resourceName = "test-resource-2"

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		nodeconfig := &configurationv1beta2.NodeConfig{}

		BeforeEach(func() {
			By("setting the correct HOSTFS_ENABLED to true")
			Expect(os.Setenv("HOSTFS_ENABLED", "true")).To(Succeed())
			By("creating the custom resource for the Kind NodeConfig")
			err := k8sClient.Get(ctx, typeNamespacedName, nodeconfig)
			if err != nil && errors.IsNotFound(err) {
				resource := &configurationv1beta2.NodeConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: configurationv1beta2.NodeConfigSpec{
						BlockInFiles: modules.BlockInFiles{
							Blocks: []modules.BlockInFile{
								{
									FileName: "/boot/test",
									Content:  "test",
								},
							},
							State: "present",
						},
						NodeSelector: []metav1.LabelSelectorRequirement{
							{
								Key:      "ready",
								Operator: metav1.LabelSelectorOpIn,
								Values: []string{
									"true",
								},
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &configurationv1beta2.NodeConfig{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance NodeConfig")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			_, err := controllerReconciler1.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).To(HaveOccurred())

			_, err = controllerReconciler2.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).To(HaveOccurred())

			By("Checking that the error status is set")
			resource := &configurationv1beta2.NodeConfig{}
			err = k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			nodeStatus, ok := resource.Status.Nodes[nodeName1]
			Expect(ok).To(BeTrue())
			Expect(nodeStatus.Status).To(Equal(configurationv1beta2.NodeStatusError))

			By("Checking that the error condition is set")
			conditionAvailable := resource.Status.Conditions.Find(configurationv1beta2.NodeConditionAvailable)
			Expect(conditionAvailable.Status).To(Equal(metav1.ConditionFalse))
			conditionInProgress := resource.Status.Conditions.Find(configurationv1beta2.NodeConditionInProgress)
			Expect(conditionInProgress.Status).To(Equal(metav1.ConditionFalse))
			conditionError := resource.Status.Conditions.Find(configurationv1beta2.NodeConditionError)
			Expect(conditionError.Status).To(Equal(metav1.ConditionTrue))
			Expect(conditionError.Reason).To(Equal("2/2 nodes in error"))
		})
	})

	Context("When reconciling a node not ready", func() {
		const resourceName = "test-resource-not-ready"

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		nodeconfig := &configurationv1beta2.NodeConfig{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind NodeConfig")
			err := k8sClient.Get(ctx, typeNamespacedName, nodeconfig)
			if err != nil && errors.IsNotFound(err) {
				resource := &configurationv1beta2.NodeConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: configurationv1beta2.NodeConfigSpec{
						NodeSelector: []metav1.LabelSelectorRequirement{
							{
								Key:      "not-ready",
								Operator: metav1.LabelSelectorOpIn,
								Values: []string{
									"true",
								},
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}

		})

		AfterEach(func() {
			resource := &configurationv1beta2.NodeConfig{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())

			By("cleanup the specific resource instance NodeConfig")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		It("shouldn't reconcile if flag is not set", func() {
			By("Trying to reconcile the request")
			_, err := controllerReconciler3.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).To(HaveOccurred())
		})

		It("should reconcile if flag is set", func() {
			By("Setting IgnoreNodeReady flag")
			controllerReconciler3.IgnoreNodeReady = true

			By("Reconciling the resource")
			_, err := controllerReconciler3.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
