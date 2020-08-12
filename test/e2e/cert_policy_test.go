// Copyright (c) 2020 Red Hat, Inc.

package e2e

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/open-cluster-management/governance-policy-propagator/test/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Test cert policy", func() {
	Describe("Test cert policy inform", func() {
		const certPolicyName string = "cert-policy"
		const certPolicyYaml string = "../resources/cert_policy/cert-policy.yaml"
		It("should be created on managed cluster", func() {
			By("Creating " + certPolicyYaml)
			utils.Kubectl("apply", "-f", certPolicyYaml, "-n", userNamespace, "--kubeconfig=../../kubeconfig_hub")
			hubPlc := utils.GetWithTimeout(clientHubDynamic, gvrPolicy, certPolicyName, userNamespace, true, defaultTimeoutSeconds)
			Expect(hubPlc).NotTo(BeNil())
			By("Patching " + certPolicyName + "-plr with decision of cluster managed")
			plr := utils.GetWithTimeout(clientHubDynamic, gvrPlacementRule, certPolicyName+"-plr", userNamespace, true, defaultTimeoutSeconds)
			plr.Object["status"] = utils.GeneratePlrStatus("managed")
			plr, err := clientHubDynamic.Resource(gvrPlacementRule).Namespace(userNamespace).UpdateStatus(plr, metav1.UpdateOptions{})
			Expect(err).To(BeNil())
			By("Checking " + certPolicyName + " on managed cluster in ns " + clusterNamespace)
			managedplc := utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, userNamespace+"."+certPolicyName, clusterNamespace, true, defaultTimeoutSeconds)
			Expect(managedplc).NotTo(BeNil())
		})
		It("the policy should be compliant as there is no certificate", func() {
			By("Checking if the status of root policy is compliant")
			yamlPlc := utils.ParseYaml("../resources/cert_policy/" + certPolicyName + "-compliant.yaml")
			Eventually(func() interface{} {
				rootPlc := utils.GetWithTimeout(clientHubDynamic, gvrPolicy, certPolicyName, userNamespace, true, defaultTimeoutSeconds)
				return rootPlc.Object["status"]
			}, defaultTimeoutSeconds, 1).Should(utils.SemanticEqual(yamlPlc.Object["status"]))
		})
		It("the policy should be noncompliant after creating a certficate that expires", func() {
			By("Creating ../resources/cert_policy/issuer.yaml in ns default")
			utils.Kubectl("apply", "-f", "../resources/cert_policy/issuer.yaml", "-n", "default", "--kubeconfig=../../kubeconfig_managed")
			By("Creating ../resources/cert_policy/certificate.yaml in ns default")
			utils.Kubectl("apply", "-f", "../resources/cert_policy/certificate.yaml", "-n", "default", "--kubeconfig=../../kubeconfig_managed")
			By("Checking if the status of root policy is noncompliant")
			yamlPlc := utils.ParseYaml("../resources/cert_policy/" + certPolicyName + "-noncompliant.yaml")
			Eventually(func() interface{} {
				rootPlc := utils.GetWithTimeout(clientHubDynamic, gvrPolicy, certPolicyName, userNamespace, true, defaultTimeoutSeconds)
				return rootPlc.Object["status"]
			}, defaultTimeoutSeconds, 1).Should(utils.SemanticEqual(yamlPlc.Object["status"]))
		})
		It("the policy should be compliant after creating a certficate that doesn't expire", func() {
			By("Creating ../resources/cert_policy/certificate_compliant.yaml in ns default")
			utils.Kubectl("apply", "-f", "../resources/cert_policy/certificate_compliant.yaml", "-n", "default", "--kubeconfig=../../kubeconfig_managed")
			By("Checking if the status of root policy is compliant")
			yamlPlc := utils.ParseYaml("../resources/cert_policy/" + certPolicyName + "-compliant.yaml")
			Eventually(func() interface{} {
				rootPlc := utils.GetWithTimeout(clientHubDynamic, gvrPolicy, certPolicyName, userNamespace, true, defaultTimeoutSeconds)
				return rootPlc.Object["status"]
			}, defaultTimeoutSeconds, 1).Should(utils.SemanticEqual(yamlPlc.Object["status"]))
		})
		It("the policy should be noncompliant after creating a CA certficate that expires", func() {
			By("Creating ../resources/cert_policy/issuer.yaml in ns default")
			utils.Kubectl("apply", "-f", "../resources/cert_policy/issuer.yaml", "-n", "default", "--kubeconfig=../../kubeconfig_managed")
			By("Creating ../resources/cert_policy/certificate_expired-ca.yaml in ns default")
			utils.Kubectl("apply", "-f", "../resources/cert_policy/certificate_expired-ca.yaml", "-n", "default", "--kubeconfig=../../kubeconfig_managed")
			By("Checking if the status of root policy is noncompliant")
			yamlPlc := utils.ParseYaml("../resources/cert_policy/" + certPolicyName + "-noncompliant.yaml")
			Eventually(func() interface{} {
				rootPlc := utils.GetWithTimeout(clientHubDynamic, gvrPolicy, certPolicyName, userNamespace, true, defaultTimeoutSeconds)
				return rootPlc.Object["status"]
			}, defaultTimeoutSeconds, 1).Should(utils.SemanticEqual(yamlPlc.Object["status"]))
		})
		It("the policy should be noncompliant after creating a certficate that has too long duration", func() {
			By("Creating ../resources/cert_policy/issuer.yaml in ns default")
			utils.Kubectl("apply", "-f", "../resources/cert_policy/issuer.yaml", "-n", "default", "--kubeconfig=../../kubeconfig_managed")
			By("Creating ../resources/cert_policy/certificate_long.yaml in ns default")
			utils.Kubectl("apply", "-f", "../resources/cert_policy/certificate_long.yaml", "-n", "default", "--kubeconfig=../../kubeconfig_managed")
			By("Checking if the status of root policy is noncompliant")
			yamlPlc := utils.ParseYaml("../resources/cert_policy/" + certPolicyName + "-noncompliant.yaml")
			Eventually(func() interface{} {
				rootPlc := utils.GetWithTimeout(clientHubDynamic, gvrPolicy, certPolicyName, userNamespace, true, defaultTimeoutSeconds)
				return rootPlc.Object["status"]
			}, defaultTimeoutSeconds, 1).Should(utils.SemanticEqual(yamlPlc.Object["status"]))
		})
		It("the policy should be noncompliant after creating a CA certficate that has too long duration", func() {
			By("Creating ../resources/cert_policy/issuer.yaml in ns default")
			utils.Kubectl("apply", "-f", "../resources/cert_policy/issuer.yaml", "-n", "default", "--kubeconfig=../../kubeconfig_managed")
			By("Creating ../resources/cert_policy/certificate_long-ca.yaml in ns default")
			utils.Kubectl("apply", "-f", "../resources/cert_policy/certificate_long-ca.yaml", "-n", "default", "--kubeconfig=../../kubeconfig_managed")
			By("Checking if the status of root policy is noncompliant")
			yamlPlc := utils.ParseYaml("../resources/cert_policy/" + certPolicyName + "-noncompliant.yaml")
			Eventually(func() interface{} {
				rootPlc := utils.GetWithTimeout(clientHubDynamic, gvrPolicy, certPolicyName, userNamespace, true, defaultTimeoutSeconds)
				return rootPlc.Object["status"]
			}, defaultTimeoutSeconds, 1).Should(utils.SemanticEqual(yamlPlc.Object["status"]))
		})
		It("the policy should be noncompliant after creating a certficate that has a DNS entry that is not allowed", func() {
			By("Creating ../resources/cert_policy/issuer.yaml in ns default")
			utils.Kubectl("apply", "-f", "../resources/cert_policy/issuer.yaml", "-n", "default", "--kubeconfig=../../kubeconfig_managed")
			By("Creating ../resources/cert_policy/certificate_allow-noncompliant.yaml in ns default")
			utils.Kubectl("apply", "-f", "../resources/cert_policy/certificate_allow-noncompliant.yaml", "-n", "default", "--kubeconfig=../../kubeconfig_managed")
			By("Checking if the status of root policy is noncompliant")
			yamlPlc := utils.ParseYaml("../resources/cert_policy/" + certPolicyName + "-noncompliant.yaml")
			Eventually(func() interface{} {
				rootPlc := utils.GetWithTimeout(clientHubDynamic, gvrPolicy, certPolicyName, userNamespace, true, defaultTimeoutSeconds)
				return rootPlc.Object["status"]
			}, defaultTimeoutSeconds, 1).Should(utils.SemanticEqual(yamlPlc.Object["status"]))
		})
		It("the policy should be noncompliant after creating a certficate with a disallowed wildcard", func() {
			By("Creating ../resources/cert_policy/issuer.yaml in ns default")
			utils.Kubectl("apply", "-f", "../resources/cert_policy/issuer.yaml", "-n", "default", "--kubeconfig=../../kubeconfig_managed")
			By("Creating ../resources/cert_policy/certificate_disallow-noncompliant.yaml in ns default")
			utils.Kubectl("apply", "-f", "../resources/cert_policy/certificate_disallow-noncompliant.yaml", "-n", "default", "--kubeconfig=../../kubeconfig_managed")
			By("Checking if the status of root policy is noncompliant")
			yamlPlc := utils.ParseYaml("../resources/cert_policy/" + certPolicyName + "-noncompliant.yaml")
			Eventually(func() interface{} {
				rootPlc := utils.GetWithTimeout(clientHubDynamic, gvrPolicy, certPolicyName, userNamespace, true, defaultTimeoutSeconds)
				return rootPlc.Object["status"]
			}, defaultTimeoutSeconds, 1).Should(utils.SemanticEqual(yamlPlc.Object["status"]))
		})
		It("should clean up", func() {
			By("Deleting " + certPolicyYaml)
			utils.Kubectl("delete", "-f", certPolicyYaml, "-n", userNamespace, "--kubeconfig=../../kubeconfig_hub")
			By("Checking if there is any policy left")
			utils.ListWithTimeout(clientHubDynamic, gvrPolicy, metav1.ListOptions{}, 0, true, defaultTimeoutSeconds)
			utils.ListWithTimeout(clientManagedDynamic, gvrPolicy, metav1.ListOptions{}, 0, true, defaultTimeoutSeconds)
			By("Checking if there is any cert policy left")
			utils.ListWithTimeout(clientManagedDynamic, gvrCertPolicy, metav1.ListOptions{}, 0, true, defaultTimeoutSeconds)
			By("Deleting ../resources/cert_policy/issuer.yaml in ns default")
			utils.Kubectl("delete", "-f", "../resources/cert_policy/issuer.yaml", "-n", "default", "--kubeconfig=../../kubeconfig_managed")
			By("Deleting ../resources/cert_policy/certificate.yaml in ns default")
			utils.Kubectl("delete", "-f", "../resources/cert_policy/certificate.yaml", "-n", "default", "--kubeconfig=../../kubeconfig_managed")
			By("Deleting cert-policy-secret")
			utils.Kubectl("delete", "secret", "cert-policy-secret", "-n", "default", "--kubeconfig=../../kubeconfig_managed")
		})
	})
})