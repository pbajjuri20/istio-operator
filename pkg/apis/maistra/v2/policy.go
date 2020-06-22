package v2

// PolicyConfig configures policy aspects of the mesh.
type PolicyConfig struct {
	// Required, the policy implementation
	Type PolicyType `json:"type,omitempty"`
	// Mixer configuration (legacy, v1)
	// .Values.mixer.policy.enabled
	Mixer *MixerPolicyConfig `json:"mixer,omitempty"`
	// Remote mixer configuration (legacy, v1)
	// .Values.mixer.policy.remotePolicyAddress
	Remote *RemotePolicyConfig `json:"remote,omitempty"`
	// Istiod policy implementation (v2)
	// XXX: is this the default policy config, i.e. what's used if mixer is not
	// being used?  Does this need to be explicit?
	Istiod *IstiodPolicyConfig `json:"istiod,omitempty"`
}

// PolicyType represents the type of policy implementation used by the mesh.
type PolicyType string

const (
	// PolicyTypeMixer represents mixer, v1 implementation
	PolicyTypeMixer PolicyType = "Mixer"
	// PolicyTypeRemote represents remote mixer, v1 implementation
	PolicyTypeRemote PolicyType = "Remote"
	// PolicyTypeIstiod represents istio, v2 implementation
	PolicyTypeIstiod PolicyType = "Istiod"
)

// MixerPolicyConfig configures a mixer implementation for policy
// .Values.mixer.policy.enabled
type MixerPolicyConfig struct {
	// EnableChecks configures whether or not policy checks should be enabled.
	// .Values.global.disablePolicyChecks | default "true" (false, inverted logic)
	// Set the following variable to false to disable policy checks by the Mixer.
	// Note that metrics will still be reported to the Mixer.
	EnableChecks bool `json:"enableChecks,omitempty"`
	// FailOpen configures policy checks to fail if mixer cannot be reached.
	// .Values.global.policyCheckFailOpen, maps to MeshConfig.policyCheckFailOpen
	// policyCheckFailOpen allows traffic in cases when the mixer policy service cannot be reached.
	// Default is false which means the traffic is denied when the client is unable to connect to Mixer.
	FailOpen bool `json:"failOpen,omitempty"`
	// Runtime configures execution aspects of the mixer deployment/pod (e.g. resources)
	Runtime *DeploymentRuntimeConfig `json:"runtime,omitempty"`
	// Adapters configures available adapters.
	Adapters *MixerPolicyAdaptersConfig `json:"adapters,omitempty"`
}

// MixerPolicyAdaptersConfig configures policy adapters for mixer.
type MixerPolicyAdaptersConfig struct {
	// UseAdapterCRDs configures mixer to support deprecated mixer CRDs.
	// .Values.mixer.policy.adapters.useAdapterCRDs, removed in istio 1.4, defaults to false
	// XXX: I don't think this should ever be used, as the CRDs were supported in istio 1.1, but removed entirely in 1.4
	UseAdapterCRDs bool `json:"useAdapterCRDs,omitempty"`
	// Kubernetesenv configures the use of the kubernetesenv adapter.
	// .Values.mixer.policy.adapters.kubernetesenv.enabled, defaults to true
	KubernetesEnv bool `json:"kubernetesenv,omitempty"`
}

// RemotePolicyConfig configures a remote mixer instance for policy
type RemotePolicyConfig struct {
	// Address represents the address of the mixer server.
	// .Values.global.remotePolicyAddress, maps to MeshConfig.mixerCheckServer
	Address string `json:"address,omitempty"`
	// CreateServices specifies whether or not a k8s Service should be created for the remote policy server.
	// .Values.global.createRemoteSvcEndpoints
	CreateService bool `json:"createService,omitempty"`
	// EnableChecks configures whether or not policy checks should be enabled.
	// .Values.global.disablePolicyChecks | default "true" (false, inverted logic)
	// Set the following variable to false to disable policy checks by the Mixer.
	// Note that metrics will still be reported to the Mixer.
	EnableChecks bool `json:"enableChecks,omitempty"`
	// FailOpen configures policy checks to fail if mixer cannot be reached.
	// .Values.global.policyCheckFailOpen, maps to MeshConfig.policyCheckFailOpen
	// policyCheckFailOpen allows traffic in cases when the mixer policy service cannot be reached.
	// Default is false which means the traffic is denied when the client is unable to connect to Mixer.
	FailOpen bool `json:"failOpen,omitempty"`
}

// IstiodPolicyConfig configures policy aspects of istiod
type IstiodPolicyConfig struct{}
