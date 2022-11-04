/*
Copyright 2022.

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

package v1alpha1

import (
	"github.com/webcenter-fr/opensearch-operator/api/shared"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OpensearchSpec defines the desired state of Opensearch
// +k8s:openapi-gen=true
type OpensearchSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	shared.ImageSpec `json:",inline"`

	// Version is the Opensearch version to use
	// Default is use the latest
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Version string `json:"version,omitempty"`

	// SetVMMaxMapCount permit to set the right value for VMMaxMapCount on node
	// It need to run pod as root with privileged option
	// Default is true
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	SetVMMaxMapCount *bool `json:"setVMMaxMapCount,omitempty"`

	// PluginsList is the list of additionnal plugin to install on each Opensearch node
	// Default is empty
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	PluginsList []string `json:"pluginsList,omitempty"`

	// GlobalNodeGroup permit to set some default parameters for each node groups
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	GlobalNodeGroup GlobalNodeGroupSpec `json:"globalNodeGroup,omitempty"`

	// NodeGroups permit to groups node per use case
	// For exemple master, data and ingest
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	NodeGroups []NodeGroupSpec `json:"nodeGroups,omitempty"`

	// Endpoint permit to set endpoints to access on Opensearch from external kubernetes
	// You can set ingress and / or load balancer
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Endpoint *EndpointSpec `json:"endpoint,omitempty"`
}

type EndpointSpec struct {
	// Ingress permit to set ingress settings
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Ingress *IngressSpec `json:"ingress,omitempty"`

	// Load balancer permit to set load balancer settings
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	LoadBalancer *LoadBalancerSpec `json:"loadBalancer,omitempty"`
}

type LoadBalancerSpec struct {
	// Enabled permit to enabled / disabled load balancer
	// Cloud provider need to support it
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Enabled bool `json:"enabled,omitempty"`

	// TargetNodeGroupName permit to define if specific node group is responsible to receive external access, like ingest nodes
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	TargetNodeGroupName string `json:"targetNodeGroupName,omitempty"`

	// Tls permit to set TLS endpoint spec
	Tls *TlsSpec `json:"tls,omitempty"`
}

type TlsSpec struct {

	// SelfSignedCertificate permit to set self signed certificate settings
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	SelfSignedCertificate *SelfSignedCertificateSpec `json:"selfSignedCertificate,omitempty"`

	// CertificateSecretRef is the secret that store your custom certificates.
	// It need to have the following keys: tls.key, tls.crt and optionally ca.crt
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	CertificateSecretRef string `json:"certificateSecretRef,omitempty"`
}

type SelfSignedCertificateSpec struct {

	// AltIps permit to set subject alt names of type ip when generate certificate
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	AltIps []string `json:"altIPs:,omitempty"`

	// AltNames permit to set subject alt names of type dns when generate certificate
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	AltNames []string `json:"altNames:,omitempty"`
}

type IngressSpec struct {

	// Enabled permit to enabled / disabled ingress
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Enabled bool `json:"enabled,omitempty"`

	// TargetNodeGroupName permit to define if specific node group is responsible to receive external access, like ingest nodes
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	TargetNodeGroupName string `json:"targetNodeGroupName,omitempty"`

	// Host is the hostname to access on Opensearch
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Host string `json:"host,omitempty"`

	// SecretRef is the secret ref that store certificates
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	SecretRef string `json:"secretRef,omitempty"`

	// Labels to set in ingress
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations to set in ingress
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// IngressSpec it merge with expected ingress spec
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	IngressSpec *networkingv1.IngressSpec `json:"ingressSpec,omitempty"`
}

type GlobalNodeGroupSpec struct {

	// AdditionalVolumes permit to use additionnal volumes
	// Default is empty
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	AdditionalVolumes []*VolumeSpec `json:"additionalVolumes,omitempty"`

	// AntiAffinity permit to set anti affinity policy
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	AntiAffinity *AntiAffinitySpec `json:"antiAffinity,omitempty"`

	// PodDisruptionBudget is the pod disruption budget policy
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	PodDisruptionBudgetSpec *policyv1.PodDisruptionBudgetSpec `json:"podDisruptionBudget,omitempty"`

	// InitContainerResources permit to set resources on init containers
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	InitContainerResources *corev1.ResourceRequirements `json:"initContainerResources,omitempty"`

	// PodTemplate is merged with expected pod
	// It usefull to add some extra properties on pod spec
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	PodTemplate *corev1.PodTemplateSpec `json:"podTemplate,omitempty"`

	// Jvm permit to set extra option on JVM like memory or proxy to download plugins
	// Becarefull with memory, not forget to set the right ressource on pod
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Jvm string `json:"jvm,omitempty"`

	// Config is the Opensearch config dedicated for this node groups like roles
	// The key is the file stored on opensearch/config
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Config map[string]string `json:"config,omitempty"`

	// SecurityRef is the secret that store the security settings
	SecurityRef string `json:"securityRef,omitempty"`

	// Labels permit to set labels on containers
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations permit to set annotation on containers
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Env permit to set some environment variable on containers
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// EnvFrom permit to set some environment variable from config map or secret
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty"`
}

type NodeGroupSpec struct {

	// Name is the the node group name
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Name string `json:"name,omitempty"`

	// Replicas is the number of replicas
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Replicas int32 `json:"replicas,omitempty"`

	// Roles is the list of Opensearch roles
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Roles []string `json:"roles,omitempty"`

	// Persistence is the spec to persist data
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Persistence *PersistenceSpec `json:"persistence,omitempty"`

	// AntiAffinity permit to set anti affinity policy
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	AntiAffinity *AntiAffinitySpec `json:"antiAffinity,omitempty"`

	// Resources permit to set ressources on Opensearch container
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// Jvm permit to set extra option on JVM like proxy to download plugins
	// Not set Xmx or Xms. It's automatically computed from Opensearch container resources
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Jvm string `json:"jvm,omitempty"`

	// Config is the Opensearch config dedicated for this node groups
	// The key is the file stored on opensearch/config
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Config map[string]string `json:"config,omitempty"`

	// SecurityRef is the secret that store the security settings
	SecurityRef string `json:"securityRef,omitempty"`

	// Tolerations permit to set toleration on pod
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// NodeSelector permit to set node selector on pod
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Labels permit to set labels on containers
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations permit to set annotation on containers
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Env permit to set some environment variable on containers
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// EnvFrom permit to set some environment variable from config map or secret
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty"`

	// PodSpec is merged with expected pod
	// It usefull to add some extra properties on pod spec
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	PodTemplate *corev1.PodTemplateSpec `json:"podSpec,omitempty"`

	// PodDisruptionBudget is the pod disruption budget policy
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	PodDisruptionBudgetSpec *policyv1.PodDisruptionBudgetSpec `json:"podDisruptionBudget,omitempty"`
}

type PersistenceSpec struct {
	// VolumeClaim is the persistent volume claim spec use by statefullset
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	VolumeClaimSpec *corev1.PersistentVolumeClaimSpec `json:"volumeClaim,omitempty"`

	// Volume is the volume source to use instead volumeClaim
	// It usefull if you should to use hostPath
	// +optional
	Volume *corev1.VolumeSource `json:"volume,omitempty"`
}

type VolumeSpec struct {

	// Name is the volume name
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Name string `json:"name,omitempty"`

	corev1.VolumeMount `json:",inline"`

	corev1.VolumeSource `json:",inline"`
}

type AntiAffinitySpec struct {

	// Type permit to set anti affinity as soft or hard
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Type string `json:"type,omitempty"`

	// TopologyKey is the topology key to use
	// Default to topology.kubernetes.io/zone
	// +optional
	TopologyKey string `json:"topologyKey,omitempty"`
}

// OpensearchStatus defines the observed state of Opensearch
type OpensearchStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Phase is the current cluster deployment phase
	Phase string `json:"phase"`

	// Url is the Opensearch endpoint
	Url string `json:"url"`

	CredentialsRef string `json:"credentialsRef"`

	// List of conditions
	// +operator-sdk:csv:customresourcedefinitions:type=status
	Conditions []metav1.Condition `json:"conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion

// Opensearch is the Schema for the opensearches API
// +operator-sdk:csv:customresourcedefinitions:resources={{None,None,None}}
// +kubebuilder:printcolumn:name="URL",type="string",JSONPath=".status.url"
// +kubebuilder:printcolumn:name="CredentialsRef",type="string",JSONPath=".status.credentialsRef"
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="Cluster deployment status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type Opensearch struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OpensearchSpec   `json:"spec,omitempty"`
	Status OpensearchStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OpensearchList contains a list of Opensearch
type OpensearchList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Opensearch `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Opensearch{}, &OpensearchList{})
}


