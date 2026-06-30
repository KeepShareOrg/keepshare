package forward

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(mr.Close)
	return redis.NewClient(&redis.Options{Addr: mr.Addr()}), mr
}

func TestDeduper_FirstSeenAndForget(t *testing.T) {
	ctx := context.Background()
	rdb, _ := newTestRedis(t)
	d := NewDeduper(rdb, time.Minute)

	first, err := d.FirstSeen(ctx, "abc")
	if err != nil || !first {
		t.Fatalf("first: ok=%v err=%v", first, err)
	}
	again, err := d.FirstSeen(ctx, "abc")
	if err != nil || again {
		t.Fatalf("second: ok=%v err=%v", again, err)
	}
	if err := d.Forget(ctx, "abc"); err != nil {
		t.Fatal(err)
	}
	third, err := d.FirstSeen(ctx, "abc")
	if err != nil || !third {
		t.Fatalf("after forget: ok=%v err=%v", third, err)
	}
}

func TestDeduper_TTLExpiry(t *testing.T) {
	ctx := context.Background()
	rdb, mr := newTestRedis(t)
	d := NewDeduper(rdb, time.Minute)

	if ok, _ := d.FirstSeen(ctx, "xyz"); !ok {
		t.Fatal("first should succeed")
	}
	mr.FastForward(2 * time.Minute)
	again, _ := d.FirstSeen(ctx, "xyz")
	if !again {
		t.Fatal("after TTL expiry, FirstSeen should succeed again")
	}
}
