//go:build linux

package enforcement

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

type fakeCommand struct {
	output []byte
	err    error
}

type fakeRunner struct {
	commands  []string
	responses []fakeCommand
	input     []byte
}

func (r *fakeRunner) Run(_ context.Context, name string, args []string, input []byte) ([]byte, error) {
	r.commands = append(r.commands, name+" "+strings.Join(args, " "))
	if input != nil {
		r.input = append([]byte(nil), input...)
	}
	if len(r.responses) == 0 {
		return nil, errors.New("unexpected command")
	}
	response := r.responses[0]
	r.responses = r.responses[1:]
	return response.output, response.err
}

func TestInstallCreatesAndVerifiesBlockingPolicy(t *testing.T) {
	runner := &fakeRunner{responses: []fakeCommand{
		{output: []byte("table inet filter\n")},
		{},
		{output: blockingPolicy},
	}}
	platform := &linuxPlatform{runner: runner, uid: func() int { return 0 }}

	if err := platform.InstallBlockingPolicy(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(runner.commands, []string{
		"nft list tables",
		"nft -f -",
		"nft list table inet snekproxy",
	}) {
		t.Fatalf("unexpected commands: %v", runner.commands)
	}
	if string(runner.input) != string(blockingPolicy) {
		t.Fatalf("unexpected policy input: %q", runner.input)
	}
}

func TestRemoveRefusesToDeleteUnrecognizedTable(t *testing.T) {
	runner := &fakeRunner{responses: []fakeCommand{
		{output: []byte("table inet snekproxy\n")},
		{output: []byte("table inet snekproxy {\n chain output { policy accept; }\n}\n")},
	}}
	platform := &linuxPlatform{runner: runner, uid: func() int { return 0 }}

	if err := platform.RemoveBlockingPolicy(context.Background()); err == nil {
		t.Fatal("expected refusal to remove unrecognized table")
	}
	if len(runner.commands) != 2 {
		t.Fatalf("expected no delete command, got %v", runner.commands)
	}
}

func TestInstallRefusesToOverwritePreexistingTable(t *testing.T) {
	runner := &fakeRunner{responses: []fakeCommand{
		{output: []byte("table inet snekproxy\n")},
		{output: []byte("table inet snekproxy {\n chain output { policy accept; }\n}\n")},
	}}
	platform := &linuxPlatform{runner: runner, uid: func() int { return 0 }}

	if err := platform.InstallBlockingPolicy(context.Background()); err == nil {
		t.Fatal("expected refusal to overwrite pre-existing table")
	}
	if len(runner.commands) != 2 {
		t.Fatalf("expected no policy-apply command, got %v", runner.commands)
	}
}

func TestRemoveDeletesOnlyVerifiedPolicy(t *testing.T) {
	runner := &fakeRunner{responses: []fakeCommand{
		{output: []byte("table inet snekproxy\n")},
		{output: blockingPolicy},
		{},
		{},
	}}
	platform := &linuxPlatform{runner: runner, uid: func() int { return 0 }}

	if err := platform.RemoveBlockingPolicy(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(runner.commands, []string{
		"nft list tables",
		"nft list table inet snekproxy",
		"nft delete table inet snekproxy",
		"nft list tables",
	}) {
		t.Fatalf("unexpected commands: %v", runner.commands)
	}
}

func TestInstallRequiresRoot(t *testing.T) {
	runner := &fakeRunner{}
	platform := &linuxPlatform{runner: runner, uid: func() int { return 1000 }}

	if err := platform.InstallBlockingPolicy(context.Background()); err == nil {
		t.Fatal("expected root requirement error")
	}
	if len(runner.commands) != 0 {
		t.Fatalf("unexpected commands for unprivileged caller: %v", runner.commands)
	}
}
