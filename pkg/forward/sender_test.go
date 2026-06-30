package forward

import (
	"bytes"
	"encoding/base64"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/KeepShareOrg/keepshare/config"
)

func TestSender_Build_PrependsNoticeWithOriginalRecipient(t *testing.T) {
	sender := NewSender("HEAD<TO>TAIL")
	b, err := sender.Build(
		config.ForwardAccount{Username: "Forwarder", Email: "fwd@x"},
		"user@example.com",
		"4pc0ai1uqv59@xx.com",
		Message{
			ID:      "1",
			From:    "noreply@accounts.mypikpak.com",
			To:      []string{"4pc0ai1uqv59@xx.com"},
			Subject: "Hello",
		},
		"Original text body",
	)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(b.Subject, "[Forward]") || !strings.Contains(b.Subject, "Hello") {
		t.Fatalf("subject unexpected: %q", b.Subject)
	}
	if b.From != "fwd@x" || b.To != "user@example.com" {
		t.Fatalf("from/to wrong: %+v", b)
	}

	raw := string(b.Body)
	if !strings.HasPrefix(raw, "From: Forwarder <fwd@x>\r\n") {
		t.Fatalf("From header missing or wrong: %q", raw[:200])
	}
	// plain part is base64-encoded; decode and assert content.
	wantText := "HEAD4pc0ai1uqv59@xx.comTAIL\nOriginal text body"
	if !strings.Contains(raw, base64.StdEncoding.EncodeToString([]byte(wantText))) {
		t.Fatalf("plain body should contain substituted notice + original body; raw=%q", raw)
	}
}

func TestSender_Build_SubstitutesOriginalSenderAndDate(t *testing.T) {
	sender := NewSender("FROM=<FROM>|DATE=<DATE>|SUBJECT=<SUBJECT>|TO=<TO>")
	sentAt := time.Date(2026, 6, 30, 8, 30, 0, 0, time.UTC)
	b, err := sender.Build(
		config.ForwardAccount{Username: "F", Email: "fwd@x"},
		"user@example.com",
		"master@xx.com",
		Message{
			ID:      "1",
			From:    "noreply@accounts.mypikpak.com",
			To:      []string{"master@xx.com"},
			Subject: "Security alert",
			Date:    sentAt,
		},
		"body",
	)
	if err != nil {
		t.Fatal(err)
	}
	want := "FROM=noreply@accounts.mypikpak.com|DATE=" + sentAt.Format(time.RFC1123Z) + "|SUBJECT=Security alert|TO=master@xx.com"
	if !strings.Contains(string(b.Body), base64.StdEncoding.EncodeToString([]byte(want+"\nbody"))) {
		t.Fatalf("notice should substitute from/date/subject/to; raw=%q", string(b.Body))
	}
}

func TestSender_Build_EmptyBodyStillSendsNotice(t *testing.T) {
	sender := NewSender("NOTICE")
	b, err := sender.Build(
		config.ForwardAccount{Username: "F", Email: "fwd@x"},
		"u@example.com",
		"m@xx",
		Message{ID: "1", From: "noreply@accounts.mypikpak.com", To: []string{"m@xx"}, Subject: "S"},
		"",
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b.Body), base64.StdEncoding.EncodeToString([]byte("NOTICE"))) {
		t.Fatalf("notice expected: %q", string(b.Body))
	}
}

func TestSender_Build_RequiresEmailAndUser(t *testing.T) {
	sender := NewSender("N")
	if _, err := sender.Build(config.ForwardAccount{}, "u@x", "m@x", Message{}, "body"); err == nil {
		t.Fatal("expected error when account.email empty")
	}
	if _, err := sender.Build(config.ForwardAccount{Email: "f@x"}, "", "m@x", Message{}, "body"); err == nil {
		t.Fatal("expected error when toUser empty")
	}
}

func TestSender_Send_SendsViaConfiguredServer(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	got := make(chan []byte, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		writeLine := func(s string) { _, _ = conn.Write([]byte(s + "\r\n")) }
		readChunk := func() {
			tmp := make([]byte, 4096)
			_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			_, _ = conn.Read(tmp)
		}
		writeLine("220 demo ESMTP")
		readChunk() // EHLO
		writeLine("250-demo")
		writeLine("250 AUTH PLAIN")
		readChunk() // AUTH PLAIN <base64>
		writeLine("235 OK")
		readChunk() // MAIL FROM
		writeLine("250 OK")
		readChunk() // RCPT TO
		writeLine("250 OK")
		readChunk() // DATA
		writeLine("354 send data")
		var data bytes.Buffer
		tmp := make([]byte, 4096)
		for {
			_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			n, err := conn.Read(tmp)
			if n > 0 {
				data.Write(tmp[:n])
			}
			if bytes.Contains(data.Bytes(), []byte("\r\n.\r\n")) {
				writeLine("250 OK")
				writeLine("221 bye")
				break
			}
			if err != nil {
				break
			}
		}
		got <- data.Bytes()
	}()

	host, portStr, _ := net.SplitHostPort(ln.Addr().String())
	port := 0
	for _, c := range portStr {
		port = port*10 + int(c-'0')
	}

	account := config.ForwardAccount{Username: "F", Email: "fwd@x", Server: host, Port: port, Password: "pw"}
	sender := NewSender("NOTICE")
	b, err := sender.Build(
		account,
		"u@example.com",
		"m@xx",
		Message{ID: "1", From: "noreply@accounts.mypikpak.com", To: []string{"m@xx"}, Subject: "S"},
		"B",
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := sender.Send(b, account); err != nil {
		t.Fatalf("send err: %v", err)
	}

	select {
	case data := <-got:
		if !bytes.Contains(data, []byte("From: F <fwd@x>")) {
			t.Fatalf("captured data lacks From header: %q", string(data))
		}
		if !bytes.Contains(data, []byte(base64.StdEncoding.EncodeToString([]byte("NOTICE\nB")))) {
			t.Fatalf("captured data lacks base64-encoded body: %q", string(data))
		}
	case <-time.After(3 * time.Second):
		t.Fatal("no data captured from server")
	}
}
