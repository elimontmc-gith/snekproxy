//go:build linux

package enforcement

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const nftTableName = "snekproxy"

var blockingPolicy = []byte(`table inet snekproxy {
 comment "managed by SnekProxy blocking-only policy v1"
 chain output {
  type filter hook output priority filter; policy drop;
  oifname "lo" accept
 }
}
`)

type commandRunner interface {
	Run(context.Context, string, []string, []byte) ([]byte, error)
}

type execCommandRunner struct{}

func (execCommandRunner) Run(ctx context.Context, name string, args []string, input []byte) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if input != nil {
		cmd.Stdin = bytes.NewReader(input)
	}
	return cmd.CombinedOutput()
}

type linuxPlatform struct {
	runner commandRunner
	uid    func() int
}

func NewPlatform() Platform {
	return &linuxPlatform{runner: execCommandRunner{}, uid: os.Geteuid}
}

func (p *linuxPlatform) CheckPrerequisites(ctx context.Context) error {
	if p.uid() != 0 {
		return errors.New("Linux enforcement requires root privileges")
	}
	if _, err := exec.LookPath("nft"); err != nil {
		return fmt.Errorf("nftables command is unavailable: %w", err)
	}
	if output, err := p.runner.Run(ctx, "nft", []string{"--version"}, nil); err != nil {
		return fmt.Errorf("check nftables: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func (p *linuxPlatform) InstallBlockingPolicy(ctx context.Context) error {
	if p.uid() != 0 {
		return errors.New("installing Linux enforcement requires root privileges")
	}
	exists, err := p.tableExists(ctx)
	if err != nil {
		return err
	}
	if exists {
		if err := p.VerifyBlockingPolicy(ctx); err != nil {
			return fmt.Errorf("refusing to replace existing inet %s table: %w", nftTableName, err)
		}
		return nil
	}

	if output, err := p.runner.Run(ctx, "nft", []string{"-f", "-"}, blockingPolicy); err != nil {
		return fmt.Errorf("apply nftables blocking policy: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return p.VerifyBlockingPolicy(ctx)
}

func (p *linuxPlatform) VerifyBlockingPolicy(ctx context.Context) error {
	if p.uid() != 0 {
		return errors.New("verifying Linux enforcement requires root privileges")
	}
	output, err := p.runner.Run(ctx, "nft", []string{"list", "table", "inet", nftTableName}, nil)
	if err != nil {
		return fmt.Errorf("list nftables table: %w: %s", err, strings.TrimSpace(string(output)))
	}
	if !sameNftPolicy(string(output), string(blockingPolicy)) {
		return errors.New("existing nftables table does not exactly match the SnekProxy blocking policy")
	}
	return nil
}

func (p *linuxPlatform) RemoveBlockingPolicy(ctx context.Context) error {
	if p.uid() != 0 {
		return errors.New("removing Linux enforcement requires root privileges")
	}
	exists, err := p.tableExists(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	if err := p.VerifyBlockingPolicy(ctx); err != nil {
		return fmt.Errorf("refusing to remove unrecognized inet %s table: %w", nftTableName, err)
	}
	if output, err := p.runner.Run(ctx, "nft", []string{"delete", "table", "inet", nftTableName}, nil); err != nil {
		return fmt.Errorf("delete nftables blocking policy: %w: %s", err, strings.TrimSpace(string(output)))
	}
	exists, err = p.tableExists(ctx)
	if err != nil {
		return fmt.Errorf("verify nftables policy removal: %w", err)
	}
	if exists {
		return errors.New("nftables blocking policy still exists after removal")
	}
	return nil
}

func (p *linuxPlatform) tableExists(ctx context.Context) (bool, error) {
	output, err := p.runner.Run(ctx, "nft", []string{"list", "tables"}, nil)
	if err != nil {
		return false, fmt.Errorf("list nftables tables: %w: %s", err, strings.TrimSpace(string(output)))
	}
	for _, line := range strings.Split(string(output), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 3 && fields[0] == "table" && fields[1] == "inet" && fields[2] == nftTableName {
			return true, nil
		}
	}
	return false, nil
}

func sameNftPolicy(actual, expected string) bool {
	normalize := func(value string) []string {
		var lines []string
		for _, line := range strings.Split(value, "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				lines = append(lines, line)
			}
		}
		return lines
	}
	return equalStrings(normalize(actual), normalize(expected))
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
