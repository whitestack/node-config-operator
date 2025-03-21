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
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	configurationv1beta2 "github.com/whitestack/node-config-operator/api/v1beta2"
	"github.com/whitestack/node-config-operator/internal/modules"
)

var _ = Describe("NodeConfig Controller", func() {
	Context("When reconciling an empty resource", func() {
		const resourceName = "test-resource"
		const nodeName = "testNode"

		ctx := context.Background()

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
			controllerReconciler := &NodeConfigReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				NodeName: nodeName,
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking that the correct status is set")
			resource := &configurationv1beta2.NodeConfig{}
			err = k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			nodeStatus, ok := resource.Status.Nodes[nodeName]
			Expect(ok).NotTo(BeFalse())
			Expect(nodeStatus.Status).To(Equal(configurationv1beta2.NodeStatusAvailable))
		})
	})

	Context("When reconciling a failing resource", func() {
		const resourceName = "test-resource-2"
		const nodeName = "testNode"

		ctx := context.Background()

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
			controllerReconciler := &NodeConfigReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				NodeName: nodeName,
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).To(HaveOccurred())

			By("Checking that the error status is set")
			resource := &configurationv1beta2.NodeConfig{}
			err = k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			nodeStatus, ok := resource.Status.Nodes[nodeName]
			Expect(ok).To(BeTrue())
			Expect(nodeStatus.Status).To(Equal(configurationv1beta2.NodeStatusError))
		})
	})
})
