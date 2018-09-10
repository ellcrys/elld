/**
This package provides functionalities for packaging a directory
into a Blockcode. It includes functions for signing, verifying,
encoding and decoding a Blockcode.
*/
package blockcode

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/fatih/structs"
	"github.com/k0kubun/pp"

	"github.com/thoas/go-funk"
	"github.com/vmihailenco/msgpack"

	"gopkg.in/asaskevich/govalidator.v4"

	"github.com/mholt/archiver"

	"github.com/ellcrys/elld/util"
)

// Lang represents a blockcode language
type Lang string

var (
	// LangGo is a Lang representing `Go` programming language
	LangGo Lang = "go"
	// LangTypescript is a Lang representing `Typescript` programming language
	LangTypescript Lang = "typescript"
	// LangPython is a Lang representing `Python` programming language
	LangPython Lang = "python"
)

// Blockcode defines a block code
type Blockcode struct {
	code     []byte
	Manifest *Manifest
}

// Manifest describes the blockcode
type Manifest struct {
	Lang        Lang     `msgpack:"lang" json:"lang"`
	LangVersion string   `msgpack:"langVer" json:"langVer"`
	PublicFuncs []string `msgpack:"publicFuncs" json:"publicFuncs"`
}

// Bytes returns the bytes representation of the manifest
func (m *Manifest) Bytes() []byte {
	return util.ObjectToBytes(m)
}

// Size returns the bytecode size
func (bc *Blockcode) Size() int {
	return len(bc.code)
}

// Bytes return bytes representation of the Blockcode
func (bc *Blockcode) Bytes() []byte {
	return util.ObjectToBytes([]interface{}{
		bc.code,
		bc.Manifest.Bytes(),
	})
}

// GetCode returns the code in its tar archived form
func (bc *Blockcode) GetCode() []byte {
	return bc.code
}

// Hash returns the SHA256 hash of the blockcode
func (bc *Blockcode) Hash() util.Hash {
	bs := bc.Bytes()
	hash := sha256.Sum256(bs)
	return util.BytesToHash(hash[:])
}

// ID returns the hex representation of Hash()
func (bc *Blockcode) ID() string {
	return bc.Hash().HexStr()
}

// Read un-tars the code into destination
func (bc *Blockcode) Read(destination string) error {

	if !util.IsPathOk(destination) {
		return fmt.Errorf("destination path does not exist")
	}

	buff := bytes.NewBuffer(bc.code)
	return archiver.Tar.Read(buff, destination)
}

// FromBytes creates a Blockcode from serialized Blockcode
func FromBytes(bs []byte) (*Blockcode, error) {

	var data []interface{}
	if err := msgpack.Unmarshal(bs, &data); err != nil {
		return nil, err
	}

	var code = data[0]
	var manifest Manifest

	// decode the manifest
	if err := msgpack.Unmarshal(data[1].([]byte), &manifest); err != nil {
		return nil, err
	}

	return &Blockcode{
		code:     code.([]byte),
		Manifest: &manifest,
	}, nil
}

// FromDir creates a Blockcode from the content of a directory
func FromDir(projectPath string) (*Blockcode, error) {

	var manifest Manifest
	var filePaths []string

	// Check whether the project path exists
	if !util.IsPathOk(projectPath) {
		return nil, fmt.Errorf("project path does not exist")
	}

	// Read the files and directories
	// of the project
	fileInfos, err := ioutil.ReadDir(projectPath)
	if err != nil {
		return nil, err
	}

	for _, f := range fileInfos {

		filePaths = append(filePaths, filepath.Join(projectPath, f.Name()))
		if f.Name() != "package.json" {
			continue
		}

		// Read the manifest and decode it
		packageJSON, err := ioutil.ReadFile(filepath.Join(projectPath, f.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read manifest. %s", err)
		}
		if err = json.Unmarshal(packageJSON, &manifest); err != nil {
			return nil, fmt.Errorf("manifest is malformed. %s", err)
		}
	}

	if structs.IsZero(manifest) {
		return nil, fmt.Errorf("'package.json' file not found in {%s}", projectPath)
	}

	if err := validateManifest(&manifest); err != nil {
		return nil, err
	}
	pp.Println(filePaths, manifest)

	var buf = bytes.NewBuffer(nil)
	if err := archiver.Tar.Write(buf, filePaths); err != nil {
		return nil, fmt.Errorf("failed to create archive. %s", err)
	}

	bc := new(Blockcode)
	bc.code = buf.Bytes()
	bc.Manifest = &manifest

	return bc, nil
}

// validateManifest
func validateManifest(m *Manifest) error {

	if govalidator.IsNull(string(m.Lang)) {
		return fmt.Errorf("manifest error: language is missing")
	}

	if !funk.Contains([]Lang{LangGo, LangTypescript, LangPython}, Lang(m.Lang)) {
		return fmt.Errorf("manifest error: language {%s} is not supported", m.Lang)
	}

	if govalidator.IsNull(m.LangVersion) {
		return fmt.Errorf("manifest error: language version is required")
	}

	pubFuncs := []string{}
	for _, f := range m.PublicFuncs {
		if f = strings.TrimSpace(f); f != "" {
			pubFuncs = append(pubFuncs, f)
		}
	}

	if len(pubFuncs) == 0 {
		return fmt.Errorf("manifest error: at least one public function is required")
	}

	return nil
}
