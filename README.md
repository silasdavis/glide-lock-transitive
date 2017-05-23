# glide-lock-transitive
glide-lock-transitive is a command line tool for locking transitive dependencies
with glide (https://github.com/Masterminds/glide). It can be used standalone 
or as a 'glide plugin'.

## Rationale
When adding or updating a dependency with glide glide is able to fetch transitive
dependencies (that is dependencies of dependencies). It will then lock a version
of any transitive dependencies into the glide.lock lock file. By doing this a 
complete set of dependencies are pinned to specific versions and can be 
reproducibly installed with `glide install`.

Glide is able to determine transitive dependency versions by looking at various 
files. For dependencies using glide itself it looks at the glide.yaml manifest 
file. The point of the glide manifest file is that it allows to specify versions
fuzzily; as tag ranges, branches, or not at all (meaning master). This means that
when you add or update a direct dependency with glide glide will fetch the latest
versions it can under the constraints of the dependency's manifest. It _will not_ 
use the versions your dependency has specified by its lock file. However your
dependency will have closed into its lock file the versions of its dependencies
it has been tested with. I almost always want to get the exact versions of my
transitive dependencies that were specified in the lock file rather than tacitly 

glide-lock-transitive is a tool to pull in a dependency and its transitive 
dependencies while respecting the explicit (commit hash) versions of those
transitive dependencies contained in the glide.lock lock file.

It works by fetching the direct dependency and merging its lock file into
the project's lock file.

## Installation

`go get -u github.com/silasdavis/glide-lock-transitive`

## Usage
This can be used as a glide plugin by running `glide lock-transitive`, though
this is identical in effect to running `glide-lock-transitive`.

Typical usage is:
 
```shell
# To get master
glide lock-transitive get github.com/foo/bar

# By version tag (when tagged as vX.Y.Z)
glide lock-transitive get github.com/foo/bar@0.2.4

# Or by commit hash
glide lock-transitive get github.com/foo/bar@47f5e645123faf25dd8bd109a0f17f77da2892cc

# Or by branch
glide lock-transitive get github.com/foo/bar@develop

# To get a specific subpackage
glide lock-transitive get github.com/foo/bar/subpackage@0.2.4
```
 
All of which will add the foo/bar dependency to the glide.yaml at the version 
specified and merge in its lock file dependencies into your project's lock file.

There is also a command for merging glide.lock files together manually, which
can be helpful for composing a lock file from a cascade of overrides.


```shell
glide-lock-transitive is a plugin for pinning transitive dependencies

Usage:
  glide lock-transitive [flags]
  glide [command]

Available Commands:
  get         Gets a remote dependency to this project along with its transitive dependencies.
  help        Help about any command
  lock-merge  Merge glide.lock files together

Flags:
  -h, --help   help for glide

Use "glide [command] --help" for more information about a command.

```
