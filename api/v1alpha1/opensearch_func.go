package v1alpha1

import (
	"fmt"
	"strings"

	"github.com/disaster37/k8sbuilder"
	"github.com/pkg/errors"
	"github.com/thoas/go-funk"
	"github.com/webcenter-fr/opensearch-operator/pkg/helper"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
)

// GetNodeNames permit to get all nodes names
// It return the list with all node names (DNS / pod name)
func (h *Opensearch) GetNodeNames() (nodeNames []string) {
	nodeNames = make([]string, 0)

	for _, nodeGroup := range h.Spec.NodeGroups {
		nodeNames = append(nodeNames, h.GetNodeGroupNodeNames(&nodeGroup)...)
	}

	return nodeNames
}

// GetSecretNameForTlsTransport permit to get the secret name that store all certificates for transport layout
// It return the secret name as string
func (h *Opensearch) GetSecretNameForTlsTransport() (secretName string) {
	return fmt.Sprintf("%s-os-tls-transport", h.Name)
}

// IsSelfManagedSecretForTlsApi return true if the operator manage the certificates for Api layout
// It return false if secret is provided
func (h *Opensearch) IsSelfManagedSecretForTlsApi() bool {
	if h.Spec.Endpoint != nil && h.Spec.Endpoint.LoadBalancer != nil && h.Spec.Endpoint.LoadBalancer.Enabled && h.Spec.Endpoint.LoadBalancer.Tls !=  nil && h.Spec.Endpoint.LoadBalancer.Tls.CertificateSecretRef != "" {
		return false
	}
	return true
}

// GetSecretNameForTlsApi permit to get the secret name that store all certificates for Api layout (Http endpoint)
// It return the secret name as string
func (h *Opensearch) GetSecretNameForTlsApi() (secretName string) {
	if h.IsSelfManagedSecretForTlsApi() {
		return fmt.Sprintf("%s-os-tls-api", h.Name)
	}

	return h.Spec.Endpoint.LoadBalancer.Tls.CertificateSecretRef
}


// GetSecretNameForAdminCredentials permit to get the secret name that store the admin credentials
func (h *Opensearch) GetSecretNameForAdminCredentials() (secretName string) {
	return fmt.Sprintf("%s-os-credential", h.Name)
}

// GetSecretNameForSecurityConfig permit to get the secret name that store the security config of Opensearch
// It use the secret declared on node group. If empty, it use secret declared on global node group
func (h *Opensearch) GetSecretNameForSecurityConfig(nodeGroup *NodeGroupSpec) (secretName string) {
	if nodeGroup.SecurityRef != "" {
		return nodeGroup.SecurityRef
	}

	if h.Spec.GlobalNodeGroup.SecurityRef != "" {
		return h.Spec.GlobalNodeGroup.SecurityRef
	}

	return ""
}

// GetNodeGroupName permit to get the node group name
func (h *Opensearch) GetNodeGroupName(nodeGroupName string) (name string) {
	return fmt.Sprintf("%s-%s-os", h.Name, nodeGroupName)
}

// GetNodeGroupNodeNames permit to get node names that composed the node group
func (h *Opensearch) GetNodeGroupNodeNames(nodeGroup *NodeGroupSpec) (nodeNames []string) {
	nodeNames = make([]string, 0, nodeGroup.Replicas)

	for i := 0; i < int(nodeGroup.Replicas); i++ {
		nodeNames = append(nodeNames, fmt.Sprintf("%s-%d", h.GetNodeGroupName(nodeGroup.Name), i))
	}

	return nodeNames
}

// GetConfigMapNameForConfig permit to get the configMap name that store the config of Opensearch
func (h *Opensearch) GetNodeGroupConfigMapName(nodeGroupName string) (configMapName string) {
	return fmt.Sprintf("%s-config", h.GetNodeGroupName(nodeGroupName))
}

// GetGlobalServiceName permit to get the global service name
func (h *Opensearch) GetGlobalServiceName() (serviceName string) {
	return fmt.Sprintf("%s-os", h.Name)
}

// GetNodeGroupServiceName permit to get the service name for specified node group name
func (h *Opensearch) GetNodeGroupServiceName(nodeGroupName string) (serviceName string) {
	return h.GetNodeGroupName(nodeGroupName)
}

// GetNodeGroupServiceNameHeadless permit to get the service name headless for specified node group name
func (h *Opensearch) GetNodeGroupServiceNameHeadless(nodeGroupName string) (serviceName string) {
	return fmt.Sprintf("%s-headless", h.GetNodeGroupName(nodeGroupName))
}

func(h *Opensearch) GetNodeGroupPDBName(nodeGroupName string) (serviceName string) {
	return h.GetNodeGroupName(nodeGroupName)
}

// IsIngressEnabled return true if ingress is enabled
func (h *Opensearch) IsIngressEnabled() bool {
	if h.Spec.Endpoint != nil && h.Spec.Endpoint.Ingress != nil && h.Spec.Endpoint.Ingress.Enabled {
		return true
	}

	return false
}

// IsLoadBalancerEnabled return true if LoadBalancer is enabled
func (h *Opensearch) IsLoadBalancerEnabled() bool {
	if h.Spec.Endpoint != nil && h.Spec.Endpoint.LoadBalancer != nil && h.Spec.Endpoint.LoadBalancer.Enabled {
		return true
	}

	return false
}

// GetContainerImage permit to get the image name
func (h *Opensearch) GetContainerImage() string {
	version := "latest"
	if h.Spec.Version != "" {
		version = h.Spec.Version
	}

	image := defaultImage
	if h.Spec.Image != "" {
		image = h.Spec.Image
	}

	return fmt.Sprintf("%s:%s", image, version)
}

// GenerateIngress permit to generate Ingress object
// It return error if ingress spec is not provided
// It return nil if ingress is disabled
func (h *Opensearch) GenerateIngress() (ingress *networkingv1.Ingress, err error) {
	if !h.IsIngressEnabled() {
		return nil, nil
	}

	if h.Spec.Endpoint.Ingress.Host == "" {
		return nil, errors.New("endpoint.ingress.host must be provided")
	}

	defaultAnnotations := map[string]string {
		"nginx.ingress.kubernetes.io/force-ssl-redirect": "true",
		"nginx.ingress.kubernetes.io/backend-protocol": "HTTPS",
	}

	pathType := networkingv1.PathTypePrefix
	targetService := h.GetGlobalServiceName()
	if h.Spec.Endpoint.Ingress.TargetNodeGroupName != "" {
		// Check the node group specified exist
		isFound := false
		for _, nodeGroup := range h.Spec.NodeGroups {
			if nodeGroup.Name == h.Spec.Endpoint.Ingress.TargetNodeGroupName {
				isFound = true
				break
			}
		}
		if !isFound {
			return nil, errors.Errorf("The target group name '%s' not found", h.Spec.Endpoint.Ingress.TargetNodeGroupName)
		}

		targetService = h.GetNodeGroupServiceName(h.Spec.Endpoint.Ingress.TargetNodeGroupName)
	}


	ingress = &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: h.Namespace,
			Name: h.Name,
			Labels: funk.UnionStringMap(h.Spec.Endpoint.Ingress.Labels, h.Labels),
			Annotations: funk.UnionStringMap(defaultAnnotations, h.Spec.Endpoint.Ingress.Annotations, h.Annotations),
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: h.Spec.Endpoint.Ingress.Host,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: targetService,
											Port: networkingv1.ServiceBackendPort{Number: 9200},
										},
									},
								},
							},
						},
					},
				},
			},
			TLS: []networkingv1.IngressTLS{
				{
					Hosts: []string{h.Spec.Endpoint.Ingress.Host},
					SecretName: h.Spec.Endpoint.Ingress.SecretRef,
				},
			},
		},
	}

	// Merge expected ingress with custom ingress spec
	if err = helper.Merge(&ingress.Spec, h.Spec.Endpoint.Ingress.IngressSpec); err != nil {
		return nil, errors.Wrap(err, "Error when merge ingress spec")
	}

	return ingress, nil
}

// GenerateServices permit to generate services
// It generate one for all cluster and for each node group
// For each node groups, it also generate headless services
func (h *Opensearch) GenerateServices() (services []*corev1.Service, err error) {
	services = make([]*corev1.Service, 0, (1 + len(h.Spec.NodeGroups)) * 2)
	var(
		service *corev1.Service
		headlessService *corev1.Service
	)

	defaultHeadlessAnnotations := map[string]string {
		"service.alpha.kubernetes.io/tolerate-unready-endpoints": "true",
	}

	// Generate cluster service
	service = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: h.Namespace,
			Name: h.GetGlobalServiceName(),
			Labels: h.Labels,
			Annotations: h.Annotations,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			SessionAffinity: corev1.ServiceAffinityNone,
			Selector: map[string]string{
				"cluster": h.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Name: "http",
					Protocol: corev1.ProtocolTCP,
					Port: 9200,
					TargetPort: intstr.FromInt(9200),
				},
				{
					Name: "transport",
					Protocol: corev1.ProtocolTCP,
					Port: 9300,
					TargetPort: intstr.FromInt(9300),
				},
			},
		},
	}

	services = append(services, service)

	// Generate service for each node group
	for _, nodeGroup :=  range h.Spec.NodeGroups {
		service = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: h.Namespace,
				Name: h.GetNodeGroupServiceName(nodeGroup.Name),
				Labels: h.Labels,
				Annotations: h.Annotations,
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeClusterIP,
				SessionAffinity: corev1.ServiceAffinityNone,
				Selector: map[string]string{
					"cluster": h.Name,
					"nodeGroup": nodeGroup.Name,
				},
				Ports: []corev1.ServicePort{
					{
						Name: "http",
						Protocol: corev1.ProtocolTCP,
						Port: 9200,
						TargetPort: intstr.FromInt(9200),
					},
					{
						Name: "transport",
						Protocol: corev1.ProtocolTCP,
						Port: 9300,
						TargetPort: intstr.FromInt(9300),
					},
				},
			},
		}

		headlessService = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: h.Namespace,
				Name: h.GetNodeGroupServiceNameHeadless(nodeGroup.Name),
				Labels: h.Labels,
				Annotations: funk.UnionStringMap(h.Annotations, defaultHeadlessAnnotations),
			},
			Spec: corev1.ServiceSpec{
				ClusterIP: "None",
				PublishNotReadyAddresses: true,
				Type: corev1.ServiceTypeClusterIP,
				SessionAffinity: corev1.ServiceAffinityNone,
				Selector: map[string]string{
					"cluster": h.Name,
					"nodeGroup": nodeGroup.Name,
				},
				Ports: []corev1.ServicePort{
					{
						Name: "http",
						Protocol: corev1.ProtocolTCP,
						Port: 9200,
						TargetPort: intstr.FromInt(9200),
					},
					{
						Name: "transport",
						Protocol: corev1.ProtocolTCP,
						Port: 9300,
						TargetPort: intstr.FromInt(9300),
					},
				},
			},
		}

		services = append(services, service, headlessService)
	}

	return services, nil
}

// GenerateLoadbalancer permit to generate Loadbalancer throught service
// It return nil if Loadbalancer is disabled
func (h *Opensearch) GenerateLoadbalancer() (service *corev1.Service, err error) {

	if !h.IsLoadBalancerEnabled() {
		return nil, nil
	}


	selector := map[string]string{
		"cluster": h.Name,
	}
	if h.Spec.Endpoint.LoadBalancer.TargetNodeGroupName != "" {
		// Check the node group specified exist
		isFound := false
		for _, nodeGroup := range h.Spec.NodeGroups {
			if nodeGroup.Name == h.Spec.Endpoint.LoadBalancer.TargetNodeGroupName {
				isFound = true
				break
			}
		}
		if !isFound {
			return nil, errors.Errorf("The target group name '%s' not found", h.Spec.Endpoint.LoadBalancer.TargetNodeGroupName)
		}

		selector["nodeGroup"] = h.Spec.Endpoint.LoadBalancer.TargetNodeGroupName
	}

	service = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: h.Namespace,
			Name: fmt.Sprintf("%s-lb", h.GetGlobalServiceName()),
			Labels: h.Labels,
			Annotations: h.Annotations,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeLoadBalancer,
			SessionAffinity: corev1.ServiceAffinityNone,
			Selector: selector,
			Ports: []corev1.ServicePort{
				{
					Name: "http",
					Protocol: corev1.ProtocolTCP,
					Port: 9200,
					TargetPort: intstr.FromInt(9200),
				},
			},
		},
	}

	return service, nil
}

// GenerateConfigMaps permit to generate config maps for each node Groups
func (h *Opensearch) GenerateConfigMaps() (configMaps []*corev1.ConfigMap, err error) {
	var (
		configMap *corev1.ConfigMap
		expectedConfig map[string]string
	)

	configMaps = make([]*corev1.ConfigMap, 0, len(h.Spec.NodeGroups))
	injectedConfigMap := map[string]string {
		"opensearch.yml": `
plugins.security.ssl.transport.keystore_type: 'PKCS12/PFX'
plugins.security.ssl.transport.keystore_filepath: 'certs/transport/${hostname}.pfx'
plugins.security.ssl.transport.truststore_type: 'PKCS12/PFX'
plugins.security.ssl.transport.truststore_filepath: 'certs/transport/${hostanme}.pfx'
plugins.security.ssl.transport.enforce_hostname_verification: true
plugins.security.ssl.http.enabled: true
plugins.security.ssl.http.keystore_type: 'PKCS12/PFX'
plugins.security.ssl.http.keystore_filepath: 'certs/http/api.pfx'
plugins.security.ssl.http.truststore_type: 'PKCS12/PFX'
plugins.security.ssl.http.truststore_filepath: 'certs/http/api.pfx'`,
	}

	for _, nodeGroup := range h.Spec.NodeGroups {
		
		if h.Spec.GlobalNodeGroup.Config != nil {
			expectedConfig, err = helper.MergeSettings(nodeGroup.Config, h.Spec.GlobalNodeGroup.Config)
			if err != nil {
				return nil, errors.Wrapf(err, "Error when merge config from global config and node group config %s", nodeGroup.Name)
			}
		} else {
			expectedConfig = nodeGroup.Config
		}

		// Inject computed config
		expectedConfig, err = helper.MergeSettings(injectedConfigMap, expectedConfig)
		if err != nil {
			return nil, errors.Wrapf(err, "Error when merge expected config with computed config on node group %s", nodeGroup.Name)
		}

		configMap = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: h.Namespace,
				Name: h.GetNodeGroupConfigMapName(nodeGroup.Name),
				Labels: h.Labels,
				Annotations: h.Annotations,
			},
			Data: expectedConfig,
		}
		configMaps = append(configMaps, configMap)
	}

	return configMaps, nil
}

// GenerateStatefullsets permit to generate statefullsets for each node groups
func (h *Opensearch) GenerateStatefullsets() (statefullsets []*appv1.StatefulSet, err error) {
	var (
		sts *appv1.StatefulSet
	)

	for _, nodeGroup := range h.Spec.NodeGroups {

		cb := k8sbuilder.NewContainerBuilder()
		ptb := k8sbuilder.NewPodTemplateBuilder()
		globalOpensearchContainer := getOpensearchContainer(h.Spec.GlobalNodeGroup.PodTemplate)
		if globalOpensearchContainer == nil {
			globalOpensearchContainer = &corev1.Container{}
		}
		localOpensearchContainer := getOpensearchContainer(nodeGroup.PodTemplate)
		if localOpensearchContainer == nil {
			localOpensearchContainer = &corev1.Container{}
		}

		// Initialise Opensearch container from user provided
		cb.WithContainer(globalOpensearchContainer).
		WithContainer(localOpensearchContainer, k8sbuilder.Merge).
		Container().Name = "opensearch"

		// Compute EnvFrom
		cb.WithEnvFrom(h.Spec.GlobalNodeGroup.EnvFrom).
		WithEnvFrom(nodeGroup.EnvFrom, k8sbuilder.Merge)

		// Compute Env
		cb.WithEnv(h.Spec.GlobalNodeGroup.Env).
		WithEnv(nodeGroup.Env, k8sbuilder.Merge).
		WithEnv(h.computeRoles(nodeGroup.Roles), k8sbuilder.Merge).
		WithEnv([]corev1.EnvVar{
			{
				Name: "node.name",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath: "metadata.name",
					},
				},
			},
			{
				Name: "host",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath: "spec.nodeName",
					},
				},
			},
			{
				Name: "OPENSEARCH_JAVA_OPTS",
				Value: h.computeJavaOpts(&nodeGroup),
			},
			{
				Name: "cluster.initial_master_nodes",
				Value: h.computeInitialMasterNodes(),
			},
			{
				Name: "discovery.seed_hosts",
				Value: h.computeDiscoverySeedHosts(),
			},
			{
				Name: "cluster.name",
				Value: h.Name,
			},
			{
				Name: "network.host",
				Value: "0.0.0.0",
			},
			{
				Name: "bootstrap.memory_lock",
				Value: "true",
			},
			{
				Name: "DISABLE_INSTALL_DEMO_CONFIG",
				Value: "true",
			},
		}, k8sbuilder.Merge)
		if len(h.Spec.NodeGroups) == 1 && h.Spec.NodeGroups[0].Replicas == 1 {
			// Cluster with only one node
			cb.WithEnv([]corev1.EnvVar{
				{
					Name: "discovery.type",
					Value: "single-node",
				},
			 }, k8sbuilder.Merge)
		}

		// Compute ports
		cb.WithPort([]corev1.ContainerPort{
			{
				Name: "http",
				ContainerPort: 9200,
				Protocol: corev1.ProtocolTCP,
			},
			{
				Name: "transport",
				ContainerPort: 9300,
				Protocol: corev1.ProtocolTCP,
			},
		}, k8sbuilder.Merge)
		
		// Compute resources
		cb.WithResource(nodeGroup.Resources, k8sbuilder.Merge)

		// Compute image
		cb.WithImage(h.GetContainerImage(), k8sbuilder.OverwriteIfDefaultValue)

		// Compute image pull policy
		cb.WithImagePullPolicy(h.Spec.ImagePullPolicy).
		WithImagePullPolicy(globalOpensearchContainer.ImagePullPolicy, k8sbuilder.Merge).
		WithImagePullPolicy(localOpensearchContainer.ImagePullPolicy, k8sbuilder.Merge)

		// Compute security context
		cb.WithSecurityContext(&corev1.SecurityContext{
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{
					"ALL",
				},
			},
			RunAsUser: pointer.Int64(1000),
			RunAsNonRoot: pointer.Bool(true),
		}, k8sbuilder.OverwriteIfDefaultValue)

		// Compute volume mount
		additionalVolumeMounts := make([]corev1.VolumeMount, 0, len(h.Spec.GlobalNodeGroup.AdditionalVolumes))
		for _, vol := range h.Spec.GlobalNodeGroup.AdditionalVolumes {
			additionalVolumeMounts = append(additionalVolumeMounts, corev1.VolumeMount{
				Name: vol.Name,
				MountPath: vol.MountPath,
				ReadOnly: vol.ReadOnly,
				SubPath: vol.SubPath,
			})
		}
		cb.WithVolumeMount(additionalVolumeMounts).
		WithVolumeMount( []corev1.VolumeMount{
			{
				Name: "node-tls",
				MountPath: "/usr/share/opensearch/config/certs/node",
			},
			{
				Name: "api-tls",
				MountPath: "/usr/share/opensearch/config/certs/api",
			},
		}, k8sbuilder.Merge)
		if nodeGroup.Persistence != nil && (nodeGroup.Persistence.Volume != nil || nodeGroup.Persistence.VolumeClaimSpec != nil) {
			cb.WithVolumeMount([]corev1.VolumeMount{
				{
					Name: "opensearch-data",
					MountPath: "/usr/share/opensearch/data",
				},
			}, k8sbuilder.Merge)
		}
		// Compute mount config maps
		configMaps, err := h.GenerateConfigMaps()
		if err != nil {
			return nil, errors.Wrap(err, "Error when generate configMaps")
		}
		for _, configMap :=  range configMaps {
			if configMap.Name == h.GetNodeGroupConfigMapName(nodeGroup.Name) {
				additionalVolumeMounts = make([]corev1.VolumeMount, 0, len(configMap.Data))
				for key :=  range configMap.Data {
					additionalVolumeMounts = append(additionalVolumeMounts, corev1.VolumeMount{
						Name: "opensearch-config",
						MountPath: fmt.Sprintf("/usr/share/opensearch/config/%s", key),
						SubPath: key,
					})
				}
				cb.WithVolumeMount(additionalVolumeMounts, k8sbuilder.Merge)
			}
		}

		// Compute liveness
		cb.WithLivenessProbe(&corev1.Probe{
			TimeoutSeconds: 5,
			PeriodSeconds: 30,
			FailureThreshold: 10,
			SuccessThreshold: 1,
			ProbeHandler: corev1.ProbeHandler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt(9300),
				},
			},
		}, k8sbuilder.OverwriteIfDefaultValue)

		// Compute readiness
		cb.WithReadinessProbe(&corev1.Probe{
			TimeoutSeconds: 5,
			PeriodSeconds: 30,
			FailureThreshold: 3,
			SuccessThreshold: 1,
			ProbeHandler: corev1.ProbeHandler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt(9200),
				},
			},
		}, k8sbuilder.OverwriteIfDefaultValue)

		// Compute startup
		cb.WithStartupProbe(&corev1.Probe{
			InitialDelaySeconds: 10,
			TimeoutSeconds: 5,
			PeriodSeconds: 10,
			FailureThreshold: 30,
			SuccessThreshold: 1,
			ProbeHandler: corev1.ProbeHandler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt(9200),
				},
			},
		}, k8sbuilder.OverwriteIfDefaultValue)

		// Add specific command to handle plugin installation
		var pluginInstallation strings.Builder
		pluginInstallation.WriteString(`
#!/usr/bin/env bash
set -euo pipefail

`)	
		for _, plugin :=  range h.Spec.PluginsList {
			pluginInstallation.WriteString(fmt.Sprintf("./bin/opensearch-plugin install -b %s\n", plugin))
		}
		pluginInstallation.WriteString("bash opensearch-docker-entrypoint.sh")
		cb.Container().Command = []string{
			"sh",
			"-c",
			pluginInstallation.String(),
		}

		// Initialise PodTemplate
		ptb.WithPodTemplateSpec(h.Spec.GlobalNodeGroup.PodTemplate).
		WithPodTemplateSpec(nodeGroup.PodTemplate, k8sbuilder.Merge)

		// Compute labels
		ptb.WithLabels(h.Labels, k8sbuilder.Merge).
		WithLabels(h.Spec.GlobalNodeGroup.Labels, k8sbuilder.Merge).
		WithLabels(nodeGroup.Labels, k8sbuilder.Merge).WithLabels(map[string]string{
			"cluster": h.Name,
			"nodeGroup": nodeGroup.Name,
		}, k8sbuilder.Merge)

		// Compute annotations
		ptb.WithAnnotations(h.Annotations, k8sbuilder.Merge).
		WithAnnotations(h.Spec.GlobalNodeGroup.Annotations, k8sbuilder.Merge).
		WithAnnotations(nodeGroup.Annotations, k8sbuilder.Merge)

		// Compute NodeSelector
		ptb.WithNodeSelector(nodeGroup.NodeSelector, k8sbuilder.Merge)

		// Compute Termination grac period
		ptb.WithTerminationGracePeriodSeconds(120, k8sbuilder.OverwriteIfDefaultValue)

		// Compute toleration
		ptb.WithTolerations(nodeGroup.Tolerations, k8sbuilder.Merge)



		// compute anti affinity
		if h.Spec.GlobalNodeGroup.AntiAffinity != nil {
			antiAffinity := corev1.PodAntiAffinity{}
			topologyKey := "kubernetes.io/hostname"
			if h.Spec.GlobalNodeGroup.AntiAffinity.TopologyKey != "" {
				topologyKey = h.Spec.GlobalNodeGroup.AntiAffinity.TopologyKey
			}
			if h.Spec.GlobalNodeGroup.AntiAffinity.Type == "hard" {
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
			} else {
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
			}
			ptb.WithAffinity(corev1.Affinity{
				PodAntiAffinity: &antiAffinity,
			}, k8sbuilder.Merge)
		}
		if nodeGroup.AntiAffinity != nil {
			antiAffinity := corev1.PodAntiAffinity{}
			topologyKey := "kubernetes.io/hostname"
			if nodeGroup.AntiAffinity.TopologyKey != "" {
				topologyKey = nodeGroup.AntiAffinity.TopologyKey
			}
			if nodeGroup.AntiAffinity.Type == "hard" {
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
			} else {
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
			}
			ptb.WithAffinity(corev1.Affinity{
				PodAntiAffinity: &antiAffinity,
			}, k8sbuilder.Merge)
		}

		// Compute containers
		ptb.WithContainers([]corev1.Container{*cb.Container()}, k8sbuilder.Merge)

		// Compute init containers
		if h.Spec.SetVMMaxMapCount == nil || *h.Spec.SetVMMaxMapCount {
			icb := k8sbuilder.NewContainerBuilder().WithContainer(&corev1.Container{
				Name: "configure-sysctl",
				Image: h.GetContainerImage(),
				ImagePullPolicy: h.Spec.ImagePullPolicy,
				SecurityContext: &corev1.SecurityContext{
					Privileged: pointer.Bool(true),
					RunAsUser: pointer.Int64(0),
				},
				Command: []string{
					"sysctl",
					"-w",
					"vm.max_map_count=262144",
				},
			})
			icb.WithResource(h.Spec.GlobalNodeGroup.InitContainerResources)

			ptb.WithInitContainers([]corev1.Container{*icb.Container()}, k8sbuilder.Merge)
		}

		// Compute volumes
		ptb.WithVolumes([]corev1.Volume{
			{
				Name: "node-tls",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: h.GetSecretNameForTlsTransport(),
					},
				},
			},
			{
				Name: "api-tls",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: h.GetSecretNameForTlsApi(),
					},
				},
			},
			{
				Name: "opensearch-config",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: h.GetNodeGroupConfigMapName(nodeGroup.Name),
						},
					},
				},
			},
		}, k8sbuilder.Merge)
		additionalVolume := make([]corev1.Volume, 0, len(h.Spec.GlobalNodeGroup.AdditionalVolumes))
		for _, vol := range h.Spec.GlobalNodeGroup.AdditionalVolumes {
			additionalVolume = append(additionalVolume, corev1.Volume{
				Name: vol.Name,
				VolumeSource: vol.VolumeSource,
			})
		}
		ptb.WithVolumes(additionalVolume, k8sbuilder.Merge)
		if h.GetSecretNameForSecurityConfig(&nodeGroup) != "" {
			ptb.WithVolumes([]corev1.Volume{
				{
					Name: "opensearch-security",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: h.GetSecretNameForSecurityConfig(&nodeGroup),
						},
					},
				},
			}, k8sbuilder.Merge)
		}
		if nodeGroup.Persistence != nil && nodeGroup.Persistence.VolumeClaimSpec == nil && nodeGroup.Persistence.Volume != nil {
			ptb.WithVolumes([]corev1.Volume{
				{
					Name: "opensearch-data",
					VolumeSource: *nodeGroup.Persistence.Volume,
				},
			}, k8sbuilder.Merge)
		}

		// Compute Security context
		ptb.WithSecurityContext(&corev1.PodSecurityContext{
			FSGroup: pointer.Int64(1000),
		}, k8sbuilder.Merge) 

		ptb.PodTemplate().Name = h.GetNodeGroupName(nodeGroup.Name)
		
		// Compute Statefullset
		sts = &appv1.StatefulSet {
			ObjectMeta: metav1.ObjectMeta{
				Namespace: h.Namespace,
				Name: h.GetNodeGroupName(nodeGroup.Name),
				Labels: h.Labels,
				Annotations: h.Annotations,
			},
			Spec: appv1.StatefulSetSpec{
				Replicas: pointer.Int32(nodeGroup.Replicas),
				// Start all node to create cluster
				PodManagementPolicy: appv1.ParallelPodManagement,
				ServiceName: h.GetNodeGroupServiceNameHeadless(nodeGroup.Name),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"cluster": h.Name,
						"nodeGroup": nodeGroup.Name,
					},		
				},

				Template: *ptb.PodTemplate(),
			},
		}

		// Compute persistence
		if nodeGroup.Persistence != nil && nodeGroup.Persistence.VolumeClaimSpec != nil {
			sts.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "opensearch-data",
					},
					Spec: *nodeGroup.Persistence.VolumeClaimSpec,
				},
			}
		} 		

		statefullsets = append(statefullsets, sts)
	}


	return statefullsets, nil
}

// GeneratePodDisruptionBudget permit to generate pod disruption budgets for each node group
func (h *Opensearch) GeneratePodDisruptionBudget() (podDisruptionBudgets []*policyv1.PodDisruptionBudget, err error) {
	podDisruptionBudgets = make([]*policyv1.PodDisruptionBudget, 0, len(h.Spec.NodeGroups))
	var (
		pdb *policyv1.PodDisruptionBudget
	)


	maxUnavailable := intstr.FromInt(1)
	for _, nodeGroup := range h.Spec.NodeGroups {
		pdb = &policyv1.PodDisruptionBudget{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: h.Namespace,
				Name: h.GetNodeGroupPDBName(nodeGroup.Name),
				Labels: h.Labels,
				Annotations: h.Annotations,
			},
			Spec: policyv1.PodDisruptionBudgetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"cluster": h.Name,
						"nodeGroup": nodeGroup.Name,
					},
				},
			},
		}

		// Merge with specified
		if err = helper.Merge(&pdb.Spec, nodeGroup.PodDisruptionBudgetSpec, funk.Get(h.Spec.GlobalNodeGroup, "PodDisruptionBudgetSpec")); err != nil {
			return nil, errors.Wrap(err, "Error when merge pod disruption spec")
		}
		if pdb.Spec.MinAvailable == nil && pdb.Spec.MaxUnavailable == nil {
			pdb.Spec.MaxUnavailable = &maxUnavailable
		}

		podDisruptionBudgets = append(podDisruptionBudgets, pdb)
	}
	
	return podDisruptionBudgets, nil
}




// isMasterRole return true if nodegroup have `cluster_manager` role
func (h *Opensearch) IsMasterRole(nodeGroup *NodeGroupSpec) bool {
	return funk.Contains(nodeGroup.Roles, "cluster_manager")
}








