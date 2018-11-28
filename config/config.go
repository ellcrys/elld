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

// DataDir manages the client's data directory
type DataDir struct {
	path string
}

// NewDataDir creates a new DataDir object
func NewDataDir(dirPath, network string) (dd *DataDir, err error) {

	// check if dirPath exists
	if len(strings.TrimSpace(dirPath)) > 0 && !util.IsPathOk(dirPath) {
		return nil, fmt.Errorf("config directory is not ok; may not exist or we " +
			"don't have enough permission")
	}

	dd = new(DataDir)
	dd.path = dirPath

	// set default config directory if not provided and attempt to create it
	if len(dd.path) == 0 {
		hd, _ := homedir.Dir()
		dd.path = path.Join(hd, ".ellcrys", network)
		os.MkdirAll(dd.path, 0700)
	}

	return
}

// Path returns the config path
func (d *DataDir) Path() string {
	return d.path
}

// creates the config (ellcrys.json) file if it does not exist.
// Returns true and nil if config file already exists, false and nil
// if config file did not exist and was created. Otherwise, returns false and error
func (d *DataDir) createConfigFileInNotExist() (bool, error) {

	cfgFile := path.Join(d.path, "ellcrys.json")

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
func (d *DataDir) Init() error {

	var err error
	if _, err = d.createConfigFileInNotExist(); err != nil {
		return err
	}

	if fullAccountDir := path.Join(d.path, AccountDirName); !util.IsPathOk(fullAccountDir) {
		os.MkdirAll(fullAccountDir, 0700)
	}

	return nil
}

// Load reads the content of the ellcrys.json file into Config struct
func (d *DataDir) Load() (*EngineConfig, error) {
	var cfg EngineConfig
	cfgr := configor.New(&configor.Config{ENVPrefix: "ELLD"})
	if err := cfgr.Load(&cfg, path.Join(d.path, "ellcrys.json")); err != nil {
		return nil, fmt.Errorf("failed to parse config file -> %s", err)
	}
	return &cfg, nil
}

// LoadDataDir loads a data directory and returns
// the engine configuration
func LoadDataDir(dataDirPath, network string) (*EngineConfig, error) {

	dataDir, err := NewDataDir(dataDirPath, network)
	if err != nil {
		return nil, err
	}

	if err := dataDir.Init(); err != nil {
		if err != nil {
			return nil, err
		}

		return nil, err
	}

	dd, err := dataDir.Load()
	if err != nil {
		return nil, err
	}

	dd.VersionInfo = new(VersionInfo)

	if err := mergo.Merge(dd, defaultConfig); err != nil {
		return nil, err
	}

	dd.SetConfigDir(dataDir.Path())

	return dd, nil
}
