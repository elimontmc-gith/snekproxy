package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"snekproxy/internal/config"
	"snekproxy/internal/contract"
	"snekproxy/internal/enforcement"
)

type options struct {
	configPath string
	enable     bool
	disable    bool
	verbose    bool
}

func parseOptions(args []string, output io.Writer) (options, error) {
	var opts options
	flags := flag.NewFlagSet("snekproxy", flag.ContinueOnError)
	flags.SetOutput(output)
	flags.StringVar(&opts.configPath, "config", "config.yaml", "configuration path")
	flags.BoolVar(&opts.enable, "enable", false, "explicitly activate blocking enforcement")
	flags.BoolVar(&opts.disable, "disable", false, "remove the owned SnekProxy blocking policy")
	flags.BoolVar(&opts.verbose, "verbose", false, "enable verbose logging")
	if err := flags.Parse(args); err != nil {
		return options{}, err
	}
	if flags.NArg() != 0 {
		return options{}, fmt.Errorf("unexpected arguments: %v", flags.Args())
	}
	if opts.enable && opts.disable {
		return options{}, errors.New("--enable and --disable cannot be used together")
	}
	return opts, nil
}

func run(args []string, stdout, stderr io.Writer) error {
	opts, err := parseOptions(args, stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	platform := enforcement.NewPlatform()
	if opts.disable {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		state, err := enforcement.Deactivate(ctx, platform)
		fmt.Fprintf(stdout, "Protection state: %s (%s)\n", state.Name, state.Reason)
		return err
	}

	cfg, err := config.Load(opts.configPath)
	if err != nil {
		return err
	}
	if cfg.Verbose || opts.verbose {
		fmt.Fprintf(stdout, "SnekProxy configuration loaded from %s\n", opts.configPath)
	}

	if !cfg.Enabled {
		state := disabledState("protection disabled by config")
		fmt.Fprintf(stdout, "Protection state: %s (%s)\n", state.Name, state.Reason)
		return nil
	}
	if !opts.enable {
		state := disabledState("explicit --enable is required to change host networking")
		fmt.Fprintf(stdout, "Protection state: %s (%s)\n", state.Name, state.Reason)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	state, err := enforcement.Activate(ctx, platform)
	cancel()
	fmt.Fprintf(stdout, "Protection state: %s (%s)\n", state.Name, state.Reason)
	if err != nil {
		return err
	}

	fmt.Fprintln(stdout, "SnekProxy is in blocking-only mode; use --disable to remove its firewall policy.")
	if cfg.Verbose || opts.verbose {
		fmt.Fprintln(stdout, "Stopping this process does not remove the firewall policy.")
	}

	waitForShutdown()
	return nil
}

func disabledState(reason string) contract.State {
	return contract.State{Name: contract.StateDisabled, Reason: reason, Timestamp: time.Now().UTC()}
}

func waitForShutdown() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
