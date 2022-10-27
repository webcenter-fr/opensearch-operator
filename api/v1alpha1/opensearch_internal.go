package v1alpha1

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/thoas/go-funk"
	"github.com/webcenter-fr/opensearch-operator/pkg/helper"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultImage = "public.ecr.aws/opensearchproject/opensearch"
)

var (
	roleList = []string {
		"cluster_manager",
		"data",
		"ingest",
		"ml",
		"remote_cluster_client",
		"transform",
	}
)

// getJavaOpts permit to get computed JAVA_OPTS
func (h * Opensearch) computeJavaOpts(nodeGroup *NodeGroupSpec) string {
	javaOpts := []string{}
	
	if h.Spec.GlobalNodeGroup.Jvm != "" {
		javaOpts = append(javaOpts, h.Spec.GlobalNodeGroup.Jvm)
	}

	if nodeGroup.Jvm != "" {
		javaOpts = append(javaOpts, nodeGroup.Jvm)
	}

	return strings.Join(javaOpts, " ")
}

// computeInitialMasterNodes create the list of all master nodes
func (h *Opensearch) computeInitialMasterNodes() string {
	masterNodes := make([]string, 0, 3)
	for _, nodeGroup := range h.Spec.NodeGroups {
		if h.IsMasterRole(&nodeGroup) {
			masterNodes = append(masterNodes, h.GetNodeGroupNodeNames(&nodeGroup)...)
		}
	}

	return strings.Join(masterNodes, " ")
}

// computeDiscoverySeedHosts create the list of all headless service of all masters node groups
func (h *Opensearch) computeDiscoverySeedHosts() string {
	serviceNames := make([]string, 0, 1)

	for _, nodeGroup := range h.Spec.NodeGroups {
		if h.IsMasterRole(&nodeGroup) {
			serviceNames = append(serviceNames, h.GetNodeGroupServiceNameHeadless(nodeGroup.Name))
		}
	}

	return strings.Join(serviceNames, " ")
}


// computeRoles permit to compute les roles of node groups
func (h *Opensearch) computeRoles(roles []string) (envs []corev1.EnvVar) {
	envs = make([]corev1.EnvVar, 0, len(roles))

	for _, role :=  range roleList {
		if funk.ContainsString(roles, role) {
			envs = append(envs, corev1.EnvVar{
				Name: fmt.Sprintf("node.%s", role),
				Value: "true",
			})
		} else {
			envs = append(envs, corev1.EnvVar{
				Name: fmt.Sprintf("node.%s", role),
				Value: "false",
			})
		}
	}


	return envs
}

// computeAntiAffinity permit to get  anti affinity spec
// Default to soft anti affinity
func (h *Opensearch) computeAntiAffinity(nodeGroup *NodeGroupSpec) (antiAffinity *corev1.PodAntiAffinity, err error) {
	var expectedAntiAffinity *AntiAffinitySpec
	
	antiAffinity = &corev1.PodAntiAffinity{}
	topologyKey := "kubernetes.io/hostname"

	// Check if need to merge anti affinity spec
	if nodeGroup.AntiAffinity != nil || h.Spec.GlobalNodeGroup.AntiAffinity != nil {
		expectedAntiAffinity = &AntiAffinitySpec{}
		if err = helper.Merge(expectedAntiAffinity, nodeGroup.AntiAffinity, funk.Get(h.Spec.GlobalNodeGroup, "AntiAffinity")); err != nil {
			return nil, errors.Wrapf(err, "Error when merge global anti affinity  with node group %s", nodeGroup.Name)
		}
	}

	// Compute the antiAffinity
	if expectedAntiAffinity != nil &&  expectedAntiAffinity.TopologyKey != "" {
		topologyKey = expectedAntiAffinity.TopologyKey
	}
	if (expectedAntiAffinity != nil && expectedAntiAffinity.Type == "hard")  {

		antiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = []corev1.PodAffinityTerm{
			{
				TopologyKey: topologyKey,
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"cluster": h.Name,
						"nodeGroup": nodeGroup.Name,
					},
				},
			},
		}

		return antiAffinity, nil
	}
		
	antiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = []corev1.WeightedPodAffinityTerm{
		{
			Weight: 10,
			PodAffinityTerm: corev1.PodAffinityTerm{
				TopologyKey: topologyKey,
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"cluster": h.Name,
						"nodeGroup": nodeGroup.Name,
					},
				},
			},
		},
	}

	return antiAffinity, nil
}

// computeEnvFroms permit to compute the envFrom list
// It just append all, without to keep unique object
func (h *Opensearch ) computeEnvFroms(nodeGroup *NodeGroupSpec) (envFroms []corev1.EnvFromSource) {

	if nodeGroup.EnvFrom != nil {
		envFroms = nodeGroup.EnvFrom
	}

	if h.Spec.GlobalNodeGroup.EnvFrom != nil {
		if envFroms == nil {
			envFroms = h.Spec.GlobalNodeGroup.EnvFrom
		} else {
			envFroms = append(envFroms, h.Spec.GlobalNodeGroup.EnvFrom...)
		}
	}

	if envFroms == nil {
		envFroms = make([]corev1.EnvFromSource, 0)
	}

	return envFroms
}




// getOpensearchContainer permit to get opensearch container containning from pod template
func getOpensearchContainer(podTemplate *corev1.PodTemplateSpec) (container *corev1.Container) {
	if podTemplate == nil {
		return nil
	}

	for _, p := range podTemplate.Spec.Containers {
		if p.Name == "opensearch" {
			return &p
		}
	}

	return nil
}