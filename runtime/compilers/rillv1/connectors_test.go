package rillv1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rilldata/rill/runtime/drivers"
	_ "github.com/rilldata/rill/runtime/drivers/gcs"
	_ "github.com/rilldata/rill/runtime/drivers/s3"
)

func TestAnalyzeConnectors(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: `
connectors:
- name: my-s3
  type: s3
`,
		// S3 source, with a custom name in rill.yaml
		`sources/foo.yaml`: `
connector: my-s3
uri: s3://path/to/foo
`,
		// GCS source, not configured with a custom name in rill.yaml
		`sources/bar.yaml`: `
connector: gcs
uri: gs://path/to/bar
`,
	})

	p, err := Parse(ctx, repo, "", nil)
	require.NoError(t, err)

	cs, err := p.AnalyzeConnectors(ctx)
	require.NoError(t, err)

	require.Len(t, cs, 2)

	require.Len(t, cs[0].Resources, 1)
	require.Equal(t, "gcs", cs[0].Name)
	require.Equal(t, "gcs", cs[0].Driver)
	require.Equal(t, false, cs[0].AnonymousAccess)
	require.Equal(t, drivers.Connectors["gcs"].Spec(), cs[0].Spec)

	require.Len(t, cs[1].Resources, 1)
	require.Equal(t, "my-s3", cs[1].Name)
	require.Equal(t, "s3", cs[1].Driver)
	require.Equal(t, false, cs[1].AnonymousAccess)
	require.Equal(t, drivers.Connectors["s3"].Spec(), cs[1].Spec)
}
