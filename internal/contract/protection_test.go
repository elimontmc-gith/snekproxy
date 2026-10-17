package contract

import "testing"

func TestPolicyBlockedWhenNotVerified(t *testing.T) {
	p := Policy{Enabled: true, PlatformVerified: false, FirewallReady: false, DNSProtected: false, ForwardingReady: false}
	state := p.Evaluate()
	if state.Name != StateErrorBlocked {
		t.Fatalf("expected %s, got %s", StateErrorBlocked, state.Name)
	}
}

func TestPolicyProtectedWhenVerified(t *testing.T) {
	p := Policy{Enabled: true, PlatformVerified: true, FirewallReady: true, DNSProtected: true, ForwardingReady: true}
	state := p.Evaluate()
	if state.Name != StateProtected {
		t.Fatalf("expected %s, got %s", StateProtected, state.Name)
	}
}

func TestPolicyBlockedWhenForwardingUnavailable(t *testing.T) {
	p := Policy{Enabled: true, PlatformVerified: true, FirewallReady: true, DNSProtected: true, ForwardingReady: false}
	state := p.Evaluate()
	if state.Name != StateErrorBlocked {
		t.Fatalf("expected %s, got %s", StateErrorBlocked, state.Name)
	}
}
