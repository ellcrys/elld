package config

import (
	"encoding/json"
	"fmt"
	"os"
	path "path/filepath"
	"strings"

	"github.com/imdario/mergo"

	"github.com/ellcrys/elld/util"
	"github.com/jinzhu/configor"

	"github.com/mitchellh/go-homedir"
)

// AccountDirName is the name of the directory for storing accounts
var AccountDirName = "accounts"

// Config holds configuration information
type Config struct {
	path string
}

// NewConfig creates a new Config object
func NewConfig(dirPath string) (cfg *Config, err error) {

	// check if dirPath exists
	if len(strings.TrimSpace(dirPath)) > 0 && !util.IsPathOk(dirPath) {
		return nil, fmt.Errorf("config directory is not ok; may not exist or we don't have enough permission")
	}

	cfg = new(Config)
	cfg.path = dirPath

	// set default config directory if not provided and attempt to create it
	if len(cfg.path) == 0 {
		hd, _ := homedir.Dir()
		cfg.path = fmt.Sprintf("%s/.ellcrys", hd)
		os.Mkdir(cfg.path, 0700)
	}

	return
}

// Path returns the config path
func (cd *Config) Path() string {
	return cd.path
}

// creates the config (ellcrys.json) file if it does not exist.
// Returns true and nil if config file already exists, false and nil
// if config file did not exist and was created. Otherwise, returns false and error
func (cd *Config) createConfigFileInNotExist() (bool, error) {

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
func (cd *Config) Init() error {

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
func (cd *Config) Load() (*EngineConfig, error) {
	var cfg EngineConfig
	cfgr := configor.New(&configor.Config{ENVPrefix: "ELLD"})
	if err := cfgr.Load(&cfg, path.Join(cd.path, "ellcrys.json")); err != nil {
		return nil, fmt.Errorf("failed to parse config file -> %s", err)
	}
	return &cfg, nil
}

// LoadCfg loads the config file
func LoadCfg(dataDirPath string) (*EngineConfig, error) {

	dataDir, err := NewConfig(dataDirPath)
	if err != nil {
		return nil, err
	}

	if err := dataDir.Init(); err != nil {
		if err != nil {
			return nil, err
		}

		return nil, err
	}

	cfg, err := dataDir.Load()
	if err != nil {
		return nil, err
	}

	cfg.VersionInfo = new(VersionInfo)

	if err := mergo.Merge(cfg, defaultConfig); err != nil {
		return nil, err
	}

	cfg.SetConfigDir(dataDir.Path())

	return cfg, nil
}
