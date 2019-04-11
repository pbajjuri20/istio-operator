package controlplane

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	"context"
	"strconv"

    istiov1alpha3 "github.com/maistra/istio-operator/pkg/apis/istio/v1alpha3"

    "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
)

var (
    // XXX: move this into a ConfigMap so users can override things if they add new types in customized charts
    // ordered by which types should be deleted, first to last
	namespacedResources = []schema.GroupVersionKind{
		schema.GroupVersionKind{Group: "autoscaling", Version: "v2beta1", Kind: "HorizontalPodAutoscaler"},
		schema.GroupVersionKind{Group: "policy", Version: "v1beta1", Kind: "PodDisruptionBudget"},
		schema.GroupVersionKind{Group: "route.openshift.io", Version: "v1", Kind: "Route"},
		schema.GroupVersionKind{Group: "apps.openshift.io", Version: "v1", Kind: "DeploymentConfig"},
		schema.GroupVersionKind{Group: "apps", Version: "v1beta1", Kind: "Deployment"},
		schema.GroupVersionKind{Group: "apps", Version: "v1beta1", Kind: "StatefulSet"},
		schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
		schema.GroupVersionKind{Group: "batch", Version: "v1", Kind: "Job"},
		schema.GroupVersionKind{Group: "extensions", Version: "v1beta1", Kind: "DaemonSet"},
		schema.GroupVersionKind{Group: "extensions", Version: "v1beta1", Kind: "Deployment"},
		schema.GroupVersionKind{Group: "extensions", Version: "v1beta1", Kind: "Ingress"},
		schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"},
		schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Endpoints"},
		schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"},
		schema.GroupVersionKind{Group: "", Version: "v1", Kind: "PersistentVolumeClaim"},
		schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
		schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Secret"},
		schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ServiceAccount"},
		schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1beta1", Kind: "RoleBinding"},
		schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "RoleBinding"},
		schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1beta1", Kind: "Role"},
		schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "Role"},
		schema.GroupVersionKind{Group: "authentication.istio.io", Version: "v1alpha1", Kind: "Policy"},
		schema.GroupVersionKind{Group: "config.istio.io", Version: "v1alpha2", Kind: "adapter"},
		schema.GroupVersionKind{Group: "config.istio.io", Version: "v1alpha2", Kind: "attributemanifest"},
		schema.GroupVersionKind{Group: "config.istio.io", Version: "v1alpha2", Kind: "handler"},
		schema.GroupVersionKind{Group: "config.istio.io", Version: "v1alpha2", Kind: "kubernetes"},
		schema.GroupVersionKind{Group: "config.istio.io", Version: "v1alpha2", Kind: "logentry"},
		schema.GroupVersionKind{Group: "config.istio.io", Version: "v1alpha2", Kind: "metric"},
		schema.GroupVersionKind{Group: "config.istio.io", Version: "v1alpha2", Kind: "rule"},
		schema.GroupVersionKind{Group: "config.istio.io", Version: "v1alpha2", Kind: "template"},
		schema.GroupVersionKind{Group: "networking.istio.io", Version: "v1alpha3", Kind: "DestinationRule"},
		schema.GroupVersionKind{Group: "networking.istio.io", Version: "v1alpha3", Kind: "EnvoyFilter"},
		schema.GroupVersionKind{Group: "networking.istio.io", Version: "v1alpha3", Kind: "Gateway"},
		schema.GroupVersionKind{Group: "networking.istio.io", Version: "v1alpha3", Kind: "VirtualService"},
	}

    // ordered by which types should be deleted, first to last
	nonNamespacedResources = []schema.GroupVersionKind{
		schema.GroupVersionKind{Group: "admissionregistration.k8s.io", Version: "v1beta1", Kind: "MutatingWebhookConfiguration"},
		schema.GroupVersionKind{Group: "admissionregistration.k8s.io", Version: "v1beta1", Kind: "ValidatingWebhookConfiguration"},
		schema.GroupVersionKind{Group: "certmanager.k8s.io", Version: "v1alpha1", Kind: "ClusterIssuer"},
		schema.GroupVersionKind{Group: "oauth.openshift.io", Version: "v1", Kind: "OAuthClient"},
		schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRole"},
		schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRoleBinding"},
		schema.GroupVersionKind{Group: "authentication.istio.io", Version: "v1alpha1", Kind: "MeshPolicy"},
    }
)

func (r *controlPlaneReconciler) prune(generation int64) error {
    allErrors := []error{}
    // special handling for launcher
    err := r.pruneResources(namespacedResources, generation, launcherProjectName)
    if err != nil {
        allErrors = append(allErrors, err)
    }
    err = r.pruneResources(namespacedResources, generation, r.instance.Namespace)
    if err != nil {
        allErrors = append(allErrors, err)
    }
    err = r.pruneResources(nonNamespacedResources, generation, "")
    if err != nil {
        allErrors = append(allErrors, err)
    }
    return utilerrors.NewAggregate(allErrors)
}

func (r *controlPlaneReconciler) pruneResources(gvks []schema.GroupVersionKind, generation int64, namespace string) error {
	allErrors := []error{}
	instanceType, _ := meta.TypeAccessor(r.instance)
    instanceGeneration := strconv.FormatInt(generation, 10)
    labelSelector := map[string]string{resourceLabel: string(istiov1alpha3.NewResourceKey(r.instance, instanceType))}
    for _, gvk := range gvks {
        objects := &unstructured.UnstructuredList{}
        objects.SetGroupVersionKind(gvk)
        err := r.client.List(context.TODO(), client.MatchingLabels(labelSelector).InNamespace(namespace), objects)
        if err != nil {
            r.log.Error(err, "Error retrieving resources to prune", "type", gvk.String())
            allErrors = append(allErrors, err)
            continue
        }
        for _, object := range objects.Items {
            if generation, ok := object.GetLabels()[resourceGenerationLabel]; !ok || generation != instanceGeneration {
                err = r.client.Delete(context.TODO(), &object, client.PropagationPolicy(metav1.DeletePropagationBackground))
                if err != nil {
                    r.log.Error(err, "Error pruning resource", "resource", istiov1alpha3.NewResourceKey(&object, &object))
                    allErrors = append(allErrors, err)
                }
            }
        }
    }
    return utilerrors.NewAggregate(allErrors)
}
