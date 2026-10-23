package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeValueExpr(t *testing.T) {
	testCases := []struct {
		in   string
		want string
	}{
		{in: "x", want: "x.value_id"},
		{in: "a.b", want: "a.b.value_id"},
		{in: "a.b.c", want: "a.b.c"},
		{in: "a.b.c.d", want: "a.b.c.d"},
	}

	for _, tc := range testCases {
		t.Run(tc.in, func(t *testing.T) {
			assert.Equal(t, tc.want, normalizeValueExpr(tc.in))
		})
	}
}

func TestExplicitValues(t *testing.T) {
	got := explicitValues([]string{"a.b", "x.y.z"})
	want := []refEntry{
		{Key: "a.b.value_id", Expr: "a.b.value_id"},
		{Key: "x.y.z", Expr: "x.y.z"},
	}

	assert.Equal(t, want, got)
	assert.Empty(t, explicitValues(nil))
}

func TestMergeRefEntries(t *testing.T) {
	got := mergeRefEntries(
		[]refEntry{{Key: "a", Expr: "a"}},
		[]refEntry{{Key: "a", Expr: "DIFFERENT"}, {Key: "b", Expr: "b"}},
	)
	want := []refEntry{
		{Key: "a", Expr: "a"},
		{Key: "b", Expr: "b"},
	}

	assert.Equal(t, want, got)
}

func TestAlignedMapEntries(t *testing.T) {
	got := alignedMapEntries([]refEntry{
		{Key: "ab", Expr: "X"},
		{Key: "a", Expr: "Y"},
	})
	want := "    \"ab\" = X\n    \"a\"  = Y\n"

	assert.Equal(t, want, got)
}

func TestBuildRefs(t *testing.T) {
	warn := func(base string) string {
		return "# NOTE: skipped " + base + ": count/for_each size is not a literal; enumerate instances manually"
	}

	testCases := []struct {
		name         string
		in           []resourceRef
		wantValues   []refEntry
		wantTargets  []refEntry
		wantWarnings []string
	}{
		{
			name: "value single",
			in:   []resourceRef{{Type: "scaleway_annotations_value", Name: "v"}},
			wantValues: []refEntry{
				{Key: "scaleway_annotations_value.v.value_id", Expr: "scaleway_annotations_value.v.value_id"},
			},
		},
		{
			name: "target single",
			in:   []resourceRef{{Type: "scaleway_instance_server", Name: "s"}},
			wantTargets: []refEntry{
				{Key: "scaleway_instance_server.s", Expr: "scaleway_instance_server.s.srn"},
			},
		},
		{
			name: "annotations resource ignored",
			in:   []resourceRef{{Type: "scaleway_annotations_binding", Name: "b"}},
		},
		{
			name: "target count",
			in:   []resourceRef{{Type: "scaleway_instance_server", Name: "s", IsSet: true, SetCount: 2, SetCountOK: true}},
			wantTargets: []refEntry{
				{Key: "scaleway_instance_server.s.0", Expr: "scaleway_instance_server.s[0].srn"},
				{Key: "scaleway_instance_server.s.1", Expr: "scaleway_instance_server.s[1].srn"},
			},
		},
		{
			name: "value count",
			in:   []resourceRef{{Type: "scaleway_annotations_value", Name: "v", IsSet: true, SetCount: 2, SetCountOK: true}},
			wantValues: []refEntry{
				{Key: "scaleway_annotations_value.v.0.value_id", Expr: "scaleway_annotations_value.v[0].value_id"},
				{Key: "scaleway_annotations_value.v.1.value_id", Expr: "scaleway_annotations_value.v[1].value_id"},
			},
		},
		{
			name:         "dynamic count warns",
			in:           []resourceRef{{Type: "scaleway_instance_server", Name: "s", IsSet: true, SetCount: 3, SetCountOK: false}},
			wantWarnings: []string{warn("scaleway_instance_server.s")},
		},
		{
			name:         "zero count warns",
			in:           []resourceRef{{Type: "scaleway_instance_server", Name: "s", IsSet: true, SetCount: 0, SetCountOK: true}},
			wantWarnings: []string{warn("scaleway_instance_server.s")},
		},
		{
			name:         "for_each warns",
			in:           []resourceRef{{Type: "scaleway_instance_server", Name: "s", IsSet: true}},
			wantWarnings: []string{warn("scaleway_instance_server.s")},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			values, targets, warnings := buildRefs(tc.in)
			assert.Equal(t, tc.wantValues, values)
			assert.Equal(t, tc.wantTargets, targets)
			assert.Equal(t, tc.wantWarnings, warnings)
		})
	}
}

func TestRenderModule(t *testing.T) {
	values := []refEntry{{Key: "scaleway_annotations_value.v.value_id", Expr: "scaleway_annotations_value.v.value_id"}}
	targets := []refEntry{{Key: "scaleway_instance_server.s", Expr: "scaleway_instance_server.s.srn"}}

	t.Run("literal values", func(t *testing.T) {
		assert.Equal(t, goldenLiteral, renderModule("", values, targets, nil))
	})

	t.Run("inherited values", func(t *testing.T) {
		assert.Equal(t, goldenInherited, renderModule("var.default_annotations_values", nil, targets, nil))
	})

	t.Run("with warnings", func(t *testing.T) {
		warn := []string{"# NOTE: skipped scaleway_k8s_cluster.k: count/for_each size is not a literal; enumerate instances manually"}
		assert.Equal(t, goldenWarnings, renderModule("", values, targets, warn))
	})
}

func TestCheckModule(t *testing.T) {
	t.Run("no file no content", func(t *testing.T) {
		ok, err := checkModule(t.TempDir(), "")
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("no file has content", func(t *testing.T) {
		ok, err := checkModule(t.TempDir(), "X")
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("file matches", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, outputFileName), []byte("X"), 0o600))
		ok, err := checkModule(dir, "X")
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("file differs", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, outputFileName), []byte("X"), 0o600))
		ok, err := checkModule(dir, "Y")
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("orphan file", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, outputFileName), []byte("X"), 0o600))
		ok, err := checkModule(dir, "")
		require.NoError(t, err)
		assert.False(t, ok)
	})
}

func TestGenerateModule(t *testing.T) {
	t.Run("value and target", func(t *testing.T) {
		dir := t.TempDir()
		cfg := "resource \"scaleway_annotations_value\" \"v\" { key = \"k\" }\n" +
			"resource \"scaleway_instance_server\" \"s\" { name = \"n\" }\n"
		require.NoError(t, os.WriteFile(filepath.Join(dir, "main.tf"), []byte(cfg), 0o600))

		got, err := generateModule(dir, nil, false, "")
		require.NoError(t, err)
		assert.Contains(t, got, "\"scaleway_annotations_value.v.value_id\" = scaleway_annotations_value.v.value_id")
		assert.Contains(t, got, "\"scaleway_instance_server.s\" = scaleway_instance_server.s.srn")
		assert.Contains(t, got, "resource \"scaleway_annotations_binding\" \"default_annotations\"")
	})

	t.Run("no resources", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "main.tf"), []byte("provider \"scaleway\" {}\n"), 0o600))

		got, err := generateModule(dir, nil, false, "")
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("inherit", func(t *testing.T) {
		dir := t.TempDir()
		cfg := "variable \"default_annotations_values\" { type = map(string) }\n" +
			"resource \"scaleway_instance_server\" \"s\" { name = \"n\" }\n"
		require.NoError(t, os.WriteFile(filepath.Join(dir, "main.tf"), []byte(cfg), 0o600))

		got, err := generateModule(dir, nil, true, "default_annotations_values")
		require.NoError(t, err)
		assert.Contains(t, got, "annotations_values = var.default_annotations_values")
	})
}

const goldenLiteral = `# Code generated by cmd/default-annotations-gen; DO NOT EDIT.
locals {
  annotations_values = {
    "scaleway_annotations_value.v.value_id" = scaleway_annotations_value.v.value_id
  }

  annotations_targets = {
    "scaleway_instance_server.s" = scaleway_instance_server.s.srn
  }

  annotations_bindings = {
    for entry in flatten([
      for v_key, v_id in local.annotations_values :
      [
        for t_key, t_srn in local.annotations_targets :
        {
          key        = "${v_key}|${t_key}"
          value_id   = v_id
          target_srn = t_srn
        }
      ]
    ])
    : entry.key => {
      value_id   = entry.value_id
      target_srn = entry.target_srn
    }
  }
}

resource "scaleway_annotations_binding" "default_annotations" {
  for_each = local.annotations_bindings

  value_id   = each.value.value_id
  target_srn = each.value.target_srn
}
`

const goldenInherited = `# Code generated by cmd/default-annotations-gen; DO NOT EDIT.
locals {
  annotations_values = var.default_annotations_values

  annotations_targets = {
    "scaleway_instance_server.s" = scaleway_instance_server.s.srn
  }

  annotations_bindings = {
    for entry in flatten([
      for v_key, v_id in local.annotations_values :
      [
        for t_key, t_srn in local.annotations_targets :
        {
          key        = "${v_key}|${t_key}"
          value_id   = v_id
          target_srn = t_srn
        }
      ]
    ])
    : entry.key => {
      value_id   = entry.value_id
      target_srn = entry.target_srn
    }
  }
}

resource "scaleway_annotations_binding" "default_annotations" {
  for_each = local.annotations_bindings

  value_id   = each.value.value_id
  target_srn = each.value.target_srn
}
`

const goldenWarnings = `# Code generated by cmd/default-annotations-gen; DO NOT EDIT.
# NOTE: skipped scaleway_k8s_cluster.k: count/for_each size is not a literal; enumerate instances manually

locals {
  annotations_values = {
    "scaleway_annotations_value.v.value_id" = scaleway_annotations_value.v.value_id
  }

  annotations_targets = {
    "scaleway_instance_server.s" = scaleway_instance_server.s.srn
  }

  annotations_bindings = {
    for entry in flatten([
      for v_key, v_id in local.annotations_values :
      [
        for t_key, t_srn in local.annotations_targets :
        {
          key        = "${v_key}|${t_key}"
          value_id   = v_id
          target_srn = t_srn
        }
      ]
    ])
    : entry.key => {
      value_id   = entry.value_id
      target_srn = entry.target_srn
    }
  }
}

resource "scaleway_annotations_binding" "default_annotations" {
  for_each = local.annotations_bindings

  value_id   = each.value.value_id
  target_srn = each.value.target_srn
}
`
