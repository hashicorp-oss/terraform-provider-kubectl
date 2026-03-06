package util

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GVRFromUnstructured extracts a canonical schema.GroupVersionResource out of the resource's
// metadata by checking it against the discovery API via a RESTMapper.
func GVRFromUnstructured(
	o *unstructured.Unstructured,
	m meta.RESTMapper,
) (schema.GroupVersionResource, error) {
	apv := o.GetAPIVersion()
	kind := o.GetKind()
	gv, err := schema.ParseGroupVersion(apv)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}
	mapping, err := m.RESTMapping(gv.WithKind(kind).GroupKind(), gv.Version)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}
	return mapping.Resource, err
}

// GVKFromTftypesObject extracts a canonical schema.GroupVersionKind out of the resource's
// metadata by checking it against the discovery API via a RESTMapper.
func GVKFromTftypesObject(in *tftypes.Value, m meta.RESTMapper) (schema.GroupVersionKind, error) {
	var obj map[string]tftypes.Value
	err := in.As(&obj)
	if err != nil {
		return schema.GroupVersionKind{}, err
	}
	var apv string
	var kind string
	err = obj["apiVersion"].As(&apv)
	if err != nil {
		return schema.GroupVersionKind{}, err
	}
	err = obj["kind"].As(&kind)
	if err != nil {
		return schema.GroupVersionKind{}, err
	}
	gv, err := schema.ParseGroupVersion(apv)
	if err != nil {
		return schema.GroupVersionKind{}, err
	}
	mappings, err := m.RESTMappings(gv.WithKind(kind).GroupKind())
	if err != nil {
		return schema.GroupVersionKind{}, err
	}
	for _, m := range mappings {
		if m.GroupVersionKind.GroupVersion().String() == apv {
			return m.GroupVersionKind, nil
		}
	}
	return schema.GroupVersionKind{}, errors.New("cannot select exact GV from REST mapper")
}

// IsResourceNamespaced determines if a resource is namespaced or cluster-level
// by querying the Kubernetes discovery API.
func IsResourceNamespaced(gvk schema.GroupVersionKind, m meta.RESTMapper) (bool, error) {
	rm, err := m.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return false, err
	}
	if rm.Scope.Name() == meta.RESTScopeNameNamespace {
		return true, nil
	}
	return false, nil
}
