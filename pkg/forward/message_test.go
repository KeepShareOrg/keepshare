package forward

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"
)

func TestMessage_Fingerprint_StableForSameInput(t *testing.T) {
	now := time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC)
	a := Message{ID: "1", Mailbox: "m", From: "f@x", To: []string{"t@x"}, Subject: "S", Date: now, PosixMillis: now.UnixMilli()}
	b := Message{ID: "1", Mailbox: "m", From: "f@x", To: []string{"t@x"}, Subject: "S", Date: now.Add(time.Hour), PosixMillis: now.Add(time.Hour).UnixMilli()}
	if a.Fingerprint() != b.Fingerprint() {
		t.Fatal("fingerprint must ignore Date/PosixMillis and depend only on id+mailbox+from+to+subject")
	}
}

func TestMessage_Fingerprint_DiffersWhenAnyFieldChanges(t *testing.T) {
	base := Message{ID: "1", Mailbox: "m", From: "f@x", To: []string{"t@x"}, Subject: "S"}
	expected := sha256Hex("1|m|f@x|t@x|S")
	if base.Fingerprint() != expected {
		t.Fatalf("want %s got %s", expected, base.Fingerprint())
	}
	if base.Fingerprint() == (Message{ID: "2", Mailbox: "m", From: "f@x", To: []string{"t@x"}, Subject: "S"}).Fingerprint() {
		t.Fatal("id change should change fingerprint")
	}
	if base.Fingerprint() == (Message{ID: "1", Mailbox: "n", From: "f@x", To: []string{"t@x"}, Subject: "S"}).Fingerprint() {
		t.Fatal("mailbox change should change fingerprint")
	}
	if base.Fingerprint() == (Message{ID: "1", Mailbox: "m", From: "g@x", To: []string{"t@x"}, Subject: "S"}).Fingerprint() {
		t.Fatal("from change should change fingerprint")
	}
	if base.Fingerprint() == (Message{ID: "1", Mailbox: "m", From: "f@x", To: []string{"u@x"}, Subject: "S"}).Fingerprint() {
		t.Fatal("to change should change fingerprint")
	}
	if base.Fingerprint() == (Message{ID: "1", Mailbox: "m", From: "f@x", To: []string{"t@x"}, Subject: "Z"}).Fingerprint() {
		t.Fatal("subject change should change fingerprint")
	}
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
