package forward

import (
	"context"
	"testing"
	"time"

	"github.com/KeepShareOrg/keepshare/config"
)

type fakeSource struct {
	msgs []Message
	i    int
}

func (f *fakeSource) Next(ctx context.Context) (Message, bool, error) {
	if f.i >= len(f.msgs) {
		<-ctx.Done()
		return Message{}, false, ctx.Err()
	}
	m := f.msgs[f.i]
	f.i++
	return m, true, nil
}

// fixedBody returns the same body text for every fetch.
func fixedBody(body string) fetchBodyFunc {
	return func(context.Context, string, string) (string, error) { return body, nil }
}

func TestRun_ForwardsAndDedups(t *testing.T) {
	rdb, _ := newTestRedis(t)
	masters := map[string]struct{}{"4pc0ai1uqv59@xx.com": {}}
	loadMasters := func(context.Context) (map[string]struct{}, error) { return masters, nil }
	lookupUser := func(context.Context, string) (string, error) { return "target@example.com", nil }
	fetchBody := fixedBody("original body")

	var sent []string
	sendFn := func(b *Built, _ config.ForwardAccount) error {
		sent = append(sent, b.To)
		return nil
	}
	rotator := NewRotator([]config.ForwardAccount{{Email: "a@x", Server: "s", Port: 587}}, time.Minute)
	dedup := NewDeduper(rdb, time.Minute)
	sender := NewSender("NOTICE-<TO>-")
	src := &fakeSource{msgs: []Message{
		// accepted, should be sent once
		{ID: "1", Mailbox: "4pc0ai1uqv59@xx.com", From: "noreply@accounts.mypikpak.com", To: []string{"4pc0ai1uqv59@xx.com"}, Subject: "Hello"},
		// duplicate (same fingerprint), should be dropped by dedup
		{ID: "1", Mailbox: "4pc0ai1uqv59@xx.com", From: "noreply@accounts.mypikpak.com", To: []string{"4pc0ai1uqv59@xx.com"}, Subject: "Hello"},
		// wrong from, dropped
		{ID: "2", Mailbox: "4pc0ai1uqv59@xx.com", From: "evil@x", To: []string{"4pc0ai1uqv59@xx.com"}, Subject: "x"},
		// wrong recipient, dropped
		{ID: "3", Mailbox: "stranger@x", From: "noreply@accounts.mypikpak.com", To: []string{"stranger@x"}, Subject: "x"},
	}}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := runWith(ctx, rotator, dedup, sender, src, loadMasters, lookupUser, fetchBody, sendFn); err != nil && err != context.DeadlineExceeded {
		t.Fatalf("run err: %v", err)
	}

	if len(sent) != 1 {
		t.Fatalf("want 1 sent email, got %d: %v", len(sent), sent)
	}
	if sent[0] != "target@example.com" {
		t.Fatalf("forwarded to wrong user: %q", sent[0])
	}
}

func TestRun_CleanExitOnCancel(t *testing.T) {
	rdb, _ := newTestRedis(t)
	rotator := NewRotator([]config.ForwardAccount{{Email: "a@x"}}, time.Minute)
	dedup := NewDeduper(rdb, time.Minute)
	src := &fakeSource{}
	loadMasters := func(context.Context) (map[string]struct{}, error) { return nil, nil }
	lookupUser := func(context.Context, string) (string, error) { return "", nil }
	sendFn := func(*Built, config.ForwardAccount) error { return nil }

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := runWith(ctx, rotator, dedup, NewSender(""), src, loadMasters, lookupUser, fixedBody(""), sendFn); err != nil {
		t.Fatalf("expected clean exit, got %v", err)
	}
}

func TestRun_AllAccountsCoolingDownDropsAndPreservesDedupKey(t *testing.T) {
	rdb, _ := newTestRedis(t)
	rotator := NewRotator([]config.ForwardAccount{{Email: "a@x"}}, time.Hour)
	rotator.MarkFailed("a@x")
	dedup := NewDeduper(rdb, time.Minute)
	src := &fakeSource{msgs: []Message{
		{ID: "1", Mailbox: "4pc0ai1uqv59@xx.com", From: "noreply@accounts.mypikpak.com", To: []string{"4pc0ai1uqv59@xx.com"}},
	}}
	sendFn := func(*Built, config.ForwardAccount) error {
		t.Fatal("send should not be called when no accounts available")
		return nil
	}
	loadMasters := func(context.Context) (map[string]struct{}, error) {
		return map[string]struct{}{"4pc0ai1uqv59@xx.com": {}}, nil
	}
	lookupUser := func(context.Context, string) (string, error) { return "target@example.com", nil }

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := runWith(ctx, rotator, dedup, NewSender(""), src, loadMasters, lookupUser, fixedBody("b"), sendFn); err != nil && err != context.DeadlineExceeded {
		t.Fatalf("err: %v", err)
	}

	// dedup key was forgotten by runWith, so a future delivery can retry.
	fp := (Message{ID: "1", Mailbox: "4pc0ai1uqv59@xx.com", From: "noreply@accounts.mypikpak.com", To: []string{"4pc0ai1uqv59@xx.com"}}).Fingerprint()
	again, err := dedup.FirstSeen(context.Background(), fp)
	if err != nil {
		t.Fatal(err)
	}
	if !again {
		t.Fatal("dedup key should have been forgotten so the next delivery can succeed")
	}
}

func TestMaskEmail(t *testing.T) {
	cases := map[string]string{
		"user@example.com": "u***@example.com",
		"a@x":              "*@x",
		"notanemail":       "***",
		"":                 "***",
	}
	for in, want := range cases {
		if got := maskEmail(in); got != want {
			t.Fatalf("maskEmail(%q) = %q, want %q", in, got, want)
		}
	}
}
