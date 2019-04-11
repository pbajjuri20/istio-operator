package controlplane

import (
	"context"
	"reflect"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"

	"github.com/ghodss/yaml"

	istiov1alpha3 "github.com/maistra/istio-operator/pkg/apis/istio/v1alpha3"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	"k8s.io/helm/pkg/manifest"
	"k8s.io/helm/pkg/releaseutil"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	resourceLabel           = "istio.openshift.io/resource"
	resourceGenerationLabel = "istio.openshift.io/resource-generation"
)

func (r *controlPlaneReconciler) processComponentManifests(componentName string) error {
	var err error
	status := r.instance.Status.FindComponentByName(componentName)
	renderings, hasRenderings := r.renderings[componentName]
	origLogger := r.log
	r.log = r.log.WithValues("Component", componentName)
	defer func() { r.log = origLogger }()
	if hasRenderings {
		r.log.Info("reconciling component resources")
		if status == nil {
			status = istiov1alpha3.NewComponentStatus()
			status.Resource = componentName
		} else {
			status.RemoveCondition(istiov1alpha3.ConditionTypeReconciled)
		}
		status, err = r.processManifests(renderings, status)
		status.ObservedGeneration = r.instance.GetGeneration()
		if err := r.processNewComponent(componentName, status); err != nil {
			r.log.Error(err, "unexpected error occurred during postprocessing of new component")
		}
		r.status.ComponentStatus = append(r.status.ComponentStatus, status)
	} else {
		r.log.Info("no renderings for component")
	}
	r.log.Info("component reconciliation complete")
	return err
}

func (r *controlPlaneReconciler) processManifests(manifests []manifest.Manifest,
	oldStatus *istiov1alpha3.ComponentStatus) (*istiov1alpha3.ComponentStatus, error) {

	allErrors := []error{}
	resourcesProcessed := map[istiov1alpha3.ResourceKey]struct{}{}
	newStatus := istiov1alpha3.NewComponentStatus()
	newStatus.StatusType = oldStatus.StatusType
	newStatus.Resource = oldStatus.Resource

	origLogger := r.log
	defer func() { r.log = origLogger }()
	for _, manifest := range manifests {
		r.log = origLogger.WithValues("manifest", manifest.Name)
		if !strings.HasSuffix(manifest.Name, ".yaml") {
			r.log.V(2).Info("Skipping rendering of manifest")
			continue
		}
		r.log.V(2).Info("Processing resources from manifest")
		// split the manifest into individual objects
		objects := releaseutil.SplitManifests(manifest.Content)
		for _, raw := range objects {
			rawJSON, err := yaml.YAMLToJSON([]byte(raw))
			if err != nil {
				r.log.Error(err, "unable to convert raw data to JSON")
				allErrors = append(allErrors, err)
				continue
			}
			obj := &unstructured.Unstructured{}
			_, _, err = unstructured.UnstructuredJSONScheme.Decode(rawJSON, nil, obj)
			if err != nil {
				r.log.Error(err, "unable to decode object into Unstructured")
				allErrors = append(allErrors, err)
				continue
			}
			err = r.processObject(obj, resourcesProcessed, oldStatus, newStatus)
			if err != nil {
				allErrors = append(allErrors, err)
			}
		}
	}

	// handle deletions
	// XXX: should these be processed in reverse order of creation?
	for index := len(oldStatus.Resources) - 1; index >= 0; index-- {
		status := oldStatus.Resources[index]
		resourceKey := istiov1alpha3.ResourceKey(status.Resource)
		if _, ok := resourcesProcessed[resourceKey]; !ok {
			r.log = origLogger.WithValues("Resource", resourceKey)
			if condition := status.GetCondition(istiov1alpha3.ConditionTypeInstalled); condition.Status != istiov1alpha3.ConditionStatusFalse {
				r.log.Info("deleting resource")
				unstructured := resourceKey.ToUnstructured()
				err := r.client.Delete(context.TODO(), unstructured, client.PropagationPolicy(metav1.DeletePropagationForeground))
				updateDeleteStatus(status, err)
				newStatus.Resources = append(newStatus.Resources, status)
				if err == nil || errors.IsNotFound(err) || errors.IsGone(err) {
					status.ObservedGeneration = 0
					// special handling
					if err := r.processDeletedObject(unstructured); err != nil {
						r.log.Error(err, "unexpected error occurred during cleanup of deleted resource")
					}
				} else {
					r.log.Error(err, "error deleting resource")
					allErrors = append(allErrors, err)
				}
			}
		}
	}
	err := utilerrors.NewAggregate(allErrors)
	if len(manifests) > 0 {
		updateReconcileStatus(&newStatus.StatusType, err)
	} else {
		updateDeleteStatus(&newStatus.StatusType, err)
	}
	return newStatus, err
}

func (r *controlPlaneReconciler) processObject(obj *unstructured.Unstructured, resourcesProcessed map[istiov1alpha3.ResourceKey]struct{},
	oldStatus *istiov1alpha3.ComponentStatus, newStatus *istiov1alpha3.ComponentStatus) error {
	origLogger := r.log
	defer func() { r.log = origLogger }()

	key := istiov1alpha3.NewResourceKey(obj, obj)
	r.log = origLogger.WithValues("Resource", key)

	if obj.GetKind() == "List" {
		allErrors := []error{}
		list, err := obj.ToList()
		if err != nil {
			r.log.Error(err, "error converting List object")
			return err
		}
		for _, item := range list.Items {
			err = r.processObject(&item, resourcesProcessed, oldStatus, newStatus)
			if err != nil {
				allErrors = append(allErrors, err)
			}
		}
		return utilerrors.NewAggregate(allErrors)
	}

	// Add owner ref
	if obj.GetNamespace() == r.instance.GetNamespace() {
		obj.SetOwnerReferences(r.ownerRefs)
	} else {
		// XXX: can't set owner reference on cross-namespace or cluster resources
	}

	// add markers
	labels := obj.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	instanceType, _ := meta.TypeAccessor(r.instance)
	labels[resourceLabel] = string(istiov1alpha3.NewResourceKey(r.instance, instanceType))
	labels[resourceGenerationLabel] = strconv.FormatInt(r.instance.GetGeneration(), 10)
	obj.SetLabels(labels)

	r.log.V(2).Info("beginning reconciliation of resource", "ResourceKey", key)

	resourcesProcessed[key] = seen
	status := oldStatus.FindResourceByKey(key)
	if status == nil {
		newResourceStatus := istiov1alpha3.NewStatus()
		status = &newResourceStatus
		status.Resource = string(key)
	}
	newStatus.Resources = append(newStatus.Resources, status)

	err := r.patchObject(obj)
	if err != nil {
		r.log.Error(err, "error patching object")
		updateReconcileStatus(status, err)
		return err
	}

	receiver := key.ToUnstructured()
	objectKey, err := client.ObjectKeyFromObject(receiver)
	if err != nil {
		r.log.Error(err, "client.ObjectKeyFromObject() failed for resource")
		// This can only happen if reciever isn't an unstructured.Unstructured
		// i.e. this should never happen
		updateReconcileStatus(status, err)
		return err
	}
	err = r.client.Get(context.TODO(), objectKey, receiver)
	if err != nil {
		if errors.IsNotFound(err) {
			r.log.Info("creating resource")
			err = r.client.Create(context.TODO(), obj)
			if err == nil {
				status.ObservedGeneration = 1
				// special handling
				if err := r.processNewObject(obj); err != nil {
					// just log for now
					r.log.Error(err, "unexpected error occurred during postprocessing of new resource")
				}
			}
		}
	} else if shouldUpdate(obj.UnstructuredContent(), receiver.UnstructuredContent()) {
		// XXX: consider using patching mechanism
		r.log.Info("updating existing resource")
		status.RemoveCondition(istiov1alpha3.ConditionTypeReconciled)
		//r.log.Info("updates not supported at this time")
		// XXX: k8s barfs on some updates: metadata.resourceVersion: Invalid value: 0x0: must be specified for an update
		obj.SetResourceVersion(receiver.GetResourceVersion())
		err = r.client.Update(context.TODO(), obj)
		if err == nil {
			status.ObservedGeneration = obj.GetGeneration()
		}
	} else {
		// need to update generation label
		labels := receiver.GetLabels()
		labels[resourceGenerationLabel] = strconv.FormatInt(r.instance.GetGeneration(), 10)
		receiver.SetLabels(labels)
		err = r.client.Update(context.TODO(), receiver)
	}
	r.log.V(2).Info("resource reconciliation complete")
	updateReconcileStatus(status, err)
	if err != nil {
		r.log.Error(err, "error occurred reconciling resource")
	}
	return err
}

// shouldUpdate checks to see if the spec fields are the same for both objects.
// if the objects don't have a spec field, it checks all other fields, skipping
// known fields that shouldn't impact updates: kind, apiVersion, metadata, and status.
func shouldUpdate(o1, o2 map[string]interface{}) bool {
	if spec1, ok1 := o1["spec"]; ok1 {
		// we assume these are the same type of object
		return reflect.DeepEqual(spec1, o2["spec"])
	}
	for key, value := range o1 {
		if key == "status" || key == "kind" || key == "apiVersion" || key == "metadata" {
			continue
		}
		if !reflect.DeepEqual(value, o2[key]) {
			return true
		}
	}
	return false
}
