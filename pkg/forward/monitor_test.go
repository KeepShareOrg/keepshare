package forward

import "testing"

func TestResolveMonitorURL(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		mail    string
		want    string
		wantErr bool
	}{
		{"empty falls back to https", "", "https://mail.keepshare.org", "wss://mail.keepshare.org/api/v1/monitor/messages", false},
		{"empty falls back to http", "", "http://localhost:9000", "ws://localhost:9000/api/v1/monitor/messages", false},
		{"explicit wins", "wss://custom/endpoint", "https://mail.keepshare.org", "wss://custom/endpoint", false},
		{"explicit URL with path kept", "wss://custom/other/path", "https://mail.keepshare.org", "wss://custom/other/path", false},
		{"invalid explicit", "://broken", "https://mail.keepshare.org", "", true},
		{"empty both", "", "", "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := resolveMonitorURL(c.in, c.mail)
			if (err != nil) != c.wantErr {
				t.Fatalf("err mismatch: got %v wantErr %v", err, c.wantErr)
			}
			if err == nil && got != c.want {
				t.Fatalf("want %q got %q", c.want, got)
			}
		})
	}
}

func TestNewMonitorRequiresURL(t *testing.T) {
	if _, err := NewMonitor(""); err == nil {
		t.Fatal("expected error for empty wsURL")
	}
	if _, err := NewMonitor("wss://x/y"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
