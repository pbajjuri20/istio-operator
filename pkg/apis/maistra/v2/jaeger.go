package v2

// JaegerTracerConfig configures Jaeger a tracer implementation
// XXX: this currently deviates from upstream, which creates a jaeger all-in-one deployment manually
type JaegerTracerConfig struct {
	// Name of Jaeger CR, Namespace must match control plane namespace
	Name string `json:"name,omitempty"`
	// Install configures a Jaeger installation, which will be created if the
	// named Jaeger resource is not present.  If null, the named Jaeger resource
	// must exist.
	Install *JaegerInstallConfig `json:"install,omitempty"`
}

// JaegerInstallConfig configures a Jaeger installation.
type JaegerInstallConfig struct {
	// Config represents the configuration of Jaeger behavior.
	Config JaegerConfig `json:"config,omitempty"`
	// Runtime configures runtime aspects of Jaeger deployment/pod
	// Used to configure resources and affinity.  runtime.pod.containers can be
	// used to override details for specific jaeger components, e.g. allInOne,
	// query, etc.  runtime.metadata.annotations maps to
	// .Values.tracing.jaeger.annotations
	Runtime *ComponentRuntimeConfig `json:"runtime,omitempty"`
	// Ingress configures k8s Ingress or OpenShift Route for Jaeger services
	// .Values.tracing.jaeger.ingress.enabled, false if null
	Ingress *JaegerIngressConfig `json:"ingress,omitempty"`
}

// JaegerConfig is used to configure the behavior of the Jaeger installation.JaegerConfig
// XXX: should this include the templates that were used before?
// XXX: the storage type is used to imply which template should be used
type JaegerConfig struct {
	// Storage represents the storage configuration for the Jaeger install
	Storage *JaegerStorageConfig `json:"storage,omitempty"`
}

// JaegerStorageConfig configures the storage used by the Jaeger installation.
type JaegerStorageConfig struct {
	// Type of storage to use
	Type JaegerStorageType `json:"type,omitempty"`
	// Memory represents configuration of in-memory storage
	// implies .Values.tracing.jaeger.template=all-in-one
	Memory *JaegerMemoryStorageConfig `json:"memory,omitempty"`
	// Elasticsearch represents configuration of elasticsearch storage
	// implies .Values.tracing.jaeger.template=production-elasticsearch
	Elasticsearch *JaegerElasticsearchStorageConfig `json:"elasticsearch,omitempty"`
}

// JaegerStorageType represents the type of storage configured for Jaeger
type JaegerStorageType string

const (
	// JaegerStorageTypeMemory represents in-memory storage
	JaegerStorageTypeMemory JaegerStorageType = "Memory"
	// JaegerStorageTypeElasticsearch represents Elasticsearch storage
	JaegerStorageTypeElasticsearch JaegerStorageType = "Elasticsearch"
)

// JaegerMemoryStorageConfig configures in-memory storage parameters for Jaeger
type JaegerMemoryStorageConfig struct {
	// MaxTraces to store
	// .Values.tracing.jaeger.memory.max_traces, defaults to 100000
	MaxTraces int64 `json:"maxTraces,omitempty"`
}

// JaegerElasticsearchStorageConfig configures elasticsearch storage parameters for Jaeger
type JaegerElasticsearchStorageConfig struct {
	// NodeCount represents the number of elasticsearch nodes to create.
	// .Values.tracing.jaeger.elasticsearch.nodeCount, defaults to 3
	NodeCount int32 `json:"nodeCount,omitempty"`
	// Storage represents storage configuration for elasticsearch.
	// .Values.tracing.jaeger.elasticsearch.storage, raw yaml
	// XXX: RawExtension?
	Storage map[string]string `json:"storage,omitempty"`
	// RedundancyPolicy configures the redundancy policy for elasticsearch
	// .Values.tracing.jaeger.elasticsearch.redundancyPolicy, raw yaml
	// XXX: RawExtension?
	RedundancyPolicy map[string]string `json:"redundancyPolicy,omitempty"`
	// IndexCleaner represents the configuration for the elasticsearch index cleaner
	// .Values.tracing.jaeger.elasticsearch.esIndexCleaner, raw yaml
	// XXX: RawExtension?
	IndexCleaner map[string]string `json:"indexCleaner,omitempty"`
	// Runtime allows for customization of the elasticsearch pods
	// used for node selector, etc., specific to elasticsearch config
	Runtime *PodRuntimeConfig `json:"runtime,omitempty"`
}

// JaegerIngressConfig configures k8s Ingress or OpenShift Route for exposing
// Jaeger services.
type JaegerIngressConfig struct {
	// Metadata represents addtional annotations/labels to be applied to the ingress/route.
	Metadata MetadataConfig `json:"metadata,omitempty"`
}
