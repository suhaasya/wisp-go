package server_test

import (
	"context"
	"testing"

	"github.com/suhaasya/wisp-go/internal/server"
)

func TestOpenURL_defaultBrowser(t *testing.T) {
	t.Parallel()
	var gotName string
	var gotArgs []string
	l := server.Launcher{
		Run: func(_ context.Context, name string, args ...string) error {
			gotName = name
			gotArgs = append([]string{}, args...)
			return nil
		},
		Platform: server.PlatformDarwin,
	}
	if err := l.OpenURL(context.Background(), "http://127.0.0.1:1/", false, ""); err != nil {
		t.Fatal(err)
	}
	if gotName != "open" || len(gotArgs) != 1 || gotArgs[0] != "http://127.0.0.1:1/" {
		t.Fatalf("got %q %v", gotName, gotArgs)
	}
}

func TestOpenURL_appModeChrome(t *testing.T) {
	t.Parallel()
	var gotArgs []string
	l := server.Launcher{
		Run: func(_ context.Context, name string, args ...string) error {
			gotArgs = append([]string{name}, args...)
			return nil
		},
		Platform: server.PlatformDarwin,
	}
	chrome := "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	if err := l.OpenURL(context.Background(), "http://127.0.0.1:9/", true, chrome); err != nil {
		t.Fatal(err)
	}
	want := []string{chrome, "--app=http://127.0.0.1:9/"}
	if len(gotArgs) != len(want) {
		t.Fatalf("got %v, want %v", gotArgs, want)
	}
	for i := range want {
		if gotArgs[i] != want[i] {
			t.Fatalf("got %v, want %v", gotArgs, want)
		}
	}
}
