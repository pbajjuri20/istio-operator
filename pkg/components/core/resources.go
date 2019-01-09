package core

import (
	"sync"
	"text/template"

	"github.com/maistra/istio-operator/pkg/components/common"
)

type templateParams struct {
	common.TemplateParams
	PriorityClassName           string
	ReplicaCount                int
	MonitoringPort              int
	ControlPlaneSecurityEnabled bool
	ConfigureValidation         bool
}

var (
	_singleton *common.Templates
	_init      sync.Once
)

func Templates() *common.Templates {
	_init.Do(func() {
		commonTemplates := common.TemplatesInstance()
		_singleton = &common.Templates{
			ServiceAccountTemplate:     commonTemplates.ServiceAccountTemplate,
			ClusterRoleBindingTemplate: commonTemplates.ClusterRoleBindingTemplate,
			ServiceTemplate:            template.New("Service.yaml"),
			DeploymentTemplate:         template.New("Deployment.yaml"),
			ClusterRoleTemplate:        template.New("ClusterRole.yaml"),
		}
		//_singleton.ServiceTemplate.Parse(serviceYamlTemplate)
		//_singleton.DeploymentTemplate.Parse(deploymentYamlTemplate)
		//_singleton.ClusterRoleTemplate.Parse(clusterRoleYamlTemplate)
	})
	return _singleton
}

// used by galley, injector and pilot deployments
const meshConfigMapYamlTemplate = `
{{- if .Values.pilot.enabled }}
apiVersion: v1
kind: ConfigMap
metadata:
  name: istio
  namespace: {{ .Release.Namespace }}
  labels:
    app: istio
data:
  mesh: |-
    # Set the following variable to true to disable policy checks by the Mixer.
    # Note that metrics will still be reported to the Mixer.
    disablePolicyChecks: {{ .Values.global.disablePolicyChecks }}

    # Set enableTracing to false to disable request tracing.
    enableTracing: {{ .Values.global.enableTracing }}

    # Set accessLogFile to empty string to disable access log.
    accessLogFile: "{{ .Values.global.proxy.accessLogFile }}"

    # Set accessLogEncoding to JSON or TEXT to configure sidecar access log
    accessLogEncoding: {{ .Values.global.proxy.accessLogEncoding }}

    #
    # Deprecated: mixer is using EDS
    {{- if or .Values.mixer.policy.enabled .Values.mixer.telemetry.enabled }}
    {{- if .Values.global.controlPlaneSecurityEnabled }}
    mixerCheckServer: istio-policy.{{ .Release.Namespace }}.svc.cluster.local:15004
    mixerReportServer: istio-telemetry.{{ .Release.Namespace }}.svc.cluster.local:15004
    {{- else }}
    mixerCheckServer: istio-policy.{{ .Release.Namespace }}.svc.cluster.local:9091
    mixerReportServer: istio-telemetry.{{ .Release.Namespace }}.svc.cluster.local:9091
    {{- end }}

    # policyCheckFailOpen allows traffic in cases when the mixer policy service cannot be reached.
    # Default is false which means the traffic is denied when the client is unable to connect to Mixer.
    policyCheckFailOpen: {{ .Values.global.policyCheckFailOpen }}
    {{- end }}

    {{- if .Values.ingress.enabled }}
    # This is the k8s ingress service name, update if you used a different name
    ingressService: istio-{{ .Values.global.k8sIngressSelector }}
    {{- end }}

    # Unix Domain Socket through which envoy communicates with NodeAgent SDS to get
    # key/cert for mTLS. Use secret-mount files instead of SDS if set to empty. 
    sdsUdsPath: {{ .Values.global.sds.udsPath }}

    # This flag is used by secret discovery service(SDS). 
    # If set to true(prerequisite: https://kubernetes.io/docs/concepts/storage/volumes/#projected), Istio will inject volumes mount 
    # for k8s service account JWT, so that K8s API server mounts k8s service account JWT to envoy container, which 
    # will be used to generate key/cert eventually. This isn't supported for non-k8s case.
    enableSdsTokenMount: {{ .Values.global.sds.enableTokenMount }}

    # The trust domain corresponds to the trust root of a system.
    # Refer to https://github.com/spiffe/spiffe/blob/master/standards/SPIFFE-ID.md#21-trust-domain
    trustDomain: {{ .Values.global.trustDomain }}

    #
    defaultConfig:
      #
      # TCP connection timeout between Envoy & the application, and between Envoys.
      connectTimeout: 10s
      #
      ### ADVANCED SETTINGS #############
      # Where should envoy's configuration be stored in the istio-proxy container
      configPath: "/etc/istio/proxy"
      binaryPath: "/usr/local/bin/envoy"
      # The pseudo service name used for Envoy.
      serviceCluster: istio-proxy
      # These settings that determine how long an old Envoy
      # process should be kept alive after an occasional reload.
      drainDuration: 45s
      parentShutdownDuration: 1m0s
      #
      # The mode used to redirect inbound connections to Envoy. This setting
      # has no effect on outbound traffic: iptables REDIRECT is always used for
      # outbound connections.
      # If "REDIRECT", use iptables REDIRECT to NAT and redirect to Envoy.
      # The "REDIRECT" mode loses source addresses during redirection.
      # If "TPROXY", use iptables TPROXY to redirect to Envoy.
      # The "TPROXY" mode preserves both the source and destination IP
      # addresses and ports, so that they can be used for advanced filtering
      # and manipulation.
      # The "TPROXY" mode also configures the sidecar to run with the
      # CAP_NET_ADMIN capability, which is required to use TPROXY.
      #interceptionMode: REDIRECT
      #
      # Port where Envoy listens (on local host) for admin commands
      # You can exec into the istio-proxy container in a pod and
      # curl the admin port (curl http://localhost:15000/) to obtain
      # diagnostic information from Envoy. See
      # https://lyft.github.io/envoy/docs/operations/admin.html
      # for more details
      proxyAdminPort: 15000
      #
      # Set concurrency to a specific number to control the number of Proxy worker threads.
      # If set to 0 (default), then start worker thread for each CPU thread/core.
      concurrency: {{ .Values.global.proxy.concurrency }}
      #
      {{- if eq .Values.global.proxy.tracer "lightstep" }}
      tracing:
        lightstep:
          # Address of the LightStep Satellite pool
          address: {{ .Values.global.tracer.lightstep.address }}
          # Access Token used to communicate with the Satellite pool
          accessToken: {{ .Values.global.tracer.lightstep.accessToken }}
          # Whether communication with the Satellite pool should be secure
          secure: {{ .Values.global.tracer.lightstep.secure }}
          # Path to the file containing the cacert to use when verifying TLS
          cacertPath: {{ .Values.global.tracer.lightstep.cacertPath }}
      {{- else if eq .Values.global.proxy.tracer "zipkin" }}
      tracing:
        zipkin:
          # Address of the Zipkin collector
        {{- if .Values.global.tracer.zipkin.address }}
          address: {{ .Values.global.tracer.zipkin.address }}
        {{- else }}
          address: zipkin.{{ .Release.Namespace }}:9411
        {{- end }}
      {{- end }}

    {{- if .Values.global.proxy.envoyStatsd.enabled }}
      #
      # Statsd metrics collector converts statsd metrics into Prometheus metrics.
      statsdUdpAddress: {{ .Values.global.proxy.envoyStatsd.host }}.{{ .Release.Namespace }}:{{ .Values.global.proxy.envoyStatsd.port }}
    {{- end }}

    {{- if .Values.global.controlPlaneSecurityEnabled }}
      #
      # Mutual TLS authentication between sidecars and istio control plane.
      controlPlaneAuthPolicy: MUTUAL_TLS
      #
      # Address where istio Pilot service is running
      discoveryAddress: istio-pilot.{{ .Release.Namespace }}:15011
    {{- else }}
      #
      # Mutual TLS authentication between sidecars and istio control plane.
      controlPlaneAuthPolicy: NONE
      #
      # Address where istio Pilot service is running
      discoveryAddress: istio-pilot.{{ .Release.Namespace }}:15010
    {{- end }}
  
  # Configuration file for the mesh networks to be used by the Split Horizon EDS.
  meshNetworks: |-
  {{- if .Values.global.meshNetworks }}
    networks:
{{ toYaml .Values.global.meshNetworks | indent 6 }}
  {{- else }}
    networks: {}
  {{- end }}
{{- end }}
`
