package authguard

import (
	"fmt"
	"testing"
	"time"
)

func TestLimiterPerClientAndWindow(t *testing.T) {
	l := New()
	now := time.Unix(1000, 0)
	l.now = func() time.Time { return now }
	for i := 0; i < perClientLimit; i++ {
		if !l.Allow("127.0.0.1:1234") {
			t.Fatal("limited early")
		}
	}
	if l.Allow("127.0.0.1:9999") {
		t.Fatal("same IP bypassed limit by changing port")
	}
	if !l.Allow("127.0.0.2:1234") {
		t.Fatal("different IP was limited")
	}
	now = now.Add(window)
	if !l.Allow("127.0.0.1:1234") {
		t.Fatal("limit did not expire")
	}
}
func TestLimiterGlobalLimit(t *testing.T) {
	l := New()
	for i := 0; i < globalLimit; i++ {
		if !l.Allow(fmt.Sprintf("10.0.%d.%d:1234", i/250, i%250)) {
			t.Fatal("limited early")
		}
	}
	if l.Allow("10.2.0.1:1234") {
		t.Fatal("global limit bypassed")
	}
}
