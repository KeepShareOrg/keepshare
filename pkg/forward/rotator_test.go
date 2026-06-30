package forward

import (
	"testing"
	"time"

	"github.com/KeepShareOrg/keepshare/config"
)

func mkAcc(email string) config.ForwardAccount {
	return config.ForwardAccount{Email: email, Server: "smtp", Port: 587, Username: "u", Password: "p"}
}

func TestRotator_RoundRobinOrder(t *testing.T) {
	r := NewRotator([]config.ForwardAccount{
		mkAcc("a@x"), mkAcc("b@x"), mkAcc("c@x"),
	}, time.Minute)

	want := []string{"a@x", "b@x", "c@x", "a@x", "b@x"}
	for i, w := range want {
		acc, more := r.Next()
		if !more {
			t.Fatalf("[%d] expected more accounts", i)
		}
		if acc.Email != w {
			t.Fatalf("[%d] want %s got %s", i, w, acc.Email)
		}
	}
}

func TestRotator_SkipCoolingDown(t *testing.T) {
	r := NewRotator([]config.ForwardAccount{
		mkAcc("a@x"), mkAcc("b@x"), mkAcc("c@x"),
	}, 5*time.Minute)

	r.MarkFailed("a@x")
	a1, more := r.Next()
	if !more || a1.Email == "a@x" {
		t.Fatalf("a@x should be skipped, got %s ok=%v", a1.Email, more)
	}
}

func TestRotator_AllCoolingDownReturnsFalse(t *testing.T) {
	r := NewRotator([]config.ForwardAccount{mkAcc("a@x")}, time.Minute)
	r.MarkFailed("a@x")
	acc, more := r.Next()
	if more || acc.Email != "" {
		t.Fatalf("expected no available account, got %+v ok=%v", acc, more)
	}
}

func TestRotator_MarkOKClearsCooldown(t *testing.T) {
	r := NewRotator([]config.ForwardAccount{mkAcc("a@x")}, time.Minute)
	r.MarkFailed("a@x")
	r.MarkOK("a@x")
	acc, more := r.Next()
	if !more || acc.Email != "a@x" {
		t.Fatalf("MarkOK should clear cooldown; got %+v ok=%v", acc, more)
	}
}
