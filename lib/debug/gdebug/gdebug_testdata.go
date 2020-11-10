package gdebug

import (
	"path/filepath"
)

// TestDataPath retrieves and returns the testdata path of current package,
// which is used for unit testing cases only.
// The optional parameter <names> specifies the its sub-folders/sub-files,
// which will be joined with current system separator and returned with the path.
func TestDataPath(names ...string) string {
	path := CallerDirectory() + string(filepath.Separator) + "testdata"
	for _, name := range names {
		path += string(filepath.Separator) + name
	}
	return path
}
