package otelgorm

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

func TestChunkBy(t *testing.T) {
	tests := []struct {
		input  string
		max    int
		output []string
	}{
		{
			"1234",
			1,
			[]string{"1", "2", "3", "4"},
		},
		{
			"1234",
			2,
			[]string{"12", "34"},
		},
		{
			"1234",
			4,
			[]string{"1234"},
		},
		{
			"1234",
			5,
			[]string{"1234"},
		},
	}

	for idx, tc := range tests {
		t.Run(fmt.Sprintf("test_%d", idx), func(tt *testing.T) {
			result := []string{}
			chunkBy(tc.input, tc.max, func(s string, option ...trace.EventOption) {
				result = append(result, s)
			})

			require.Equal(tt, tc.output, result)
		})
	}
}

func TestWithOmitVariablesFromTrace(t *testing.T) {
	ctx := context.Background()
	shouldOmit, _ := ctx.Value(omitVarsKey).(bool)
	require.NotEqual(t, shouldOmit, true)
	ctx = WithOmitVariablesFromTrace(context.Background())
	shouldOmit, _ = ctx.Value(omitVarsKey).(bool)
	require.Equal(t, shouldOmit, true)
}
