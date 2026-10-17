package contract

import "time"

const (
	StateDisabled = "DISABLED"
	StateActivating = "ACTIVATING"
	StateProtected = "PROTECTED"
	StateDegradedBlocked = "DEGRADED_BLOCKED"
	StateStopping = "STOPPING"
	StateRecovering = "RECOVERING"
	StateErrorBlocked = "ERROR_BLOCKED"
)

const (
	StatusPass = "PASS"
	StatusFail = "FAIL"
	StatusNotVerified = "NOT VERIFIED"
	StatusNotApplicable = "NOT APPLICABLE"
)

type State struct {
	Name      string
	Reason    string
	Timestamp time.Time
}

type Policy struct {
	Enabled          bool
	ForwardingReady  bool
	FirewallReady    bool
	DNSProtected     bool
	PlatformVerified bool
	Reason           string
}

func (p Policy) Evaluate() State {
	if !p.Enabled {
		return State{Name: StateDisabled, Reason: "protection disabled by config", Timestamp: time.Now().UTC()}
	}

	if !p.PlatformVerified {
		return State{Name: StateErrorBlocked, Reason: "platform enforcement is not verified", Timestamp: time.Now().UTC()}
	}

	if !p.FirewallReady || !p.DNSProtected {
		return State{Name: StateDegradedBlocked, Reason: "required enforcement policy is incomplete", Timestamp: time.Now().UTC()}
	}

	if !p.ForwardingReady {
		return State{Name: StateErrorBlocked, Reason: "forwarding engine not ready; fail closed", Timestamp: time.Now().UTC()}
	}

	return State{Name: StateProtected, Reason: "all required enforcement checks passed", Timestamp: time.Now().UTC()}
}

func (p Policy) IsBlocked() bool {
	state := p.Evaluate()
	return state.Name == StateDegradedBlocked || state.Name == StateErrorBlocked
}
