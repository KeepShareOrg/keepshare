package log

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRequestID(t *testing.T) {
	for i := 0; i < 20; i++ {
		id := NewRequestID()
		t.Log(len(id), id)
	}
}

func TestRequestIDFromContext(t *testing.T) {
	const r = "1hhma73iq_ddatjsovu1"
	ctx := DataContext(context.Background(), DataContextOptions{RequestID: r})
	id, tm := RequestIDFromContext(ctx)
	t.Logf("go id: %s, tm: %s", id, tm)
	assert.Equal(t, int64(1702629707), tm.Unix())
	assert.Equal(t, r, id)
}
