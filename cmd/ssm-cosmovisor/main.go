package main

import (
	"fmt"
	scv "github.com/blockpane/ssm-cosmovisor"
	"github.com/cosmos/cosmos-sdk/cosmovisor"
	"github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor/cmd"
	"github.com/cosmos/cosmos-sdk/cosmovisor/errors"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.SetFlags(log.Lshortfile)

	if os.Getenv("USE_SSM_KEY") == "true" || os.Getenv("USE_SSM_KEY") == "TRUE" {
		pk := scv.MustGetKey()
		e := scv.BackupOrig()
		if e != nil {
			log.Fatal(e)
		}

		// ensure original key is restored on exit
		exiting := make(chan os.Signal, 1) // can also trigger a restore if closed
		done := make(chan interface{}, 1)
		signal.Notify(exiting, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
		go func() {
			<-exiting
			restoreErr := scv.RestoreOrig()
			if restoreErr != nil {
				log.Println(restoreErr)
			}
			close(done)
		}()

		// provide the key once:
		go func() {
			selfNotify := func(err error) {
				if err != nil {
					_ = log.Output(2, err.Error())
					close(exiting)
					<-done
					os.Exit(1)
				}
			}
			e = scv.WritePipeOnce(pk)
			selfNotify(e)
			e = scv.WriteStrippedKey(*pk)
			selfNotify(e)
			pk = nil
		}()
	} else {
		log.Println("not fetching private consensus key from parameter store.")
	}

	// Now startup cosmovisor normally, everything below this line is copied from cosmovisor's main.go:
	cosmovisor.SetupLogging()
	if err := Run(os.Args[1:]); err != nil {
		cosmovisor.Logger.Error().Err(err).Msg("")
		os.Exit(1)
	}
}

// Run is the main loop, but returns an error
func Run(args []string) error {
	cmd.RunCosmovisorCommands(args)

	cfg, cerr := cosmovisor.GetConfigFromEnv()
	if cerr != nil {
		switch err := cerr.(type) {
		case *errors.MultiError:
			cosmovisor.Logger.Error().Msg("multiple configuration errors found:")
			for i, e := range err.GetErrors() {
				cosmovisor.Logger.Error().Err(e).Msg(fmt.Sprintf("  %d:", i+1))
			}
		default:
			cosmovisor.Logger.Error().Err(err).Msg("configuration error:")
		}
		return cerr
	}
	cosmovisor.Logger.Info().Msg("Configuration is valid:\n" + cfg.DetailString())
	launcher, err := cosmovisor.NewLauncher(cfg)
	if err != nil {
		return err
	}

	doUpgrade, err := launcher.Run(args, os.Stdout, os.Stderr)
	// if RestartAfterUpgrade, we launch after a successful upgrade (only condition LaunchProcess returns nil)
	for cfg.RestartAfterUpgrade && err == nil && doUpgrade {
		cosmovisor.Logger.Info().Str("app", cfg.Name).Msg("upgrade detected, relaunching")
		doUpgrade, err = launcher.Run(args, os.Stdout, os.Stderr)
	}
	if doUpgrade && err == nil {
		cosmovisor.Logger.Info().Msg("upgrade detected, DAEMON_RESTART_AFTER_UPGRADE is off. Verify new upgrade and start cosmovisor again.")
	}

	return err
}
