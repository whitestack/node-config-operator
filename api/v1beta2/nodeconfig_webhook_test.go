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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/whitestack/node-config-operator/internal/modules"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("NodeConfig Webhook", func() {
	var (
		validator NodeConfigValidator
	)
	Context("When creating NodeConfig under Validating Webhook", func() {
		BeforeEach(func() {
			vSettings := validatorSettings{modulePresent: true}
			validator = NodeConfigValidator{k8sClient, vSettings}
			Expect(validator).NotTo(BeNil())
		})
		It("Should only allow one NodeConfig with the same modules and same nodeSelector", func() {
			nc1 := NodeConfig{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-node-config",
					Namespace: "default",
				},
				Spec: NodeConfigSpec{
					Hosts: modules.Hosts{
						Hosts: []modules.Host{
							{
								Hostname: "test.com",
								IP:       "10.0.0.2",
							},
						},
						State: "present",
					},
				},
			}
			nc2 := NodeConfig{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-node-config2",
					Namespace: "default",
				},
				Spec: NodeConfigSpec{
					Hosts: modules.Hosts{
						Hosts: []modules.Host{
							{
								Hostname: "test.com",
								IP:       "10.0.0.2",
							},
						},
						State: "present",
					},
				},
			}

			err := k8sClient.Create(ctx, &nc1)
			Expect(err).NotTo(HaveOccurred())
			_, err = validator.ValidateCreate(ctx, &nc2)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("hosts module already defined"))
		})

		It("Should allow NodeConfigs with different NodeSelectors", func() {
			nc1 := NodeConfig{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-node-config-selector-1",
					Namespace: "default",
				},
				Spec: NodeConfigSpec{
					Hosts: modules.Hosts{
						Hosts: []modules.Host{
							{
								Hostname: "test.com",
								IP:       "10.0.0.2",
							},
						},
						State: "present",
					},
					NodeSelector: []v1.LabelSelectorRequirement{
						{
							Key:      "test",
							Operator: v1.LabelSelectorOpIn,
							Values:   []string{"test"},
						},
					},
				},
			}
			nc2 := NodeConfig{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-node-config-selector-2",
					Namespace: "default",
				},
				Spec: NodeConfigSpec{
					Hosts: modules.Hosts{
						Hosts: []modules.Host{
							{
								Hostname: "test.com",
								IP:       "10.0.0.2",
							},
						},
						State: "present",
					},
					NodeSelector: []v1.LabelSelectorRequirement{
						{
							Key:      "test",
							Operator: v1.LabelSelectorOpIn,
							Values:   []string{"test2"},
						},
					},
				},
			}

			err := k8sClient.Create(ctx, &nc1)
			Expect(err).NotTo(HaveOccurred())
			_, err = validator.ValidateCreate(ctx, &nc2)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
