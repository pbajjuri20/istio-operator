package conversion

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	v1 "github.com/maistra/istio-operator/pkg/apis/maistra/v1"
	v2 "github.com/maistra/istio-operator/pkg/apis/maistra/v2"
	"github.com/maistra/istio-operator/pkg/controller/versions"
)

func populateTelemetryValues(in *v2.ControlPlaneSpec, values map[string]interface{}) error {
	telemetry := in.Telemetry
	if telemetry == nil || in.Telemetry.Type == "" {
		return nil
	}

	istiod := !(in.Version == "" || in.Version == versions.V1_0.String() || in.Version == versions.V1_1.String())

	if err := setHelmStringValue(values, "telemetry.implementation", string(in.Telemetry.Type)); err != nil {
		return nil
	}

	if telemetry.Type == v2.TelemetryTypeNone {
		if istiod {
			if err := setHelmBoolValue(values, "telemetry.enabled", false); err != nil {
				return err
			}
			if err := setHelmBoolValue(values, "telemetry.v1.enabled", false); err != nil {
				return err
			}
			if err := setHelmBoolValue(values, "telemetry.v2.enabled", false); err != nil {
				return err
			}
		}
		return setHelmBoolValue(values, "mixer.telemetry.enabled", false)
	}

	switch telemetry.Type {
	case v2.TelemetryTypeMixer:
		return populateMixerTelemetryValues(in, istiod, values)
	case v2.TelemetryTypeRemote:
		return populateRemoteTelemetryValues(in, istiod, values)
	case v2.TelemetryTypeIstiod:
		return populateIstiodTelemetryValues(in, values)
	}

	if istiod {
		return setHelmBoolValue(values, "telemetry.enabled", false)
	}
	setHelmBoolValue(values, "mixer.telemetry.enabled", false)
	return fmt.Errorf("Unknown telemetry type: %s", telemetry.Type)
}

func populateMixerTelemetryValues(in *v2.ControlPlaneSpec, istiod bool, values map[string]interface{}) error {
	mixer := in.Telemetry.Mixer
	if mixer == nil {
		mixer = &v2.MixerTelemetryConfig{}
	}

	// Make sure mixer is enabled
	if err := setHelmBoolValue(values, "mixer.enabled", true); err != nil {
		return err
	}

	v1TelemetryValues := make(map[string]interface{})
	if err := setHelmBoolValue(v1TelemetryValues, "enabled", true); err != nil {
		return err
	}

	if err := populateTelemetryBatchingValues(mixer.Batching, v1TelemetryValues); err != nil {
		return nil
	}

	if mixer.SessionAffinity != nil {
		if err := setHelmBoolValue(v1TelemetryValues, "sessionAffinityEnabled", *mixer.SessionAffinity); err != nil {
			return err
		}
	}

	if mixer.Adapters != nil {
		telemetryAdaptersValues := make(map[string]interface{})
		if mixer.Adapters.UseAdapterCRDs != nil {
			if err := setHelmBoolValue(telemetryAdaptersValues, "useAdapterCRDs", *mixer.Adapters.UseAdapterCRDs); err != nil {
				return err
			}
		}
		if mixer.Adapters.KubernetesEnv != nil {
			if err := setHelmBoolValue(telemetryAdaptersValues, "kubernetesenv.enabled", *mixer.Adapters.KubernetesEnv); err != nil {
				return err
			}
		}
		if mixer.Adapters.Stdio == nil {
			if err := setHelmBoolValue(telemetryAdaptersValues, "stdio.enabled", false); err != nil {
				return err
			}
		} else {
			if err := setHelmBoolValue(telemetryAdaptersValues, "stdio.enabled", true); err != nil {
				return err
			}
			if err := setHelmBoolValue(telemetryAdaptersValues, "stdio.outputAsJson", mixer.Adapters.Stdio.OutputAsJSON); err != nil {
				return err
			}
		}
		if mixer.Adapters.Prometheus == nil {
			if err := setHelmBoolValue(telemetryAdaptersValues, "prometheus.enabled", false); err != nil {
				return err
			}
		} else {
			if err := setHelmBoolValue(telemetryAdaptersValues, "prometheus.enabled", true); err != nil {
				return err
			}
			if mixer.Adapters.Prometheus.MetricsExpiryDuration != "" {
				if err := setHelmStringValue(telemetryAdaptersValues, "prometheus.metricsExpiryDuration", mixer.Adapters.Prometheus.MetricsExpiryDuration); err != nil {
					return err
				}
			}
		}
		if mixer.Adapters.Stackdriver == nil {
			if err := setHelmBoolValue(telemetryAdaptersValues, "stackdriver.enabled", false); err != nil {
				return err
			}
		} else {
			stackdriver := mixer.Adapters.Stackdriver
			if err := setHelmBoolValue(telemetryAdaptersValues, "stackdriver.enabled", true); err != nil {
				return err
			}
			if err := setHelmBoolValue(telemetryAdaptersValues, "stackdriver.contextGraph.enabled", stackdriver.EnableContextGraph); err != nil {
				return err
			}
			if err := setHelmBoolValue(telemetryAdaptersValues, "stackdriver.logging.enabled", stackdriver.EnableLogging); err != nil {
				return err
			}
			if err := setHelmBoolValue(telemetryAdaptersValues, "stackdriver.metrics.enabled", stackdriver.EnableMetrics); err != nil {
				return err
			}
			if stackdriver.Auth != nil {
				auth := stackdriver.Auth
				if err := setHelmBoolValue(telemetryAdaptersValues, "stackdriver.auth.appCredentials", auth.AppCredentials); err != nil {
					return err
				}
				if err := setHelmStringValue(telemetryAdaptersValues, "stackdriver.auth.apiKey", auth.APIKey); err != nil {
					return err
				}
				if err := setHelmStringValue(telemetryAdaptersValues, "stackdriver.auth.serviceAccountPath", auth.ServiceAccountPath); err != nil {
					return err
				}
			}
			if stackdriver.Tracer != nil {
				tracer := mixer.Adapters.Stackdriver.Tracer
				if err := setHelmIntValue(telemetryAdaptersValues, "stackdriver.tracer.sampleProbability", int64(tracer.SampleProbability)); err != nil {
					return err
				}
			}
		}
		if len(telemetryAdaptersValues) > 0 {
			if err := setHelmValue(values, "mixer.adapters", telemetryAdaptersValues); err != nil {
				return err
			}
		}
	}

	// set the telemetry values
	if istiod {
		var v2TelemetryValues map[string]interface{}
		if rawTelemetryValues, ok, err := unstructured.NestedFieldNoCopy(values, "telemetry"); ok {
			if v2TelemetryValues, ok = rawTelemetryValues.(map[string]interface{}); !ok {
				v2TelemetryValues = make(map[string]interface{})
			}
		} else if err != nil {
			return nil
		} else {
			v2TelemetryValues = make(map[string]interface{})
		}
		if err := setHelmBoolValue(v2TelemetryValues, "enabled", true); err != nil {
			return err
		}
		if err := setHelmBoolValue(v2TelemetryValues, "v1.enabled", true); err != nil {
			return err
		}
		if err := setHelmBoolValue(v2TelemetryValues, "v2.enabled", false); err != nil {
			return err
		}

		if err := setHelmValue(values, "telemetry", v2TelemetryValues); err != nil {
			return err
		}
		if len(v1TelemetryValues) > 0 {
			if err := setHelmValue(values, "mixer.telemetry", v1TelemetryValues); err != nil {
				return err
			}
		}
	} else {
		if len(v1TelemetryValues) > 0 {
			if err := setHelmValue(values, "mixer.telemetry", v1TelemetryValues); err != nil {
				return err
			}
		}
	}

	return nil
}

func populateTelemetryBatchingValues(in *v2.TelemetryBatchingConfig, telemetryBatchingValues map[string]interface{}) error {
	if in == nil {
		return nil
	}
	if in.MaxTime != "" {
		if err := setHelmStringValue(telemetryBatchingValues, "reportBatchMaxTime", in.MaxTime); err != nil {
			return err
		}
	}
	if in.MaxEntries != nil {
		return setHelmIntValue(telemetryBatchingValues, "reportBatchMaxEntries", int64(*in.MaxEntries))
	}
	return nil
}

func populateRemoteTelemetryValues(in *v2.ControlPlaneSpec, istiod bool, values map[string]interface{}) error {
	remote := in.Telemetry.Remote
	if remote == nil {
		remote = &v2.RemoteTelemetryConfig{}
	}

	// Make sure mixer is disabled
	if err := setHelmBoolValue(values, "mixer.enabled", false); err != nil {
		return err
	}

	if err := setHelmStringValue(values, "global.remoteTelemetryAddress", remote.Address); err != nil {
		return err
	}
	// XXX: this applies to both policy and telemetry
	if err := setHelmBoolValue(values, "global.createRemoteSvcEndpoints", remote.CreateService); err != nil {
		return err
	}

	v1TelemetryValues := make(map[string]interface{})
	if err := setHelmBoolValue(v1TelemetryValues, "enabled", true); err != nil {
		return err
	}

	if err := populateTelemetryBatchingValues(remote.Batching, v1TelemetryValues); err != nil {
		return nil
	}

	// set the telemetry values
	if istiod {
		var v2TelemetryValues map[string]interface{}
		if rawTelemetryValues, ok, err := unstructured.NestedFieldNoCopy(values, "telemetry"); ok {
			if v2TelemetryValues, ok = rawTelemetryValues.(map[string]interface{}); !ok {
				v2TelemetryValues = make(map[string]interface{})
			}
		} else if err != nil {
			return nil
		} else {
			v2TelemetryValues = make(map[string]interface{})
		}
		if err := setHelmBoolValue(v2TelemetryValues, "enabled", true); err != nil {
			return err
		}
		if err := setHelmBoolValue(v2TelemetryValues, "v1.enabled", true); err != nil {
			return err
		}
		if err := setHelmBoolValue(v2TelemetryValues, "v2.enabled", false); err != nil {
			return err
		}

		if err := setHelmValue(values, "telemetry", v2TelemetryValues); err != nil {
			return err
		}
		if len(v1TelemetryValues) > 0 {
			if err := setHelmValue(values, "mixer.telemetry", v1TelemetryValues); err != nil {
				return err
			}
		}
	} else {
		if len(v1TelemetryValues) > 0 {
			if err := setHelmValue(values, "mixer.telemetry", v1TelemetryValues); err != nil {
				return err
			}
		}
	}

	return nil
}

func populateIstiodTelemetryValues(in *v2.ControlPlaneSpec, values map[string]interface{}) error {
	istiod := in.Telemetry.Istiod
	if istiod == nil {
		istiod = &v2.IstiodTelemetryConfig{}
	}

	// Make sure mixer is disabled
	if err := setHelmBoolValue(values, "mixer.enabled", false); err != nil {
		return err
	}

	var telemetryValues map[string]interface{}
	if rawTelemetryValues, ok, err := unstructured.NestedFieldNoCopy(values, "telemetry"); ok {
		if telemetryValues, ok = rawTelemetryValues.(map[string]interface{}); !ok {
			telemetryValues = make(map[string]interface{})
		}
	} else if err != nil {
		return nil
	} else {
		telemetryValues = make(map[string]interface{})
	}
	if err := setHelmBoolValue(telemetryValues, "enabled", true); err != nil {
		return err
	}
	if err := setHelmBoolValue(telemetryValues, "v1.enabled", false); err != nil {
		return err
	}
	if err := setHelmBoolValue(telemetryValues, "v2.enabled", true); err != nil {
		return err
	}

	// Adapters
	if istiod.MetadataExchange != nil {
		me := istiod.MetadataExchange
		if err := setHelmBoolValue(telemetryValues, "v2.metadataExchange.wasmEnabled", me.WASMEnabled); err != nil {
			return err
		}
	}

	if istiod.PrometheusFilter != nil {
		prometheus := istiod.PrometheusFilter
		if err := setHelmBoolValue(telemetryValues, "v2.prometheus.enabled", true); err != nil {
			return err
		}
		if err := setHelmBoolValue(telemetryValues, "v2.prometheus.wasmEnabled", prometheus.WASMEnabled); err != nil {
			return err
		}
		if err := setHelmBoolValue(values, "meshConfig.enablePrometheusMerge", prometheus.Scrape); err != nil {
			return err
		}
	}

	if istiod.StackDriverFilter != nil {
		stackdriver := istiod.StackDriverFilter
		if err := setHelmBoolValue(telemetryValues, "v2.stackdriver.enabled", true); err != nil {
			return err
		}
		if err := setHelmBoolValue(telemetryValues, "v2.stackdriver.logging", stackdriver.Logging); err != nil {
			return err
		}
		if err := setHelmBoolValue(telemetryValues, "v2.stackdriver.monitoring", stackdriver.Monitoring); err != nil {
			return err
		}
		if err := setHelmBoolValue(telemetryValues, "v2.stackdriver.topology", stackdriver.Topology); err != nil {
			return err
		}
		if err := setHelmBoolValue(telemetryValues, "v2.stackdriver.disableOutbound", stackdriver.DisableOutbound); err != nil {
			return err
		}
		if err := setHelmValue(telemetryValues, "v2.stackdriver.configOverride", stackdriver.ConfigOverride.GetContent()); err != nil {
			return err
		}
	}

	if istiod.AccessLogTelemetryFilter != nil {
		accessLog := istiod.AccessLogTelemetryFilter
		if err := setHelmBoolValue(telemetryValues, "v2.accessLogPolicy.enabled", true); err != nil {
			return err
		}
		if err := setHelmStringValue(telemetryValues, "v2.accessLogPolicy.logWindowDuration", accessLog.LogWindowDuration); err != nil {
			return err
		}
	}

	// set the telemetry values
	if len(telemetryValues) > 0 {
		if err := setHelmValue(values, "telemetry", telemetryValues); err != nil {
			return err
		}
	}

	return nil
}

func populateTelemetryConfig(in *v1.HelmValues, out *v2.ControlPlaneSpec, version versions.Version) error {
	var telemetryType v2.TelemetryType
	if telemetryTypeStr, ok, err := in.GetString("telemetry.implementation"); ok && telemetryTypeStr != "" {
		switch v2.TelemetryType(telemetryTypeStr) {
		case v2.TelemetryTypeIstiod:
			telemetryType = v2.TelemetryTypeIstiod
		case v2.TelemetryTypeMixer:
			telemetryType = v2.TelemetryTypeMixer
		case v2.TelemetryTypeRemote:
			telemetryType = v2.TelemetryTypeRemote
		case v2.TelemetryTypeNone:
			telemetryType = v2.TelemetryTypeNone
		default:
			return fmt.Errorf("unkown telemetry.implementation specified: %s", telemetryTypeStr)
		}
	} else if err != nil {
		return err
	} else {
		// figure out what we're installing
		if v2Enabled, v2EnabledSet, err := in.GetBool("telemetry.v2.enabled"); v2EnabledSet && v2Enabled {
			telemetryType = v2.TelemetryTypeIstiod
		} else if err != nil {
			return err
		} else if mixerTelemetryEnabled, mixerTelemetryEnabledSet, err := in.GetBool("mixer.telemetry.enabled"); err == nil {
			// installing some form of mixer based telemetry
			if mixerEnabled, mixerEnabledSet, err := in.GetBool("mixer.enabled"); err == nil {
				if !mixerEnabledSet || !mixerTelemetryEnabledSet {
					// assume no telemetry to configure
					return nil
				}
				if mixerEnabled {
					if mixerTelemetryEnabled {
						// installing mixer telemetry
						telemetryType = v2.TelemetryTypeMixer
					} else {
						// mixer telemetry disabled
						telemetryType = v2.TelemetryTypeNone
					}
				} else if mixerTelemetryEnabled {
					// using remote mixer telemetry
					telemetryType = v2.TelemetryTypeRemote
				} else {
					switch version {
					case versions.V1_0, versions.V1_1:
						// telemetry disabled
						telemetryType = v2.TelemetryTypeNone
					case versions.V2_0:
						if v2EnabledSet {
							telemetryType = v2.TelemetryTypeNone
						} else {
							telemetryType = v2.TelemetryTypeIstiod
						}
					default:
						return fmt.Errorf("unknown version: %s", version.String())
					}
				}
			} else {
				return err
			}
		} else {
			return err
		}
	}
	if telemetryType == "" {
		return fmt.Errorf("Could not determine policy type")
	}

	out.Telemetry = &v2.TelemetryConfig{
		Type: telemetryType,
	}
	switch telemetryType {
	case v2.TelemetryTypeIstiod:
		config := &v2.IstiodTelemetryConfig{}
		if applied, err := populateIstiodTelemetryConfig(in, config); err != nil {
			return err
		} else if applied {
			out.Telemetry.Istiod = config
		}
	case v2.TelemetryTypeMixer:
		config := &v2.MixerTelemetryConfig{}
		if applied, err := populateMixerTelemetryConfig(in, config); err != nil {
			return err
		} else if applied {
			out.Telemetry.Mixer = config
		}
	case v2.TelemetryTypeRemote:
		config := &v2.RemoteTelemetryConfig{}
		if applied, err := populateRemoteTelemetryConfig(in, config); err != nil {
			return err
		} else if applied {
			out.Telemetry.Remote = config
		}
	case v2.TelemetryTypeNone:
		// no configuration to set
	}

	return nil
}

func populateMixerTelemetryConfig(in *v1.HelmValues, out *v2.MixerTelemetryConfig) (bool, error) {
	setValues := false

	rawMixerValues, ok, err := in.GetMap("mixer")
	if err != nil {
		return false, err
	} else if !ok || len(rawMixerValues) == 0 {
		rawMixerValues = make(map[string]interface{})
	}
	mixerValues := v1.NewHelmValues(rawMixerValues)

	rawV1TelemetryValues, ok, err := mixerValues.GetMap("telemetry")
	if err != nil {
		return false, err
	} else if !ok || len(rawV1TelemetryValues) == 0 {
		rawV1TelemetryValues = make(map[string]interface{})
	}
	v1TelemetryValues := v1.NewHelmValues(rawV1TelemetryValues)

	if sessionAffinityEnabled, ok, err := v1TelemetryValues.GetBool("sessionAffinityEnabled"); ok {
		out.SessionAffinity = &sessionAffinityEnabled
		setValues = true
	} else if err != nil {
		return false, nil
	}

	batching := &v2.TelemetryBatchingConfig{}
	if applied, err := populateTelemetryBatchingConfig(v1TelemetryValues, batching); err != nil {
		return false, nil
	} else if applied {
		setValues = true
		out.Batching = batching
	}

	var telemetryAdaptersValues *v1.HelmValues
	if rawAdaptersValues, ok, err := mixerValues.GetMap("adapters"); ok {
		telemetryAdaptersValues = v1.NewHelmValues(rawAdaptersValues)
	} else if err != nil {
		return false, err
	}

	if telemetryAdaptersValues != nil {
		adapters := &v2.MixerTelemetryAdaptersConfig{}
		setAdapters := false
		if useAdapterCRDs, ok, err := telemetryAdaptersValues.GetBool("useAdapterCRDs"); ok {
			adapters.UseAdapterCRDs = &useAdapterCRDs
			setAdapters = true
		} else if err != nil {
			return false, err
		}
		if kubernetesenv, ok, err := telemetryAdaptersValues.GetBool("kubernetesenv.enabled"); ok {
			adapters.KubernetesEnv = &kubernetesenv
			setAdapters = true
		} else if err != nil {
			return false, err
		}
		if stdio, ok, err := telemetryAdaptersValues.GetBool("stdio.enabled"); ok && stdio {
			adapters.Stdio = &v2.MixerTelemetryStdioConfig{}
			if outputAsJSON, ok, err := telemetryAdaptersValues.GetBool("stdio.outputAsJson"); ok {
				adapters.Stdio.OutputAsJSON = outputAsJSON
			} else if err != nil {
				return false, err
			}
			setAdapters = true
		} else if err != nil {
			return false, err
		}
		if prometheus, ok, err := telemetryAdaptersValues.GetBool("prometheus.enabled"); ok && prometheus {
			adapters.Prometheus = &v2.MixerTelemetryPrometheusConfig{}
			if metricsExpiryDuration, ok, err := telemetryAdaptersValues.GetString("prometheus.metricsExpiryDuration"); ok {
				adapters.Prometheus.MetricsExpiryDuration = metricsExpiryDuration
			} else if err != nil {
				return false, err
			}
			setAdapters = true
		} else if err != nil {
			return false, err
		}
		if stackdriver, ok, err := telemetryAdaptersValues.GetBool("stackdriver.enabled"); ok && stackdriver {
			adapters.Stackdriver = &v2.MixerTelemetryStackdriverConfig{}
			setAdapters = true
			if contextGraph, ok, err := telemetryAdaptersValues.GetBool("stackdriver.contextGraph.enabled"); ok {
				adapters.Stackdriver.EnableContextGraph = contextGraph
			} else if err != nil {
				return false, err
			}
			if logging, ok, err := telemetryAdaptersValues.GetBool("stackdriver.logging.enabled"); ok {
				adapters.Stackdriver.EnableLogging = logging
			} else if err != nil {
				return false, err
			}
			if metrics, ok, err := telemetryAdaptersValues.GetBool("stackdriver.metrics.enabled"); ok {
				adapters.Stackdriver.EnableMetrics = metrics
			} else if err != nil {
				return false, err
			}
			auth := &v2.MixerTelemetryStackdriverAuthConfig{}
			setAuth := false
			if appCredentials, ok, err := telemetryAdaptersValues.GetBool("stackdriver.auth.appCredentials"); ok {
				auth.AppCredentials = appCredentials
				setAuth = true
			} else if err != nil {
				return false, err
			}
			if apiKey, ok, err := telemetryAdaptersValues.GetString("stackdriver.auth.apiKey"); ok {
				auth.APIKey = apiKey
				setAuth = true
			} else if err != nil {
				return false, err
			}
			if serviceAccountPath, ok, err := telemetryAdaptersValues.GetString("stackdriver.auth.serviceAccountPath"); ok {
				auth.ServiceAccountPath = serviceAccountPath
				setAuth = true
			} else if err != nil {
				return false, err
			}
			if setAuth {
				adapters.Stackdriver.Auth = auth
			}
			if sampleProbability, ok, err := telemetryAdaptersValues.GetInt64("stackdriver.tracer.sampleProbability"); ok {
				adapters.Stackdriver.Tracer = &v2.MixerTelemetryStackdriverTracerConfig{
					SampleProbability: int(sampleProbability),
				}
			} else if err != nil {
				return false, err
			}
		} else if err != nil {
			return false, err
		}
		if setAdapters {
			out.Adapters = adapters
			setValues = true
		}
	}

	return setValues, nil
}

func populateTelemetryBatchingConfig(in *v1.HelmValues, out *v2.TelemetryBatchingConfig) (bool, error) {
	setValues := false
	if reportBatchMaxTime, ok, err := in.GetString("reportBatchMaxTime"); ok {
		out.MaxTime = reportBatchMaxTime
		setValues = true
	} else if err != nil {
		return false, err
	}
	if rawReportBatchMaxEntries, ok, err := in.GetInt64("reportBatchMaxEntries"); ok {
		reportBatchMaxEntries := int32(rawReportBatchMaxEntries)
		out.MaxEntries = &reportBatchMaxEntries
		setValues = true
	} else if err != nil {
		return false, err
	}

	return setValues, nil
}

func populateRemoteTelemetryConfig(in *v1.HelmValues, out *v2.RemoteTelemetryConfig) (bool, error) {
	setValues := false

	if remoteTelemetryAddress, ok, err := in.GetString("global.remoteTelemetryAddress"); ok {
		out.Address = remoteTelemetryAddress
		setValues = true
	} else if err != nil {
		return false, err
	}
	if createRemoteSvcEndpoints, ok, err := in.GetBool("global.createRemoteSvcEndpoints"); ok {
		out.CreateService = createRemoteSvcEndpoints
		setValues = true
	} else if err != nil {
		return false, err
	}

	rawV1TelemetryValues, ok, err := in.GetMap("mixer.telemetry")
	if err != nil {
		return false, err
	} else if !ok || len(rawV1TelemetryValues) == 0 {
		rawV1TelemetryValues = make(map[string]interface{})
	}
	v1TelemetryValues := v1.NewHelmValues(rawV1TelemetryValues)

	batching := &v2.TelemetryBatchingConfig{}
	if applied, err := populateTelemetryBatchingConfig(v1TelemetryValues, batching); err != nil {
		return false, nil
	} else if applied {
		out.Batching = batching
		setValues = true
	}

	return setValues, nil
}

func populateIstiodTelemetryConfig(in *v1.HelmValues, out *v2.IstiodTelemetryConfig) (bool, error) {
	setValues := false

	rawTelemetryValues, ok, err := in.GetMap("telemetry")
	if err != nil {
		return false, err
	} else if !ok || len(rawTelemetryValues) == 0 {
		rawTelemetryValues = make(map[string]interface{})
	}
	telemetryValues := v1.NewHelmValues(rawTelemetryValues)

	// Adapters
	if metadataExchangeWASM, ok, err := telemetryValues.GetBool("v2.metadataExchange.wasmEnabled"); ok {
		out.MetadataExchange = &v2.MetadataExchangeConfig{
			WASMEnabled: metadataExchangeWASM,
		}
		setValues = true
	} else if err != nil {
		return false, err
	}

	if prometheus, ok, err := telemetryValues.GetBool("v2.prometheus.enabled"); ok && prometheus {
		out.PrometheusFilter = &v2.PrometheusFilterConfig{}
		setValues = true
		if wasmEnabled, ok, err := telemetryValues.GetBool("v2.prometheus.wasmEnabled"); ok {
			out.PrometheusFilter.WASMEnabled = wasmEnabled
		} else if err != nil {
			return false, err
		}
		if wasmEnabled, ok, err := telemetryValues.GetBool("v2.prometheus.wasmEnabled"); ok {
			out.PrometheusFilter.WASMEnabled = wasmEnabled
		} else if err != nil {
			return false, err
		}
		if enablePrometheusMerge, ok, err := in.GetBool("meshConfig.enablePrometheusMerge"); ok {
			out.PrometheusFilter.Scrape = enablePrometheusMerge
		} else if err != nil {
			return false, err
		}
	} else if err != nil {
		return false, err
	}

	if stackdriver, ok, err := telemetryValues.GetBool("v2.stackdriver.enabled"); ok && stackdriver {
		out.StackDriverFilter = &v2.StackDriverFilterConfig{}
		setValues = true
		if logging, ok, err := telemetryValues.GetBool("v2.stackdriver.logging"); ok {
			out.StackDriverFilter.Logging = logging
		} else if err != nil {
			return false, err
		}
		if monitoring, ok, err := telemetryValues.GetBool("v2.stackdriver.monitoring"); ok {
			out.StackDriverFilter.Monitoring = monitoring
		} else if err != nil {
			return false, err
		}
		if topology, ok, err := telemetryValues.GetBool("v2.stackdriver.topology"); ok {
			out.StackDriverFilter.Topology = topology
		} else if err != nil {
			return false, err
		}
		if disableOutbound, ok, err := telemetryValues.GetBool("v2.stackdriver.disableOutbound"); ok {
			out.StackDriverFilter.DisableOutbound = disableOutbound
		} else if err != nil {
			return false, err
		}
		if configOverride, ok, err := telemetryValues.GetMap("v2.stackdriver.configOverride"); ok && len(configOverride) > 0 {
			out.StackDriverFilter.ConfigOverride = v1.NewHelmValues(configOverride)
		} else if err != nil {
			return false, err
		}
	} else if err != nil {
		return false, err
	}

	if accessLogPolicy, ok, err := telemetryValues.GetBool("v2.accessLogPolicy.enabled"); ok && accessLogPolicy {
		out.AccessLogTelemetryFilter = &v2.AccessLogTelemetryFilterConfig{}
		setValues = true
		if logWindowDuration, ok, err := telemetryValues.GetString("v2.accessLogPolicy.logWindowDuration"); ok {
			out.AccessLogTelemetryFilter.LogWindowDuration = logWindowDuration
		} else if err != nil {
			return false, err
		}
	} else if err != nil {
		return false, err
	}

	return setValues, nil
}
