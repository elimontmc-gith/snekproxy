package enforcement

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"snekproxy/internal/contract"
)

type fakePlatform struct {
	calls []string
	fail  string
}

func (p *fakePlatform) CheckPrerequisites(context.Context) error {
	return p.call("check")
}

func (p *fakePlatform) InstallBlockingPolicy(context.Context) error {
	return p.call("install")
}

func (p *fakePlatform) VerifyBlockingPolicy(context.Context) error {
	return p.call("verify")
}

func (p *fakePlatform) RemoveBlockingPolicy(context.Context) error {
	return p.call("remove")
}

func (p *fakePlatform) call(name string) error {
	p.calls = append(p.calls, name)
	if p.fail == name {
		return errors.New("platform operation failed")
	}
	return nil
}

func TestActivateVerifiesBlockingPolicyAndNeverClaimsProtected(t *testing.T) {
	platform := &fakePlatform{}

	state, err := Activate(context.Background(), platform)
	if err != nil {
		t.Fatal(err)
	}
	if state.Name != contract.StateErrorBlocked {
		t.Fatalf("expected %s, got %s", contract.StateErrorBlocked, state.Name)
	}
	if !reflect.DeepEqual(platform.calls, []string{"check", "install", "verify"}) {
		t.Fatalf("unexpected operation order: %v", platform.calls)
	}
}

func TestActivateStopsAtFirstFailure(t *testing.T) {
	for _, test := range []struct {
		fail  string
		calls []string
	}{
		{fail: "check", calls: []string{"check"}},
		{fail: "install", calls: []string{"check", "install"}},
		{fail: "verify", calls: []string{"check", "install", "verify"}},
	} {
		t.Run(test.fail, func(t *testing.T) {
			platform := &fakePlatform{fail: test.fail}
			state, err := Activate(context.Background(), platform)
			if err == nil {
				t.Fatal("expected activation error")
			}
			if state.Name != contract.StateErrorBlocked {
				t.Fatalf("expected blocked state, got %s", state.Name)
			}
			if !reflect.DeepEqual(platform.calls, test.calls) {
				t.Fatalf("unexpected operation order: %v", platform.calls)
			}
		})
	}
}

func TestDeactivateReportsDisabledAfterCleanup(t *testing.T) {
	platform := &fakePlatform{}

	state, err := Deactivate(context.Background(), platform)
	if err != nil {
		t.Fatal(err)
	}
	if state.Name != contract.StateDisabled {
		t.Fatalf("expected %s, got %s", contract.StateDisabled, state.Name)
	}
	if !reflect.DeepEqual(platform.calls, []string{"remove"}) {
		t.Fatalf("unexpected operation order: %v", platform.calls)
	}
}

func TestDeactivateReportsBlockedWhenCleanupFails(t *testing.T) {
	platform := &fakePlatform{fail: "remove"}

	state, err := Deactivate(context.Background(), platform)
	if err == nil {
		t.Fatal("expected cleanup error")
	}
	if state.Name != contract.StateErrorBlocked {
		t.Fatalf("expected %s, got %s", contract.StateErrorBlocked, state.Name)
	}
}
