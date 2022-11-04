package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/webcenter-fr/opensearch-operator/pkg/test"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
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
					"opensearch.yml": `node.value: test
node.value2: test`,
					"log4j.yml": "log.test: test\n",
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

	configMaps, err := o.GenerateConfigMaps()
	assert.NoError(t, err)
	test.EqualFromYamlFile(t, "../../fixture/api/os-configmap.yml", configMaps[0])
}

func TestGenerateIngress(t *testing.T) {
	var (
		err error
		o *Opensearch
		i *networkingv1.Ingress
	)
 
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


	assert.NoError(t, err)
	test.EqualFromYamlFile(t, "../../fixture/api/os-ingress-with-target.yml", i)

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

	assert.NoError(t, err)
	test.EqualFromYamlFile(t, "../../fixture/api/os-ingress-without-target.yml", i)

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

	assert.NoError(t, err)
	test.EqualFromYamlFile(t, "../../fixture/api/os-ingress-with-all-options.yml", i)

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
	_, err = o.GenerateIngress()
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

func TestGenerateStatefullset(t *testing.T) {

	var (
		o *Opensearch
		err error
		sts []*appv1.StatefulSet
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
	assert.NoError(t, err)
	test.EqualFromYamlFile(t, "../../fixture/api/os-statefullset-all.yml", sts[0])

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

	assert.NoError(t, err)
	test.EqualFromYamlFile(t, "../../fixture/api/os-statefullset-master.yml", sts[0])
	test.EqualFromYamlFile(t, "../../fixture/api/os-statefullset-data.yml", sts[1])
	test.EqualFromYamlFile(t, "../../fixture/api/os-statefullset-client.yml", sts[2])
}