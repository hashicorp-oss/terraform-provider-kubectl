// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kubectl

import (
	"testing"
)

func TestDeepMergeMaps_basic(t *testing.T) {
	base := map[string]any{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]any{
			"name":      "test",
			"namespace": "default",
		},
		"data": map[string]any{
			"key1": "value1",
		},
	}

	overlay := map[string]any{
		"data": map[string]any{
			"key2": "value2",
		},
	}

	deepMergeMaps(base, overlay)

	data := base["data"].(map[string]any) //nolint:forcetypeassert
	if data["key1"] != "value1" {
		t.Errorf("expected key1=value1, got %v", data["key1"])
	}
	if data["key2"] != "value2" {
		t.Errorf("expected key2=value2, got %v", data["key2"])
	}
}

func TestDeepMergeMaps_overlayOverridesBase(t *testing.T) {
	base := map[string]any{
		"data": map[string]any{
			"key1": "original",
		},
	}

	overlay := map[string]any{
		"data": map[string]any{
			"key1": "overridden",
		},
	}

	deepMergeMaps(base, overlay)

	data := base["data"].(map[string]any) //nolint:forcetypeassert
	if data["key1"] != "overridden" {
		t.Errorf("expected key1=overridden, got %v", data["key1"])
	}
}

func TestDeepMergeMaps_nestedMerge(t *testing.T) {
	base := map[string]any{
		"metadata": map[string]any{
			"name": "test",
			"labels": map[string]any{
				"app": "myapp",
			},
		},
	}

	overlay := map[string]any{
		"metadata": map[string]any{
			"labels": map[string]any{
				"env": "prod",
			},
		},
	}

	deepMergeMaps(base, overlay)

	labels := base["metadata"].(map[string]any)["labels"].(map[string]any) //nolint:forcetypeassert
	if labels["app"] != "myapp" {
		t.Errorf("expected app=myapp, got %v", labels["app"])
	}
	if labels["env"] != "prod" {
		t.Errorf("expected env=prod, got %v", labels["env"])
	}
}

func TestDeepMergeMaps_nonMapOverwritesMap(t *testing.T) {
	base := map[string]any{
		"data": map[string]any{
			"key1": "value1",
		},
	}

	overlay := map[string]any{
		"data": "scalar-value",
	}

	deepMergeMaps(base, overlay)

	if base["data"] != "scalar-value" {
		t.Errorf("expected data=scalar-value, got %v", base["data"])
	}
}

func TestDeepMergeMaps_addsNewTopLevelKeys(t *testing.T) {
	base := map[string]any{
		"existing": "value",
	}

	overlay := map[string]any{
		"newkey": "newvalue",
	}

	deepMergeMaps(base, overlay)

	if base["existing"] != "value" {
		t.Errorf("expected existing=value, got %v", base["existing"])
	}
	if base["newkey"] != "newvalue" {
		t.Errorf("expected newkey=newvalue, got %v", base["newkey"])
	}
}

func TestDeepMergeMaps_emptyOverlay(t *testing.T) {
	base := map[string]any{
		"key": "value",
	}

	deepMergeMaps(base, map[string]any{})

	if base["key"] != "value" {
		t.Errorf("expected key=value, got %v", base["key"])
	}
}

func TestExtractLeafPaths_flat(t *testing.T) {
	m := map[string]any{
		"key1": "value1",
		"key2": "value2",
	}

	paths := extractLeafPaths(m, "")

	expected := []string{"key1", "key2"}
	if len(paths) != len(expected) {
		t.Fatalf("expected %d paths, got %d: %v", len(expected), len(paths), paths)
	}
	for i, p := range paths {
		if p != expected[i] {
			t.Errorf("expected path[%d]=%s, got %s", i, expected[i], p)
		}
	}
}

func TestExtractLeafPaths_nested(t *testing.T) {
	m := map[string]any{
		"data": map[string]any{
			"password": "secret",
		},
		"metadata": map[string]any{
			"labels": map[string]any{
				"env": "prod",
			},
		},
	}

	paths := extractLeafPaths(m, "")

	expected := []string{"data.password", "metadata.labels.env"}
	if len(paths) != len(expected) {
		t.Fatalf("expected %d paths, got %d: %v", len(expected), len(paths), paths)
	}
	for i, p := range paths {
		if p != expected[i] {
			t.Errorf("expected path[%d]=%s, got %s", i, expected[i], p)
		}
	}
}

func TestExtractLeafPaths_withPrefix(t *testing.T) {
	m := map[string]any{
		"password": "secret",
	}

	paths := extractLeafPaths(m, "data")

	if len(paths) != 1 || paths[0] != "data.password" {
		t.Errorf("expected [data.password], got %v", paths)
	}
}

func TestExtractLeafPaths_empty(t *testing.T) {
	paths := extractLeafPaths(map[string]any{}, "")

	if len(paths) != 0 {
		t.Errorf("expected empty paths, got %v", paths)
	}
}

func TestComputeManifestWoChecksum_deterministic(t *testing.T) {
	m := map[string]any{
		"data": map[string]any{
			"password": "secret123",
		},
	}

	c1 := computeManifestWoChecksum(m)
	c2 := computeManifestWoChecksum(m)

	if c1 == "" {
		t.Fatal("expected non-empty checksum")
	}
	if c1 != c2 {
		t.Errorf("expected deterministic checksum, got %s != %s", c1, c2)
	}
}

func TestComputeManifestWoChecksum_changesWithInput(t *testing.T) {
	m1 := map[string]any{"data": map[string]any{"password": "secret1"}}
	m2 := map[string]any{"data": map[string]any{"password": "secret2"}}

	c1 := computeManifestWoChecksum(m1)
	c2 := computeManifestWoChecksum(m2)

	if c1 == c2 {
		t.Errorf("expected different checksums for different inputs, got %s", c1)
	}
}
