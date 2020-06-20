package v2

// TelemetryConfig for the mesh
type TelemetryConfig struct {
	// Type of telemetry implementation to use.
	Type TelemetryType `json:"type,omitempty"`
	// Mixer represents legacy, v1 telemetry.
	// implies .Values.telemetry.v1.enabled, if not null
	Mixer *MixerTelemetryConfig `json:"mixer,omitempty"`
	// Remote represents a remote, legacy, v1 telemetry.
	Remote *RemoteTelemetryConfig `json:"remote,omitempty"`
	// Istiod represents istiod, v2 telemetry
	Istiod *IstiodTelemetryConfig `json:"istiod,omitempty"`
}

// TelemetryType represents the telemetry implementation used.
type TelemetryType string

const (
	// TelemetryTypeMixer represents mixer telemetry, v1
	TelemetryTypeMixer TelemetryType = "Mixer"
	// TelemetryTypeRemote represents remote mixer telemetry server, v1
	TelemetryTypeRemote TelemetryType = "Remote"
	// TelemetryTypeIstiod represents istio, v2
	TelemetryTypeIstiod TelemetryType = "Istiod"
)

// MixerTelemetryConfig is the configuration for legacy, v1 mixer telemetry.
// .Values.telemetry.v1.enabled
type MixerTelemetryConfig struct {
	// SessionAffinity configures session affinity for sidecar telemetry connections.
	// .Values.mixer.telemetry.sessionAffinityEnabled, maps to MeshConfig.sidecarToTelemetrySessionAffinity
	SessionAffinity bool `json:"sessionAffinity,omitempty"`
	// Batching settings used when sending telemetry.
	Batching TelemetryBatchingConfig `json:"batching,omitempty"`
	// Runtime configuration to apply to the mixer telemetry deployment.
	Runtime *DeploymentRuntimeConfig `json:"runtime,omitempty"`
	// Adapters configures the adapters used by mixer telemetry.
	Adapters *MixerTelemetryAdaptersConfig `json:"adapters,omitempty"`
}

// TelemetryBatchingConfig configures how telemetry data is batched.
type TelemetryBatchingConfig struct {
	// MaxEntries represents the maximum number of entries to collect before sending them to mixer.
	// .Values.mixer.telemetry.reportBatchMaxEntries, maps to MeshConfig.reportBatchMaxEntries
	// Set reportBatchMaxEntries to 0 to use the default batching behavior (i.e., every 100 requests).
	// A positive value indicates the number of requests that are batched before telemetry data
	// is sent to the mixer server
	MaxEntries int32 `json:"maxEntries,omitempty"`
	// MaxTime represents the maximum amount of time to hold entries before sending them to mixer.
	// .Values.mixer.telemetry.reportBatchMaxTime, maps to MeshConfig.reportBatchMaxTime
	// Set reportBatchMaxTime to 0 to use the default batching behavior (i.e., every 1 second).
	// A positive time value indicates the maximum wait time since the last request will telemetry data
	// be batched before being sent to the mixer server
	MaxTime string `json:"maxTime,omitempty"`
}

// MixerTelemetryAdaptersConfig is the configuration for mixer telemetry adapters.
type MixerTelemetryAdaptersConfig struct {
	// UseAdapterCRDs specifies whether or not mixer should support deprecated CRDs.
	// .Values.mixer.adapters.useAdapterCRDs, removed in istio 1.4, defaults to false
	// XXX: i think this can be removed completely
	UseAdapterCRDs bool `json:"useAdapterCRDs,omitempty"`
	// KubernetesEnv enables support for the kubernetesenv adapter.
	// .Values.mixer.adapters.kubernetesenv.enabled, defaults to true
	KubernetesEnv bool `json:"kubernetesenv,omitempty"`
	// Stdio enables and configures the stdio adapter.
	// .Values.mixer.adapters.stdio.enabled, defaults to false (null)
	Stdio *MixerTelemetryStdioConfig `json:"stdio,omitempty"`
	// Prometheus enables and configures the prometheus adapter.
	// .Values.mixer.adapters.prometheus.enabled, defaults to true (non-null)
	Prometheus *MixerTelemetryPrometheusConfig `json:"prometheus,omitempty"`
	// Stackdriver enables and configures the stackdriver apdater.
	// .Values.mixer.adapters.stackdriver.enabled, defaults to false (null)
	Stackdriver *MixerTelemetryStackdriverConfig `json:"stackdriver,omitempty"`
}

// MixerTelemetryStdioConfig configures the stdio adapter for mixer telemetry.
type MixerTelemetryStdioConfig struct {
	// OutputAsJSON if true.
	// .Values.mixer.adapters.stdio.outputAsJson, defaults to false
	OutputAsJSON bool `json:"outputAsJSON,omitempty"`
}

// MixerTelemetryPrometheusConfig configures the prometheus adapter for mixer telemetry.
type MixerTelemetryPrometheusConfig struct {
	// MetricsExpiryDuration is the duration to hold metrics.
	// .Values.mixer.adapters.prometheus.metricsExpiryDuration, defaults to 10m
	MetricsExpiryDuration string `json:"metricsExpiryDuration,omitempty"`
}

// MixerTelemetryStackdriverConfig configures the stackdriver adapter for mixer telemetry.
type MixerTelemetryStackdriverConfig struct {
	// Auth configuration for stackdriver adapter
	Auth *MixerTelemetryStackdriverAuthConfig `json:"auth,omitempty"`
	// Tracer configuration for stackdriver adapter
	// .Values.mixer.adapters.stackdriver.tracer.enabled, defaults to false (null)
	Tracer *MixerTelemetryStackdriverTracerConfig `json:"tracer,omitempty"`
	// EnableContextGraph for stackdriver adapter
	// .Values.mixer.adapters.stackdriver.contextGraph.enabled, defaults to false
	EnableContextGraph bool `json:"enableContextGraph,omitempty"`
	// EnableLogging for stackdriver adapter
	// .Values.mixer.adapters.stackdriver.logging.enabled, defaults to true
	EnableLogging bool `json:"enableLogging,omitempty"`
	// EnableMetrics for stackdriver adapter
	// .Values.mixer.adapters.stackdriver.metrics.enabled, defaults to true
	EnableMetrics bool `json:"enableMetrics,omitempty"`
}

// MixerTelemetryStackdriverAuthConfig is the auth config for stackdriver.  Only one field may be set
type MixerTelemetryStackdriverAuthConfig struct {
	// AppCredentials if true, use default app credentials.
	// .Values.mixer.adapters.stackdriver.auth.appCredentials, defaults to false
	AppCredentials bool `json:"appCredentials,omitempty"`
	// APIKey use the specified key.
	// .Values.mixer.adapters.stackdriver.auth.apiKey
	APIKey string `json:"apiKey,omitempty"`
	// ServiceAccountPath use the path to the service account.
	// .Values.mixer.adapters.stackdriver.auth.serviceAccountPath
	ServiceAccountPath string `json:"serviceAccountPath,omitempty"`
}

// MixerTelemetryStackdriverTracerConfig tracer config for stackdriver mixer telemetry adapter
type MixerTelemetryStackdriverTracerConfig struct {
	// SampleProbability to use for tracer data.
	// .Values.mixer.adapters.stackdriver.tracer.sampleProbability
	SampleProbability int `json:"sampleProbability,omitempty"`
}

// RemoteTelemetryConfig configures a remote, legacy, v1 mixer telemetry.
// .Values.telemetry.v1.enabled true
type RemoteTelemetryConfig struct {
	// Address is the address of the remote telemetry server
	// .Values.global.remoteTelemetryAddress, maps to MeshConfig.mixerReportServer
	Address string `json:"address,omitempty"`
	// CreateServices for the remote server.
	// .Values.global.createRemoteSvcEndpoints
	CreateServices bool `json:"createServices,omitempty"`
	// Batching settings used when sending telemetry.
	Batching TelemetryBatchingConfig `json:"batching,omitempty"`
}

// IstiodTelemetryConfig configures v2 telemetry using istiod
// .Values.telemetry.v2.enabled
type IstiodTelemetryConfig struct {
	// MetadataExchange configuration for v2 telemetry.
	// always enabled
	MetadataExchange *MetadataExchangeConfig `json:"metadataExchange,omitempty"`
	// PrometheusFilter configures the prometheus filter for v2 telemetry.
	// .Values.telemetry.v2.prometheus.enabled
	PrometheusFilter *PrometheusFilterConfig `json:"prometheusFilter,omitempty"`
	// StackDriverFilter configures the stackdriver filter for v2 telemetry.
	// .Values.telemetry.v2.stackdriver.enabled
	StackDriverFilter *StackDriverFilterConfig `json:"stackDriverFilter,omitempty"`
	// AccessLogTelemetryFilter configures the access logging filter for v2 telemetry.
	// .Values.telemetry.v2.accessLogPolicy.enabled
	AccessLogTelemetryFilter *AccessLogTelemetryFilterConfig `json:"accessLogTelemetryFilter,omitempty"`
}

// MetadataExchangeConfig for v2 telemetry.
type MetadataExchangeConfig struct {
	// WASMEnabled for metadata exchange.
	// .Values.telemetry.v2.metadataExchange.wasmEnabled
	// Indicates whether to enable WebAssembly runtime for metadata exchange filter.
	WASMEnabled bool `json:"wasmEnabled,omitempty"`
}

// PrometheusFilterConfig for v2 telemetry.
// previously enablePrometheusMerge
// annotates injected pods with prometheus.io annotations (scrape, path, port)
// overridden through prometheus.istio.io/merge-metrics
type PrometheusFilterConfig struct {
	// Scrape metrics from the pod if true.
	// defaults to true
	Scrape bool `json:"scrape,omitempty"`
	// WASMEnabled for prometheus filter.
	// Indicates whether to enable WebAssembly runtime for stats filter.
	WASMEnabled bool `json:"wasmEnabled,omitempty"`
}

// StackDriverFilterConfig for v2 telemetry.
type StackDriverFilterConfig struct {
	// all default to false
	Logging         bool              `json:"logging,omitempty"`
	Monitoring      bool              `json:"monitoring,omitempty"`
	Topology        bool              `json:"topology,omitempty"`
	DisableOutbound bool              `json:"disableOutbound,omitempty"`
	ConfigOverride  map[string]string `json:"configOverride,omitempty"`
}

// AccessLogTelemetryFilterConfig for v2 telemetry.
type AccessLogTelemetryFilterConfig struct {
	// LogWindoDuration configures the log window duration for access logs.
	// defaults to 43200s
	// To reduce the number of successful logs, default log window duration is
	// set to 12 hours.
	LogWindoDuration string `json:"logWindowDuration,omitempty"`
}
