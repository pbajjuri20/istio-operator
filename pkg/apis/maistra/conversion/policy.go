package conversion

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	v2 "github.com/maistra/istio-operator/pkg/apis/maistra/v2"
	"github.com/maistra/istio-operator/pkg/controller/versions"
)

// XXX: Not all of the settings are mapped correctly, as there are differences
// between v1.0/v1.1 and v2.0

func populatePolicyValues(in *v2.ControlPlaneSpec, values map[string]interface{}) error {
	// Cluster settings
	if in.Policy == nil {
		return nil
	}
	if in.Policy.Type == v2.PolicyTypeNone {
		return setHelmBoolValue(values, "mixer.policy.enabled", false)
	}

	istiod := !(in.Version == "" || in.Version == versions.V1_0.String() || in.Version == versions.V1_1.String())
	if in.Policy.Type == "" {
		if istiod {
			in.Policy.Type = v2.PolicyTypeIstiod
		} else {
			in.Policy.Type = v2.PolicyTypeMixer
		}
	}
	switch in.Policy.Type {
	case v2.PolicyTypeMixer:
		return populateMixerPolicyValues(in, istiod, values)
	case v2.PolicyTypeRemote:
		return populateRemotePolicyValues(in, values)
	case v2.PolicyTypeIstiod:
		return populateIstiodPolicyValues(in, values)
	}
	setHelmBoolValue(values, "mixer.policy.enabled", false)
	return fmt.Errorf("Unknown policy type: %s", in.Policy.Type)
}

func populateMixerPolicyValues(in *v2.ControlPlaneSpec, istiod bool, values map[string]interface{}) error {
	mixer := in.Policy.Mixer
	if mixer == nil {
		mixer = &v2.MixerPolicyConfig{}
	}

	// Make sure mixer is enabled
	if err := setHelmBoolValue(values, "mixer.enabled", true); err != nil {
		return err
	}

	policyValues := make(map[string]interface{})
	if err := setHelmBoolValue(policyValues, "enabled", true); err != nil {
		return err
	}
	if mixer.EnableChecks != nil {
		if err := setHelmBoolValue(values, "global.disablePolicyChecks", !*mixer.EnableChecks); err != nil {
			return err
		}
	}
	if mixer.FailOpen != nil {
		if err := setHelmBoolValue(values, "global.policyCheckFailOpen", *mixer.FailOpen); err != nil {
			return err
		}
	}

	if mixer.Adapters != nil {
		adaptersValues := make(map[string]interface{})
		if mixer.Adapters.UseAdapterCRDs != nil {
			if err := setHelmBoolValue(adaptersValues, "useAdapterCRDs", *mixer.Adapters.UseAdapterCRDs); err != nil {
				return err
			}
		}
		if mixer.Adapters.KubernetesEnv != nil {
			if err := setHelmBoolValue(adaptersValues, "kubernetesenv.enabled", *mixer.Adapters.KubernetesEnv); err != nil {
				return err
			}
		}
		if len(adaptersValues) > 0 {
			if istiod {
				if err := setHelmValue(policyValues, "adapters", adaptersValues); err != nil {
					return err
				}
			} else {
				if err := setHelmValue(values, "mixer.adapters", adaptersValues); err != nil {
					return err
				}
			}
		}
	}

	// Deployment specific settings
	runtime := mixer.Runtime
	if runtime != nil {
		if err := populateRuntimeValues(runtime, policyValues); err != nil {
			return err
		}

		// set image and resources
		if runtime.Container != nil {
			if runtime.Container.Image != "" {
				if istiod {
					if err := setHelmStringValue(policyValues, "image", runtime.Container.Image); err != nil {
						return err
					}
				} else {
					// XXX: this applies to both policy and telemetry in pre 1.6
					if err := setHelmStringValue(values, "mixer.image", runtime.Container.Image); err != nil {
						return err
					}
				}
			}
			if runtime.Container.Resources != nil {
				if resourcesValues, err := toValues(runtime.Container.Resources); err == nil {
					if len(resourcesValues) > 0 {
						if err := setHelmValue(policyValues, "resources", resourcesValues); err != nil {
							return err
						}
					}
				} else {
					return err
				}
			}
		}
	}

	if !istiod {
		// move podAnnotations, nodeSelector, podAntiAffinityLabelSelector, and
		// podAntiAffinityTermLabelSelector from mixer.policy to mixer for v1.0 and v1.1
		// Note, these may overwrite settings specified in telemetry
		if podAnnotations, found, _ := unstructured.NestedFieldCopy(policyValues, "podAnnotations"); found {
			if err := setHelmValue(values, "mixer.podAnnotations", podAnnotations); err != nil {
				return err
			}
		}
		if nodeSelector, found, _ := unstructured.NestedFieldCopy(policyValues, "nodeSelector"); found {
			if err := setHelmValue(values, "mixer.nodeSelector", nodeSelector); err != nil {
				return err
			}
		}
		if podAntiAffinityLabelSelector, found, _ := unstructured.NestedFieldCopy(policyValues, "podAntiAffinityLabelSelector"); found {
			if err := setHelmValue(values, "mixer.podAntiAffinityLabelSelector", podAntiAffinityLabelSelector); err != nil {
				return err
			}
		}
		if podAntiAffinityTermLabelSelector, found, _ := unstructured.NestedFieldCopy(policyValues, "podAntiAffinityTermLabelSelector"); found {
			if err := setHelmValue(values, "mixer.podAntiAffinityTermLabelSelector", podAntiAffinityTermLabelSelector); err != nil {
				return err
			}
		}
	}

	// set the policy values
	if len(policyValues) > 0 {
		if err := setHelmValue(values, "mixer.policy", policyValues); err != nil {
			return err
		}
	}

	return nil
}

func populateRemotePolicyValues(in *v2.ControlPlaneSpec, values map[string]interface{}) error {
	remote := in.Policy.Remote
	if remote == nil {
		remote = &v2.RemotePolicyConfig{}
	}

	// Make sure mixer is disabled
	if err := setHelmBoolValue(values, "mixer.enabled", false); err != nil {
		return err
	}
	if err := setHelmBoolValue(values, "mixer.policy.enabled", true); err != nil {
		return err
	}

	if err := setHelmStringValue(values, "global.remotePolicyAddress", remote.Address); err != nil {
		return err
	}
	// XXX: this applies to both policy and telemetry
	if err := setHelmBoolValue(values, "global.createRemoteSvcEndpoints", remote.CreateService); err != nil {
		return err
	}
	if remote.EnableChecks != nil {
		if err := setHelmBoolValue(values, "global.disablePolicyChecks", !*remote.EnableChecks); err != nil {
			return err
		}
	}
	if remote.FailOpen != nil {
		if err := setHelmBoolValue(values, "global.policyCheckFailOpen", *remote.FailOpen); err != nil {
			return err
		}
	}

	return nil
}

func populateIstiodPolicyValues(in *v2.ControlPlaneSpec, values map[string]interface{}) error {
	if err := setHelmBoolValue(values, "mixer.enabled", false); err != nil {
		return err
	}
	if err := setHelmBoolValue(values, "mixer.policy.enabled", false); err != nil {
		return err
	}
	return nil
}
