package main

import (
	"os"
	"os/exec"
	"sync"
	"testing"
)

func TestValidatePortRejectsLowPort(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") == "1" && os.Getenv("HELPER") == "validatePortLow" {
		validatePort(80)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestValidatePortRejectsLowPort")
	cmd.Env = append(os.Environ(),
		"GO_WANT_HELPER_PROCESS=1",
		"HELPER=validatePortLow",
	)
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected validatePort to exit for low port")
	}
}

func TestValidateProtectionReturnsEmptyWhenDisabled(t *testing.T) {
	if got := validateProtection(false, ""); got != "" {
		t.Fatalf("expected empty secret when protection is disabled, got %q", got)
	}
}

func TestValidateProtectionUsesProvidedKey(t *testing.T) {
	if got := validateProtection(true, "abc"); got != "abc" {
		t.Fatalf("expected provided secret, got %q", got)
	}
}

func TestValidateProtectionUsesEnvKey(t *testing.T) {
	t.Setenv(secretKeyEnv, "from-env")
	if got := validateProtection(true, ""); got != "from-env" {
		t.Fatalf("expected env secret, got %q", got)
	}
}

func TestValidateProtectionRejectsMissingSecret(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") == "1" && os.Getenv("HELPER") == "validateProtectionMissing" {
		_ = os.Unsetenv(secretKeyEnv)
		validateProtection(true, "")
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestValidateProtectionRejectsMissingSecret")
	cmd.Env = append(os.Environ(),
		"GO_WANT_HELPER_PROCESS=1",
		"HELPER=validateProtectionMissing",
		secretKeyEnv+"=",
	)
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected validateProtection to exit when secret is missing")
	}
}

func TestSetupSyncStoresWaitGroup(t *testing.T) {
	_, ctx, cancel, wg := setupSync()
	defer cancel()

	if wg == nil {
		t.Fatal("expected wait group")
	}
	value := ctx.Value(wgKey)
	typed, ok := value.(*sync.WaitGroup)
	if !ok || typed == nil {
		t.Fatalf("expected wait group in context, got %T", value)
	}
}
