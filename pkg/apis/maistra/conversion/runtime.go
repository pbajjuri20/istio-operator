package conversion

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	v1 "github.com/maistra/istio-operator/pkg/apis/maistra/v1"
	v2 "github.com/maistra/istio-operator/pkg/apis/maistra/v2"
)

func populateControlPlaneRuntimeValues(runtime *v2.ControlPlaneRuntimeConfig, values map[string]interface{}) error {
	if runtime == nil {
		return nil
	}

	if runtime.Defaults != nil {
		// defaultNodeSelector, defaultTolerations, defaultPodDisruptionBudget, priorityClassName
		deployment := runtime.Defaults.Deployment
		if deployment != nil {
			if deployment.PodDisruption != nil {
				if err := setHelmBoolValue(values, "global.defaultPodDisruptionBudget.enabled", true); err != nil {
					return err
				}
				if deployment.PodDisruption.MinAvailable != nil {
					if deployment.PodDisruption.MinAvailable.Type == intstr.Int {
						if err := setHelmIntValue(values, "global.defaultPodDisruptionBudget.minAvailable",
							int64(deployment.PodDisruption.MinAvailable.IntValue())); err != nil {
							return err
						}
					} else {
						if err := setHelmStringValue(values, "global.defaultPodDisruptionBudget.minAvailable",
							deployment.PodDisruption.MinAvailable.String()); err != nil {
							return err
						}
					}
				}
				if deployment.PodDisruption.MaxUnavailable != nil {
					if deployment.PodDisruption.MaxUnavailable.Type == intstr.Int {
						if err := setHelmIntValue(values, "global.defaultPodDisruptionBudget.maxUnavailable",
							int64(deployment.PodDisruption.MaxUnavailable.IntValue())); err != nil {
							return err
						}
					} else {
						if err := setHelmStringValue(values, "global.defaultPodDisruptionBudget.maxUnavailable",
							deployment.PodDisruption.MaxUnavailable.String()); err != nil {
							return err
						}
					}
				}
			}
		}
		pod := runtime.Defaults.Pod
		if pod != nil {
			if len(pod.NodeSelector) > 0 {
				if err := setHelmStringMapValue(values, "global.defaultNodeSelector", pod.NodeSelector); err != nil {
					return err
				}
			}
			if len(pod.Tolerations) > 0 {
				untypedSlice := make([]interface{}, len(pod.Tolerations))
				for index, toleration := range pod.Tolerations {
					untypedSlice[index] = toleration
				}
				if tolerations, err := sliceToValues(untypedSlice); err == nil {
					if len(tolerations) > 0 {
						if err := setHelmValue(values, "global.defaultTolerations", tolerations); err != nil {
							return err
						}
					}
				} else {
					return err
				}
			}
			if pod.PriorityClassName != "" {
				if err := setHelmStringValue(values, "global.priorityClassName", pod.PriorityClassName); err != nil {
					return err
				}
			}
		}
		container := runtime.Defaults.Container
		if container != nil {
			if container.ImagePullPolicy != "" {
				if err := setHelmStringValue(values, "global.imagePullPolicy", string(container.ImagePullPolicy)); err != nil {
					return err
				}
			}
			if len(container.ImagePullSecrets) > 0 {
				pullSecretsValues := make([]string, 0)
				for _, secret := range container.ImagePullSecrets {
					pullSecretsValues = append(pullSecretsValues, secret.Name)
				}
				if err := setHelmStringSliceValue(values, "global.imagePullSecrets", pullSecretsValues); err != nil {
					return err
				}
			}
			if container.ImageRegistry != "" {
				if err := setHelmStringValue(values, "global.hub", container.ImageRegistry); err != nil {
					return err
				}
			}
			if container.ImageTag != "" {
				if err := setHelmStringValue(values, "global.tag", container.ImageTag); err != nil {
					return err
				}
			}
			if container.Resources != nil {
				if resourcesValues, err := toValues(container.Resources); err == nil {
					if len(resourcesValues) > 0 {
						if err := setHelmValue(values, "global.defaultResources", resourcesValues); err != nil {
							return err
						}
					}
				} else {
					return err
				}
			}
		}
	}

	for component, config := range runtime.Components {
		componentValues := make(map[string]interface{})
		if err := populateRuntimeValues(&config, componentValues); err == nil {
			for key, value := range componentValues {
				if err := setHelmValue(values, string(component)+"."+key, value); err != nil {
					return err
				}
			}
		} else {
			return err
		}
	}
	return nil
}

func populateDefaultContainerValues(containers map[string]v2.ContainerConfig, values map[string]interface{}) error {
	if containers == nil {
		return nil
	}
	if defaultContainer, ok := containers["default"]; ok {
		return populateContainerConfigValues(&defaultContainer, values)
	}
	return nil
}

func populateRuntimeValues(runtime *v2.ComponentRuntimeConfig, componentValues map[string]interface{}) error {
	if runtime == nil {
		return nil
	}
	if err := populateDeploymentHelmValues(runtime.Deployment, componentValues); err != nil {
		return err
	}
	if err := populatePodHelmValues(runtime.Pod, componentValues); err != nil {
		return err
	}
	if err := populateContainerConfigValues(runtime.Container, componentValues); err != nil {
		return err
	}
	if runtime.Deployment != nil {
	}

	return nil
}

func populateDeploymentHelmValues(deployment *v2.DeploymentRuntimeConfig, componentValues map[string]interface{}) error {
	if deployment == nil {
		return nil
	}
	if deployment.Replicas != nil {
		if err := setHelmIntValue(componentValues, "replicaCount", int64(*deployment.Replicas)); err != nil {
			return err
		}
	}
	// labels are populated from Service.Metadata.Labels
	if deployment.Strategy != nil && deployment.Strategy.RollingUpdate != nil {
		if deployment.Strategy.RollingUpdate.MaxSurge != nil {
			if deployment.Strategy.RollingUpdate.MaxSurge.Type == intstr.Int {
				if err := setHelmIntValue(componentValues, "rollingMaxSurge", int64(deployment.Strategy.RollingUpdate.MaxSurge.IntValue())); err != nil {
					return err
				}
			} else {
				if err := setHelmStringValue(componentValues, "rollingMaxSurge", deployment.Strategy.RollingUpdate.MaxSurge.String()); err != nil {
					return err
				}
			}
		}
		if deployment.Strategy.RollingUpdate.MaxUnavailable != nil {
			if deployment.Strategy.RollingUpdate.MaxUnavailable.Type == intstr.Int {
				if err := setHelmIntValue(componentValues, "rollingMaxUnavailable", int64(deployment.Strategy.RollingUpdate.MaxUnavailable.IntValue())); err != nil {
					return err
				}
			} else {
				if err := setHelmStringValue(componentValues, "rollingMaxUnavailable", deployment.Strategy.RollingUpdate.MaxUnavailable.String()); err != nil {
					return err
				}
			}
		}
	}
	if err := populateAutoscalingHelmValues(deployment.AutoScaling, componentValues); err != nil {
		return err
	}
	return nil
}

func populatePodHelmValues(pod *v2.PodRuntimeConfig, componentValues map[string]interface{}) error {
	if pod == nil {
		return nil
	}
	if len(pod.Metadata.Annotations) > 0 {
		if err := setHelmStringMapValue(componentValues, "podAnnotations", pod.Metadata.Annotations); err != nil {
			return err
		}
	}
	if len(pod.Metadata.Labels) > 0 {
		if err := setHelmStringMapValue(componentValues, "podLabels", pod.Metadata.Labels); err != nil {
			return err
		}
	}
	if pod.PriorityClassName != "" {
		// XXX: this is only available with global.priorityClassName
		if err := setHelmStringValue(componentValues, "priorityClassName", pod.PriorityClassName); err != nil {
			return err
		}
	}

	// Scheduling
	if len(pod.NodeSelector) > 0 {
		if err := setHelmStringMapValue(componentValues, "nodeSelector", pod.NodeSelector); err != nil {
			return err
		}
	}
	if pod.Affinity != nil {
		// NodeAffinity is not supported, only NodeSelector may be used.
		// PodAffinity is not supported.
		if len(pod.Affinity.PodAntiAffinity.RequiredDuringScheduling) > 0 {
			podAntiAffinityLabelSelector := convertAntiAffinityTermsToHelmValues(pod.Affinity.PodAntiAffinity.RequiredDuringScheduling)
			if len(podAntiAffinityLabelSelector) > 0 {
				if err := setHelmValue(componentValues, "podAntiAffinityLabelSelector", podAntiAffinityLabelSelector); err != nil {
					return err
				}
			}
		}
		if len(pod.Affinity.PodAntiAffinity.PreferredDuringScheduling) > 0 {
			podAntiAffinityTermLabelSelector := convertAntiAffinityTermsToHelmValues(pod.Affinity.PodAntiAffinity.PreferredDuringScheduling)
			if len(podAntiAffinityTermLabelSelector) > 0 {
				if err := setHelmValue(componentValues, "podAntiAffinityTermLabelSelector", podAntiAffinityTermLabelSelector); err != nil {
					return err
				}
			}
		}
	}
	if len(pod.Tolerations) > 0 {
		untypedSlice := make([]interface{}, len(pod.Tolerations))
		for index, toleration := range pod.Tolerations {
			untypedSlice[index] = toleration
		}
		if tolerations, err := sliceToValues(untypedSlice); err == nil {
			if len(tolerations) > 0 {
				if err := setHelmValue(componentValues, "tolerations", tolerations); err != nil {
					return err
				}
			}
		} else {
			return err
		}
	}
	return nil
}

func convertAntiAffinityTermsToHelmValues(terms []v2.PodAntiAffinityTerm) []interface{} {
	termsValues := make([]interface{}, 0)
	for _, term := range terms {
		termsValues = append(termsValues, map[string]interface{}{
			"key":         term.Key,
			"operator":    string(term.Operator),
			"values":      strings.Join(term.Values, ","),
			"topologyKey": term.TopologyKey,
		})
	}
	return termsValues
}

func populateAutoscalingHelmValues(autoScalerConfg *v2.AutoScalerConfig, componentValues map[string]interface{}) error {
	if autoScalerConfg == nil {
		if err := setHelmBoolValue(componentValues, "autoscaleEnabled", false); err != nil {
			return err
		}
	} else {
		if err := setHelmBoolValue(componentValues, "autoscaleEnabled", true); err != nil {
			return err
		}
		if autoScalerConfg.MinReplicas != nil {
			if err := setHelmIntValue(componentValues, "autoscaleMin", int64(*autoScalerConfg.MinReplicas)); err != nil {
				return err
			}
		}
		if autoScalerConfg.MaxReplicas != nil {
			if err := setHelmIntValue(componentValues, "autoscaleMax", int64(*autoScalerConfg.MaxReplicas)); err != nil {
				return err
			}
		}
		if autoScalerConfg.TargetCPUUtilizationPercentage != nil {
			if err := setHelmIntValue(componentValues, "cpu.targetAverageUtilization", int64(*autoScalerConfg.TargetCPUUtilizationPercentage)); err != nil {
				return err
			}
		}
	}
	return nil
}

func populateContainerConfigValues(containerConfig *v2.ContainerConfig, componentValues map[string]interface{}) error {
	if containerConfig == nil {
		return nil
	}
	if containerConfig.ImagePullPolicy != "" {
		if err := setHelmStringValue(componentValues, "imagePullPolicy", string(containerConfig.ImagePullPolicy)); err != nil {
			return err
		}
	}
	if len(containerConfig.ImagePullSecrets) > 0 {
		pullSecretsValues := make([]string, 0)
		for _, secret := range containerConfig.ImagePullSecrets {
			pullSecretsValues = append(pullSecretsValues, secret.Name)
		}
		if err := setHelmStringSliceValue(componentValues, "imagePullSecrets", pullSecretsValues); err != nil {
			return err
		}
	}
	if containerConfig.ImageRegistry != "" {
		if err := setHelmStringValue(componentValues, "hub", containerConfig.ImageRegistry); err != nil {
			return err
		}
	}
	if containerConfig.Image != "" {
		if err := setHelmStringValue(componentValues, "image", containerConfig.Image); err != nil {
			return err
		}
	}
	if containerConfig.ImageTag != "" {
		if err := setHelmStringValue(componentValues, "tag", containerConfig.ImageTag); err != nil {
			return err
		}
	}
	if containerConfig.Resources != nil {
		if resourcesValues, err := toValues(containerConfig.Resources); err == nil {
			if len(resourcesValues) > 0 {
				if err := setHelmValue(componentValues, "resources", resourcesValues); err != nil {
					return err
				}
			}
		} else {
			return err
		}
	}

	return nil
}

func runtimeValuesToComponentRuntimeConfig(in *v1.HelmValues, out *v2.ComponentRuntimeConfig) (bool, error) {
	setValues := false

	deployment := out.Deployment
	if deployment == nil {
		deployment = &v2.DeploymentRuntimeConfig{}
	}
	if applied, err := runtimeValuesToDeploymentRuntimeConfig(in, deployment); err == nil {
		if applied {
			setValues = true
			out.Deployment = deployment
		}
	} else {
		return false, err
	}

	pod := out.Pod
	if pod == nil {
		pod = &v2.PodRuntimeConfig{}
	}
	if applied, err := runtimeValuesToPodRuntimeConfig(in, pod); err == nil {
		if applied {
			setValues = true
			out.Pod = pod
		}
	} else {
		return false, err
	}

	container := out.Container
	if container == nil {
		container = &v2.ContainerConfig{}
	}
	if applied, err := populateContainerConfig(in, container); err == nil {
		if applied {
			setValues = true
			out.Container = container
		}
	} else {
		return false, err
	}

	return setValues, nil
}

func runtimeValuesToDeploymentRuntimeConfig(in *v1.HelmValues, out *v2.DeploymentRuntimeConfig) (bool, error) {
	setValues := false
	if replicaCount, ok, err := in.GetInt64("replicaCount"); ok {
		replicas := int32(replicaCount)
		out.Replicas = &replicas
		setValues = true
	} else if err != nil {
		return false, err
	}
	rollingUpdate := &appsv1.RollingUpdateDeployment{}
	setRollingUpdate := false
	if rollingMaxSurgeString, ok, err := in.GetString("rollingMaxSurge"); ok {
		maxSurge := intstr.FromString(rollingMaxSurgeString)
		rollingUpdate.MaxSurge = &maxSurge
		setRollingUpdate = true
	} else if err != nil {
		if rollingMaxSurgeInt, ok, err := in.GetInt64("rollingMaxSurge"); ok {
			maxSurge := intstr.FromInt(int(rollingMaxSurgeInt))
			rollingUpdate.MaxSurge = &maxSurge
			setRollingUpdate = true
		} else if err != nil {
			return false, err
		}
	}
	if rollingMaxUnavailableString, ok, err := in.GetString("rollingMaxUnavailable"); ok {
		maxUnavailable := intstr.FromString(rollingMaxUnavailableString)
		rollingUpdate.MaxUnavailable = &maxUnavailable
		setRollingUpdate = true
	} else if err != nil {
		if rollingMaxUnavailableInt, ok, err := in.GetInt64("rollingMaxUnavailable"); ok {
			maxUnavailable := intstr.FromInt(int(rollingMaxUnavailableInt))
			rollingUpdate.MaxUnavailable = &maxUnavailable
			setRollingUpdate = true
		} else if err != nil {
			return false, err
		}
	}
	if setRollingUpdate {
		out.Strategy = &appsv1.DeploymentStrategy{
			RollingUpdate: rollingUpdate,
		}
		setValues = true
	}
	if applied, err := runtimeValuesToAutoscalingConfig(in, out); err == nil {
		if applied {
			setValues = true
		}
	} else {
		return false, err
	}

	return setValues, nil
}

func runtimeValuesToPodRuntimeConfig(in *v1.HelmValues, out *v2.PodRuntimeConfig) (bool, error) {
	setValues := false
	if rawAnnotations, ok, err := in.GetMap("podAnnotations"); ok && len(rawAnnotations) > 0 {
		if err := setMetadataAnnotations(rawAnnotations, &out.Metadata); err != nil {
			return false, err
		}
		setValues = true
	} else if err != nil {
		return false, err
	}
	if rawLabels, ok, err := in.GetMap("podLabels"); ok && len(rawLabels) > 0 {
		if err := setMetadataLabels(rawLabels, &out.Metadata); err != nil {
			return false, err
		}
		setValues = true
	} else if err != nil {
		return false, err
	}
	if priorityClassName, ok, err := in.GetString("priorityClassName"); ok {
		out.PriorityClassName = priorityClassName
		setValues = true
	} else if err != nil {
		return false, err
	}

	// Scheduling
	if rawSelector, ok, err := in.GetMap("nodeSelector"); ok && len(rawSelector) > 0 {
		out.NodeSelector = make(map[string]string)
		for key, value := range rawSelector {
			if stringValue, ok := value.(string); ok {
				out.NodeSelector[key] = stringValue
			} else {
				return false, fmt.Errorf("non string value in nodeSelector definition")
			}
		}
		setValues = true
	} else if err != nil {
		return false, err
	}
	if rawAntiAffinityLabelSelector, ok, err := in.GetSlice("podAntiAffinityLabelSelector"); ok && len(rawAntiAffinityLabelSelector) > 0 {
		if out.Affinity == nil {
			out.Affinity = &v2.Affinity{}
		} else {
			// clear this out
			out.Affinity.PodAntiAffinity.RequiredDuringScheduling = nil
		}
		for _, rawSelector := range rawAntiAffinityLabelSelector {
			if selectorValues, ok := rawSelector.(map[string]interface{}); ok {
				term := v2.PodAntiAffinityTerm{}
				if err := affinityTermValuesToAntiAffinityTerm(v1.NewHelmValues(selectorValues), &term); err != nil {
					return false, err
				}
				out.Affinity.PodAntiAffinity.RequiredDuringScheduling = append(out.Affinity.PodAntiAffinity.RequiredDuringScheduling, term)
			} else {
				return false, fmt.Errorf("could not cast podAntiAffinityLabelSelector value to map[string]interface{}")
			}
		}
		setValues = true
	} else if err != nil {
		return false, err
	}
	if rawAntiAffinityLabelSelector, ok, err := in.GetSlice("podAntiAffinityTermLabelSelector"); ok && len(rawAntiAffinityLabelSelector) > 0 {
		if out.Affinity == nil {
			out.Affinity = &v2.Affinity{}
		} else {
			// clear this out
			out.Affinity.PodAntiAffinity.PreferredDuringScheduling = nil
		}
		for _, rawSelector := range rawAntiAffinityLabelSelector {
			if selectorValues, ok := rawSelector.(map[string]interface{}); ok {
				term := v2.PodAntiAffinityTerm{}
				if err := affinityTermValuesToAntiAffinityTerm(v1.NewHelmValues(selectorValues), &term); err != nil {
					return false, err
				}
				out.Affinity.PodAntiAffinity.PreferredDuringScheduling = append(out.Affinity.PodAntiAffinity.PreferredDuringScheduling, term)
			} else {
				return false, fmt.Errorf("could not cast podAntiAffinityTermLabelSelector value to map[string]interface{}")
			}
		}
		setValues = true
	} else if err != nil {
		return false, err
	}
	if tolerationsValues, ok, err := in.GetSlice("tolerations"); ok && len(tolerationsValues) > 0 {
		for _, tolerationValues := range tolerationsValues {
			toleration := corev1.Toleration{}
			if err := fromValues(tolerationValues, &toleration); err != nil {
				return false, err
			}
			out.Tolerations = append(out.Tolerations, toleration)
		}
		setValues = true
	} else if err != nil {
		return false, err
	}
	return setValues, nil
}

func affinityTermValuesToAntiAffinityTerm(in *v1.HelmValues, term *v2.PodAntiAffinityTerm) error {
	if key, ok, err := in.GetString("key"); ok {
		term.Key = key
	} else if err != nil {
		return err
	}
	if operator, ok, err := in.GetString("operator"); ok {
		term.Operator = metav1.LabelSelectorOperator(operator)
	} else if err != nil {
		return err
	}
	if topologyKey, ok, err := in.GetString("topologyKey"); ok {
		term.TopologyKey = topologyKey
	} else if err != nil {
		return err
	}
	if values, ok, err := in.GetString("values"); ok {
		term.Values = strings.Split(values, ",")
	} else if err != nil {
		return err
	}
	return nil
}

func runtimeValuesToAutoscalingConfig(in *v1.HelmValues, out *v2.DeploymentRuntimeConfig) (bool, error) {
	if enabled, ok, err := in.GetBool("autoscaleEnabled"); ok {
		if !enabled {
			// return true to support round-tripping
			return true, nil
		}
	} else if err != nil {
		return false, err
	}
	autoScaling := &v2.AutoScalerConfig{}
	setValues := false
	if minReplicas64, ok, err := in.GetInt64("autoscaleMin"); ok {
		minReplicas := int32(minReplicas64)
		autoScaling.MinReplicas = &minReplicas
		setValues = true
	} else if err != nil {
		return false, err
	}
	if maxReplicas64, ok, err := in.GetInt64("autoscaleMax"); ok {
		maxReplicas := int32(maxReplicas64)
		autoScaling.MaxReplicas = &maxReplicas
		setValues = true
	} else if err != nil {
		return false, err
	}
	if cpuUtilization64, ok, err := in.GetInt64("cpu.targetAverageUtilization"); ok {
		cpuUtilization := int32(cpuUtilization64)
		autoScaling.TargetCPUUtilizationPercentage = &cpuUtilization
		setValues = true
	} else if err != nil {
		return false, err
	}
	if setValues {
		out.AutoScaling = autoScaling
	}
	return setValues, nil
}

func populateControlPlaneRuntimeConfig(in *v1.HelmValues, out *v2.ControlPlaneSpec) (bool, error) {
	rawGlobalValues, ok, err := in.GetMap("global")
	if err != nil {
		return false, err
	} else if !ok || len(rawGlobalValues) == 0 {
		return false, nil
	}
	globalValues := v1.NewHelmValues(rawGlobalValues)
	runtime := &v2.ControlPlaneRuntimeConfig{}
	setRuntime := false

	defaults := &v2.DefaultRuntimeConfig{}
	setDefaults := false

	if rawPDBValues, ok, err := globalValues.GetMap("defaultPodDisruptionBudget"); ok && len(rawPDBValues) > 0 {
		pdbValues := v1.NewHelmValues(rawPDBValues)
		if pdbEnabled, ok, err := pdbValues.GetBool("enabled"); ok && pdbEnabled {
			defaults.Deployment = &v2.CommonDeploymentRuntimeConfig{
				PodDisruption: &v2.PodDisruptionBudget{},
			}
			setDefaults = true
			if err := fromValues(pdbValues, defaults.Deployment.PodDisruption); err != nil {
				return false, err
			}
		} else if err != nil {
			return false, err
		}
	} else if err != nil {
		return false, err
	}

	pod := &v2.CommonPodRuntimeConfig{}
	setPod := false
	if nodeSelector, ok, err := globalValues.GetMap("defaultNodeSelector"); ok && len(nodeSelector) > 0 {
		pod.NodeSelector = make(map[string]string)
		setPod = true
		if err := fromValues(nodeSelector, &pod.NodeSelector); err != nil {
			return false, err
		}
	} else if err != nil {
		return false, err
	}
	if tolerations, ok, err := globalValues.GetSlice("defaultTolerations"); ok && len(tolerations) > 0 {
		pod.Tolerations = make([]corev1.Toleration, len(tolerations))
		setPod = true
		for index, tolerationValues := range tolerations {
			if err := fromValues(tolerationValues, &pod.Tolerations[index]); err != nil {
				return false, err
			}
		}
	} else if err != nil {
		return false, err
	}
	if priorityClassName, ok, err := globalValues.GetString("priorityClassName"); ok {
		pod.PriorityClassName = priorityClassName
		setPod = true
	} else if err != nil {
		return false, err
	}

	if setPod {
		defaults.Pod = pod
		setDefaults = true
	}

	container := &v2.CommonContainerConfig{}
	if applied, err := populateCommonContainerConfig(globalValues, container); err != nil {
		return false, err
	} else if applied {
		defaults.Container = container
		setDefaults = true
	}
	// global resources use a different key
	if resourcesValues, ok, err := globalValues.GetMap("defaultResources"); ok && len(resourcesValues) > 0 {
		container.Resources = &corev1.ResourceRequirements{}
		if err := fromValues(resourcesValues, container.Resources); err != nil {
			return false, err
		}
		if defaults.Container == nil {
			defaults.Container = container
		}
		setDefaults = true
	} else if err != nil {
		return false, err
	}

	runtime.Components = make(map[v2.ControlPlaneComponentName]v2.ComponentRuntimeConfig)
	for _, component := range v2.ControlPlaneComponentNames {
		if componentValues, ok, err := in.GetMap(string(component)); ok {
			componentConfig := v2.ComponentRuntimeConfig{}
			if applied, err := runtimeValuesToComponentRuntimeConfig(v1.NewHelmValues(componentValues), &componentConfig); err != nil {
				return false, err
			} else if applied {
				runtime.Components[component] = componentConfig
				setRuntime = true
			}
		} else if err != nil {
			return false, err
		}
	}
	if len(runtime.Components) == 0 {
		runtime.Components = nil
	}

	if setDefaults {
		runtime.Defaults = defaults
		setRuntime = true
	}

	if setRuntime {
		out.Runtime = runtime
	}

	return setRuntime, nil
}

func populateContainerConfig(in *v1.HelmValues, out *v2.ContainerConfig) (bool, error) {
	setContainer := false
	if image, ok, err := in.GetString("image"); ok {
		out.Image = image
		setContainer = true
	} else if err != nil {
		return false, err
	}
	if applied, err := populateCommonContainerConfig(in, &out.CommonContainerConfig); err == nil {
		setContainer = setContainer || applied
	} else {
		return false, err
	}
	return setContainer, nil
}

func populateCommonContainerConfig(in *v1.HelmValues, out *v2.CommonContainerConfig) (bool, error) {
	setContainer := false
	if imagePullPolicy, ok, err := in.GetString("imagePullPolicy"); ok {
		out.ImagePullPolicy = corev1.PullPolicy(imagePullPolicy)
		setContainer = true
	} else if err != nil {
		return false, err
	}
	if imagePullSecrets, ok, err := in.GetStringSlice("imagePullSecrets"); ok && len(imagePullSecrets) > 0 {
		out.ImagePullSecrets = make([]corev1.LocalObjectReference, len(imagePullSecrets))
		setContainer = true
		for index, pullSecret := range imagePullSecrets {
			out.ImagePullSecrets[index].Name = pullSecret
		}
	} else if err != nil {
		return false, err
	}
	if hub, ok, err := in.GetString("hub"); ok {
		out.ImageRegistry = hub
		setContainer = true
	} else if err != nil {
		return false, err
	}
	if tag, ok, err := in.GetString("tag"); ok {
		out.ImageTag = tag
		setContainer = true
	} else if err != nil {
		return false, err
	}
	if resourcesValues, ok, err := in.GetMap("resources"); ok && len(resourcesValues) > 0 {
		out.Resources = &corev1.ResourceRequirements{}
		if err := fromValues(resourcesValues, out.Resources); err != nil {
			return false, err
		}
		setContainer = true
	} else if err != nil {
		return false, err
	}

	return setContainer, nil
}

func populateComponentServiceValues(serviceConfig *v2.ComponentServiceConfig, componentServiceValues map[string]interface{}) error {
	if len(serviceConfig.Metadata.Annotations) > 0 {
		if err := setHelmStringMapValue(componentServiceValues, "service.annotations", serviceConfig.Metadata.Annotations); err != nil {
			return err
		}
	}
	if len(serviceConfig.Metadata.Labels) > 0 {
		if err := setHelmStringMapValue(componentServiceValues, "service.labels", serviceConfig.Metadata.Labels); err != nil {
			return err
		}
	}
	if serviceConfig.NodePort != nil {
		if *serviceConfig.NodePort == 0 {
			if err := setHelmBoolValue(componentServiceValues, "service.nodePort.enabled", false); err != nil {
				return err
			}
		} else {
			if err := setHelmBoolValue(componentServiceValues, "service.nodePort.enabled", true); err != nil {
				return err
			}
			if err := setHelmIntValue(componentServiceValues, "service.nodePort.port", int64(*serviceConfig.NodePort)); err != nil {
				return err
			}
		}
	}
	ingressValues := make(map[string]interface{})
	if err := populateAddonIngressValues(serviceConfig.Ingress, ingressValues); err == nil {
		if len(ingressValues) > 0 {
			if err := setHelmValue(componentServiceValues, "ingress", ingressValues); err != nil {
				return err
			}
			// patch up contextPath, as some addons expect this at root, as
			// opposed to under ingress
			if contextPath, ok := ingressValues["contextPath"]; ok {
				if err := setHelmValue(componentServiceValues, "contextPath", contextPath); err != nil {
					return err
				}
			}
		}
	} else {
		return err
	}
	return nil
}

func populateComponentServiceConfig(in *v1.HelmValues, out *v2.ComponentServiceConfig) (bool, error) {
	setValues := false
	if rawAnnotations, ok, err := in.GetMap("service.annotations"); ok && len(rawAnnotations) > 0 {
		if err := setMetadataAnnotations(rawAnnotations, &out.Metadata); err != nil {
			return false, err
		}
		setValues = true
	} else if err != nil {
		return false, err
	}
	if rawLabels, ok, err := in.GetMap("service.labels"); ok && len(rawLabels) > 0 {
		if err := setMetadataLabels(rawLabels, &out.Metadata); err != nil {
			return false, err
		}
		setValues = true
	} else if err != nil {
		return false, err
	}
	if enabled, ok, err := in.GetBool("service.nodePort.enabled"); ok {
		if !enabled {
			nodePort := int32(0)
			out.NodePort = &nodePort
			setValues = true
		} else if rawNodePort, ok, err := in.GetInt64("service.nodePort.port"); ok {
			nodePort := int32(rawNodePort)
			out.NodePort = &nodePort
			setValues = true
		} else if err != nil {
			return false, err
		}
	} else if err != nil {
		return false, err
	}
	if rawIngressValues, ok, err := in.GetMap("ingress"); ok && len(rawIngressValues) > 0 {
		ingressValues := v1.NewHelmValues(rawIngressValues)
		ingress := &v2.ComponentIngressConfig{}
		// patch up contextPath, as some addons have this at root, as opposed to
		// under ingress
		if _, hasContextPath := rawIngressValues["contextPath"]; !hasContextPath {
			if contextPath, ok, err := in.GetString("contextPath"); ok && contextPath != "" {
				rawIngressValues["contextPath"] = contextPath
			} else if err != nil {
				return false, err
			}
		}
		if applied, err := populateAddonIngressConfig(ingressValues, ingress); err != nil {
			return false, err
		} else if applied {
			out.Ingress = ingress
			setValues = true
		}
	} else if err != nil {
		return false, err
	}

	return setValues, nil
}
