// Copyright (c) 2020 Red Hat, Inc.

package e2e

import (
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/open-cluster-management/governance-policy-propagator/test/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// GetClusterLevelWithTimeout keeps polling to get the object for timeout seconds until wantFound is met (true for found, false for not found)
func GetClusterLevelWithTimeout(
	clientHubDynamic dynamic.Interface,
	gvr schema.GroupVersionResource,
	name string,
	wantFound bool,
	timeout int,
) *unstructured.Unstructured {
	if timeout < 1 {
		timeout = 1
	}
	var obj *unstructured.Unstructured

	Eventually(func() error {
		var err error
		namespace := clientHubDynamic.Resource(gvr)
		obj, err = namespace.Get(name, metav1.GetOptions{})
		if wantFound && err != nil {
			return err
		}
		if !wantFound && err == nil {
			return fmt.Errorf("expected to return IsNotFound error")
		}
		if !wantFound && err != nil && !errors.IsNotFound(err) {
			return err
		}
		return nil
	}, timeout, 1).Should(BeNil())
	if wantFound {
		return obj
	}
	return nil
}

var _ = Describe("Test gatekeeper", func() {
	Describe("Test gatekeeper policy creation", func() {
		const GKPolicyName string = "policy-gatekeeper"
		const GKPolicyYaml string = "../resources/gatekeeper/policy-gatekeeper.yaml"
		const cfgpolKRLName string = "policy-gatekeeper-k8srequiredlabels"
		const cfgpolauditName string = "policy-gatekeeper-audit"
		const cfgpoladmissionName string = "policy-gatekeeper-admission"
		const NSYamlFail string = "../resources/gatekeeper/ns-create-invalid.yaml"
		It("should deploy gatekeeper release on managed cluster", func() {
			configCRD := GetClusterLevelWithTimeout(clientManagedDynamic, gvrCRD, "configs.config.gatekeeper.sh", true, defaultTimeoutSeconds)
			Expect(configCRD).NotTo(BeNil())
			cpsCRD := GetClusterLevelWithTimeout(clientManagedDynamic, gvrCRD, "constraintpodstatuses.status.gatekeeper.sh", true, defaultTimeoutSeconds)
			Expect(cpsCRD).NotTo(BeNil())
			ctpsCRD := GetClusterLevelWithTimeout(clientManagedDynamic, gvrCRD, "constrainttemplatepodstatuses.status.gatekeeper.sh", true, defaultTimeoutSeconds)
			Expect(ctpsCRD).NotTo(BeNil())
			ctCRD := GetClusterLevelWithTimeout(clientManagedDynamic, gvrCRD, "constrainttemplates.templates.gatekeeper.sh", true, defaultTimeoutSeconds)
			Expect(ctCRD).NotTo(BeNil())
		})
		It("configurationPolicies should be created on managed", func() {
			By("Creating policy on hub")
			utils.Kubectl("apply", "-f", GKPolicyYaml, "-n", "default", "--kubeconfig=../../kubeconfig_hub")
			hubPlc := utils.GetWithTimeout(clientHubDynamic, gvrPolicy, GKPolicyName, "default", true, defaultTimeoutSeconds)
			Expect(hubPlc).NotTo(BeNil())
			By("Patching " + GKPolicyName + " pr with decision of cluster managed")
			plr := utils.GetWithTimeout(clientHubDynamic, gvrPlacementRule, "placement-"+GKPolicyName, "default", true, defaultTimeoutSeconds)
			plr.Object["status"] = utils.GeneratePlrStatus("managed")
			plr, err := clientHubDynamic.Resource(gvrPlacementRule).Namespace("default").UpdateStatus(plr, metav1.UpdateOptions{})
			Expect(err).To(BeNil())
			By("Checking configpolicies on managed")
			krl := utils.GetWithTimeout(clientManagedDynamic, gvrConfigurationPolicy, cfgpolKRLName, clusterNamespace, true, defaultTimeoutSeconds)
			Expect(krl).NotTo(BeNil())
			audit := utils.GetWithTimeout(clientManagedDynamic, gvrConfigurationPolicy, cfgpolauditName, clusterNamespace, true, defaultTimeoutSeconds)
			Expect(audit).NotTo(BeNil())
			admission := utils.GetWithTimeout(clientManagedDynamic, gvrConfigurationPolicy, cfgpoladmissionName, clusterNamespace, true, defaultTimeoutSeconds)
			Expect(admission).NotTo(BeNil())
		})
		It("K8sRequiredLabels ns-must-have-gk should be created on managed", func() {
			By("Checking if K8sRequiredLabels exists")
			k8srequiredlabelsCRD := GetClusterLevelWithTimeout(clientManagedDynamic, gvrCRD, "k8srequiredlabels.constraints.gatekeeper.sh", true, defaultTimeoutSeconds)
			Expect(k8srequiredlabelsCRD).NotTo(BeNil())
			By("Checking if ns-must-have-gk CR exists")
			nsMustHaveGkCR := GetClusterLevelWithTimeout(clientManagedDynamic, gvrK8sRequiredLabels, "ns-must-have-gk", true, defaultTimeoutSeconds)
			Expect(nsMustHaveGkCR).NotTo(BeNil())
		})
		It("K8sRequiredLabels ns-must-have-gk should be properly enforced for audit", func() {
			By("Checking if ns-must-have-gk status field has been updated")
			Eventually(func() interface{} {
				nsMustHaveGkCR := GetClusterLevelWithTimeout(clientManagedDynamic, gvrK8sRequiredLabels, "ns-must-have-gk", true, defaultTimeoutSeconds)
				return nsMustHaveGkCR.Object["status"]
			}, defaultTimeoutSeconds, 1).ShouldNot(BeNil())
			By("Checking if ns-must-have-gk status.totalViolations field has been updated")
			Eventually(func() interface{} {
				nsMustHaveGkCR := GetClusterLevelWithTimeout(clientManagedDynamic, gvrK8sRequiredLabels, "ns-must-have-gk", true, defaultTimeoutSeconds)
				return nsMustHaveGkCR.Object["status"].(map[string]interface{})["totalViolations"]
			}, defaultTimeoutSeconds*2, 1).ShouldNot(BeNil())
			By("Checking if ns-must-have-gk status.violations field has been updated")
			Eventually(func() interface{} {
				nsMustHaveGkCR := GetClusterLevelWithTimeout(clientManagedDynamic, gvrK8sRequiredLabels, "ns-must-have-gk", true, defaultTimeoutSeconds)
				return nsMustHaveGkCR.Object["status"].(map[string]interface{})["violations"]
			}, defaultTimeoutSeconds, 1).ShouldNot(BeNil())
		})
		It("K8sRequiredLabels ns-must-have-gk should be properly enforced for admission", func() {
			By("Checking if ns-must-have-gk status.byPod field size is two")
			Eventually(func() interface{} {
				nsMustHaveGkCR := GetClusterLevelWithTimeout(clientManagedDynamic, gvrK8sRequiredLabels, "ns-must-have-gk", true, defaultTimeoutSeconds)
				return len(nsMustHaveGkCR.Object["status"].(map[string]interface{})["byPod"].([]interface{}))
			}, defaultTimeoutSeconds, 1).Should(Equal(2))
		})
		It("should generate statuses properly on hub", func() {
			By("Checking if status for policy template policy-gatekeeper-k8srequiredlabels is compliant")
			Eventually(func() interface{} {
				plc := utils.GetWithTimeout(clientHubDynamic, gvrPolicy, "default."+GKPolicyName, clusterNamespace, true, defaultTimeoutSeconds)
				details := plc.Object["status"].(map[string]interface{})["details"].([]interface{})
				return details[0].(map[string]interface{})["compliant"]
			}, defaultTimeoutSeconds, 1).Should(Equal("Compliant"))
			By("Checking if violation message for policy template policy-gatekeeper-audit is correct")
			Eventually(func() interface{} {
				plc := utils.GetWithTimeout(clientHubDynamic, gvrPolicy, "default."+GKPolicyName, clusterNamespace, true, defaultTimeoutSeconds)
				details := plc.Object["status"].(map[string]interface{})["details"].([]interface{})
				return details[1].(map[string]interface{})["history"].([]interface{})[0].(map[string]interface{})["message"]
			}, defaultTimeoutSeconds, 1).Should(Equal("NonCompliant; violation - k8srequiredlabels `ns-must-have-gk` does not exist as specified"))
			By("Checking if violation message for policy template policy-gatekeeper-admission is correct")
			Eventually(func() interface{} {
				plc := utils.GetWithTimeout(clientHubDynamic, gvrPolicy, "default."+GKPolicyName, clusterNamespace, true, defaultTimeoutSeconds)
				details := plc.Object["status"].(map[string]interface{})["details"].([]interface{})
				return details[2].(map[string]interface{})["history"].([]interface{})[0].(map[string]interface{})["message"]
			}, defaultTimeoutSeconds, 1).Should(Equal("Compliant; notification - no instances of `events` exist as specified, therefore this Object template is compliant"))
		})
		It("Creating an invalid ns should generate a violation message", func() {
			By("Creating invalid namespace on managed")
			// keep trying until the create was rejected by gatekeeper
			Eventually(func() interface{} {
				out, _ := exec.Command("kubectl", "create", "ns", "e2etestfail", "--kubeconfig=../../kubeconfig_managed").CombinedOutput()
				return string(out)
			}, defaultTimeoutSeconds, 1).Should(ContainSubstring("denied by ns-must-have-gk"))
			By("Checking if status for policy template policy-gatekeeper-admission is noncompliant")
			Eventually(func() interface{} {
				plc := utils.GetWithTimeout(clientHubDynamic, gvrPolicy, "default."+GKPolicyName, clusterNamespace, true, defaultTimeoutSeconds)
				details := plc.Object["status"].(map[string]interface{})["details"].([]interface{})
				return details[2].(map[string]interface{})["compliant"]
			}, defaultTimeoutSeconds, 1).Should(Equal("NonCompliant"))
			By("Checking if violation message for policy template policy-gatekeeper-admission is correct")
			Eventually(func() interface{} {
				plc := utils.GetWithTimeout(clientHubDynamic, gvrPolicy, "default."+GKPolicyName, clusterNamespace, true, defaultTimeoutSeconds)
				details := plc.Object["status"].(map[string]interface{})["details"].([]interface{})
				return details[2].(map[string]interface{})["history"].([]interface{})[0].(map[string]interface{})["message"]
			}, defaultTimeoutSeconds, 1).Should(ContainSubstring("NonCompliant; violation - events exist:"))
		})
		It("should create relatedObjects properly on managed", func() {
			By("Checking configurationpolicies on managed")
			Eventually(func() interface{} {
				plc := utils.GetWithTimeout(clientManagedDynamic, gvrConfigurationPolicy, cfgpolauditName, clusterNamespace, true, defaultTimeoutSeconds)
				ro := plc.Object["status"].(map[string]interface{})["relatedObjects"].([]interface{})
				return ro[0].(map[string]interface{})["object"].(map[string]interface{})["metadata"].(map[string]interface{})["name"]
			}, defaultTimeoutSeconds, 1).Should(Equal("ns-must-have-gk"))
		})
		It("should clean up", func() {
			By("Deleting gatekeeper policy on hub")
			utils.Kubectl("delete", "-f", GKPolicyYaml, "-n", "default", "--kubeconfig=../../kubeconfig_hub")
			By("Checking if there is any policy left")
			utils.ListWithTimeout(clientHubDynamic, gvrPolicy, metav1.ListOptions{}, 0, true, defaultTimeoutSeconds)
			utils.ListWithTimeout(clientManagedDynamic, gvrPolicy, metav1.ListOptions{}, 0, true, defaultTimeoutSeconds)
			By("Checking if there is any configuration policy left")
			utils.ListWithTimeout(clientManagedDynamic, gvrConfigurationPolicy, metav1.ListOptions{}, 0, true, defaultTimeoutSeconds)
			By("Deleting gatekeeper ConstraintTemplate and K8sRequiredLabels")
			utils.Kubectl("delete", "K8sRequiredLabels", "--all", "--kubeconfig=../../kubeconfig_managed")
			utils.Kubectl("delete", "crd", "k8srequiredlabels.constraints.gatekeeper.sh", "--kubeconfig=../../kubeconfig_managed")
		})
	})
})