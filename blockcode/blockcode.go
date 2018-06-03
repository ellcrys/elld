/**
This package provides functionalities for packaging a directory
into a Blockcode. It includes functions for signing, verifying,
encoding and decoding a Blockcode.
*/
package blockcode

import (
	"bytes"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/thoas/go-funk"

	"gopkg.in/asaskevich/govalidator.v4"

	"github.com/mholt/archiver"

	"github.com/ellcrys/druid/util"
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

type asn1Blockcode struct {
	Code     []byte
	Manifest []byte
}

// Manifest describes the blockcode
type Manifest struct {
	Lang        Lang     `json:"lang"`
	LangVersion string   `json:"langVer"`
	PublicFuncs []string `json:"publicFuncs"`
}

// Len returns the bytecode size
func (bc *Blockcode) Len() int {
	return len(bc.code)
}

// Bytes return the ASN.1 marshalled representation of the Blockcode
func (bc *Blockcode) Bytes() []byte {

	manifestBS, _ := json.Marshal(bc.Manifest)
	result, err := asn1.Marshal(asn1Blockcode{
		bc.code,
		manifestBS,
	})
	if err != nil {
		panic(err)
	}
	return result
}

// GetCode returns the code in its tar archived form
func (bc *Blockcode) GetCode() []byte {
	return bc.code
}

// Hash returns the SHA256 hash of the blockcode
func (bc *Blockcode) Hash() []byte {
	bs := bc.Bytes()
	hash := sha256.Sum256(bs)
	return hash[:]
}

// ID returns the hex representation of Hash()
func (bc *Blockcode) ID() string {
	return hex.EncodeToString(bc.Hash())
}

// Read un-tars the code into destination
func (bc *Blockcode) Read(destination string) error {

	if !util.IsPathOk(destination) {
		return fmt.Errorf("destination path does not exist")
	}

	buff := bytes.NewBuffer(bc.code)
	return archiver.Tar.Read(buff, destination)
}

// FromBytes creates a Blockcode from a slice of bytes produced by Blockcode.Bytes
func FromBytes(bs []byte) (*Blockcode, error) {

	var asn1Bc asn1Blockcode
	if _, err := asn1.Unmarshal(bs, &asn1Bc); err != nil {
		return nil, err
	}

	var manifest Manifest
	if err := json.Unmarshal(asn1Bc.Manifest, &manifest); err != nil {
		return nil, err
	}

	return &Blockcode{
		code:     asn1Bc.Code,
		Manifest: &manifest,
	}, nil
}

// FromDir creates a Blockcode from the content of a directory
func FromDir(projectPath string) (*Blockcode, error) {

	var manifest Manifest
	var manifestFileExists = false
	var filePaths []string

	if !util.IsPathOk(projectPath) {
		return nil, fmt.Errorf("project path does not exist")
	}

	fileInfos, err := ioutil.ReadDir(projectPath)
	if err != nil {
		return nil, err
	}

	for _, f := range fileInfos {
		filePaths = append(filePaths, filepath.Join(projectPath, f.Name()))

		if f.Name() == "package.json" {
			manifestFileExists = true
			packageJSON, err := ioutil.ReadFile(filepath.Join(projectPath, f.Name()))
			if err != nil {
				return nil, fmt.Errorf("failed to read manifest. %s", err)
			}
			if err = json.Unmarshal(packageJSON, &manifest); err != nil {
				return nil, fmt.Errorf("manifest is malformed. %s", err)
			}
		}
	}

	if !manifestFileExists {
		return nil, fmt.Errorf("'package.json' file not found in {%s}", projectPath)
	}

	if err := validateManifest(&manifest); err != nil {
		return nil, err
	}

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
