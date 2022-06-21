package internal

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tenntenn/mocknn/internal/goflags"
	"github.com/tenntenn/mocknn/internal/overlay"
	"go.uber.org/multierr"
	"golang.org/x/tools/go/packages"
)

const (
	ExitSuccess = 0
	ExitError   = 1
)

type Mocknn struct {
	Version   string
	Dir       string
	Output    io.Writer
	ErrOutput io.Writer
	Input     io.Reader
}

func Main(version string, args []string) int {
	m := &Mocknn{
		Version:   version,
		Dir:       ".",
		Output:    os.Stdout,
		ErrOutput: os.Stderr,
		Input:     os.Stdin,
	}
	return m.Main(args)
}

func (m *Mocknn) Main(args []string) int {
	if err := m.Run(args); err != nil {
		fmt.Fprintln(m.ErrOutput, "mocknn:", err)
		return ExitError
	}
	return ExitSuccess
}

func (m *Mocknn) Run(args []string) error {

	if len(args) == 0 {
		args = []string{"."}
	}

	switch args[0] {
	case "-v":
		if _, err := fmt.Fprintln(os.Stdout, "mocknn", m.Version); err != nil {
			return err
		}
	case "test":
		if err := m.testWithMock(args[1:]); err != nil {
			return err
		}
	default:
		if err := m.printOverlayJSON(args); err != nil {
			return err
		}
	}

	return nil
}

func (m *Mocknn) load(patterns []string) (_ []*packages.Package, rerr error) {
	config := &packages.Config{
		Dir:   m.Dir,
		Tests: true,
		Mode: packages.NeedName | packages.NeedTypes |
			packages.NeedSyntax | packages.NeedTypesInfo |
			packages.NeedModule,
	}

	pkgs, err := packages.Load(config, patterns...)
	if err != nil {
		return nil, err
	}

	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		for _, err := range pkg.Errors {
			rerr = multierr.Append(rerr, err)
		}
	})

	if rerr != nil {
		return nil, rerr
	}

	return pkgs, nil
}

func (m *Mocknn) testWithMock(args []string) (rerr error) {

	tmpdir, err := os.MkdirTemp("", "mocknn-*")
	if err != nil {
		return err
	}
	defer func() {
		rerr = multierr.Append(rerr, os.RemoveAll(tmpdir))
	}()

	var (
		flagOverlay string
	)

	flags := flag.NewFlagSet("mocknn test", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&flagOverlay, "overlay", "", "overlay json")
	goflags.All(flags)
	if err := flags.Parse(args); err != nil {
		return err
	}

	var initOverlay *packages.OverlayJSON
	if flagOverlay != "" {
		if err := json.NewDecoder(strings.NewReader(flagOverlay)).Decode(&initOverlay); err != nil {
			return err
		}
	}

	pkgs, err := m.load(flags.Args())
	if err != nil {
		return err
	}

	g := &overlay.Generator{
		Dir:     tmpdir,
		Pkgs:    pkgs,
		Overlay: initOverlay,
	}

	overlayJSON, err := g.Generate()
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(tmpdir, "overlay.json"))
	if err != nil {
		return err
	}

	if err := json.NewEncoder(f).Encode(overlayJSON); err != nil {
		return err
	}
	defer func() {
		rerr = multierr.Append(rerr, f.Close())
	}()

	opts := make([]string, 0, flags.NFlag())
	flags.Visit(func(f *flag.Flag) {
		opts = append(opts, fmt.Sprintf("-%s=%v", f.Name, f.Value))
	})

	goargs := append([]string{"test", "-overlay", f.Name()}, append(opts, flags.Args()...)...)
	cmd := exec.Command("go", goargs...)
	cmd.Stdout = m.Output
	cmd.Stderr = m.ErrOutput
	cmd.Stdin = m.Input
	cmd.Dir = m.Dir

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (m *Mocknn) printOverlayJSON(args []string) (rerr error) {

	pkgs, err := m.load(args)
	if err != nil {
	}

	g := &overlay.Generator{
		Pkgs: pkgs,
	}

	overlayJSON, err := g.Generate()
	if err != nil {
		return err
	}

	f, err := os.CreateTemp("", "mocknn-overlay-*.json")
	if err != nil {
		return err
	}
	defer func() {
		rerr = multierr.Append(rerr, f.Close())
	}()

	if err := json.NewEncoder(f).Encode(overlayJSON); err != nil {
		return err
	}

	if _, err := fmt.Fprint(m.Output, f.Name()); err != nil {
		return err
	}

	return nil
}
