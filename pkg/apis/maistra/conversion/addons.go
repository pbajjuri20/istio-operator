package conversion

import (
	"fmt"

	v1 "github.com/maistra/istio-operator/pkg/apis/maistra/v1"
	v2 "github.com/maistra/istio-operator/pkg/apis/maistra/v2"
)

func populateAddonsValues(in *v2.ControlPlaneSpec, values map[string]interface{}) error {
	if in.Addons == nil {
		return nil
	}

	if in.Addons.Visualization != nil {
		// do kiali first, so it doesn't override any kiali.* settings added by other addons (e.g. prometheus, grafana, jaeger)
		// XXX: not sure how important this is, as these settings should be updated as part of reconcilation
		if in.Addons.Visualization.Kiali != nil {
			if err := populateKialiAddonValues(in.Addons.Visualization.Kiali, values); err != nil {
				return err
			}
		}
		if in.Addons.Visualization.Grafana != nil {
			if err := populateGrafanaAddonValues(in.Addons.Visualization.Grafana, values); err != nil {
				return err
			}
		}
	}

	if in.Addons.Metrics != nil {
		if in.Addons.Metrics.Prometheus != nil {
			if err := populatePrometheusAddonValues(in, values); err != nil {
				return err
			}
		}
	}

	// Tracing
	if in.Addons.Tracing != nil {
		if in.Addons.Tracing.Sampling != nil {
			if err := setHelmFloatValue(values, "pilot.traceSampling", float64(*in.Addons.Tracing.Sampling)/100.0); err != nil {
				return err
			}
		}
		switch in.Addons.Tracing.Type {
		case v2.TracerTypeNone:
			if err := setHelmBoolValue(values, "tracing.enabled", false); err != nil {
				return err
			}
			if err := setHelmBoolValue(values, "global.enableTracing", false); err != nil {
				return err
			}
			if err := setHelmStringValue(values, "tracing.provider", "none"); err != nil {
				return err
			}
			if err := setHelmStringValue(values, "global.proxy.tracer", "none"); err != nil {
				return err
			}
		case v2.TracerTypeJaeger:
			if err := setHelmValue(values, "tracing.provider", "jaeger"); err != nil {
				return err
			}
			if err := setHelmBoolValue(values, "tracing.enabled", true); err != nil {
				return err
			}
			if err := setHelmBoolValue(values, "global.enableTracing", true); err != nil {
				return err
			}
			if err := setHelmStringValue(values, "global.proxy.tracer", "jaeger"); err != nil {
				return err
			}
		case "":
			// nothing to do
		default:
			return fmt.Errorf("Unknown tracer type: %s", in.Addons.Tracing.Type)
		}

		// always add values, even if they're not enabled to support profiles
		if err := populateJaegerAddonValues(in.Addons.Tracing.Jaeger, values); err != nil {
			return err
		}
	}

	if in.Addons.Misc != nil {
		if in.Addons.Misc.ThreeScale != nil {
			if err := populateThreeScaleAddonValues(in.Addons.Misc.ThreeScale, values); err != nil {
				return err
			}
		}
	}

	return nil
}

func populateAddonIngressValues(ingress *v2.ComponentIngressConfig, addonIngressValues map[string]interface{}) error {
	if ingress == nil {
		return nil
	}
	if ingress.Enabled != nil {
		if err := setHelmBoolValue(addonIngressValues, "enabled", *ingress.Enabled); err != nil {
			return err
		}
		if !*ingress.Enabled {
			return nil
		}
	}

	if ingress.ContextPath != "" {
		if err := setHelmStringValue(addonIngressValues, "contextPath", ingress.ContextPath); err != nil {
			return err
		}
	}
	if len(ingress.Hosts) > 0 {
		if err := setHelmStringSliceValue(addonIngressValues, "hosts", ingress.Hosts); err != nil {
			return err
		}
	}
	if len(ingress.Metadata.Annotations) > 0 {
		if err := setHelmStringMapValue(addonIngressValues, "annotations", ingress.Metadata.Annotations); err != nil {
			return err
		}
	}
	if len(ingress.Metadata.Labels) > 0 {
		if err := setHelmStringMapValue(addonIngressValues, "labels", ingress.Metadata.Labels); err != nil {
			return err
		}
	}
	if len(ingress.TLS.GetContent()) > 0 {
		if err := setHelmValue(addonIngressValues, "tls", ingress.TLS.GetContent()); err != nil {
			return err
		}
	}
	return nil
}

func populateAddonsConfig(in *v1.HelmValues, out *v2.ControlPlaneSpec) error {
	addonsConfig := &v2.AddonsConfig{}
	setAddons := false
	visualization := &v2.VisualizationAddonsConfig{}
	setVisualization := false
	kiali := &v2.KialiAddonConfig{}
	if updated, err := populateKialiAddonConfig(in, kiali); updated {
		visualization.Kiali = kiali
		setVisualization = true
	} else if err != nil {
		return err
	}
	prometheus := &v2.PrometheusAddonConfig{}
	if updated, err := populatePrometheusAddonConfig(in, prometheus); updated {
		addonsConfig.Metrics = &v2.MetricsAddonsConfig{
			Prometheus: prometheus,
		}
		setAddons = true
	} else if err != nil {
		return err
	}
	tracing := &v2.TracingConfig{}
	if updated, err := populateTracingAddonConfig(in, tracing); updated {
		addonsConfig.Tracing = tracing
		setAddons = true
	} else if err != nil {
		return err
	}
	grafana := &v2.GrafanaAddonConfig{}
	if updated, err := populateGrafanaAddonConfig(in, grafana); updated {
		visualization.Grafana = grafana
		setVisualization = true
	} else if err != nil {
		return err
	}
	misc := &v2.MiscAddonsConfig{}
	if updated, err := populateMiscAddonsConfig(in, misc); updated {
		addonsConfig.Misc = misc
		setAddons = true
	} else if err != nil {
		return err
	}

	if setVisualization {
		addonsConfig.Visualization = visualization
		setAddons = true
	}
	if setAddons {
		out.Addons = addonsConfig
	}

	// HACK - remove grafana component's runtime env, as it is incorporated into
	// the grafana config directly
	if out.Runtime != nil && out.Runtime.Components != nil {
		if grafanaComponentConfig, ok := out.Runtime.Components[v2.ControlPlaneComponentNameGrafana]; ok && grafanaComponentConfig.Container != nil {
			grafanaComponentConfig.Container.Env = nil
		}
	}

	return nil
}

func populateMiscAddonsConfig(in *v1.HelmValues, out *v2.MiscAddonsConfig) (bool, error) {
	miscConfig := out
	if err := populateThreeScaleAddonConfig(in, miscConfig); err != nil {
		return false, err
	}

	return miscConfig.ThreeScale != nil, nil
}

func populateAddonIngressConfig(in *v1.HelmValues, out *v2.ComponentIngressConfig) (bool, error) {
	setValues := false
	if enabled, ok, err := in.GetBool("enabled"); ok {
		out.Enabled = &enabled
		setValues = true
	} else if err != nil {
		return false, err
	}

	if contextPath, ok, err := in.GetString("contextPath"); ok {
		out.ContextPath = contextPath
		setValues = true
	} else if err != nil {
		return false, err
	}
	if hosts, ok, err := in.GetStringSlice("hosts"); ok {
		out.Hosts = append([]string{}, hosts...)
		setValues = true
	} else if err != nil {
		return false, err
	}

	if rawAnnotations, ok, err := in.GetMap("annotations"); ok && len(rawAnnotations) > 0 {
		if err := setMetadataAnnotations(rawAnnotations, &out.Metadata); err != nil {
			return false, err
		}
		setValues = true
	} else if err != nil {
		return false, err
	}

	if rawLabels, ok, err := in.GetMap("labels"); ok && len(rawLabels) > 0 {
		if err := setMetadataLabels(rawLabels, &out.Metadata); err != nil {
			return false, err
		}
		setValues = true
	} else if err != nil {
		return false, err
	}

	if tls, ok, err := in.GetMap("tls"); ok && len(tls) > 0 {
		out.TLS = v1.NewHelmValues(tls)
		setValues = true
	} else if err != nil {
		return false, err
	}

	return setValues, nil
}
