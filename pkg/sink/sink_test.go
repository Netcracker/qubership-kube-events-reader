package sink

import (
	"regexp"
	"testing"

	"github.com/Netcracker/qubership-kube-events-reader/pkg/filter"
	"github.com/Netcracker/qubership-kube-events-reader/pkg/test"
	"github.com/stretchr/testify/assert"
)

var sinkTest = &Sink{
	Match: []*Rule{
		{
			Kind: regexp.MustCompile("Pod|Deployment"),
		},
	},
	Exclude: []*Rule{
		{
			Type: regexp.MustCompile("Normal"),
		},
	},
}

func TestSink_IsEventAllowed(t *testing.T) {
	assert.True(t, sinkTest.IsEventAllowed(test.EventDeploymentMonitoring))
	assert.False(t, sinkTest.IsEventAllowed(test.EventPodLogging))
	assert.True(t, sinkTest.IsEventAllowed(test.EventPodTracing))
	assert.False(t, sinkTest.IsEventAllowed(test.EventPvcMonitoring))
}

func Test_initializeSinkWithFilters_Nil(t *testing.T) {
	sinkInitialized := initializeSinkWithFilters(nil)
	assert.NotNil(t, sinkInitialized)
	assert.Equal(t, 0, len(sinkInitialized.Match))
	assert.Equal(t, 0, len(sinkInitialized.Exclude))

	assert.True(t, sinkInitialized.IsEventAllowed(test.EventDeploymentMonitoring))
	assert.True(t, sinkInitialized.IsEventAllowed(test.EventPodLogging))
	assert.True(t, sinkInitialized.IsEventAllowed(test.EventPodTracing))
	assert.True(t, sinkInitialized.IsEventAllowed(test.EventPvcMonitoring))
}

func Test_initializeSinkWithFilters_Empty(t *testing.T) {
	var filtersSink = filter.Sink{
		Name:    "logs",
		Exclude: []filter.EventMatch{},
	}
	sinkInitialized := initializeSinkWithFilters(&filtersSink)
	assert.NotNil(t, sinkInitialized)
	assert.Equal(t, 0, len(sinkInitialized.Match))
	assert.Equal(t, 0, len(sinkInitialized.Exclude))

	assert.True(t, sinkInitialized.IsEventAllowed(test.EventDeploymentMonitoring))
	assert.True(t, sinkInitialized.IsEventAllowed(test.EventPodLogging))
	assert.True(t, sinkInitialized.IsEventAllowed(test.EventPodTracing))
	assert.True(t, sinkInitialized.IsEventAllowed(test.EventPvcMonitoring))
}

var filtersSinkMatchAndExclude = filter.Sink{
	Name: "logs",
	Exclude: []filter.EventMatch{
		{
			Type: "Normal",
		},
	},
	Match: []filter.EventMatch{
		{
			Kind: "Pod",
		},
		{
			Namespace: "tracing|monitoring",
		},
	},
}

func Test_initializeSinkWithFilters_MatchAndExclude(t *testing.T) {

	sinkInitialized := initializeSinkWithFilters(&filtersSinkMatchAndExclude)
	assert.NotNil(t, sinkInitialized)
	assert.Equal(t, 1, len(sinkInitialized.Exclude))
	assert.Equal(t, 2, len(sinkInitialized.Match))

	assert.True(t, sinkInitialized.IsEventAllowed(test.EventDeploymentMonitoring))
	assert.False(t, sinkInitialized.IsEventAllowed(test.EventPodLogging))
	assert.True(t, sinkInitialized.IsEventAllowed(test.EventPodTracing))
	assert.True(t, sinkInitialized.IsEventAllowed(test.EventPvcMonitoring))
}

func Test_initializeSinkWithFilters_Match(t *testing.T) {
	var filtersSink = filter.Sink{
		Name: "logs",
		Match: []filter.EventMatch{
			{
				Kind: "Deployment",
			},
			{
				Namespace: "logging",
			},
		},
	}
	sinkInitialized := initializeSinkWithFilters(&filtersSink)
	assert.NotNil(t, sinkInitialized)
	assert.Equal(t, 0, len(sinkInitialized.Exclude))
	assert.Equal(t, 2, len(sinkInitialized.Match))

	assert.True(t, sinkInitialized.IsEventAllowed(test.EventDeploymentMonitoring))
	assert.True(t, sinkInitialized.IsEventAllowed(test.EventPodLogging))
	assert.False(t, sinkInitialized.IsEventAllowed(test.EventPodTracing))
	assert.False(t, sinkInitialized.IsEventAllowed(test.EventPvcMonitoring))
}
