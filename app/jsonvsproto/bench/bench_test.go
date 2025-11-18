package bench

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkProtoMarshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ProtoMarshal(Sample)
	}
}

func BenchmarkJSONMarshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = JSONMarshal(Sample)
	}
}

func BenchmarkProtoUnmarshal(b *testing.B) {
	data, err := ProtoMarshal(Sample)
	require.NoError(b, err)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ProtoUnmarshal(data)
	}
}

func BenchmarkJSONUnmarshal(b *testing.B) {
	data, err := JSONMarshal(Sample)
	require.NoError(b, err)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = JSONUnmarshal(data)
	}
}

func TestSizeComparison(t *testing.T) {
	p, err := ProtoMarshal(Sample)
	require.NoError(t, err)
	j, err := JSONMarshal(Sample)
	require.NoError(t, err)

	t.Logf("protobuf size: %d bytes", len(p))
	t.Logf("json size:     %d bytes", len(j))

	// expected protobuf to be smaller or equal in size for this struct
	if len(p) >= len(j) {
		t.Log("note: protobuf is not smaller for this sample; results may vary by content")
	}
}
