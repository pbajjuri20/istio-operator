package conversion

import (
	"reflect"
	"testing"

	v1 "github.com/maistra/istio-operator/pkg/apis/maistra/v1"
	v2 "github.com/maistra/istio-operator/pkg/apis/maistra/v2"
	"github.com/maistra/istio-operator/pkg/controller/versions"
)

var (
	kialiTestNodePort = int32(12345)
)

var kialiTestCases = []conversionTestCase{
	{
		name: "nil." + versions.V2_0.String(),
		spec: &v2.ControlPlaneSpec{
			Version: versions.V2_0.String(),
			Addons: &v2.AddonsConfig{
				Visualization: v2.VisualizationAddonsConfig{
					Kiali: nil,
				},
			},
		},
		isolatedIstio: v1.NewHelmValues(map[string]interface{}{}),
		completeIstio: v1.NewHelmValues(map[string]interface{}{
			"global": map[string]interface{}{
				"useMCP": true,
				"multiCluster": map[string]interface{}{
					"enabled": false,
				},
				"meshExpansion": map[string]interface{}{
					"enabled": false,
					"useILB":  false,
				},
			},
		}),
	},
	{
		name: "defaults." + versions.V2_0.String(),
		spec: &v2.ControlPlaneSpec{
			Version: versions.V2_0.String(),
			Addons: &v2.AddonsConfig{
				Visualization: v2.VisualizationAddonsConfig{
					Kiali: &v2.KialiAddonConfig{},
				},
			},
		},
		isolatedIstio: v1.NewHelmValues(map[string]interface{}{}),
		completeIstio: v1.NewHelmValues(map[string]interface{}{
			"global": map[string]interface{}{
				"useMCP": true,
				"multiCluster": map[string]interface{}{
					"enabled": false,
				},
				"meshExpansion": map[string]interface{}{
					"enabled": false,
					"useILB":  false,
				},
			},
		}),
	},
	{
		name: "simple." + versions.V2_0.String(),
		spec: &v2.ControlPlaneSpec{
			Version: versions.V2_0.String(),
			Addons: &v2.AddonsConfig{
				Visualization: v2.VisualizationAddonsConfig{
					Kiali: &v2.KialiAddonConfig{
						Enablement: v2.Enablement{
							Enabled: &featureEnabled,
						},
						Name: "my-kiali",
					},
				},
			},
		},
		isolatedIstio: v1.NewHelmValues(map[string]interface{}{
			"kiali": map[string]interface{}{
				"enabled":      true,
				"resourceName": "my-kiali",
			},
		}),
		completeIstio: v1.NewHelmValues(map[string]interface{}{
			"global": map[string]interface{}{
				"useMCP": true,
				"multiCluster": map[string]interface{}{
					"enabled": false,
				},
				"meshExpansion": map[string]interface{}{
					"enabled": false,
					"useILB":  false,
				},
			},
		}),
	},
	{
		name: "install.defaults." + versions.V2_0.String(),
		spec: &v2.ControlPlaneSpec{
			Version: versions.V2_0.String(),
			Addons: &v2.AddonsConfig{
				Visualization: v2.VisualizationAddonsConfig{
					Kiali: &v2.KialiAddonConfig{
						Name:    "my-kiali",
						Install: &v2.KialiInstallConfig{},
					},
				},
			},
		},
		isolatedIstio: v1.NewHelmValues(map[string]interface{}{
			"kiali": map[string]interface{}{
				"resourceName": "my-kiali",
			},
		}),
		completeIstio: v1.NewHelmValues(map[string]interface{}{
			"global": map[string]interface{}{
				"useMCP": true,
				"multiCluster": map[string]interface{}{
					"enabled": false,
				},
				"meshExpansion": map[string]interface{}{
					"enabled": false,
					"useILB":  false,
				},
			},
		}),
	},
	{
		name: "install.config.simple." + versions.V2_0.String(),
		spec: &v2.ControlPlaneSpec{
			Version: versions.V2_0.String(),
			Addons: &v2.AddonsConfig{
				Visualization: v2.VisualizationAddonsConfig{
					Kiali: &v2.KialiAddonConfig{
						Name: "my-kiali",
						Install: &v2.KialiInstallConfig{
							Dashboard: v2.KialiDashboardConfig{
								EnableGrafana:    &featureEnabled,
								EnablePrometheus: &featureEnabled,
								EnableTracing:    &featureDisabled,
								ViewOnly:         &featureEnabled,
							},
						},
					},
				},
			},
		},
		isolatedIstio: v1.NewHelmValues(map[string]interface{}{
			"kiali": map[string]interface{}{
				"resourceName": "my-kiali",
				"dashboard": map[string]interface{}{
					"enableGrafana":    true,
					"enablePrometheus": true,
					"enableTracing":    false,
					"viewOnlyMode":     true,
				},
			},
		}),
		completeIstio: v1.NewHelmValues(map[string]interface{}{
			"global": map[string]interface{}{
				"useMCP": true,
				"multiCluster": map[string]interface{}{
					"enabled": false,
				},
				"meshExpansion": map[string]interface{}{
					"enabled": false,
					"useILB":  false,
				},
			},
		}),
	},
	{
		name: "install.service.misc." + versions.V2_0.String(),
		spec: &v2.ControlPlaneSpec{
			Version: versions.V2_0.String(),
			Addons: &v2.AddonsConfig{
				Visualization: v2.VisualizationAddonsConfig{
					Kiali: &v2.KialiAddonConfig{
						Name: "my-kiali",
						Install: &v2.KialiInstallConfig{
							Service: v2.ComponentServiceConfig{
								Metadata: v2.MetadataConfig{
									Annotations: map[string]string{
										"some-service-annotation": "service-annotation-value",
									},
									Labels: map[string]string{
										"some-service-label": "service-label-value",
									},
								},
							},
						},
					},
				},
			},
		},
		isolatedIstio: v1.NewHelmValues(map[string]interface{}{
			"kiali": map[string]interface{}{
				"resourceName": "my-kiali",
				"service": map[string]interface{}{
					"annotations": map[string]interface{}{
						"some-service-annotation": "service-annotation-value",
					},
					"labels": map[string]interface{}{
						"some-service-label": "service-label-value",
					},
				},
			},
		}),
		completeIstio: v1.NewHelmValues(map[string]interface{}{
			"global": map[string]interface{}{
				"useMCP": true,
				"multiCluster": map[string]interface{}{
					"enabled": false,
				},
				"meshExpansion": map[string]interface{}{
					"enabled": false,
					"useILB":  false,
				},
			},
		}),
	},
	{
		name: "install.service.ingress.defaults." + versions.V2_0.String(),
		spec: &v2.ControlPlaneSpec{
			Version: versions.V2_0.String(),
			Addons: &v2.AddonsConfig{
				Visualization: v2.VisualizationAddonsConfig{
					Kiali: &v2.KialiAddonConfig{
						Name: "my-kiali",
						Install: &v2.KialiInstallConfig{
							Service: v2.ComponentServiceConfig{
								Ingress: &v2.ComponentIngressConfig{},
							},
						},
					},
				},
			},
		},
		isolatedIstio: v1.NewHelmValues(map[string]interface{}{
			"kiali": map[string]interface{}{
				"resourceName": "my-kiali",
			},
		}),
		completeIstio: v1.NewHelmValues(map[string]interface{}{
			"global": map[string]interface{}{
				"useMCP": true,
				"multiCluster": map[string]interface{}{
					"enabled": false,
				},
				"meshExpansion": map[string]interface{}{
					"enabled": false,
					"useILB":  false,
				},
			},
		}),
	},
	{
		name: "install.service.ingress.full." + versions.V2_0.String(),
		spec: &v2.ControlPlaneSpec{
			Version: versions.V2_0.String(),
			Addons: &v2.AddonsConfig{
				Visualization: v2.VisualizationAddonsConfig{
					Kiali: &v2.KialiAddonConfig{
						Name: "my-kiali",
						Install: &v2.KialiInstallConfig{
							Service: v2.ComponentServiceConfig{
								Ingress: &v2.ComponentIngressConfig{
									Enablement: v2.Enablement{
										Enabled: &featureEnabled,
									},
									ContextPath: "/kiali",
									Hosts: []string{
										"kiali.example.com",
									},
									Metadata: v2.MetadataConfig{
										Annotations: map[string]string{
											"ingress-annotation": "ingress-annotation-value",
										},
										Labels: map[string]string{
											"ingress-label": "ingress-label-value",
										},
									},
									TLS: v1.NewHelmValues(map[string]interface{}{
										"termination": "reencrypt",
									}),
								},
							},
						},
					},
				},
			},
		},
		isolatedIstio: v1.NewHelmValues(map[string]interface{}{
			"kiali": map[string]interface{}{
				"resourceName": "my-kiali",
				"contextPath": "/kiali",
				"ingress": map[string]interface{}{
					"enabled":     true,
					"contextPath": "/kiali",
					"annotations": map[string]interface{}{
						"ingress-annotation": "ingress-annotation-value",
					},
					"labels": map[string]interface{}{
						"ingress-label": "ingress-label-value",
					},
					"hosts": []interface{}{
						"kiali.example.com",
					},
					"tls": map[string]interface{}{
						"termination": "reencrypt",
					},
				},
			},
		}),
		completeIstio: v1.NewHelmValues(map[string]interface{}{
			"global": map[string]interface{}{
				"useMCP": true,
				"multiCluster": map[string]interface{}{
					"enabled": false,
				},
				"meshExpansion": map[string]interface{}{
					"enabled": false,
					"useILB":  false,
				},
			},
		}),
	},
	{
		name: "install.service.nodeport." + versions.V2_0.String(),
		spec: &v2.ControlPlaneSpec{
			Version: versions.V2_0.String(),
			Addons: &v2.AddonsConfig{
				Visualization: v2.VisualizationAddonsConfig{
					Kiali: &v2.KialiAddonConfig{
						Name: "my-kiali",
						Install: &v2.KialiInstallConfig{
							Service: v2.ComponentServiceConfig{
								NodePort: &kialiTestNodePort,
							},
						},
					},
				},
			},
		},
		isolatedIstio: v1.NewHelmValues(map[string]interface{}{
			"kiali": map[string]interface{}{
				"resourceName": "my-kiali",
				"service": map[string]interface{}{
					"nodePort": map[string]interface{}{
						"enabled": true,
						"port":    12345,
					},
				},
			},
		}),
		completeIstio: v1.NewHelmValues(map[string]interface{}{
			"global": map[string]interface{}{
				"useMCP": true,
				"multiCluster": map[string]interface{}{
					"enabled": false,
				},
				"meshExpansion": map[string]interface{}{
					"enabled": false,
					"useILB":  false,
				},
			},
		}),
	},
}

func TestKialiConversionFromV2(t *testing.T) {
	for _, tc := range kialiTestCases {
		t.Run(tc.name, func(t *testing.T) {
			specCopy := tc.spec.DeepCopy()
			helmValues := v1.NewHelmValues(make(map[string]interface{}))
			if err := populateAddonsValues(specCopy, helmValues.GetContent()); err != nil {
				t.Fatalf("error converting to values: %s", err)
			}
			if !reflect.DeepEqual(tc.isolatedIstio.DeepCopy(), helmValues.DeepCopy()) {
				t.Errorf("unexpected output converting v2 to values:\n\texpected:\n%#v\n\tgot:\n%#v", tc.isolatedIstio.GetContent(), helmValues.GetContent())
			}
			specv2 := &v2.ControlPlaneSpec{}
			// use expected values
			helmValues = tc.isolatedIstio.DeepCopy()
			mergeMaps(tc.completeIstio.DeepCopy().GetContent(), helmValues.GetContent())
			if err := populateAddonsConfig(helmValues.DeepCopy(), specv2); err != nil {
				t.Fatalf("error converting from values: %s", err)
			}
			assertEquals(t, tc.spec.Addons, specv2.Addons)
		})
	}
}
