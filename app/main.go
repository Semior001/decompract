package main

// golangci-lint warns on the use of go-flags without alias
//noinspection GoRedundantImportAlias
import (
	"fmt"
	"os"

	"github.com/Semior001/decompract/app/cmd"
	log "github.com/go-pkgz/lgr"
	flags "github.com/jessevdk/go-flags"
)

// Opts describes cli arguments and flags to execute a command
type Opts struct {
	cmd.Server
	Dbg bool `long:"dbg" env:"DEBUG" description:"turn on debug mode"`
}

var version = "unknown"

func main() {
	fmt.Printf("decompract version: %s\n", version)
	var opts Opts
	p := flags.NewParser(&opts, flags.Default)

	// after failure command does not return non-zero code
	if _, err := p.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	setupLog(opts.Dbg)

	// commands implements CommonOptionsCommander to allow passing set of extra options defined for all commands
	opts.SetCommon(cmd.CommonOpts{
		Version: version,
	})

	if err := opts.Execute(os.Args[1:]); err != nil {
		log.Printf("[ERROR] failed to execute command %+v", err)
	}
}

func setupLog(dbg bool) {
	if dbg {
		log.Setup(log.Debug, log.CallerFile, log.CallerFunc, log.Msec, log.LevelBraces)
		return
	}
	log.Setup(log.Msec, log.LevelBraces)
}
