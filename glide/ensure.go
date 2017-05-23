package glide

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
// The main thing it does differently is to avoid updating the entire tree
// and it will update the version of a dependency that is already present in the
// config

// Ensure fetches one or more dependencies and installs them. Ensuring they are
// present at at the desired version
//
// This includes resolving dependency resolution and re-generating the lock file.
// each name can be an entry such as github.com/silasdavis/foo@v2.3.0
// (to specify a version) or github.com/silasdavis/foo without
func Ensure(name string, installer *repo.Installer) {
	cache.SystemLock()

	action.EnsureGopath()
	action.EnsureVendorDir()
	conf := action.EnsureConfig()
	glidefile, err := gpath.Glide()
	if err != nil {
		msg.Die("Could not find Glide file: %s", err)
	}

	// Add the packages to the config.
	if err2 := ensurePackageInConfig(conf, name); err2 != nil {
		msg.Die("Failed to get new packages: %s", err2)
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

// ensurePackageInConfig adds the given packages to the config file at the
// specified versions.
//
// Along the way it:
// - ensures that this package is not in the ignore list
// - splits version of of package name and adds the version attribute
// - separates repo from packages
// - generates a list of subpackages
func ensurePackageInConfig(conf *cfg.Config, name string) error {
	var version string
	parts := strings.Split(name, "@")
	if len(parts) > 1 {
		name = parts[0]
		version = parts[1]
	}

	msg.Info("Attempting to get package %s", name)

	root, subpkg := util.NormalizeName(name)
	if len(root) == 0 {
		return fmt.Errorf("Package name is required for %q.", name)
	}

	dep := conf.Imports.Get(root)
	if dep == nil {
		dep = &cfg.Dependency{
			Name: root,
		}
	} else {
		// Check if the subpackage is present.
		if subpkg != "" {
			if dep.HasSubpackage(subpkg) {
				msg.Warn("--> Package %q is already in glide.yaml. Skipping", name)
			} else {
				dep.Subpackages = append(dep.Subpackages, subpkg)
				msg.Info("--> Adding sub-package %s to existing import %s", subpkg, root)
			}
		} else if version != "" {
			msg.Info("--> Setting package %s to version %s", root, version)
			dep.Reference = version
		} else {
			msg.Warn("--> Package %q is already in glide.yaml. Skipping", root)
		}
		return nil
	}

	if conf.HasIgnore(root) {
		msg.Warn("--> Package %q is set to be ignored in glide.yaml. Skipping", root)
		return nil
	}

	if version != "" {
		dep.Reference = version
	}
	if subpkg != "" {
		dep.Subpackages = []string{subpkg}
	}

	if dep.Reference != "" {
		msg.Info("--> Adding %s to your configuration with the version %s", dep.Name, dep.Reference)
	} else {
		msg.Info("--> Adding %s to your configuration", dep.Name)
	}

	conf.Imports = append(conf.Imports, dep)
	return nil
}
