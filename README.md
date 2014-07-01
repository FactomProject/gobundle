# gobundle

###### v1.1

Provides a tool &amp; API for bundling resources with Go

Gobundle provides a command for packing resources into a go executable.
Given a list of packages, for each package, it finds all of the files in
`<pkg dir>/bundle/`, tars and gzips them, writing the resulting data
to a byte array in `<pkg dir>/bundle.go`. The available flags are:

  * --file: if present, the data will be written to the argument as a tar. The resulting file will be an
actual tar, as opposed to the golang byte array literal that is written to
bundle.go. This is primarily for testing purposes.

To use the gobundle package, `import "github.com/firelizzard18/gobundle"`,
set `gobundle.Setup.Application.Name` to your application name, and call
`gobundle.Init()` (which calls `flag.Parse()`). The available flags are:

  * --conf: sets the configuration directory path; default value depends on the OS
  * --data: sets the data directory path; default value depends on the OS
  * --unpack &lt;arg&gt;:
    * suppress: Don't unpack
    * unpack[,force]: Do unpack
    * detect[,force]: Unpack if the conf and data are non-extant
    * only[,force]: Unpack and exit
    * `force` forces unpack, overwriting extant files

## Change log

#### v1.1

Added --file for the command.

#### v1.0

Initial Release
