package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
	cli "github.com/urfave/cli"
	gx "github.com/whyrusleeping/gx/gxutil"
)

var (
	pm *gx.PM
)

const Version = "0.0.1"

func main() {
	cfg, err := gx.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	pm, err = gx.NewPM(cfg)
	if err != nil {
		log.Fatal(err)
	}

	app := cli.NewApp()
	app.Author = "bruinxs"
	app.Version = Version
	app.Usage = "keep gx package same"
	app.Action = keep

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func keep(c *cli.Context) error {
	if len(c.Args()) == 0 {
		return errors.New("requires a package reference")
	}

	pkg, err := loadPackageFile(gx.PkgFileName)
	if err != nil {
		return err
	}

	ipath, err := gx.InstallPath(pkg.Language, "", false)
	if err != nil {
		return err
	}

	depname := c.Args().First()
	dephash, err := pm.ResolveDepName(depname)
	if err != nil {
		return err
	}

	deppkg, err := pm.GetPackageTo(dephash, filepath.Join(ipath, "gx", "ipfs", dephash))
	if err != nil {
		return err
	}

	depdeps := make(map[string]*gx.Dependency, len(deppkg.Dependencies))
	for _, d := range deppkg.Dependencies {
		depdeps[d.Name] = d
	}

	for _, d := range pkg.Dependencies {
		if dep, ok := depdeps[d.Name]; ok {
			if dep.Hash != d.Hash {
				if err := update(d.Hash, dep.Hash); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func loadPackageFile(path string) (*gx.Package, error) {
	if path == gx.PkgFileName {
		root, err := gx.GetPackageRoot()
		if err != nil {
			return nil, err
		}

		path = filepath.Join(root, gx.PkgFileName)
	}

	var pkg gx.Package
	err := gx.LoadPackageFile(&pkg, path)
	if err != nil {
		return nil, err
	}

	return &pkg, nil
}

func update(old, new string) error {
	cmd := exec.Command("gx", "update", old, new)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("gx-go", "hook", "post-update", old, new)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
