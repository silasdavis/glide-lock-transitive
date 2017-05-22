// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package glide_lock_transitive

import (
	"crypto/sha256"
	"fmt"
	"sort"

	"github.com/Masterminds/glide/cfg"
)

// Merges two glide lock files together, letting dependencies from 'base' be overwritten
// by those from 'override'. Returns the resultant glide lock file bytes
func MergeGlideLockFiles(baseLockFile, overrideLockFile *cfg.Lockfile) (*cfg.Lockfile, error) {
	imports := make(map[string]*cfg.Lock, len(baseLockFile.Imports))
	devImports := make(map[string]*cfg.Lock, len(baseLockFile.DevImports))
	// Copy the base dependencies into a map
	for _, lock := range baseLockFile.Imports {
		imports[lock.Name] = lock
	}
	for _, lock := range baseLockFile.DevImports {
		devImports[lock.Name] = lock
	}
	// Override base dependencies and add any extra ones
	for _, lock := range overrideLockFile.Imports {
		imports[lock.Name] = mergeLocks(imports[lock.Name], lock)
	}
	for _, lock := range overrideLockFile.DevImports {
		devImports[lock.Name] = mergeLocks(imports[lock.Name], lock)
	}

	deps := make([]*cfg.Dependency, 0, len(imports))
	devDeps := make([]*cfg.Dependency, 0, len(devImports))

	// Flatten to Dependencies
	for _, lock := range imports {
		deps = append(deps, pinnedDependencyFromLock(lock))
	}

	for _, lock := range devImports {
		devDeps = append(devDeps, pinnedDependencyFromLock(lock))
	}

	hasher := sha256.New()
	hasher.Write(([]byte)(baseLockFile.Hash))
	hasher.Write(([]byte)(overrideLockFile.Hash))

	return cfg.NewLockfile(deps, devDeps, fmt.Sprintf("%x", hasher.Sum(nil)))
}

func mergeLocks(baseLock, overrideLock *cfg.Lock) *cfg.Lock {
	lock := overrideLock.Clone()
	if baseLock == nil {
		return lock
	}

	// Merge and dedupe subpackages
	subpackages := make([]string, 0, len(lock.Subpackages)+len(baseLock.Subpackages))
	for _, sp := range lock.Subpackages {
		subpackages = append(subpackages, sp)
	}
	for _, sp := range baseLock.Subpackages {
		subpackages = append(subpackages, sp)
	}

	sort.Stable(sort.StringSlice(subpackages))

	dedupeSubpackages := make([]string, 0, len(subpackages))

	lastSp := ""
	elided := 0
	for _, sp := range subpackages {
		if lastSp == sp {
			elided++
		} else {
			dedupeSubpackages = append(dedupeSubpackages, sp)
			lastSp = sp
		}
	}
	lock.Subpackages = dedupeSubpackages[:len(dedupeSubpackages)-elided]
	return lock
}

func pinnedDependencyFromLock(lock *cfg.Lock) *cfg.Dependency {
	dep := cfg.DependencyFromLock(lock)
	dep.Pin = lock.Version
	return dep
}
