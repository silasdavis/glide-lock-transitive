# glide-lock-transitive
glide-lock-transitive is a command line tool for locking transitive dependencies
with glide (https://github.com/Masterminds/glide). It can be used standalone 
or as a 'glide plugin'.

## Rationale
When adding or updating a dependency with glide glide is able to fetch transitive
dependencies (that is dependencies of dependencies). It will then lock a version
of any transitive dependencies into the glide.lock lockfile. By doing this a 
complete set of dependencies are pinned to specific versions and can be 
reproducibly installed with `glide install`.

Glide is able to determine transitive dependency versions by looking at various 
files. For dependencies using glide itself it looks at the glide.yaml manifest 
file. The point of the glide manifest file is that it allows to specify versions
fuzzily; as tag ranges, branches, or not at all (meaning master). This means that
when you add or update a direct dependency with glide glide will fetch the latest
versions it can under the constraints of the dependency's manifest. It _will not_ 
use the versions your dependency has specified by its lockfile. However your
dependency will have closed into its lockfile the versions of its dependencies
it has been tested with. I almost always want to get the exact versions of my
transitive dependencies that were specified in the lockfile rather than tacitly 

glide-lock-transitive is a tool to pull in a dependency and its transitive 
dependencies while respecting the explicit (commit hash) versions of those
transitive dependencies contained in the glide.lock lockfile.

It works by fetching the direct dependency and merging its lockfile into
the project's lockfile.

## Installation

go get -u github.com/silasdavis/glide-lock-transitive

## Usage
This can be used as a glide plugin by running `glide lock-transitive`, though
this is identical in effect to running `glide-lock-transitive`.


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