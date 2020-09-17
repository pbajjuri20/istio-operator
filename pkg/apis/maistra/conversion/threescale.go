package conversion

import (
	v1 "github.com/maistra/istio-operator/pkg/apis/maistra/v1"
	v2 "github.com/maistra/istio-operator/pkg/apis/maistra/v2"
)

func populateThreeScaleAddonValues(threeScale *v2.ThreeScaleConfig, values map[string]interface{}) (reterr error) {
	if threeScale == nil {
		return nil
	}
	threeScaleValues := make(map[string]interface{})
	defer func() {
		if reterr == nil {
			if len(threeScaleValues) > 0 {
				if err := setHelmValue(values, "3scale", threeScaleValues); err != nil {
					reterr = err
				}
			}
		}
	}()

	if threeScale.Enabled != nil {
		if err := setHelmBoolValue(threeScaleValues, "enabled", *threeScale.Enabled); err != nil {
			return err
		}
	}
	if threeScale.ListenAddr != nil {
		if err := setHelmIntValue(threeScaleValues, "PARAM_THREESCALE_LISTEN_ADDR", int64(*threeScale.ListenAddr)); err != nil {
			return err
		}
	}
	if threeScale.LogGRPC != nil {
		if err := setHelmBoolValue(threeScaleValues, "PARAM_THREESCALE_LOG_GRPC", *threeScale.LogGRPC); err != nil {
			return err
		}
	}
	if threeScale.LogJSON != nil {
		if err := setHelmBoolValue(threeScaleValues, "PARAM_THREESCALE_LOG_JSON", *threeScale.LogJSON); err != nil {
			return err
		}
	}
	if threeScale.LogLevel != "" {
		if err := setHelmStringValue(threeScaleValues, "PARAM_THREESCALE_LOG_LEVEL", threeScale.LogLevel); err != nil {
			return err
		}
	}
	if threeScale.Metrics != nil {
		metrics := threeScale.Metrics
		if metrics.Port != nil {
			if err := setHelmIntValue(threeScaleValues, "PARAM_THREESCALE_METRICS_PORT", int64(*metrics.Port)); err != nil {
				return err
			}
		}
		if metrics.Report != nil {
			if err := setHelmBoolValue(threeScaleValues, "PARAM_THREESCALE_REPORT_METRICS", *metrics.Report); err != nil {
				return err
			}
		}
	}
	if threeScale.System != nil {
		system := threeScale.System
		if system.CacheMaxSize != nil {
			if err := setHelmIntValue(threeScaleValues, "PARAM_THREESCALE_CACHE_ENTRIES_MAX", int64(*system.CacheMaxSize)); err != nil {
				return err
			}
		}
		if system.CacheRefreshRetries != nil {
			if err := setHelmIntValue(threeScaleValues, "PARAM_THREESCALE_CACHE_REFRESH_RETRIES", int64(*system.CacheRefreshRetries)); err != nil {
				return err
			}
		}
		if system.CacheRefreshInterval != nil {
			if err := setHelmIntValue(threeScaleValues, "PARAM_THREESCALE_CACHE_REFRESH_SECONDS", int64(*system.CacheRefreshInterval)); err != nil {
				return err
			}
		}
		if system.CacheTTL != nil {
			if err := setHelmIntValue(threeScaleValues, "PARAM_THREESCALE_CACHE_TTL_SECONDS", int64(*system.CacheTTL)); err != nil {
				return err
			}
		}
	}
	if threeScale.Client != nil {
		client := threeScale.Client
		if client.AllowInsecureConnections != nil {
			if err := setHelmBoolValue(threeScaleValues, "PARAM_THREESCALE_ALLOW_INSECURE_CONN", *client.AllowInsecureConnections); err != nil {
				return err
			}
		}
		if client.Timeout != nil {
			if err := setHelmIntValue(threeScaleValues, "PARAM_THREESCALE_CLIENT_TIMEOUT_SECONDS", int64(*client.Timeout)); err != nil {
				return err
			}
		}
	}
	if threeScale.GRPC != nil {
		grpc := threeScale.GRPC
		if grpc.MaxConnTimeout != nil {
			if err := setHelmIntValue(threeScaleValues, "PARAM_THREESCALE_GRPC_CONN_MAX_SECONDS", int64(*grpc.MaxConnTimeout)); err != nil {
				return err
			}
		}
	}
	if threeScale.Backend != nil {
		backend := threeScale.Backend
		if backend.EnableCache != nil {
			if err := setHelmBoolValue(threeScaleValues, "PARAM_THREESCALE_USE_CACHED_BACKEND", *backend.EnableCache); err != nil {
				return err
			}
		}
		if backend.CacheFlushInterval != nil {
			if err := setHelmIntValue(threeScaleValues, "PARAM_THREESCALE_BACKEND_CACHE_FLUSH_INTERVAL_SECONDS", int64(*backend.CacheFlushInterval)); err != nil {
				return err
			}
		}
		if backend.PolicyFailClosed != nil {
			if err := setHelmBoolValue(threeScaleValues, "PARAM_THREESCALE_BACKEND_CACHE_POLICY_FAIL_CLOSED", *backend.PolicyFailClosed); err != nil {
				return err
			}
		}
	}

	return nil
}

func populateThreeScaleAddonConfig(in *v1.HelmValues, out *v2.MiscAddonsConfig) error {
	rawThreeScaleValues, ok, err := in.GetMap("3scale")
	if err != nil {
		return err
	} else if !ok || len(rawThreeScaleValues) == 0 {
		// nothing to do
		return nil
	}

	if out.ThreeScale == nil {
		out.ThreeScale = &v2.ThreeScaleConfig{}
	}
	threeScale := out.ThreeScale
	threeScaleValues := v1.NewHelmValues(rawThreeScaleValues)

	if enabled, ok, err := threeScaleValues.GetBool("enabled"); ok {
		threeScale.Enabled = &enabled
	} else if err != nil {
		return err
	}
	if rawListenAddr, ok, err := threeScaleValues.GetInt64("PARAM_THREESCALE_LISTEN_ADDR"); ok {
		listernAddr := int32(rawListenAddr)
		threeScale.ListenAddr = &listernAddr
	} else if err != nil {
		return err
	}
	if logGRPC, ok, err := threeScaleValues.GetBool("PARAM_THREESCALE_LOG_GRPC"); ok {
		threeScale.LogGRPC = &logGRPC
	} else if err != nil {
		return err
	}
	if logJSON, ok, err := threeScaleValues.GetBool("PARAM_THREESCALE_LOG_JSON"); ok {
		threeScale.LogJSON = &logJSON
	} else if err != nil {
		return err
	}
	if logLevel, ok, err := threeScaleValues.GetString("PARAM_THREESCALE_LOG_LEVEL"); ok {
		threeScale.LogLevel = logLevel
	} else if err != nil {
		return err
	}

	metrics := &v2.ThreeScaleMetricsConfig{}
	setMetrics := false
	if rawPort, ok, err := threeScaleValues.GetInt64("PARAM_THREESCALE_METRICS_PORT"); ok {
		port := int32(rawPort)
		metrics.Port = &port
		setMetrics = true
	} else if err != nil {
		return err
	}
	if report, ok, err := threeScaleValues.GetBool("PARAM_THREESCALE_REPORT_METRICS"); ok {
		metrics.Report = &report
		setMetrics = true
	} else if err != nil {
		return err
	}
	if setMetrics {
		threeScale.Metrics = metrics
	}

	system := &v2.ThreeScaleSystemConfig{}
	setSystem := false
	if cacheMaxSize, ok, err := threeScaleValues.GetInt64("PARAM_THREESCALE_CACHE_ENTRIES_MAX"); ok {
		system.CacheMaxSize = &cacheMaxSize
		setSystem = true
	} else if err != nil {
		return err
	}
	if rawCacheRefreshRetries, ok, err := threeScaleValues.GetInt64("PARAM_THREESCALE_CACHE_REFRESH_RETRIES"); ok {
		cacheRefreshRetries := int32(rawCacheRefreshRetries)
		system.CacheRefreshRetries = &cacheRefreshRetries
		setSystem = true
	} else if err != nil {
		return err
	}
	if rawCacheRefreshInterval, ok, err := threeScaleValues.GetInt64("PARAM_THREESCALE_CACHE_REFRESH_SECONDS"); ok {
		cacheRefreshInterval := int32(rawCacheRefreshInterval)
		system.CacheRefreshInterval = &cacheRefreshInterval
		setSystem = true
	} else if err != nil {
		return err
	}
	if rawCacheTTL, ok, err := threeScaleValues.GetInt64("PARAM_THREESCALE_CACHE_TTL_SECONDS"); ok {
		cacheTTL := int32(rawCacheTTL)
		system.CacheTTL = &cacheTTL
		setSystem = true
	} else if err != nil {
		return err
	}
	if setSystem {
		threeScale.System = system
	}

	client := &v2.ThreeScaleClientConfig{}
	setClient := true
	if allowInsecureConnections, ok, err := threeScaleValues.GetBool("PARAM_THREESCALE_ALLOW_INSECURE_CONN"); ok {
		client.AllowInsecureConnections = &allowInsecureConnections
		setClient = true
	} else if err != nil {
		return err
	}
	if rawTimeout, ok, err := threeScaleValues.GetInt64("PARAM_THREESCALE_CLIENT_TIMEOUT_SECONDS"); ok {
		timeout := int32(rawTimeout)
		client.Timeout = &timeout
		setClient = true
	} else if err != nil {
		return err
	}
	if setClient {
		threeScale.Client = client
	}

	if rawMaxConnTimeout, ok, err := threeScaleValues.GetInt64("PARAM_THREESCALE_GRPC_CONN_MAX_SECONDS"); ok {
		maxConnTimeout := int32(rawMaxConnTimeout)
		threeScale.GRPC = &v2.ThreeScaleGRPCConfig{
			MaxConnTimeout: &maxConnTimeout,
		}
	} else if err != nil {
		return err
	}

	backend := &v2.ThreeScaleBackendConfig{}
	setBackend := false
	if enableCache, ok, err := threeScaleValues.GetBool("PARAM_THREESCALE_USE_CACHED_BACKEND"); ok {
		backend.EnableCache = &enableCache
		setBackend = true
	} else if err != nil {
		return err
	}
	if rawCacheFlushInterval, ok, err := threeScaleValues.GetInt64("PARAM_THREESCALE_BACKEND_CACHE_FLUSH_INTERVAL_SECONDS"); ok {
		cacheFlushInterval := int32(rawCacheFlushInterval)
		backend.CacheFlushInterval = &cacheFlushInterval
		setBackend = true
	} else if err != nil {
		return err
	}
	if policyFailClosed, ok, err := threeScaleValues.GetBool("PARAM_THREESCALE_BACKEND_CACHE_POLICY_FAIL_CLOSED"); ok {
		backend.PolicyFailClosed = &policyFailClosed
		setBackend = true
	} else if err != nil {
		return err
	}
	if setBackend {
		threeScale.Backend = backend
	}

	return nil
}
