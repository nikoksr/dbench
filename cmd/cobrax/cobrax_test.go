package cobrax

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

var testArgs = []string{"testArg1", "testArg2"}

func TestHooks(t *testing.T) {
	t.Parallel()

	testCmd := &cobra.Command{}
	counters := make([]int, 3)

	handlerFunc := func(cmd *cobra.Command, args []string) {
		assert.Equal(t, cmd, testCmd)
		assert.Equal(t, args, testArgs)
		counters[0]++
	}

	handlerFunc2 := func(cmd *cobra.Command, args []string) {
		assert.Equal(t, cmd, testCmd)
		assert.Equal(t, args, testArgs)
		counters[1]++
	}

	chained := Hooks(handlerFunc, handlerFunc2)

	chained(testCmd, testArgs)

	assert.Equal(t, counters[0], 1)
	assert.Equal(t, counters[1], 1)
}

func TestHooksE(t *testing.T) {
	t.Parallel()

	testCmd := &cobra.Command{}
	counters := make([]int, 3)

	handlerFunc := func(cmd *cobra.Command, args []string) error {
		assert.Equal(t, cmd, testCmd)
		assert.Equal(t, args, testArgs)
		counters[0]++
		return nil
	}

	handlerFunc2 := func(cmd *cobra.Command, args []string) error {
		assert.Equal(t, cmd, testCmd)
		assert.Equal(t, args, testArgs)
		counters[1]++
		return nil
	}

	chained := HooksE(handlerFunc, handlerFunc2)

	err := chained(testCmd, testArgs)

	assert.Nil(t, err)
	assert.Equal(t, counters[0], 1)
	assert.Equal(t, counters[1], 1)
}

func TestHooksEWithError(t *testing.T) {
	t.Parallel()

	testCmd := &cobra.Command{}
	counters := make([]int, 3)

	handlerFunc := func(cmd *cobra.Command, args []string) error {
		assert.Equal(t, cmd, testCmd)
		assert.Equal(t, args, testArgs)
		counters[0]++
		return assert.AnError
	}

	handlerFunc2 := func(cmd *cobra.Command, args []string) error {
		assert.Equal(t, cmd, testCmd)
		assert.Equal(t, args, testArgs)
		counters[1]++
		return nil
	}

	chained := HooksE(handlerFunc, handlerFunc2)

	err := chained(testCmd, testArgs)

	assert.Equal(t, err, assert.AnError)
	assert.Equal(t, counters[0], 1)
	assert.Equal(t, counters[1], 0)
}
