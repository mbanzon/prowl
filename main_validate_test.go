package main

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

func TestSetupSyncStoresWaitGroup(t *testing.T) {
	_, ctx, cancel, wg := setupSync()
	defer cancel()

	ctxWG, ok := ctx.Value(wgKey).(*sync.WaitGroup)
	if !ok || ctxWG != wg {
		t.Fatalf("expected wait group stored in context")
	}
}

func TestValidatePortWith(t *testing.T) {
	var called bool
	var message string
	fatalf := func(args ...any) {
		called = true
		message = fmt.Sprint(args...)
	}

	validatePortWith(5001, fatalf)
	if called {
		t.Fatalf("unexpected fatal for valid port")
	}

	validatePortWith(80, fatalf)
	if !called {
		t.Fatalf("expected fatal for privileged port")
	}
	if message == "" {
		t.Fatalf("expected fatal message to be set")
	}
}

func TestValidateProtectionWith(t *testing.T) {
	var called bool
	fatalf := func(args ...any) {
		called = true
	}

	if got := validateProtectionWith(false, "", func(string) string { return "" }, fatalf); got != "" {
		t.Fatalf("expected empty secret when protection is disabled, got %q", got)
	}
	if called {
		t.Fatalf("unexpected fatal when protection is disabled")
	}

	called = false
	if got := validateProtectionWith(true, "inline", func(string) string { return "" }, fatalf); got != "inline" {
		t.Fatalf("expected inline secret, got %q", got)
	}
	if called {
		t.Fatalf("unexpected fatal with inline secret")
	}

	called = false
	if got := validateProtectionWith(true, "", func(string) string { return "env-secret" }, fatalf); got != "env-secret" {
		t.Fatalf("expected env secret, got %q", got)
	}
	if called {
		t.Fatalf("unexpected fatal with env secret")
	}

	called = false
	if got := validateProtectionWith(true, "", func(string) string { return "" }, fatalf); got != "" {
		t.Fatalf("expected empty result on missing secret, got %q", got)
	}
	if !called {
		t.Fatalf("expected fatal when secret is missing")
	}
}

func TestValidatePortWrapper(t *testing.T) {
	validatePort(5001)
}

func TestValidateProtectionWrapper(t *testing.T) {
	if got := validateProtection(false, ""); got != "" {
		t.Fatalf("expected empty secret when protection is disabled, got %q", got)
	}

	if err := os.Setenv(secretKeyEnv, "env-wrapper-secret"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	defer os.Unsetenv(secretKeyEnv)

	if got := validateProtection(true, ""); got != "env-wrapper-secret" {
		t.Fatalf("expected env secret, got %q", got)
	}
}

func TestShutdownContextTimeout(t *testing.T) {
	ctx, cancel := shutdownContext()
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatalf("expected shutdown context to have deadline")
	}
	if time.Until(deadline) <= 0 {
		t.Fatalf("expected shutdown deadline in the future")
	}
}
