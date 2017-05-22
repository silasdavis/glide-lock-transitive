package action

import (
	"fmt"
	"strings"

	"github.com/Masterminds/glide/action"
	"github.com/Masterminds/glide/cache"
	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/repo"
	"github.com/Masterminds/glide/util"
)

// This is modified version of github.com/Masterminds/glide/action/get.go

// Get fetches one or more dependencies and installs.
//
// This includes resolving dependency resolution and re-generating the lock file.
func Get(names []string, installer *repo.Installer, testDeps bool) {
	cache.SystemLock()

	action.EnsureGopath()
	action.EnsureVendorDir()
	conf := action.EnsureConfig()
	glidefile, err := gpath.Glide()
	if err != nil {
		msg.Die("Could not find Glide file: %s", err)
	}

	// Add the packages to the config.
	if count, err2 := addPkgsToConfig(conf, names, testDeps); err2 != nil {
		msg.Die("Failed to get new packages: %s", err2)
	} else if count == 0 {
		msg.Warn("Nothing to do")
		return
	}

	// Fetch the new packages. Can't resolve versions via installer.Update if
	// get is called while the vendor/ directory is empty so we checkout
	// everything.
	err = installer.Checkout(conf)
	if err != nil {
		msg.Die("Failed to checkout packages: %s", err)
	}

	// Prior to resolving dependencies we need to start working with a clone
	// of the conf because we'll be making real changes to it.
	confcopy := conf.Clone()

	// Set Reference
	if err := repo.SetReference(confcopy, installer.ResolveTest); err != nil {
		msg.Err("Failed to set references: %s", err)
	}

	err = installer.Export(confcopy)
	if err != nil {
		msg.Die("Unable to export dependencies to vendor directory: %s", err)
	}

	// Write YAML
	if err := conf.WriteFile(glidefile); err != nil {
		msg.Die("Failed to write glide YAML file: %s", err)
	}

}

// addPkgsToConfig adds the given packages to the config file.
//
// Along the way it:
// - ensures that this package is not in the ignore list
// - checks to see if this is already in the dependency list.
// - splits version of of package name and adds the version attribute
// - separates repo from packages
// - generates a list of subpackages
func addPkgsToConfig(conf *cfg.Config, names []string, testDeps bool) (int, error) {

	if len(names) == 1 {
		msg.Info("Preparing to install %d package.", len(names))
	} else {
		msg.Info("Preparing to install %d packages.", len(names))
	}
	numAdded := 0
	for _, name := range names {
		var version string
		parts := strings.Split(name, "#")
		if len(parts) > 1 {
			name = parts[0]
			version = parts[1]
		}

		msg.Info("Attempting to get package %s", name)

		root, subpkg := util.NormalizeName(name)
		if len(root) == 0 {
			return 0, fmt.Errorf("Package name is required for %q.", name)
		}

		if conf.HasDependency(root) {

			var moved bool
			var dep *cfg.Dependency
			// Move from DevImports to Imports
			if !testDeps && !conf.Imports.Has(root) && conf.DevImports.Has(root) {
				dep = conf.DevImports.Get(root)
				conf.Imports = append(conf.Imports, dep)
				conf.DevImports = conf.DevImports.Remove(root)
				moved = true
				numAdded++
				msg.Info("--> Moving %s from testImport to import", root)
			} else if testDeps && conf.Imports.Has(root) {
				msg.Warn("--> Test dependency %s already listed as import", root)
			}

			// Check if the subpackage is present.
			if subpkg != "" {
				if dep == nil {
					dep = conf.Imports.Get(root)
					if dep == nil && testDeps {
						dep = conf.DevImports.Get(root)
					}
				}
				if dep.HasSubpackage(subpkg) {
					if !moved {
						msg.Warn("--> Package %q is already in glide.yaml. Skipping", name)
					}
				} else {
					dep.Subpackages = append(dep.Subpackages, subpkg)
					msg.Info("--> Adding sub-package %s to existing import %s", subpkg, root)
					numAdded++
				}
			} else if !moved {
				msg.Warn("--> Package %q is already in glide.yaml. Skipping", root)
			}
			continue
		}

		if conf.HasIgnore(root) {
			msg.Warn("--> Package %q is set to be ignored in glide.yaml. Skipping", root)
			continue
		}

		dep := &cfg.Dependency{
			Name: root,
		}

		if version != "" {
			dep.Reference = version
		}

		if len(subpkg) > 0 {
			dep.Subpackages = []string{subpkg}
		}

		if dep.Reference != "" {
			msg.Info("--> Adding %s to your configuration with the version %s", dep.Name, dep.Reference)
		} else {
			msg.Info("--> Adding %s to your configuration", dep.Name)
		}

		if testDeps {
			conf.DevImports = append(conf.DevImports, dep)
		} else {
			conf.Imports = append(conf.Imports, dep)
		}
		numAdded++
	}
	return numAdded, nil
}
