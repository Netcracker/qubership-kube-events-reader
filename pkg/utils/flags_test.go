package utils

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestNamespaceFlagsType_Set(t *testing.T) {
	namespaceFlags := NamespaceFlagsType{}
	assert.NoError(t, namespaceFlags.Set("namespace-valid"))
	assert.Equal(t, 1, len(namespaceFlags))
	err := namespaceFlags.Set("non valid-namespace")
	assert.NotNil(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), "namespace is not valid"))
	assert.Equal(t, 1, len(namespaceFlags))
}

func TestNamespaceFlagsType_String(t *testing.T) {
	namespaceFlags := NamespaceFlagsType{}
	assert.NoError(t, namespaceFlags.Set("logging"))
	assert.NoError(t, namespaceFlags.Set("monitoring"))
	assert.NoError(t, namespaceFlags.Set("logging"))
	assert.Equal(t, "logging,monitoring", namespaceFlags.String())
}

func TestSinksFlagsType_Set(t *testing.T) {
	sinkFlags := SinksFlagsType{}
	assert.NoError(t, sinkFlags.Set("metrics"))
	assert.Equal(t, 1, len(sinkFlags))
	err := sinkFlags.Set("logs1")
	assert.NotNil(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), "output value is not valid"))
	assert.Equal(t, 1, len(sinkFlags))
	assert.NoError(t, sinkFlags.Set("metrics"))
	assert.Equal(t, 1, len(sinkFlags))
	assert.NoError(t, sinkFlags.Set("logs"))
	assert.Equal(t, 2, len(sinkFlags))
}

func TestSinksFlagsType_String(t *testing.T) {
	sinkFlags := SinksFlagsType{}
	assert.NoError(t, sinkFlags.Set("metrics"))
	assert.NoError(t, sinkFlags.Set("logs"))
	assert.NoError(t, sinkFlags.Set("metrics"))
	assert.Equal(t, "metrics,logs", sinkFlags.String())
}
