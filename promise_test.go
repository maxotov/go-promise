package promise

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_WaitForValue_timeout_occurred(t *testing.T) {
	promises := NewInt64MultiPromises()

	ctx := context.Background()
	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Millisecond*5)
	defer cancel()

	value, err := promises.WaitForValue(ctxWithTimeout, "test_key")
	require.EqualError(t, err, "context deadline exceeded")
	require.EqualValues(t, 0, value)

	pp := promises.Get("test_key")
	require.Equal(t, 1, len(pp.promises))

	require.False(t, pp.resolved)
	require.Equal(t, int64(0), pp.value)
}

func Test_AddAndGet(t *testing.T) {
	promises := NewInt64MultiPromises()

	p := promises.AddAndGet("test_key")
	require.EqualValues(t, 1, cap(p))

	p.Resolve(12)

	requireValueInChannel(t, 12, p)
	requireChannelIsClosed(t, p)
}

func Test_GetAll(t *testing.T) {
	promises := NewInt64MultiPromises()

	key := "test_key"
	p1 := promises.AddAndGet(key)
	require.EqualValues(t, 1, cap(p1))

	p2 := promises.AddAndGet(key)
	require.EqualValues(t, 1, cap(p2))

	all := promises.Get(key)
	require.EqualValues(t, &Int64Promises{promises: []Int64Promise{p1, p2}}, all)
}

func Test_ResolveAll(t *testing.T) {
	promises := NewInt64MultiPromises()

	key := "test_key"
	p1 := promises.AddAndGet(key)
	require.EqualValues(t, 1, cap(p1))

	p2 := promises.AddAndGet(key)
	require.EqualValues(t, 1, cap(p2))

	promises.Resolve(key, 12)

	pp := promises.Get(key)
	require.True(t, pp.resolved)
	require.Equal(t, int64(12), pp.value)

	requireValueInChannel(t, 12, p1)
	requireChannelIsClosed(t, p1)

	requireValueInChannel(t, 12, p2)
	requireChannelIsClosed(t, p2)
}

func requireValueInChannel(t *testing.T, expected int64, ch chan int64) {
	v, ok := <-ch
	require.True(t, ok)
	require.EqualValues(t, expected, v)
}

func requireChannelIsClosed(t *testing.T, ch chan int64) {
	t.Helper()
	_, ok := <-ch
	require.False(t, ok)
}
