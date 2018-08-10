package util

import "go/build"

// DirToImport converts a directory path on the local machine to a
// Go import path (usually relative to the $GOPATH/src directory)
//
// For example,
//     /Users/user/workspace/Go/github.com/mokiat/gostub
// will be converted to
//     github.com/mokiat/gostub
// should GOPATH include the location
//     /Users/user/workspace/Go
func DirToImport(p string) (string, error) {
	pkg, err := build.ImportDir(p, build.FindOnly)
	if err != nil {
		return "", err
	}
	return pkg.ImportPath, nil
}

// ImportToDir converts an import location to a directory path on
// the local machine.
//
// For example,
//     github.com/mokiat/gostub
// will be converted to
//     /Users/user/workspace/Go/github.com/mokiat/gostub
// should GOPATH be equal to
//     /Users/user/workspace/Go
func ImportToDir(imp string) (string, error) {
	pkg, err := build.Import(imp, "", build.FindOnly)
	if err != nil {
		return "", err
	}
	return pkg.Dir, nil
}
