//go:build !linux

package enforcement

import (
	"context"
	"errors"
)

type unsupportedPlatform struct{}

func NewPlatform() Platform {
	return unsupportedPlatform{}
}

func (unsupportedPlatform) CheckPrerequisites(context.Context) error {
	return errors.New("platform enforcement is not implemented")
}

func (unsupportedPlatform) InstallBlockingPolicy(context.Context) error {
	return errors.New("platform enforcement is not implemented")
}

func (unsupportedPlatform) VerifyBlockingPolicy(context.Context) error {
	return errors.New("platform enforcement is not implemented")
}

func (unsupportedPlatform) RemoveBlockingPolicy(context.Context) error {
	return errors.New("platform enforcement is not implemented")
}
