package main

import (
    "flag"
    "fmt"
    "os"
    "time"
)

const (
    stateDisabled = "DISABLED"
    stateActivating = "ACTIVATING"
    stateProtected = "PROTECTED"
    stateDegradedBlocked = "DEGRADED_BLOCKED"
    stateStopping = "STOPPING"
    stateRecovering = "RECOVERING"
    stateErrorBlocked = "ERROR_BLOCKED"
)

type config struct {
    Enable bool
    ConfigPath string
    Verbose bool
}

type protectionState struct {
    State string
    Reason string
    Timestamp time.Time
}

type serviceRuntime struct {
    state protectionState
    config config
}

func parseConfig() config {
    cfg := config{}
    flag.BoolVar(&cfg.Enable, "enable", false, "enable protection")
    flag.BoolVar(&cfg.Verbose, "verbose", false, "enable verbose logging")
    flag.StringVar(&cfg.ConfigPath, "config", "config.yaml", "configuration path")
    flag.Parse()
    return cfg
}

func (s *serviceRuntime) setState(state, reason string) {
    s.state = protectionState{State: state, Reason: reason, Timestamp: time.Now().UTC()}
}

func (s *serviceRuntime) ensureFailClosed() error {
    if !s.config.Enable {
        s.setState(stateDisabled, "protection disabled by config")
        return nil
    }

    // Safety-first default: do not claim protection unless enforcement is validated.
    // This scaffold intentionally prefers a blocked state rather than allowing direct traffic.
    s.setState(stateErrorBlocked, "enforcement not verified on this platform")
    return fmt.Errorf("protection not verified: requires Linux firewall, NetworkManager, and TUN enforcement validation")
}

func main() {
    cfg := parseConfig()
    rt := serviceRuntime{config: cfg}

    fmt.Printf("SnekProxy starting with config path=%s\n", cfg.ConfigPath)

    if err := rt.ensureFailClosed(); err != nil {
        if cfg.Verbose {
            fmt.Fprintf(os.Stderr, "fail-closed state: %v\n", err)
        }
        fmt.Printf("Protection state: %s (%s)\n", rt.state.State, rt.state.Reason)
        os.Exit(1)
    }

    rt.setState(stateProtected, "placeholder: platform enforcement has not been verified in this environment")
    fmt.Printf("Protection state: %s (%s)\n", rt.state.State, rt.state.Reason)
}
