// +build !ignore_autogenerated

// Code generated by conversion-gen. DO NOT EDIT.

package conversion

import (
	unsafe "unsafe"

	v1 "github.com/maistra/istio-operator/pkg/apis/maistra/v1"
	v2 "github.com/maistra/istio-operator/pkg/apis/maistra/v2"
	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

func init() {
	localSchemeBuilder.Register(RegisterConversions)
}

// RegisterConversions adds conversion functions to the given scheme.
// Public to allow building arbitrary schemes.
func RegisterConversions(s *runtime.Scheme) error {
	if err := s.AddGeneratedConversionFunc((*v1.ServiceMeshControlPlane)(nil), (*v2.ServiceMeshControlPlane)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_ServiceMeshControlPlane_To_v2_ServiceMeshControlPlane(a.(*v1.ServiceMeshControlPlane), b.(*v2.ServiceMeshControlPlane), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*v2.ServiceMeshControlPlane)(nil), (*v1.ServiceMeshControlPlane)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v2_ServiceMeshControlPlane_To_v1_ServiceMeshControlPlane(a.(*v2.ServiceMeshControlPlane), b.(*v1.ServiceMeshControlPlane), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*v1.ServiceMeshControlPlaneList)(nil), (*v2.ServiceMeshControlPlaneList)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_ServiceMeshControlPlaneList_To_v2_ServiceMeshControlPlaneList(a.(*v1.ServiceMeshControlPlaneList), b.(*v2.ServiceMeshControlPlaneList), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*v2.ServiceMeshControlPlaneList)(nil), (*v1.ServiceMeshControlPlaneList)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v2_ServiceMeshControlPlaneList_To_v1_ServiceMeshControlPlaneList(a.(*v2.ServiceMeshControlPlaneList), b.(*v1.ServiceMeshControlPlaneList), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*v1.ControlPlaneSpec)(nil), (*v2.ControlPlaneSpec)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_ControlPlaneSpec_To_v2_ControlPlaneSpec(a.(*v1.ControlPlaneSpec), b.(*v2.ControlPlaneSpec), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*v1.ControlPlaneStatus)(nil), (*v2.ControlPlaneStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_ControlPlaneStatus_To_v2_ControlPlaneStatus(a.(*v1.ControlPlaneStatus), b.(*v2.ControlPlaneStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*v2.ControlPlaneSpec)(nil), (*v1.ControlPlaneSpec)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v2_ControlPlaneSpec_To_v1_ControlPlaneSpec(a.(*v2.ControlPlaneSpec), b.(*v1.ControlPlaneSpec), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*v2.ControlPlaneStatus)(nil), (*v1.ControlPlaneStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v2_ControlPlaneStatus_To_v1_ControlPlaneStatus(a.(*v2.ControlPlaneStatus), b.(*v1.ControlPlaneStatus), scope)
	}); err != nil {
		return err
	}
	return nil
}

func autoConvert_v1_ControlPlaneSpec_To_v2_ControlPlaneSpec(in *v1.ControlPlaneSpec, out *v2.ControlPlaneSpec, s conversion.Scope) error {
	// WARNING: in.Template requires manual conversion: does not exist in peer-type
	out.Profiles = *(*[]string)(unsafe.Pointer(&in.Profiles))
	out.Version = in.Version
	// WARNING: in.NetworkType requires manual conversion: does not exist in peer-type
	// WARNING: in.Istio requires manual conversion: does not exist in peer-type
	// WARNING: in.ThreeScale requires manual conversion: does not exist in peer-type
	return nil
}

func autoConvert_v2_ControlPlaneSpec_To_v1_ControlPlaneSpec(in *v2.ControlPlaneSpec, out *v1.ControlPlaneSpec, s conversion.Scope) error {
	out.Profiles = *(*[]string)(unsafe.Pointer(&in.Profiles))
	out.Version = in.Version
	// WARNING: in.Cluster requires manual conversion: does not exist in peer-type
	// WARNING: in.General requires manual conversion: does not exist in peer-type
	// WARNING: in.Policy requires manual conversion: does not exist in peer-type
	// WARNING: in.Proxy requires manual conversion: does not exist in peer-type
	// WARNING: in.Security requires manual conversion: does not exist in peer-type
	// WARNING: in.Telemetry requires manual conversion: does not exist in peer-type
	// WARNING: in.Tracing requires manual conversion: does not exist in peer-type
	// WARNING: in.Gateways requires manual conversion: does not exist in peer-type
	// WARNING: in.Runtime requires manual conversion: does not exist in peer-type
	// WARNING: in.Addons requires manual conversion: does not exist in peer-type
	// WARNING: in.TechPreviews requires manual conversion: does not exist in peer-type
	return nil
}

func autoConvert_v1_ControlPlaneStatus_To_v2_ControlPlaneStatus(in *v1.ControlPlaneStatus, out *v2.ControlPlaneStatus, s conversion.Scope) error {
	out.StatusBase = in.StatusBase
	out.StatusType = in.StatusType
	out.ObservedGeneration = in.ObservedGeneration
	// WARNING: in.ReconciledVersion requires manual conversion: does not exist in peer-type
	out.ComponentStatusList = in.ComponentStatusList
	// WARNING: in.LastAppliedConfiguration requires manual conversion: does not exist in peer-type
	return nil
}

func autoConvert_v2_ControlPlaneStatus_To_v1_ControlPlaneStatus(in *v2.ControlPlaneStatus, out *v1.ControlPlaneStatus, s conversion.Scope) error {
	out.StatusBase = in.StatusBase
	out.StatusType = in.StatusType
	out.ObservedGeneration = in.ObservedGeneration
	// WARNING: in.OperatorVersion requires manual conversion: does not exist in peer-type
	// WARNING: in.ChartVersion requires manual conversion: does not exist in peer-type
	out.ComponentStatusList = in.ComponentStatusList
	// WARNING: in.AppliedSpec requires manual conversion: does not exist in peer-type
	// WARNING: in.AppliedValues requires manual conversion: does not exist in peer-type
	return nil
}

func autoConvert_v1_ServiceMeshControlPlane_To_v2_ServiceMeshControlPlane(in *v1.ServiceMeshControlPlane, out *v2.ServiceMeshControlPlane, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	if err := Convert_v1_ControlPlaneSpec_To_v2_ControlPlaneSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	if err := Convert_v1_ControlPlaneStatus_To_v2_ControlPlaneStatus(&in.Status, &out.Status, s); err != nil {
		return err
	}
	return nil
}

// Convert_v1_ServiceMeshControlPlane_To_v2_ServiceMeshControlPlane is an autogenerated conversion function.
func Convert_v1_ServiceMeshControlPlane_To_v2_ServiceMeshControlPlane(in *v1.ServiceMeshControlPlane, out *v2.ServiceMeshControlPlane, s conversion.Scope) error {
	return autoConvert_v1_ServiceMeshControlPlane_To_v2_ServiceMeshControlPlane(in, out, s)
}

func autoConvert_v2_ServiceMeshControlPlane_To_v1_ServiceMeshControlPlane(in *v2.ServiceMeshControlPlane, out *v1.ServiceMeshControlPlane, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	if err := Convert_v2_ControlPlaneSpec_To_v1_ControlPlaneSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	if err := Convert_v2_ControlPlaneStatus_To_v1_ControlPlaneStatus(&in.Status, &out.Status, s); err != nil {
		return err
	}
	return nil
}

// Convert_v2_ServiceMeshControlPlane_To_v1_ServiceMeshControlPlane is an autogenerated conversion function.
func Convert_v2_ServiceMeshControlPlane_To_v1_ServiceMeshControlPlane(in *v2.ServiceMeshControlPlane, out *v1.ServiceMeshControlPlane, s conversion.Scope) error {
	return autoConvert_v2_ServiceMeshControlPlane_To_v1_ServiceMeshControlPlane(in, out, s)
}

func autoConvert_v1_ServiceMeshControlPlaneList_To_v2_ServiceMeshControlPlaneList(in *v1.ServiceMeshControlPlaneList, out *v2.ServiceMeshControlPlaneList, s conversion.Scope) error {
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]v2.ServiceMeshControlPlane, len(*in))
		for i := range *in {
			if err := Convert_v1_ServiceMeshControlPlane_To_v2_ServiceMeshControlPlane(&(*in)[i], &(*out)[i], s); err != nil {
				return err
			}
		}
	} else {
		out.Items = nil
	}
	return nil
}

// Convert_v1_ServiceMeshControlPlaneList_To_v2_ServiceMeshControlPlaneList is an autogenerated conversion function.
func Convert_v1_ServiceMeshControlPlaneList_To_v2_ServiceMeshControlPlaneList(in *v1.ServiceMeshControlPlaneList, out *v2.ServiceMeshControlPlaneList, s conversion.Scope) error {
	return autoConvert_v1_ServiceMeshControlPlaneList_To_v2_ServiceMeshControlPlaneList(in, out, s)
}

func autoConvert_v2_ServiceMeshControlPlaneList_To_v1_ServiceMeshControlPlaneList(in *v2.ServiceMeshControlPlaneList, out *v1.ServiceMeshControlPlaneList, s conversion.Scope) error {
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]v1.ServiceMeshControlPlane, len(*in))
		for i := range *in {
			if err := Convert_v2_ServiceMeshControlPlane_To_v1_ServiceMeshControlPlane(&(*in)[i], &(*out)[i], s); err != nil {
				return err
			}
		}
	} else {
		out.Items = nil
	}
	return nil
}

// Convert_v2_ServiceMeshControlPlaneList_To_v1_ServiceMeshControlPlaneList is an autogenerated conversion function.
func Convert_v2_ServiceMeshControlPlaneList_To_v1_ServiceMeshControlPlaneList(in *v2.ServiceMeshControlPlaneList, out *v1.ServiceMeshControlPlaneList, s conversion.Scope) error {
	return autoConvert_v2_ServiceMeshControlPlaneList_To_v1_ServiceMeshControlPlaneList(in, out, s)
}
