package forward

import (
	"strings"
	"testing"
)

func TestFilter_AcceptExactRecipient(t *testing.T) {
	masters := map[string]struct{}{"4pc0ai1uqv59@xx.com": {}}
	matched, reason, ok := Filter(Message{
		From: "noreply@accounts.mypikpak.com",
		To:   []string{"4pc0ai1uqv59@xx.com"},
	}, masters)
	if !ok || reason != "" {
		t.Fatalf("want accepted, got ok=%v reason=%q", ok, reason)
	}
	if matched != "4pc0ai1uqv59@xx.com" {
		t.Fatalf("matched master mismatch: %q", matched)
	}
}

func TestFilter_AcceptWhenOneOfManyRecipientsMatches(t *testing.T) {
	masters := map[string]struct{}{"master@xx.com": {}}
	matched, _, ok := Filter(Message{
		From: "noreply@accounts.mypikpak.com",
		To:   []string{"someone@else.com", "master@xx.com"},
	}, masters)
	if !ok || matched != "master@xx.com" {
		t.Fatalf("should accept when any recipient matches; matched=%q ok=%v", matched, ok)
	}
}

func TestFilter_RejectWrongFrom(t *testing.T) {
	masters := map[string]struct{}{"a@x": {}}
	_, _, ok := Filter(Message{From: "evil@x", To: []string{"a@x"}}, masters)
	if ok {
		t.Fatal("must reject non-noreply from")
	}
}

func TestFilter_RejectUnknownRecipient(t *testing.T) {
	masters := map[string]struct{}{"known@x": {}}
	_, reason, ok := Filter(Message{From: "noreply@accounts.mypikpak.com", To: []string{"stranger@y"}}, masters)
	if ok {
		t.Fatal("must reject unknown recipient")
	}
	if !strings.Contains(reason, "recipient") {
		t.Fatalf("reason should mention recipient: %q", reason)
	}
}

func TestFilter_AcceptByLocalPart(t *testing.T) {
	masters := map[string]struct{}{"alice@somewhere.io": {}}
	matched, _, ok := Filter(Message{
		From: "noreply@accounts.mypikpak.com",
		To:   []string{"alice@otherhost.io"},
	}, masters)
	if !ok {
		t.Fatal("should accept by local-part match")
	}
	if matched != "alice@somewhere.io" {
		t.Fatalf("local-part match should return the DB master email, got %q", matched)
	}
}

func TestFilter_CaseInsensitiveFrom(t *testing.T) {
	masters := map[string]struct{}{"a@x": {}}
	_, _, ok := Filter(Message{From: "NoReply@Accounts.MyPikPak.com", To: []string{"a@x"}}, masters)
	if !ok {
		t.Fatal("from regex is case-insensitive")
	}
}
