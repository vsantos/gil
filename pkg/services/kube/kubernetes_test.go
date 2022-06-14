package kube

import (
	"context"
	"gil/calculator"
	"testing"

	// client "github.com/kubernetes-sdk-for-go-101/pkg/client"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/magiconair/properties/assert"
)

func TestPercentageChange(t *testing.T) {
	assert.Equal(t, PercentageChange(10, 100), float32(10))
	assert.Equal(t, PercentageChange(15, 372), float32(55.800003))
	assert.Equal(t, PercentageChange(37, 3120), float32(1154.4))
	assert.Equal(t, PercentageChange(14.13, 123), float32(17.379902))
	assert.Equal(t, PercentageChange(71.38, 15), float32(10.706999))
	assert.Equal(t, PercentageChange(83.11, 1), float32(0.8311))
}

func TestCalculateCPUPercentageOfUsage(t *testing.T) {
	assert.Equal(t, CalculateCPUPercentageOfUsage(0.3, 16), float32(1.8750001))
	assert.Equal(t, CalculateCPUPercentageOfUsage(0.1, 16), float32(0.625))
	assert.Equal(t, CalculateCPUPercentageOfUsage(1, 8), float32(12.5))
	assert.Equal(t, CalculateCPUPercentageOfUsage(0.43, 4), float32(10.75))
}
func TestCalculateMemPercentageOfUsage(t *testing.T) {
	assert.Equal(t, CalculateMemPercentageOfUsage(float32(8589934592), float32(2)), float32(400))
	assert.Equal(t, CalculateMemPercentageOfUsage(float32(8589934592), float32(8)), float32(100))
	assert.Equal(t, CalculateMemPercentageOfUsage(float32(268435456), float32(16)), float32(1.56))
	assert.Equal(t, CalculateMemPercentageOfUsage(float32(2147483648), float32(2.1)), float32(95.24))
}

func TestCalculatePodPriceByUsage(t *testing.T) {
	type pricesCase struct {
		percent  float32
		prices   calculator.NodePrice
		expected calculator.NodePrice
	}

	pCases := []pricesCase{
		{
			percent: 0.3,
			prices: calculator.NodePrice{
				Hourly:  0.94,
				Daily:   1.31,
				Weekly:  5.123,
				Monthly: 141.1231,
				Yearly:  12345.21,
			},
			expected: calculator.NodePrice{
				Hourly:  0.0028,
				Daily:   0.0039,
				Weekly:  0.0154,
				Monthly: 0.4234,
				Yearly:  37.0356,
			},
		},
		{
			percent: 13.41,
			prices: calculator.NodePrice{
				Hourly:  1.12,
				Daily:   91.12,
				Weekly:  102.12,
				Monthly: 12.1,
				Yearly:  41.12,
			},
			expected: calculator.NodePrice{
				Hourly:  0.1502,
				Daily:   12.2192,
				Weekly:  13.6943,
				Monthly: 1.6226,
				Yearly:  5.5142,
			},
		},
		{
			percent: 100,
			prices: calculator.NodePrice{
				Hourly:  312,
				Daily:   7488,
				Weekly:  52416,
				Monthly: 1572480,
				Yearly:  18869760,
			},
			expected: calculator.NodePrice{
				Hourly:  312,
				Daily:   7488,
				Weekly:  52416,
				Monthly: 1572480,
				Yearly:  18869760,
			},
		},
	}

	for _, tc := range pCases {
		assert.Equal(t, CalculatePodPriceByUsage(tc.percent, tc.prices), tc.expected)
	}
}

func TestGetCPURequest(t *testing.T) {
	type resourceCases struct {
		deployment  v1.Deployment
		expectedCPU float32
		err         error
	}

	// d.Spec.Template.Spec.Containers[0]
	tCases := []resourceCases{
		{
			deployment: v1.Deployment{
				Spec: v1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Resources: corev1.ResourceRequirements{},
								},
							},
						},
					},
				},
			},
			expectedCPU: 0,
			err:         nil,
		},
		{
			deployment: v1.Deployment{
				Spec: v1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Resources: corev1.ResourceRequirements{
										Requests: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("1"),
											corev1.ResourceMemory: resource.MustParse("10Gi"),
										},
									},
								},
							},
						},
					},
				},
			},
			expectedCPU: 1,
			err:         nil,
		},
		{
			deployment: v1.Deployment{
				Spec: v1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Resources: corev1.ResourceRequirements{
										Requests: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("1000m"),
											corev1.ResourceMemory: resource.MustParse("10Gi"),
										},
									},
								},
							},
						},
					},
				},
			},
			expectedCPU: 1,
			err:         nil,
		},
		{
			deployment: v1.Deployment{
				Spec: v1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Resources: corev1.ResourceRequirements{
										Requests: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("0.1"),
											corev1.ResourceMemory: resource.MustParse("10Gi"),
										},
									},
								},
							},
						},
					},
				},
			},
			expectedCPU: 0.1,
			err:         nil,
		},
		{
			deployment: v1.Deployment{
				Spec: v1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Resources: corev1.ResourceRequirements{
										Requests: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("100m"),
											corev1.ResourceMemory: resource.MustParse("10Gi"),
										},
									},
								},
							},
						},
					},
				},
			},
			expectedCPU: 0.1,
			err:         nil,
		},
	}

	for _, tc := range tCases {
		result, err := GetCPURequest(tc.deployment)
		assert.Equal(t, err, nil)
		assert.Equal(t, result, tc.expectedCPU)
	}

}

func TestGetMemoryRequest(t *testing.T) {
	type resourceCases struct {
		deployment          v1.Deployment
		expectedMemoryBytes int64
		err                 error
	}

	// d.Spec.Template.Spec.Containers[0]
	tCases := []resourceCases{
		{
			deployment: v1.Deployment{
				Spec: v1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Resources: corev1.ResourceRequirements{},
								},
							},
						},
					},
				},
			},
			expectedMemoryBytes: 0,
			err:                 nil,
		},
		{
			deployment: v1.Deployment{
				Spec: v1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Resources: corev1.ResourceRequirements{
										Requests: corev1.ResourceList{
											corev1.ResourceMemory: resource.MustParse("1Gi"),
										},
									},
								},
							},
						},
					},
				},
			},
			expectedMemoryBytes: 1073741824,
			err:                 nil,
		},
		{
			deployment: v1.Deployment{
				Spec: v1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Resources: corev1.ResourceRequirements{
										Requests: corev1.ResourceList{
											corev1.ResourceMemory: resource.MustParse("1024Mi"),
										},
									},
								},
							},
						},
					},
				},
			},
			expectedMemoryBytes: 1073741824,
			err:                 nil,
		},
	}

	for _, tc := range tCases {
		result, err := GetMemoryRequest(tc.deployment)
		assert.Equal(t, err, nil)
		assert.Equal(t, result, tc.expectedMemoryBytes)
	}
}

func TestGetNodes(t *testing.T) {
	var c kubernetes.Interface
	c = fake.NewSimpleClientset(&corev1.Node{
		TypeMeta: metav1.TypeMeta{
			Kind: "node",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "fake-node-hostname",
		},
		Spec: corev1.NodeSpec{},
	})

	nodes, err := GetNodes(c, context.TODO())
	assert.Equal(t, err, nil)
	assert.Equal(t, len(nodes), 1)
	assert.Equal(t, nodes[0].Name, "fake-node-hostname")
}
