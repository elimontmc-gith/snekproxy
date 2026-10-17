package enforcement

import (
	"context"
	"fmt"
	"time"

	"snekproxy/internal/contract"
)

type Platform interface {
	CheckPrerequisites(context.Context) error
	InstallBlockingPolicy(context.Context) error
	VerifyBlockingPolicy(context.Context) error
	RemoveBlockingPolicy(context.Context) error
}

func Activate(ctx context.Context, platform Platform) (contract.State, error) {
	if err := platform.CheckPrerequisites(ctx); err != nil {
		return blockedState("platform prerequisites failed; no protection can be claimed"), fmt.Errorf("check enforcement prerequisites: %w", err)
	}
	if err := platform.InstallBlockingPolicy(ctx); err != nil {
		return blockedState("blocking policy could not be installed; protection is not verified"), fmt.Errorf("install blocking policy: %w", err)
	}
	if err := platform.VerifyBlockingPolicy(ctx); err != nil {
		return blockedState("blocking policy could not be verified; protection is not verified"), fmt.Errorf("verify blocking policy: %w", err)
	}

	return blockedState("blocking policy verified; forwarding is not implemented, so non-loopback traffic remains blocked"), nil
}

func Deactivate(ctx context.Context, platform Platform) (contract.State, error) {
	if err := platform.RemoveBlockingPolicy(ctx); err != nil {
		state := blockedState("could not safely remove SnekProxy blocking policy")
		return state, fmt.Errorf("remove blocking policy: %w", err)
	}
	return contract.State{
		Name:      contract.StateDisabled,
		Reason:    "SnekProxy blocking policy removed by explicit request",
		Timestamp: time.Now().UTC(),
	}, nil
}

func blockedState(reason string) contract.State {
	return contract.State{
		Name:      contract.StateErrorBlocked,
		Reason:    reason,
		Timestamp: time.Now().UTC(),
	}
}
