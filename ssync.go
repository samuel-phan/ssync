package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/google/goterm/term"
)

type RsyncOpts struct {
	ProjectDir string
	Nodes      []string
	Excludes   []string
	Delete     bool
	SudoUser   string
	DryRun     bool
}

const (
	deleteFlagName        = "delete"
	sudoUserFlagName      = "sudo-user"
	sudoUserFlagShortName = "su"
	dryRunFlagName        = "dry-run"
	dryRunFlagShortName   = "n"
)

var (
	initFlag     = flag.Bool("init", false, "init ssync configuration")
	deleteFlag   = flag.Bool(deleteFlagName, false, "delete extraneous files from destination dirs")
	sudoUserFlag string
	dryRunFlag   bool

	// Print debug logs
	Debug    = false
	myLogger = log.New(os.Stderr, "", 0)
)

func debugf(format string, args ...interface{}) {
	if !Debug {
		return
	}
	s := fmt.Sprintf(format, args...)
	myLogger.Print(term.Bluef(s))
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			`Usage: %s [options]

%s -init [DIR]
    Initialize the project configuration in the DIR directory. By default, DIR
    is the current directory.

%s [options]
    Rsync to the nodes found from the project configuration.

%s [options] NODE1:/path1 NODE2:/path2 NODE3: ...
    Rsync to the given nodes. The rsync options come from the project
    configuration if found; the user configuration, otherwise.

Options:
`, os.Args[0], os.Args[0], os.Args[0], os.Args[0])
		flag.PrintDefaults()
	}
	// TODO: add exclude & other rsync flags
	flag.StringVar(&sudoUserFlag, sudoUserFlagName, "", "destination user ownership")
	flag.StringVar(&sudoUserFlag, sudoUserFlagShortName, "", "destination user ownership")

	flag.BoolVar(&dryRunFlag, dryRunFlagName, false, "show what would have been transferred")
	flag.BoolVar(&dryRunFlag, dryRunFlagShortName, false, "show what would have been transferred")

	flag.BoolVar(&Debug, "debug", false, "print debug messages")
	flag.BoolVar(&Debug, "d", false, "print debug messages")
	flag.Parse()

	// Load user config
	userConf, err := loadOrInitUserConf()
	if err != nil {
		log.Fatal(err)
	}

	// Eval flags
	if *initFlag {
		var projectDir string
		if len(flag.Args()) >= 1 {
			projectDir = flag.Args()[0]
		} else {
			projectDir = "."
		}

		err := initProjectConf(projectDir, userConf)
		if err != nil {
			log.Fatal(err)
		} else {
			return
		}
	}

	// Prepare arguments for rsync
	rsyncOpts := RsyncOpts{DryRun: dryRunFlag}

	// Load project config if possible
	projectDir, err := findProjectDir()
	if errors.Is(err, errProjectDirNotFound) {
		// Project dir not found
		rsyncOpts.ProjectDir, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		// get exclude opt from user config
		rsyncOpts.Excludes = userConf.Excludes
	} else {
		// Project dir found
		projectConf, err := loadProjectConf(projectDir)
		if err != nil {
			log.Fatal(err)
		}

		rsyncOpts.ProjectDir = projectDir
		rsyncOpts.Nodes = projectConf.Nodes
		rsyncOpts.Excludes = projectConf.Excludes
		rsyncOpts.Delete = projectConf.Delete
		rsyncOpts.SudoUser = projectConf.SudoUser
	}

	// check nodes arg
	if len(flag.Args()) > 0 {
		rsyncOpts.Nodes = flag.Args()
	}

	// check delete arg
	if isFlagPassed(deleteFlagName) {
		rsyncOpts.Delete = *deleteFlag
	}

	// check sudo-user arg
	if isFlagPassed(sudoUserFlagName) || isFlagPassed(sudoUserFlagShortName) {
		rsyncOpts.SudoUser = sudoUserFlag
	}

	err = rsync(projectDir, rsyncOpts)
	if err != nil {
		log.Fatal(err)
	}
}

func rsync(projectDir string, opts RsyncOpts) error {
	debugf("rsync opts: %v", opts)

	// Check opts
	if len(opts.Nodes) == 0 {
		return fmt.Errorf("no target node provided")
	}

	rsyncArgs := []string{"-avzP"}

	// excludes
	for _, ex := range opts.Excludes {
		rsyncArgs = append(rsyncArgs, "--exclude", ex)
	}

	// delete
	if opts.Delete {
		rsyncArgs = append(rsyncArgs, "--delete")
	}

	// sudo-user
	if opts.SudoUser != "" {
		rsyncArgs = append(rsyncArgs, "--rsync-path", fmt.Sprintf("sudo -u %s rsync", opts.SudoUser))
	}

	// dry-run
	if opts.DryRun {
		rsyncArgs = append(rsyncArgs, "--dry-run")
	}

	// nodes
	rsyncArgs = append(rsyncArgs, opts.ProjectDir)
	rsyncArgs = append(rsyncArgs, opts.Nodes...)

	sb := []string{"rsync"}
	for _, s := range rsyncArgs {
		if strings.Contains(s, " ") {
			sb = append(sb, fmt.Sprintf("\"%s\"", s))
		} else {
			sb = append(sb, s)
		}
	}
	fmt.Printf("Running: %s\n", strings.Join(sb, " "))
	cmd := exec.Command("rsync", rsyncArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	} else {
		return nil
	}
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
