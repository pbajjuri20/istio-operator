package conversion

import (
	v1 "github.com/maistra/istio-operator/pkg/apis/maistra/v1"
	v2 "github.com/maistra/istio-operator/pkg/apis/maistra/v2"
	"github.com/maistra/istio-operator/pkg/controller/versions"
)

func populatePrometheusAddonValues(in *v2.ControlPlaneSpec, values map[string]interface{}) (reterr error) {
	prometheus := in.Addons.Metrics.Prometheus
	if prometheus == nil {
		return nil
	}
	prometheusValues := make(map[string]interface{})
	if prometheus.Enabled != nil {
		if err := setHelmBoolValue(prometheusValues, "enabled", *prometheus.Enabled); err != nil {
			return err
		}
	}
	defer func() {
		if reterr == nil {
			if len(prometheusValues) > 0 {
				if err := setHelmValue(values, "prometheus", prometheusValues); err != nil {
					reterr = err
				}
			}
		}
	}()
	// install takes precedence
	if prometheus.Install == nil {
		if prometheus.Address != nil {
			return setHelmStringValue(values, "kiali.prometheusAddr", *prometheus.Address)
		}
		return nil
	}

	if prometheus.Install.Config.Retention != "" {
		if err := setHelmStringValue(prometheusValues, "retention", prometheus.Install.Config.Retention); err != nil {
			return err
		}
	}
	if prometheus.Install.Config.ScrapeInterval != "" {
		if err := setHelmStringValue(prometheusValues, "scrapeInterval", prometheus.Install.Config.ScrapeInterval); err != nil {
			return err
		}
	}
	if prometheus.Install.UseTLS != nil {
		if in.Version == "" || in.Version == versions.V1_0.String() || in.Version == versions.V1_1.String() {
			if err := setHelmBoolValue(prometheusValues, "security.enabled", *prometheus.Install.UseTLS); err != nil {
				return err
			}
		} else {
			if err := setHelmBoolValue(prometheusValues, "provisionPrometheusCert", *prometheus.Install.UseTLS); err != nil {
				return err
			}
		}
	}

	if err := populateComponentServiceValues(&prometheus.Install.Service, prometheusValues); err != nil {
		return err
	}

	return nil
}

func populatePrometheusAddonConfig(in *v1.HelmValues, out *v2.PrometheusAddonConfig) (bool, error) {
	rawPrometheusValues, ok, err := in.GetMap("prometheus")
	if err != nil {
		return false, err
	} else if !ok || len(rawPrometheusValues) == 0 {
		// nothing to do
		// check to see if grafana.Address should be set
		if address, ok, err := in.GetString("kiali.prometheusAddr"); ok {
			// If grafana URL is set, assume we're using an existing grafana install
			out.Address = &address
			return true, nil
		} else if err != nil {
			return false, err
		}
		return false, nil
	}
	prometheusValues := v1.NewHelmValues(rawPrometheusValues)

	prometheus := out

	if enabled, ok, err := prometheusValues.GetBool("enabled"); ok {
		prometheus.Enabled = &enabled
	} else if err != nil {
		return false, err
	}

	install := &v2.PrometheusInstallConfig{}
	setInstall := false

	if retention, ok, err := prometheusValues.GetString("retention"); ok {
		install.Config.Retention = retention
		setInstall = true
	} else if err != nil {
		return false, err
	}
	if scrapeInterval, ok, err := prometheusValues.GetString("scrapeInterval"); ok {
		install.Config.ScrapeInterval = scrapeInterval
		setInstall = true
	} else if err != nil {
		return false, err
	}

	if securityEnabled, ok, err := prometheusValues.GetBool("security.enabled"); ok {
		// v1_0 and v1_0
		install.UseTLS = &securityEnabled
		setInstall = true
	} else if err != nil {
		return false, err
	} else if provisionPrometheusCert, ok, err := prometheusValues.GetBool("provisionPrometheusCert"); ok {
		// v2_0
		install.UseTLS = &provisionPrometheusCert
		setInstall = true
	} else if err != nil {
		return false, err
	}
	if applied, err := populateComponentServiceConfig(prometheusValues, &install.Service); err == nil {
		setInstall = setInstall || applied
	} else {
		return false, err
	}

	if setInstall {
		prometheus.Install = install
	}

	return true, nil
}
