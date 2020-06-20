package v2

import (
	corev1 "k8s.io/api/core/v1"
)

type ProxyConfig struct {
	// XXX: should this be independent of global logging?  previously, this was
	// only exposed through proxy settings and there was no separate logging for
	// control plane components (e.g. pilot, mixer, etc.).
	// e.g. .Values.global.proxy.logLevel
	Logging LoggingConfig `json:"logging,omitempty"`
	// Networking represents network settings to be configured for the sidecars.
	Networking ProxyNetworkingConfig `json:"networking,omitempty"`
	Runtime    ProxyRuntimeConfig
	// maps to defaultConfig.proxyAdminPort, defaults to 15000
	AdminPort int32
	// .Values.global.proxy.concurrency, maps to defaultConfig.concurrency
	// XXX: removed in 1.7
	// XXX: this is defaulted to 2 in our values.yaml, but should probably be 0
	Concurrency int32
}

// ProxyNetworkingConfig is used to configure networking aspects of the sidecar.
type ProxyNetworkingConfig struct {
	// maps to meshConfig.defaultConfig.connectionTimeout, defaults to 10s
	// XXX: currently not exposed through values.yaml
	ConnectionTimeout string `json:"connectionTimeout,omitempty"`
	// Initialization is used to specify how the pod's networking through the
	// proxy is initialized.  This configures the use of CNI or an init container.
	Initialization ProxyNetworkInitConfig `json:"initialization,omitempty"`
	// TrafficControl configures what network traffic is routed through the proxy.
	TrafficControl ProxyTrafficControlConfig `json:"trafficControl,omitempty"`
	// Protocol configures how the sidecar works with applicaiton protocols.
	Protocol ProxyNetworkProtocolConfig `json:"protocol,omitempty"`
	// DNS configures aspects of the sidecar's usage of DNS
	DNS      ProxyDNSConfig `json:"protocol,omitempty"`
}

// ProxyNetworkInitConfig is used to configure how the pod's networking through
// the proxy is initialized.
type ProxyNetworkInitConfig struct {
	// Type of the network initialization implementation.
	Type ProxyNetworkInitType `json:"type,omitempty"`
	// CNI configures the use of CNI for initializing the pod's networking.
	// istio_cni.enabled = true, if CNI is used
	CNI *ProxyCNIConfig `json:"cni,omitempty"`
	// InitContainer configures the use of a pod init container for initializing
	// the pod's networking.
	// istio_cni.enabled = false, if InitContainer is used
	InitContainer *ProxyInitContainerConfig `json:"initContainer,omitempty"`
}

type ProxyNetworkInitType string

const (
	ProxyNetworkInitTypeCNI           ProxyNetworkInitType = "CNI"
	ProxyNetworkInitTypeInitContainer ProxyNetworkInitType = "InitContainer"
)

type ProxyCNIConfig struct {
	// TODO: add runtime configuration
	Runtime *ProxyCNIRuntimeConfig
}

type ProxyCNIRuntimeConfig struct {
	ContainerConfig
	PriorityClassName string `json:"priorityClassName,omitempty" protobuf:"bytes,24,opt,name=priorityClassName"`
}

type ProxyInitContainerConfig struct {
	// TODO: add runtime configuration
	Runtime *ContainerConfig
}

// ProxyTrafficControlConfig configures what and how traffic is routed through
// the sidecar.
type ProxyTrafficControlConfig struct {
	// Inbound configures what inbound traffic is routed through the sidecar
	// traffic.sidecar.istio.io/includeInboundPorts defaults to * (all ports)
	Inbound ProxyInboundTrafficControlConfig `json:"inbound,omitempty"`
	// Outbound configures what outbound traffic is routed through the sidecar.
	Outbound ProxyOutboundTrafficControlConfig `json:"outbound,omitempty"`
}

// ProxyNetworkInterceptionMode represents the InterceptMode types.
type ProxyNetworkInterceptionMode string

const (
	// ProxyNetworkInterceptionModeRedirect requests iptables use REDIRECT to route inbound traffic through the sidecar.
	ProxyNetworkInterceptionModeRedirect ProxyNetworkInterceptionMode = "REDIRECT"
	// ProxyNetworkInterceptionModeTProxy requests iptables use TPROXY to route inbound traffic through the sidecar.
	ProxyNetworkInterceptionModeTProxy ProxyNetworkInterceptionMode = "TPROXY"
)

// ProxyInboundTrafficControlConfig configures what inbound traffic is
// routed through the sidecar.
type ProxyInboundTrafficControlConfig struct {
	// InterceptionMode specifies how traffic is directed through the sidecar.
	// maps to meshConfig.defaultConfig.interceptionMode, overridden by sidecar.istio.io/interceptionMode
	// XXX: currently not configurable through values.yaml
	InterceptionMode ProxyNetworkInterceptionMode `json:"interceptionMode,omitempty"`
	// IncludedPorts to be routed through the sidecar. * or comma separated list of integers
	// .Values.global.proxy.includeInboundPorts, defaults to * (all ports), overridden by traffic.sidecar.istio.io/includeInboundPorts
	IncludedPorts []string `json:"includedPorts,omitempty"`
}

// ProxyOutboundTrafficControlConfig configure what outbound traffic is routed
// through the sidecar
type ProxyOutboundTrafficControlConfig struct {
	// IncludedIPRanges specifies which outbound IP ranges should be routed through the sidecar.
	// .Values.global.proxy.includeIPRanges, overridden by traffic.sidecar.istio.io/includeOutboundIPRanges
	// * or comma separated list of CIDR
	IncludedIPRanges []string `json:"includedIPRanges,omitempty"`
	// ExcludedIPRanges specifies which outbound IP ranges should _not_ be routed through the sidecar.
	// .Values.global.proxy.excludeIPRanges, overridden by traffic.sidecar.istio.io/excludeOutboundIPRanges
	// * or comma separated list of CIDR
	ExcludedIPRanges []string `json:"excludedIPRanges,omitempty"`
	// ExcludedPorts specifies which outbound ports should _not_ be routed through the sidecar.
	// .Values.global.proxy.excludeOutboundPorts, overridden by traffic.sidecar.istio.io/excludeOutboundPorts
	// comma separated list of integers
	ExcludedPorts []int32 `json:"excludedPorts,omitempty"`
	// Policy specifies what outbound traffic is allowed through the sidecar.
	// .Values.global.outboundTrafficPolicy.mode
	Policy ProxyOutboundTrafficPolicy `json:"policy,omitempty"`
}

// ProxyOutboundTrafficPolicy represents the outbound traffic policy type.
type ProxyOutboundTrafficPolicy string

const (
	// ProxyOutboundTrafficPolicyAllowAny allows all traffic through the sidecar.
	ProxyOutboundTrafficPolicyAllowAny ProxyOutboundTrafficPolicy = "ALLOW_ANY"
	// ProxyOutboundTrafficPolicyRegistryOnly only allows traffic destined for a
	// service in the service registry through the sidecar.  This limits outbound
	// traffic to only other services in the mesh.
	ProxyOutboundTrafficPolicyRegistryOnly ProxyOutboundTrafficPolicy = "REGISTRY_ONLY"
)

// ProxyNetworkProtocolConfig configures the sidecar's protocol handling.
type ProxyNetworkProtocolConfig struct {
	// DetectionTimeout specifies how much time the sidecar will spend determining
	// the protocol being used for the connection before reverting to raw TCP.
	// .Values.global.proxy.protocolDetectionTimeout, maps to protocolDetectionTimeout
	DetectionTimeout string `json:"detectionTimeout,omitempty"`
	// Debug configures debugging capabilities for the connection.
	Debug ProxyNetworkProtocolDebugConfig `json:"debug,omitempty"`
}

// ProxyNetworkProtocolDebugConfig specifies configuration for protocol debugging.
type ProxyNetworkProtocolDebugConfig struct {
	// EnableInboundSniffing enables protocol sniffing on inbound traffic.
	// .Values.pilot.enableProtocolSniffingForInbound
	EnableInboundSniffing bool `json:"enableInboudSniffing,omitempty"`
	// EnableOutboundSniffing enables protocol sniffing on outbound traffic.
	// .Values.pilot.enableProtocolSniffingForOutbound
	EnableOutboundSniffing bool `json:"enableOutboundSniffing,omitempty"`
}

// ProxyDNSConfig is used to configure aspects of the sidecar's DNS usage.
type ProxyDNSConfig struct {
	// SearchSuffixes are additional search suffixes to be used when resolving
	// names.
	// .Values.global.podDNSSearchNamespaces
	// Custom DNS config for the pod to resolve names of services in other
	// clusters. Use this to add additional search domains, and other settings.
	// see
	// https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#dns-config
	// This does not apply to gateway pods as they typically need a different
	// set of DNS settings than the normal application pods (e.g., in
	// multicluster scenarios).
	// NOTE: If using templates, follow the pattern in the commented example below.
	//    podDNSSearchNamespaces:
	//    - global
	//    - "{{ valueOrDefault .DeploymentMeta.Namespace \"default\" }}.global"
	SearchSuffixes []string `json:"searchSuffixes,omitempty"`
}

type ProxyRuntimeConfig struct {
	Readiness ProxyReadinessConfig
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
}

type ProxyReadinessConfig struct {
	// .Values.sidecarInjectorWebhook.rewriteAppHTTPProbe, defaults to false
	// rewrite probes for application pods to route through sidecar
	RewriteApplicationProbes bool
	// .Values.global.proxy.statusPort, overridden by status.sidecar.istio.io/port, defaults to 15020
	// Default port for Pilot agent health checks. A value of 0 will disable health checking.
	// XXX: this has no affect on which port is actually used for status.
	StatusPort int32
	// .Values.global.proxy.readinessInitialDelaySeconds, overridden by readiness.status.sidecar.istio.io/initialDelaySeconds, defaults to 1
	InitialDelaySeconds int32
	// .Values.global.proxy.readinessPeriodSeconds, overridden by readiness.status.sidecar.istio.io/periodSeconds, defaults to 2
	PeriodSeconds int32
	// .Values.global.proxy.readinessFailureThreshold, overridden by readiness.status.sidecar.istio.io/failureThreshold, defaults to 30
	FailureThreshold int32
}
