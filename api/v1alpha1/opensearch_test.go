package v1alpha1

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/yaml"
)

func TestGetNodeGroupName(t *testing.T) {
	o := &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Replicas: 3,
				},
				{
					Name: "data",
					Replicas: 1,
				},
			},
		},
	}


	assert.Equal(t, "test-master-os", o.GetNodeGroupName(o.Spec.NodeGroups[0].Name))
}

func TestGetNodeNames(t *testing.T) {
	o := &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Replicas: 3,
				},
				{
					Name: "data",
					Replicas: 1,
				},
			},
		},
	}

	expectedList := []string {
		"test-master-os-0",
		"test-master-os-1",
		"test-master-os-2",
		"test-data-os-0",
	}

	assert.Equal(t, expectedList, o.GetNodeNames())
}

func TestGetNodeGroupNodeNames(t *testing.T) {
	o := &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Replicas: 3,
				},
				{
					Name: "data",
					Replicas: 1,
				},
			},
		},
	}

	expectedList := []string {
		"test-master-os-0",
		"test-master-os-1",
		"test-master-os-2",
	}

	assert.Equal(t, expectedList, o.GetNodeGroupNodeNames(&o.Spec.NodeGroups[0]))
}

func TestGetSecretNameForTlsTransport(t *testing.T) {
	o := &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Replicas: 3,
				},
				{
					Name: "data",
					Replicas: 1,
				},
			},
		},
	}

	assert.Equal(t, "test-os-tls-transport", o.GetSecretNameForTlsTransport())
}

func TestGetSecretNameForTlsApi(t *testing.T) {
	o := &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{},
	}

	// With default value
	assert.Equal(t, "test-os-tls-api", o.GetSecretNameForTlsApi())

	// When specify Load balancer secret
	o.Spec.Endpoint = &EndpointSpec{
		LoadBalancer: &LoadBalancerSpec{
			Enabled: true,
			Tls: &TlsSpec{
				CertificateSecretRef: "my-secret",
			},
		},
	}
	assert.Equal(t, "my-secret", o.GetSecretNameForTlsApi())
}

func TestGetSecretNameForAdminCredentials(t *testing.T) {

	o := &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{},
	}

	assert.Equal(t, "test-os-credential", o.GetSecretNameForAdminCredentials())
	
}

func TestIsSelfManagedSecretForTlsApi(t *testing.T) {
	// With default settings
	o := &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{},
	}
	assert.True(t, o.IsSelfManagedSecretForTlsApi())

	// When Load balancer is enabled but without specify secrets
	o.Spec.Endpoint = &EndpointSpec{
		LoadBalancer: &LoadBalancerSpec{
			Enabled: true,
			Tls: &TlsSpec{
			},
		},
	}
	assert.True(t, o.IsSelfManagedSecretForTlsApi())

	// When load balancer is enabled and pecify secrets
	o.Spec.Endpoint.LoadBalancer.Tls.CertificateSecretRef = "my-secret"
	assert.False(t, o.IsSelfManagedSecretForTlsApi())

	// When load balancer is disabled and pecify secrets
	o.Spec.Endpoint.LoadBalancer.Enabled = false
	assert.True(t, o.IsSelfManagedSecretForTlsApi())

}

func TestGetNodeGroupConfigMapName(t *testing.T) {
	o := &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Replicas: 3,
				},
				{
					Name: "data",
					Replicas: 1,
				},
			},
		},
	}

	assert.Equal(t, "test-master-os-config", o.GetNodeGroupConfigMapName(o.Spec.NodeGroups[0].Name))
}

func TestGetGlobalServiceName(t *testing.T) {
	o := &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{},
	}

	assert.Equal(t, "test-os", o.GetGlobalServiceName())
}

func TestGetNodeGroupServiceName(t *testing.T) {
	o := &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Replicas: 3,
				},
				{
					Name: "data",
					Replicas: 1,
				},
			},
		},
	}

	assert.Equal(t, "test-master-os", o.GetNodeGroupServiceName(o.Spec.NodeGroups[0].Name))
}

func TestGetNodeGroupServiceNameHeadless(t *testing.T) {
	o := &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Replicas: 3,
				},
				{
					Name: "data",
					Replicas: 1,
				},
			},
		},
	}

	assert.Equal(t, "test-master-os-headless", o.GetNodeGroupServiceNameHeadless(o.Spec.NodeGroups[0].Name))
}

func TestGetNodeGroupPDBName(t *testing.T) {
	o := &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Replicas: 3,
				},
				{
					Name: "data",
					Replicas: 1,
				},
			},
		},
	}

	assert.Equal(t, "test-master-os", o.GetNodeGroupPDBName(o.Spec.NodeGroups[0].Name))
}

func TestIsIngressEnabled(t *testing.T) {

	// With default values
	o := &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{},
	}
	assert.False(t, o.IsIngressEnabled())

	// When Ingress is specified but disabled
	o.Spec.Endpoint = &EndpointSpec{
		Ingress: &IngressSpec{
			Enabled: false,
		},
	}
	assert.False(t, o.IsIngressEnabled())

	// When ingress is enabled
	o.Spec.Endpoint.Ingress.Enabled = true

}

func TestIsLoadBalancerEnabled(t *testing.T) {
	// With default values
	o := &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{},
	}
	assert.False(t, o.IsLoadBalancerEnabled())

	// When Load balancer is specified but disabled
	o.Spec.Endpoint = &EndpointSpec{
		LoadBalancer: &LoadBalancerSpec{
			Enabled: false,
		},
	}
	assert.False(t, o.IsLoadBalancerEnabled())

	// When Load balancer is specified and enabled
	o.Spec.Endpoint.LoadBalancer.Enabled = true
	assert.True(t, o.IsLoadBalancerEnabled())

}

func TestGetContainerImage(t *testing.T) {
	// With default values
	o := &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{},
	}
	assert.Equal(t, "public.ecr.aws/opensearchproject/opensearch:latest", o.GetContainerImage())

	// When version is specified
	o.Spec.Version = "v1"
	assert.Equal(t, "public.ecr.aws/opensearchproject/opensearch:v1", o.GetContainerImage())

	// When image is overwriten
	o.Spec.Image = "my-image"
	assert.Equal(t, "my-image:v1", o.GetContainerImage())
}

func TestGetConfigMaps(t *testing.T) {

	o := &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
			Labels: map[string]string{
				"label1": "value1",
			},
			Annotations: map[string]string{
				"anno1": "value1",
			},
		},
		Spec: OpensearchSpec{
			GlobalNodeGroup: GlobalNodeGroupSpec{
				Config: map[string]string{
					"opensearch.yml": `
node.value: test
node.value2: test`,
					"log4j.yml": `
log.test: test`,
				},
			},
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Config: map[string]string{
						"opensearch.yml": `
key.value: fake
node:
  value2: test2
  name: test
  roles:
    - 'master'`,
					},
				},
			},
		},
	}

	expectedConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test-master-os-config",
			Labels: map[string]string{
				"label1": "value1",
			},
			Annotations: map[string]string{
				"anno1": "value1",
			},
		},
		Data: map[string]string{
			"log4j.yml": `
log.test: test`,
			"opensearch.yml": `key:
    value: fake
node:
    name: test
    roles:
        - master
    value: test
    value2: test2
plugins:
    security:
        ssl:
            http:
                enabled: true
                keystore_filepath: certs/http/api.pfx
                keystore_type: PKCS12/PFX
                truststore_filepath: certs/http/api.pfx
                truststore_type: PKCS12/PFX
            transport:
                enforce_hostname_verification: true
                keystore_filepath: certs/transport/${hostname}.pfx
                keystore_type: PKCS12/PFX
                truststore_filepath: certs/transport/${hostanme}.pfx
                truststore_type: PKCS12/PFX
`,
		},
	}

	configMaps, err := o.GenerateConfigMaps()

	assert.NoError(t, err)
	assert.Equal(t, expectedConfigMap, configMaps[0])
}

func TestGenerateIngress(t *testing.T) {
	var (
		err error
		o *Opensearch
		i *networkingv1.Ingress
		expectedIngress *networkingv1.Ingress
	)

	pathType := networkingv1.PathTypePrefix
 
	// With default values
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{},
	}
	i, err = o.GenerateIngress()
	assert.NoError(t, err)
	assert.Nil(t, i)

	// When ingress is disabled
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			Endpoint: &EndpointSpec{
				Ingress: &IngressSpec{
					Enabled: false,
				},
			},
		},
	}
	i, err = o.GenerateIngress()
	assert.NoError(t, err)
	assert.Nil(t, i)

	// When ingress is enabled
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			Endpoint: &EndpointSpec{
				Ingress: &IngressSpec{
					Enabled: true,
					TargetNodeGroupName: "master",
					Host: "my-test.cluster.local",
				},
			},
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Replicas: 3,
				},
				{
					Name: "data",
					Replicas: 1,
				},
			},
		},
	}
	i, err = o.GenerateIngress()

	expectedIngress = &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
			Namespace: "default",
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/force-ssl-redirect": "true",
				"nginx.ingress.kubernetes.io/backend-protocol": "HTTPS",
			},
			Labels: map[string]string{},	
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: "my-test.cluster.local",
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path: "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "test-master-os",
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
					Hosts: []string{
						"my-test.cluster.local",
					},
				},
			},
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, expectedIngress, i)

	// When ingress is enabled without specify TargetNodeGroupName
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			Endpoint: &EndpointSpec{
				Ingress: &IngressSpec{
					Enabled: true,
					Host: "my-test.cluster.local",
				},
			},
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Replicas: 3,
				},
				{
					Name: "data",
					Replicas: 1,
				},
			},
		},
	}
	i, err = o.GenerateIngress()

	expectedIngress = &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
			Namespace: "default",
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/force-ssl-redirect": "true",
				"nginx.ingress.kubernetes.io/backend-protocol": "HTTPS",
			},
			Labels: map[string]string{},	
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: "my-test.cluster.local",
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path: "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "test-os",
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
					Hosts: []string{
						"my-test.cluster.local",
					},
				},
			},
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, expectedIngress, i)

	// When ingress is enabled and specify all options
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
			Labels: map[string]string{
				"globalLabel": "globalLabel",
			},
			Annotations: map[string]string{
				"globalAnnotation": "globalAnnotation",
			},
		},
		Spec: OpensearchSpec{
			Endpoint: &EndpointSpec{
				Ingress: &IngressSpec{
					Enabled: true,
					TargetNodeGroupName: "master",
					Host: "my-test.cluster.local",
					SecretRef: "my-secret",
					Labels: map[string]string{
						"ingressLabel": "ingressLabel",
					},
					Annotations: map[string]string{
						"annotationLabel": "annotationLabel",
					},
					IngressSpec: &networkingv1.IngressSpec{
						IngressClassName: pointer.String("toto"),
					},
				},
			},
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Replicas: 3,
				},
				{
					Name: "data",
					Replicas: 1,
				},
			},
		},
	}
	i, err = o.GenerateIngress()

	expectedIngress = &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
			Namespace: "default",
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/force-ssl-redirect": "true",
				"nginx.ingress.kubernetes.io/backend-protocol": "HTTPS",
				"globalAnnotation": "globalAnnotation",
				"annotationLabel": "annotationLabel",
			},
			Labels: map[string]string{
				"globalLabel": "globalLabel",
				"ingressLabel": "ingressLabel",
			},	
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: pointer.String("toto"),
			Rules: []networkingv1.IngressRule{
				{
					Host: "my-test.cluster.local",
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path: "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "test-master-os",
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
					Hosts: []string{
						"my-test.cluster.local",
					},
					SecretName: "my-secret",
				},
			},
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, expectedIngress, i)

	// When target nodeGroup not exist
	// When ingress is enabled
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			Endpoint: &EndpointSpec{
				Ingress: &IngressSpec{
					Enabled: true,
					TargetNodeGroupName: "master",
					Host: "my-test.cluster.local",
				},
			},
			NodeGroups: []NodeGroupSpec{
				{
					Name: "data",
					Replicas: 1,
				},
			},
		},
	}
	i, err = o.GenerateIngress()
	assert.Error(t, err)
}


func TestGenerateServices(t *testing.T) {

	var (
		err error
		services []*corev1.Service
		expectedService *corev1.Service
		o *Opensearch
	)
	// With default values
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{},
	}

	expectedService = &corev1.Service{
		 ObjectMeta: metav1.ObjectMeta{
			 Name: "test-os",
			 Namespace: "default",
		 },
		 Spec: corev1.ServiceSpec{
			 Type: corev1.ServiceTypeClusterIP,
			 SessionAffinity: corev1.ServiceAffinityNone,
			 Selector: map[string]string{
				 "cluster": "test",
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

	services, err = o.GenerateServices()
	assert.NoError(t, err)
	assert.NotEmpty(t, services)
	assert.Equal(t, 1, len(services))
	assert.Equal(t, expectedService, services[0])

	// When nodes Groups
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
				},
			},
		},
	}

	expectedService = &corev1.Service{
		 ObjectMeta: metav1.ObjectMeta{
			 Name: "test-master-os",
			 Namespace: "default",
		 },
		 Spec: corev1.ServiceSpec{
			 Type: corev1.ServiceTypeClusterIP,
			 SessionAffinity: corev1.ServiceAffinityNone,
			 Selector: map[string]string{
				 "cluster": "test",
				 "nodeGroup": "master",
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

	services, err = o.GenerateServices()
	assert.NoError(t, err)
	assert.NotEmpty(t, services)
	assert.Equal(t, 3, len(services))
	assert.Equal(t, expectedService, services[1])

	expectedService = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-master-os-headless",
			Namespace: "default",
			Annotations: map[string]string{
				"service.alpha.kubernetes.io/tolerate-unready-endpoints": "true",
			},
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "None",
			Type: corev1.ServiceTypeClusterIP,
			SessionAffinity: corev1.ServiceAffinityNone,
			PublishNotReadyAddresses: true,
			Selector: map[string]string{
				"cluster": "test",
				"nodeGroup": "master",
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

 assert.Equal(t, expectedService, services[2])

}

func TestGenerateLoadbalancer(t *testing.T) {

	var (
		err error
		service *corev1.Service
		expectedService *corev1.Service
		o *Opensearch
	)
	
	// With default values
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{},
	}

	service, err = o.GenerateLoadbalancer()
	assert.NoError(t, err)
	assert.Nil(t, service)

	// When load balancer is disabled
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			Endpoint: &EndpointSpec{
				LoadBalancer: &LoadBalancerSpec{
					Enabled: false,
				},
			},
		},
	}

	service, err = o.GenerateLoadbalancer()
	assert.NoError(t, err)
	assert.Nil(t, service)

	// When load balancer is enabled
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Replicas: 3,
				},
				{
					Name: "data",
					Replicas: 1,
				},
			},
			Endpoint: &EndpointSpec{
				LoadBalancer: &LoadBalancerSpec{
					Enabled: true,
					TargetNodeGroupName: "master",
				},
			},
		},
	}

	expectedService = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-os-lb",
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeLoadBalancer,
			SessionAffinity: corev1.ServiceAffinityNone,
			Selector: map[string]string{
				"cluster": "test",
				"nodeGroup": "master",
			},
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

	service, err = o.GenerateLoadbalancer()
	assert.NoError(t, err)
	assert.Equal(t, expectedService, service)

	// When load balancer is enabled without target node group
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Replicas: 3,
				},
				{
					Name: "data",
					Replicas: 1,
				},
			},
			Endpoint: &EndpointSpec{
				LoadBalancer: &LoadBalancerSpec{
					Enabled: true,
				},
			},
		},
	}

	expectedService = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-os-lb",
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeLoadBalancer,
			SessionAffinity: corev1.ServiceAffinityNone,
			Selector: map[string]string{
				"cluster": "test",
			},
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

	service, err = o.GenerateLoadbalancer()
	assert.NoError(t, err)
	assert.Equal(t, expectedService, service)

	// When load balancer is enabled with target node group that not exist
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			NodeGroups: []NodeGroupSpec{
				{
					Name: "data",
					Replicas: 1,
				},
			},
			Endpoint: &EndpointSpec{
				LoadBalancer: &LoadBalancerSpec{
					Enabled: true,
					TargetNodeGroupName: "master",
				},
			},
		},
	}

	expectedService = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-os-lb",
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeLoadBalancer,
			SessionAffinity: corev1.ServiceAffinityNone,
			Selector: map[string]string{
				"cluster": "test",
				"nodeGroup": "master",
			},
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

	_, err = o.GenerateLoadbalancer()
	assert.Error(t, err)
}


func TestGeneratePodDisruptionBudget(t *testing.T) {

	var (
		err error
		pdbs []*policyv1.PodDisruptionBudget
		expectedPdb *policyv1.PodDisruptionBudget
		o *Opensearch
	)

	maxUnavailable := intstr.FromInt(1)
	
	// With default values
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{},
	}

	pdbs, err = o.GeneratePodDisruptionBudget()
	assert.NoError(t, err)
	assert.Empty(t, pdbs)

	// When pdb spec not provided, default
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Replicas: 1,
				},
			},
		},
	}

	expectedPdb = &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-master-os",
			Namespace: "default",
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"cluster": "test",
					"nodeGroup": "master",
				},
			},
			MaxUnavailable: &maxUnavailable,
		},
	}

	pdbs, err = o.GeneratePodDisruptionBudget()
	assert.NoError(t, err)
	assert.Equal(t,1, len(pdbs))
	assert.Equal(t, expectedPdb, pdbs[0])

	// When Pdb is defined on global
	minUnavailable := intstr.FromInt(0)
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Replicas: 1,
				},
			},
			GlobalNodeGroup: GlobalNodeGroupSpec{
				PodDisruptionBudgetSpec: &policyv1.PodDisruptionBudgetSpec{
					MinAvailable: &minUnavailable,
					MaxUnavailable: nil,
				},
			},
		},
	}

	expectedPdb = &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-master-os",
			Namespace: "default",
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"cluster": "test",
					"nodeGroup": "master",
				},
			},
			MinAvailable: &minUnavailable,
		},
	}

	pdbs, err = o.GeneratePodDisruptionBudget()
	assert.NoError(t, err)
	assert.Equal(t,1, len(pdbs))
	assert.Equal(t, expectedPdb, pdbs[0])

	// When Pdb is defined on nodeGroup
	minUnavailable = intstr.FromInt(10)
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Replicas: 1,
					PodDisruptionBudgetSpec: &policyv1.PodDisruptionBudgetSpec{
						MinAvailable: &minUnavailable,
						MaxUnavailable: nil,
					},
				},
			},
			GlobalNodeGroup: GlobalNodeGroupSpec{
				PodDisruptionBudgetSpec: &policyv1.PodDisruptionBudgetSpec{
					MinAvailable: &minUnavailable,
					MaxUnavailable: nil,
				},
			},
		},
	}

	expectedPdb = &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-master-os",
			Namespace: "default",
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"cluster": "test",
					"nodeGroup": "master",
				},
			},
			MinAvailable: &minUnavailable,
		},
	}

	pdbs, err = o.GeneratePodDisruptionBudget()
	assert.NoError(t, err)
	assert.Equal(t,1, len(pdbs))
	assert.Equal(t, expectedPdb, pdbs[0])

}

func TestComputeJavaOpts(t *testing.T) {

	var o *Opensearch

	// With default values
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Replicas: 1,
				},
			},
		},
	}

	assert.Empty(t, o.computeJavaOpts(&o.Spec.NodeGroups[0]))

	// With global values
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			GlobalNodeGroup: GlobalNodeGroupSpec{
				Jvm: "-param1=1",
			},
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Replicas: 1,
				},
			},
		},
	}

	assert.Equal(t, "-param1=1", o.computeJavaOpts(&o.Spec.NodeGroups[0]))

	// With global and node group values
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			GlobalNodeGroup: GlobalNodeGroupSpec{
				Jvm: "-param1=1",
			},
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Replicas: 1,
					Jvm: "-xmx1G -xms1G",
				},
			},
		},
	}

	assert.Equal(t, "-param1=1 -xmx1G -xms1G", o.computeJavaOpts(&o.Spec.NodeGroups[0]))
}

func TestIsMasterRole(t *testing.T) {

	var o *Opensearch

	// With only master role
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Replicas: 1,
					Roles: []string{
						"cluster_manager",
					},
				},
			},
		},
	}

	assert.True(t, o.IsMasterRole(&o.Spec.NodeGroups[0]))

	// With multiple role
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Replicas: 1,
					Roles: []string{
						"data",
						"cluster_manager",
						"ingest",
					},
				},
			},
		},
	}

	assert.True(t, o.IsMasterRole(&o.Spec.NodeGroups[0]))

	// Without master role
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Replicas: 1,
					Roles: []string{
						"data",
						"ingest",
					},
				},
			},
		},
	}

	assert.False(t, o.IsMasterRole(&o.Spec.NodeGroups[0]))
}

func TestComputeInitialMasterNodes(t *testing.T) {
	var (
		o *Opensearch
	)

	// With only one master
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Replicas: 3,
					Roles: []string{
						"cluster_manager",
						"data",
						"ingest",
					},
				},
			},
		},
	}

	assert.Equal(t, "test-master-os-0 test-master-os-1 test-master-os-2", o.computeInitialMasterNodes())

	// With multiple node groups
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			NodeGroups: []NodeGroupSpec{
				{
					Name: "all",
					Replicas: 3,
					Roles: []string{
						"cluster_manager",
						"data",
						"ingest",
					},
				},
				{
					Name: "master",
					Replicas: 3,
					Roles: []string{
						"cluster_manager",
					},
				},
			},
		},
	}

	assert.Equal(t, "test-all-os-0 test-all-os-1 test-all-os-2 test-master-os-0 test-master-os-1 test-master-os-2", o.computeInitialMasterNodes())
}

func TestComputeDiscoverySeedHosts(t *testing.T) {
	var (
		o *Opensearch
	)

	// With only one master
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Replicas: 3,
					Roles: []string{
						"cluster_manager",
						"data",
						"ingest",
					},
				},
			},
		},
	}

	assert.Equal(t, "test-master-os-headless", o.computeDiscoverySeedHosts())

	// With multiple node groups
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			NodeGroups: []NodeGroupSpec{
				{
					Name: "all",
					Replicas: 3,
					Roles: []string{
						"cluster_manager",
						"data",
						"ingest",
					},
				},
				{
					Name: "master",
					Replicas: 3,
					Roles: []string{
						"cluster_manager",
					},
				},
			},
		},
	}

	assert.Equal(t, "test-all-os-headless test-master-os-headless", o.computeDiscoverySeedHosts())
}

func TestComputeRoles(t *testing.T) {
	roles := []string {
		"cluster_manager",
	}

	o := &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{},
	}

	expectedEnvs := []corev1.EnvVar {
		{
			Name: "node.cluster_manager",
			Value: "true",
		},
		{
			Name: "node.data",
			Value: "false",
		},
		{
			Name: "node.ingest",
			Value: "false",
		},
		{
			Name: "node.ml",
			Value: "false",
		},
		{
			Name: "node.remote_cluster_client",
			Value: "false",
		},
		{
			Name: "node.transform",
			Value: "false",
		},
	}

	assert.Equal(t, expectedEnvs, o.computeRoles(roles))
}

func TestComputeAntiAffinity(t *testing.T) {

	var (
		o *Opensearch
		expectedAntiAffinity *corev1.PodAntiAffinity
		err error
		antiAffinity *corev1.PodAntiAffinity
	)

	// With default values
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Replicas: 1,
				},
			},
		},
	}

	expectedAntiAffinity = &corev1.PodAntiAffinity{
		PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
			{
				Weight: 10,
				PodAffinityTerm: corev1.PodAffinityTerm{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"cluster": "test",
							"nodeGroup": "master",
						},
					},
					TopologyKey: "kubernetes.io/hostname",
				},
			},
		},
	}

	antiAffinity, err = o.computeAntiAffinity(&o.Spec.NodeGroups[0])
	assert.NoError(t, err )
	assert.Equal(t, expectedAntiAffinity, antiAffinity)


	// With global anti affinity
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			GlobalNodeGroup: GlobalNodeGroupSpec{
				AntiAffinity: &AntiAffinitySpec{
					Type: "hard",
					TopologyKey: "topology.kubernetes.io/zone",
				},
			},
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Replicas: 1,
				},
			},
		},
	}

	expectedAntiAffinity = &corev1.PodAntiAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
			{
				TopologyKey: "topology.kubernetes.io/zone",
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"cluster": "test",
						"nodeGroup": "master",
					},
				},
			},
		},
	}

	antiAffinity, err = o.computeAntiAffinity(&o.Spec.NodeGroups[0])
	assert.NoError(t, err )
	assert.Equal(t, expectedAntiAffinity, antiAffinity)

	// With global and node group anti affinity
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			GlobalNodeGroup: GlobalNodeGroupSpec{
				AntiAffinity: &AntiAffinitySpec{
					Type: "soft",
					TopologyKey: "topology.kubernetes.io/zone",
				},
			},
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Replicas: 1,
					AntiAffinity: &AntiAffinitySpec{
						Type: "hard",
					},
				},
			},
		},
	}

	expectedAntiAffinity = &corev1.PodAntiAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
			{
				TopologyKey: "topology.kubernetes.io/zone",
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"cluster": "test",
						"nodeGroup": "master",
					},
				},
			},
		},
	}

	antiAffinity, err = o.computeAntiAffinity(&o.Spec.NodeGroups[0])
	assert.NoError(t, err )
	assert.Equal(t, expectedAntiAffinity, antiAffinity)
}

func TestComputeEnvFroms(t *testing.T) {
	var (
		o *Opensearch
		expectedEnvFroms []corev1.EnvFromSource
	)

	// With default values
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Replicas: 1,
				},
			},
		},
	}

	assert.Empty(t, o.computeEnvFroms(&o.Spec.NodeGroups[0]))

	// When global envFrom
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			GlobalNodeGroup: GlobalNodeGroupSpec{
				EnvFrom: []corev1.EnvFromSource{
					{
						ConfigMapRef: &corev1.ConfigMapEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "test",
							},
						},
					},
				},
			},
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Replicas: 1,
				},
			},
		},
	}

	expectedEnvFroms = []corev1.EnvFromSource{
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "test",
				},
			},
		},
	}


	assert.Equal(t, expectedEnvFroms, o.computeEnvFroms(&o.Spec.NodeGroups[0]))

	// When global envFrom and node group envFrom
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			GlobalNodeGroup: GlobalNodeGroupSpec{
				EnvFrom: []corev1.EnvFromSource{
					{
						ConfigMapRef: &corev1.ConfigMapEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "test",
							},
						},
					},
					{
						ConfigMapRef: &corev1.ConfigMapEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "test2",
							},
						},
					},
				},
			},
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Replicas: 1,
					EnvFrom: []corev1.EnvFromSource{
						{
							ConfigMapRef: &corev1.ConfigMapEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "test",
								},
							},
						},
						{
							ConfigMapRef: &corev1.ConfigMapEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "test3",
								},
							},
						},
					},
				},
			},
		},
	}

	expectedEnvFroms = []corev1.EnvFromSource{
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "test",
				},
			},
		},
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "test3",
				},
			},
		},
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "test",
				},
			},
		},
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "test2",
				},
			},
		},
		
	}

	assert.Equal(t, expectedEnvFroms, o.computeEnvFroms(&o.Spec.NodeGroups[0]))
}

func TestGenerateStatefullset(t *testing.T) {

	var (
		o *Opensearch
		//expectedSts *appv1.StatefulSet
		err error
		sts []*appv1.StatefulSet
	)

	toYaml := func(s *appv1.StatefulSet) string {
		b, err := yaml.Marshal(s)
		if err != nil {
			panic(err)
		}

		return string(b)
	}

	// With default values
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			NodeGroups: []NodeGroupSpec{
				{
					Name: "all",
					Replicas: 1,
					Roles: []string{
						"cluster_manager",
						"data",
						"ingest",
					},
				},
			},
		},
	}

	sts, err = o.GenerateStatefullsets()

	spew.Print(toYaml(sts[0]))
	assert.NoError(t, err)
	//assert.Equal(t, "", toYaml(sts[0]))

	// With complex config
	o = &Opensearch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: OpensearchSpec{
			Version: "2.3.0",
			PluginsList: []string{
			  "repository-s3",
			},
			SetVMMaxMapCount: pointer.Bool(true),
			GlobalNodeGroup: GlobalNodeGroupSpec{
				AdditionalVolumes: []*VolumeSpec{
					{
						Name: "snapshot",
						VolumeMount: corev1.VolumeMount{
							MountPath: "/mnt/snapshot",
						},
						VolumeSource: corev1.VolumeSource{
							NFS: &corev1.NFSVolumeSource{
								Server: "nfsserver",
								Path: "/snapshot",
							},
						},
					},
				},
				InitContainerResources: &corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse("100m"),
						corev1.ResourceMemory: resource.MustParse("100Mi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse("300m"),
						corev1.ResourceMemory: resource.MustParse("500Mi"),
					},
				},
				SecurityRef: "opensearch-security",
				AntiAffinity: &AntiAffinitySpec{
					TopologyKey: "rack",
					Type: "hard",
				},
				Config: map[string]string{
					"log4.yaml": "my log4j",
				},
			},
			NodeGroups: []NodeGroupSpec{
				{
					Name: "master",
					Replicas: 3,
					Roles: []string{
						"cluster_manager",
					},
					Persistence: &PersistenceSpec{
						VolumeClaimSpec: &corev1.PersistentVolumeClaimSpec{
							StorageClassName: pointer.String("local-path"),
							AccessModes: []corev1.PersistentVolumeAccessMode{
								corev1.ReadWriteOnce,
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceStorage: resource.MustParse("5Gi"),
								},
							},
						},
					},
					Jvm: "-Xms1g -Xmx1g",
					Resources: &corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("1Gi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("2"),
							corev1.ResourceMemory: resource.MustParse("2Gi"),
						},
					},
				},
				{
					Name: "data",
					Replicas: 3,
					Roles: []string{
						"data",
					},
					Persistence: &PersistenceSpec{
						Volume: &corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/data/opensearch",
							},
						},
					},
					Jvm: "-Xms30g -Xmx30g",
					Resources: &corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("5"),
							corev1.ResourceMemory: resource.MustParse("30Gi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("8"),
							corev1.ResourceMemory: resource.MustParse("64Gi"),
						},
					},
					NodeSelector: map[string]string{
						"project": "opensearch",
					},
					Tolerations: []corev1.Toleration{
						{
							Key: "project",
							Operator: corev1.TolerationOpEqual,
							Value: "opensearch",
							Effect: corev1.TaintEffectNoSchedule,
						},
					},
				},
				{
					Name: "client",
					Replicas: 2,
					Roles: []string{
						"ingest",
					},
					Persistence: &PersistenceSpec{
						VolumeClaimSpec: &corev1.PersistentVolumeClaimSpec{
							StorageClassName: pointer.String("local-path"),
							AccessModes: []corev1.PersistentVolumeAccessMode{
								corev1.ReadWriteOnce,
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceStorage: resource.MustParse("5Gi"),
								},
							},
						},
					},
					Jvm: "-Xms2g -Xmx2g",
					Resources: &corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("2"),
							corev1.ResourceMemory: resource.MustParse("2Gi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("4"),
							corev1.ResourceMemory: resource.MustParse("4Gi"),
						},
					},
				},
			},
		},
	}

	sts, err = o.GenerateStatefullsets()

	spew.Print(toYaml(sts[1]))
	assert.NoError(t, err)
	assert.Fail(t, "test")
}