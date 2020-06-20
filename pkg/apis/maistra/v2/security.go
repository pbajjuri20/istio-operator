package v2

// SecurityConfig specifies security aspects of the control plane.
type SecurityConfig struct {
	// MutualTLS configures mutual TLS for the control plane and mesh
	MutualTLS MutualTLSConfig `json:"mutualTLS,omitempty"`
}

// MutualTLSConfig specifies mutual TLS configuration for the control plane and mesh.
type MutualTLSConfig struct {
	// Auto configures the mesh to automatically detect whether or not mutual
	// TLS is required for a specific connection.
	// .Values.global.mtls.auto
	Auto bool `json:"auto,omitempty"`
	// Enable mutual TLS by default.
	// .Values.global.mtls.enabled
	Enable bool `json:"enable,omitempty"`
	// Trust configures trust aspects associated with mutual TLS clients.
	Trust TrustConfig `json:"trust,omitempty"`
	// CertificateAuthority configures the certificate authority used by the
	// control plane to create and sign client certs and server keys.
	CertificateAuthority CertificateAuthorityConfig `json:"certificateAuthority,omitempty"`
	// Identity configures the types of user tokens used by clients.
	Identity IdentityConfig `json:"identity,omitempty"`
	// ControlPlane configures mutual TLS for control plane communication.
	ControlPlane ControlPlaneMTLSConfig `json:"identity,omitempty"`
}

// TrustConfig configures trust aspects associated with mutual TLS clients
type TrustConfig struct {
	// Domain specifies the trust domain to be used by the mesh.
	//.Values.global.trustDomain, maps to trustDomain
	// The trust domain corresponds to the trust root of a system.
	// Refer to https://github.com/spiffe/spiffe/blob/master/standards/SPIFFE-ID.md#21-trust-domain
	// XXX: can this be consolidated with clusterDomainSuffix?
	Domain string `json:"domain,omitempty"`
	// AdditionalDomains are additional SPIFFE trust domains that are accepted as trusted.
	// .Values.global.trustDomainAliases, maps to trustDomainAliases
	//  Any service with the identity "td1/ns/foo/sa/a-service-account", "td2/ns/foo/sa/a-service-account",
	//  or "td3/ns/foo/sa/a-service-account" will be treated the same in the Istio mesh.
	AdditionalDomains []string `json:"additionalDomains,omitempty"`
}

// CertificateAuthorityConfig configures the certificate authority implementation
// used by the control plane.
type CertificateAuthorityConfig struct {
	// Type is the certificate authority to use.
	// .Values.global.pilotCertProvider (istiod, kubernetes, custom)
	Type CertificateAuthorityType `json:"type,omitempty"`
	// Istiod is the configuration for Istio's internal certificate authority implementation.
	// each of these produces a CAEndpoint, i.e. CA_ADDR
	Istiod *IstiodCertificateAuthorityConfig `json:"istiod,omitempty"`
	// Custom is the configuration for a custom certificate authority.
	Custom *CustomCertificateAuthorityConfig `json:"custom,omitempty"`
}

// CertificateAuthorityType represents the type of CertificateAuthority implementation.
type CertificateAuthorityType string

const (
	// CertificateAuthorityTypeIstiod represents Istio's internal certificate authority implementation
	CertificateAuthorityTypeIstiod CertificateAuthorityType = "Istiod"
	// CertificateAuthorityTypeCustom represents a custom certificate authority implementation
	CertificateAuthorityTypeCustom CertificateAuthorityType = "Custom"
)

// IstiodCertificateAuthorityConfig is the configuration for Istio's internal
// certificate authority implementation.
type IstiodCertificateAuthorityConfig struct {
	// Type of certificate signer to use.
	// .Values.global.jwtPolicy, local=first-party-jwt, external=third-party-jwt
	Type IstioCertificateSignerType `json:"type,omitempty"`
	// SelfSigned configures istiod to generate and use a self-signed certificate for the root.
	SelfSigned *IstioSelfSignedCertificateSignerConfig `json:"selfSigned,omitempty"`
	// PrivateKey configures istiod to use a user specified private key/cert when signing certificates.
	PrivateKey *IstioPrivateKeyCertificateSignerConfig `json:"privateKey,omitempty"`
	// WorkloadCertTTLDefault is the default TTL for generated workload
	// certificates.  Used if not specified in CSR (<= 0)
	// env DEFAULT_WORKLOAD_CERT_TTL
	// defaults to 24 hours
	WorkloadCertTTLDefault string `json:"workloadCertTTLDefault,omitempty"`
	// WorkloadCertTTLMax is the maximum TTL for generated workload certificates.
	// env MAX_WORKLOAD_CERT_TTL
	// defaults to 90 days
	WorkloadCertTTLMax string `json:"workloadCertTTLMax,omitempty"`
}

// IstioCertificateSignerType represents the certificate signer implementation used by istiod.
type IstioCertificateSignerType string

const (
	// IstioCertificateSignerTypePrivateKey is the signer type used when signing with a user specified private key.
	IstioCertificateSignerTypePrivateKey IstioCertificateSignerType = "PrivateKey"
	// IstioCertificateSignerTypeSelfSigned is the signer type used when signing with a generated, self-signed certificate.
	IstioCertificateSignerTypeSelfSigned IstioCertificateSignerType = "SelfSigned"
)

// IstioSelfSignedCertificateSignerConfig is the configuration for using a
// self-signed root certificate.
type IstioSelfSignedCertificateSignerConfig struct {
	// TTL for self-signed root certificate
	// env CITADEL_SELF_SIGNED_CA_CERT_TTL
	// default is 10 years
	TTL string `json:"ttl,omitempty"`
	// GracePeriod percentile for self-signed cert
	// env CITADEL_SELF_SIGNED_ROOT_CERT_GRACE_PERIOD_PERCENTILE
	// default is 20%
	GracePeriod string `json:"gracePeriod,omitempty"`
	// CheckPeriod is the interval with which certificate is checked for rotation
	// env CITADEL_SELF_SIGNED_ROOT_CERT_CHECK_INTERVAL
	// default is 1 hour, zero or negative value disables cert rotation
	CheckPeriod string `json:"checkPeriod,omitempty"`
	// EnableJitter to use jitter for cert rotation
	// env CITADEL_ENABLE_JITTER_FOR_ROOT_CERT_ROTATOR
	// defaults to true
	EnableJitter bool `json:"enableJitter,omitempty"`
	// Org is the Org value in the certificate.
	// XXX: currently uses TrustDomain.  I don't think this is configurable.
	Org string `json:"org,omitempty"`
}

// IstioPrivateKeyCertificateSignerConfig is the configuration when using a user
// supplied private key/cert for signing.
// XXX: nothing in here is currently configurable, except RootCADir
type IstioPrivateKeyCertificateSignerConfig struct {
	// hard coded to use a secret named cacerts
	EncryptionSecret string `json:"encryptionSecret,omitempty"`
	// ROOT_CA_DIR, defaults to /etc/cacerts
	// Mount directory for encryption secret
	// XXX: currently, not configurable in the charts
	RootCADir string `json:"rootCADir,omitempty"`
	// hard coded to ca-key.pem
	SigningKeyFile string `json:"signingKeyFile,omitempty"`
	// hard coded to ca-cert.pem
	SigningCertFile string `json:"signingCertFile,omitempty"`
	// hard coded to root-cert.pem
	RootCertFile string `json:"rootCertFile,omitempty"`
	// hard coded to cert-chain.pem
	CertChainFile string `json:"certChainFile,omitempty"`
}

// CustomCertificateAuthorityConfig is the configuration for a custom
// certificate authority.
type CustomCertificateAuthorityConfig struct {
	// Address is the grpc address for an Istio compatible certificate authority endpoint.
	// .Values.global.caAddress
	// XXX: assumption is this is a grpc endpoint that provides methods like istio.v1.auth.IstioCertificateService/CreateCertificate
	Address string `json:"address,omitempty"`
}

// IdentityConfig configures the types of user tokens used by clients
type IdentityConfig struct {
	// Type is the type of identity tokens being used.
	// .Values.global.jwtPolicy
	Type IdentityConfigType `json:"type,omitempty"`
	// Kubernetes configures istiod to use Kubernetes service account tokens to
	// identify users.
	Kubernetes *KubernetesIdentityConfig `json:"kubernetes,omitempty"`
	// ThirdParty configures istiod to use a third-party token provider for
	// identifying users.
	ThirdParty *ThirdPartyIdentityConfig `json:"thirdParty,omitempty"`
}

// IdentityConfigType represents the identity implementation being used.
type IdentityConfigType string

const (
	// IdentityConfigTypeKubernetes specifies Kubernetes as the token provider.
	IdentityConfigTypeKubernetes IdentityConfigType = "Kubernetes" // first-party-jwt
	// IdentityConfigTypeThirdParty specifies a third-party token provider.
	IdentityConfigTypeThirdParty IdentityConfigType = "ThirdParty" // third-party-jwt
)

// KubernetesIdentityConfig is the Kubernetes identity configuration settings.
// implies jwtPolicy=first-party-jwt, uses /var/run/secrets/kubernetes.io/serviceaccount/token
type KubernetesIdentityConfig struct {
}

// ThirdPartyIdentityConfig configures a third-party token provider for use with
// istiod.
type ThirdPartyIdentityConfig struct {
	// TokenPath is the path to the token used to identify the workload.
	// default /var/run/secrets/tokens/istio-token
	// XXX: projects service account token with specified audience (istio-ca)
	// XXX: not configurable
	TokenPath string `json:"tokenPath,omitempty"`
	// Issuer is the URL of the issuer.
	// env TOKEN_ISSUER, defaults to iss in specified token
	Issuer string `json:"issuer,omitempty"`
	// Audience is the audience for whom the token is intended.
	// env AUDIENCE
	// .Values.global.sds.token.aud, defaults to istio-ca
	Audience string `json:"audience,omitempty"`
}

// ControlPlaneMTLSConfig is the mutual TLS configuration specific to the
// control plane.
type ControlPlaneMTLSConfig struct {
	// Enable mutual TLS for the control plane components.
	// .Values.global.controlPlaneSecurityEnabled
	Enable bool `json:"enable,omitempty"`
	// CertProvider is the certificate authority used to generate the serving
	// certificates for the control plane components.
	// .Values.global.pilotCertProvider
	// Provider used to generate serving certs for istiod (pilot)
	CertProvider ControlPlaneCertProviderType `json:"certProvider,omitempty"`
}

// ControlPlaneCertProviderType represents the provider used to generate serving
// certificates for the control plane.
type ControlPlaneCertProviderType string

const (
	// ControlPlaneCertProviderTypeIstiod identifies istiod as the provider generating the serving certifications.
	ControlPlaneCertProviderTypeIstiod ControlPlaneCertProviderType = "Istiod"
	// ControlPlaneCertProviderTypeKubernetes identifies Kubernetes as the provider generating the serving certificates.
	ControlPlaneCertProviderTypeKubernetes ControlPlaneCertProviderType = "Kubernetes"
	// ControlPlaneCertProviderTypeCustom identifies a custom provider has generated the serving certificates.
	// XXX: Not quite sure what this means. Presumably, the key and cert chain have been mounted specially
	ControlPlaneCertProviderTypeCustom ControlPlaneCertProviderType = "Custom"
)
