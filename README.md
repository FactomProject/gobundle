# gobundle

v1.1

Provides a tool &amp; API for bundling resources with Go

## Change log

### v1.1

Added --file for the command. If set to a non empty value, this flag will
redirect output from bundle.go to a tar file. The resulting file will be an
actual tar, as opposed to the golang byte array literal that is written to
bundle.go. This is primarily for testing purposes.

### v1.0

Initial Release
