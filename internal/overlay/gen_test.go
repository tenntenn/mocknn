package overlay_test

import (
	"flag"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tenntenn/golden"
	"github.com/tenntenn/mocknn/internal/overlay"
	"golang.org/x/tools/go/packages"
)

var (
	flagUpdate bool
)

func init() {
	flag.BoolVar(&flagUpdate, "update", false, "update golden files")
}

func TestGenerator_Generate(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		wantErr bool
	}{
		"normal": {false},
	}

	goldendir := filepath.Join(testdata(t), "golden")

	for name, tt := range cases {
		name, tt := name, tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tmpdir := t.TempDir()
			g := &overlay.Generator{
				Pkgs: load(t, name),
				Dir:  tmpdir,
			}

			got, err := g.Generate()

			switch {
			case err == nil && tt.wantErr:
				t.Fatal("expected error does not occur")
			case err != nil && !tt.wantErr:
				t.Fatal("unexpected error:", err)
			}

			for key, val := range got.Replace {
				got.Replace[key] = strings.TrimPrefix(val, tmpdir+string([]rune{filepath.Separator}))
			}

			gotDir := strings.ReplaceAll(golden.Txtar(t, tmpdir), "-- "+filepath.ToSlash(tmpdir), "-- ")

			name := strings.ReplaceAll(t.Name(), "/", "-")
			if flagUpdate {
				golden.Update(t, goldendir, name+"-overlay-json", got)
				golden.Update(t, goldendir, name+"-mock-files", gotDir)
				return
			}

			if diff := golden.Diff(t, goldendir, name+"-overlay-json", got); diff != "" {
				t.Errorf("overlayJSON\n:%s", diff)
			}

			if diff := golden.Diff(t, goldendir, name+"-mock-files", gotDir); diff != "" {
				t.Errorf("mock files\n:%s", diff)
			}
		})
	}
}

func testdata(t *testing.T) string {
	dir, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	return dir
}

func load(t *testing.T, pkg string) []*packages.Package {
	config := &packages.Config{
		Dir:   filepath.Join(testdata(t), "src", pkg),
		Tests: true,
		Mode: packages.NeedName | packages.NeedTypes |
			packages.NeedSyntax | packages.NeedTypesInfo |
			packages.NeedModule,
	}

	pkgs, err := packages.Load(config, "./...")
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		for _, err := range pkg.Errors {
			t.Fatal("unexpected error:", pkg, err)
		}
	})

	return pkgs
}
