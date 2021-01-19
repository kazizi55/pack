package pack_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/heroku/color"
	"github.com/pelletier/go-toml"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	"github.com/buildpacks/pack"
	"github.com/buildpacks/pack/internal/dist"
	h "github.com/buildpacks/pack/testhelpers"
)

func TestCreateBuildpack(t *testing.T) {
	color.Disable(true)
	defer color.Disable(false)
	spec.Run(t, "create_builder", testCreateBuildpack, spec.Parallel(), spec.Report(report.Terminal{}))
}

func testCreateBuildpack(t *testing.T, when spec.G, it spec.S) {
	var (
		subject *pack.Client
		tmpDir  string
	)

	it.Before(func() {
		var err error

		tmpDir, err = ioutil.TempDir("", "create-buildpack-test")
		h.AssertNil(t, err)

		subject, err = pack.NewClient()
		h.AssertNil(t, err)
	})

	when("#CreateBuildpack", func() {
		it("should create bash scripts", func() {
			err := subject.CreateBuildpack(context.TODO(), pack.CreateBuildpackOptions{
				Path:     tmpDir,
				Language: "bash",
				ID:       "example/my-cnb",
				Stacks: []dist.Stack{
					{
						ID:     "some-stack",
						Mixins: []string{"some-mixin"},
					},
				},
			})
			h.AssertNil(t, err)

			info, err := os.Stat(filepath.Join(tmpDir, "bin/build"))
			h.AssertFalse(t, os.IsNotExist(err))
			h.AssertTrue(t, info.Mode()&0100 != 0)

			info, err = os.Stat(filepath.Join(tmpDir, "bin/detect"))
			h.AssertFalse(t, os.IsNotExist(err))
			h.AssertTrue(t, info.Mode()&0100 != 0)

			assertBuildpackToml(t, tmpDir, "example/my-cnb")
		})
	})
}

func assertBuildpackToml(t *testing.T, path string, id string) {
	buildpackTOML := filepath.Join(path, "buildpack.toml")
	_, err := os.Stat(buildpackTOML)
	h.AssertFalse(t, os.IsNotExist(err))

	f, err := os.Open(buildpackTOML)
	h.AssertNil(t, err)
	var buildpackDescriptor dist.BuildpackDescriptor
	err = toml.NewDecoder(f).Decode(&buildpackDescriptor)
	h.AssertNil(t, err)
	defer f.Close()

	fmt.Printf("%s\n", buildpackDescriptor)
	h.AssertEq(t, buildpackDescriptor.Info.ID, "example/my-cnb")
}
