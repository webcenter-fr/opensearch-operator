package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)


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