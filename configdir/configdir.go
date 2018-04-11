package configdir

import (
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
)

// ConfigDir represents the clients configuration director and
// provides methods for creating, accessing and manipulating its content
type ConfigDir struct {
	path string
}

// NewConfigDir creates a new ConfigDir object
func NewConfigDir(dirPath string) (cfgDir *ConfigDir, err error) {

	// check if dirPath exists
	if !dirOk(dirPath) {
		return nil, fmt.Errorf("config directory is not ok; may not exist or we don't have enough permission")
	}

	cfgDir = new(ConfigDir)
	cfgDir.path = dirPath

	// set default config directory if not provided
	if len(cfgDir.path) == 0 {
		hd, _ := homedir.Dir()
		cfgDir.path = fmt.Sprintf("%s/.ellcrys", hd)
	}

	return
}

// Init creates the ~./ellcrys directory if it does not exists.
// It will include necessary config files, database if they are missing.
func (h *ConfigDir) Init() {

}

// dirOk checks if a directory exist and whether
// there are no permission errors
func dirOk(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}
