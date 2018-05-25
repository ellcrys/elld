package configdir

import (
	"encoding/json"
	"fmt"
	"os"
	path "path/filepath"
	"strings"

	"github.com/ellcrys/druid/util"
	"github.com/jinzhu/configor"

	"github.com/mitchellh/go-homedir"
)

// AccountDirName is the name of the directory for storing accounts
var AccountDirName = "accounts"

// ConfigDir represents the clients configuration director and
// provides methods for creating, accessing and manipulating its content
type ConfigDir struct {
	path string
}

// NewConfigDir creates a new ConfigDir object
func NewConfigDir(dirPath string) (cfgDir *ConfigDir, err error) {

	// check if dirPath exists
	if len(strings.TrimSpace(dirPath)) > 0 && !util.IsPathOk(dirPath) {
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

	if util.IsPathOk(cfgFile) {
		return true, nil
	}

	cfg, err := os.Create(cfgFile)
	if err != nil {
		return false, fmt.Errorf("failed to create config file at config directory")
	}
	defer cfg.Close()

	jsonEnc := json.NewEncoder(cfg)
	jsonEnc.SetIndent("", "\t")
	if err := jsonEnc.Encode(defaultConfig); err != nil {
		return false, fmt.Errorf("failed to encode default config -> %s", err)
	}

	return false, nil
}

// Init creates required files and directories
func (cd *ConfigDir) Init() error {

	var err error
	if _, err = cd.createConfigFileInNotExist(); err != nil {
		return err
	}

	if fullAccountDir := path.Join(cd.path, AccountDirName); !util.IsPathOk(fullAccountDir) {
		os.MkdirAll(fullAccountDir, 0700)
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

// LoadCfg loads the config file
func LoadCfg(cfgDirPath string) (*Config, error) {

	cfgDir, err := NewConfigDir(cfgDirPath)
	if err != nil {
		return nil, err
	}

	if err := cfgDir.Init(); err != nil {
		if err != nil {
			return nil, err
		}

		return nil, err
	}

	cfg, err := cfgDir.Load()
	if err != nil {
		return nil, err
	}

	cfg.SetConfigDir(cfgDir.Path())

	return cfg, nil
}
