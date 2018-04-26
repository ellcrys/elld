package configdir

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/jinzhu/configor"

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
	if len(strings.TrimSpace(dirPath)) > 0 && !isPathOk(dirPath) {
		return nil, fmt.Errorf("config directory is not ok; may not exist or we don't have enough permission")
	}

	cfgDir = new(ConfigDir)
	cfgDir.path = dirPath

	// set default config directory if not provided and attempt to create it
	if len(cfgDir.path) == 0 {
		hd, _ := homedir.Dir()
		cfgDir.path = fmt.Sprintf("%s/.ellcrys", hd)
		os.Mkdir(cfgDir.path, 0700)
	}

	return
}

// Path returns the config path
func (cd *ConfigDir) Path() string {
	return cd.path
}

// creates the config (ellcrys.json) file if it does not exist.
// Returns true and nil if config file already exists, false and nil
// if config file did not exist and was created. Otherwise, returns false and error
func (cd *ConfigDir) createConfigFileInNotExist() (bool, error) {

	cfgFile := path.Join(cd.path, "ellcrys.json")

	if isPathOk(cfgFile) {
		return true, nil
	}

	cfg, err := os.Create(cfgFile)
	if err != nil {
		return false, fmt.Errorf("failed to create config file at config directory")
	}
	defer cfg.Close()

	if err := json.NewEncoder(cfg).Encode(defaultConfig); err != nil {
		return false, fmt.Errorf("failed to encode default config -> %s", err)
	}

	return false, nil
}

// Init creates the ~./ellcrys directory if it does not exists.
// It will include necessary config files, database if they are missing.
func (cd *ConfigDir) Init() error {
	var err error
	if _, err = cd.createConfigFileInNotExist(); err != nil {
		return err
	}
	return nil
}

// Load reads the content of the ellcrys.json file into Config struct
func (cd *ConfigDir) Load() (*Config, error) {
	var cfg Config
	if err := configor.Load(&cfg, path.Join(cd.path, "ellcrys.json")); err != nil {
		return nil, fmt.Errorf("failed to parse config file -> %s", err)
	}
	return &cfg, nil
}

// isPathOk checks if a directory exist and whether
// there are no permission errors
func isPathOk(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}
