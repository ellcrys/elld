// Package blockcode provides functionalities for packaging a directory
// into a Blockcode. It includes functions for signing, verifying,
// encoding and decoding a Blockcode.
package blockcode

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/structs"
	"github.com/thoas/go-funk"
	"github.com/vmihailenco/msgpack"

	"gopkg.in/asaskevich/govalidator.v4"

	"github.com/ellcrys/elld/util"
)

// Lang represents a blockcode language
type Lang string

// FileInfo represents a file in a directory
type FileInfo struct {
	Content []byte `json:"content"`
	Path    string `json:"path"`
}

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
	return util.BytesToHash(util.Blake2b256(hash[:]))
}

// ID returns the hex representation of Hash()
func (bc *Blockcode) ID() string {
	return bc.Hash().HexStr()
}

// Deflate decompresses the block code into
// destination while maintaining the folder
// structure it had during compression.
// It will return error if a directory of
// a file already exists or if unable to
// create the directory.
func (bc *Blockcode) Deflate(dest string) error {

	// if !util.IsPathOk(destination) {
	// 	return fmt.Errorf("destination path does not exist")
	// }

	// buff := bytes.NewBuffer(bc.code)
	// return archiver.Tar.Read(buff, destination)
	return nil
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

	// Check whether the project path exists
	if !util.IsPathOk(projectPath) {
		return nil, fmt.Errorf("project path does not exist")
	}

	// Traverse the project path in lexical order.
	// Collect the content of each file.
	var filesData []FileInfo
	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {

		if info.IsDir() || err != nil {
			return err
		}

		bs, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		filesData = append(filesData, FileInfo{
			Content: bs,
			Path:    path,
		})

		// process package.json
		if strings.Index(strings.ToLower(info.Name()), "package.json") != -1 {
			jsonDec := json.NewDecoder(bytes.NewBuffer(bs))
			if err = jsonDec.Decode(&manifest); err != nil {
				return fmt.Errorf("failed to decode manifest: %s", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if structs.IsZero(manifest) {
		return nil, fmt.Errorf("'package.json' file not found in {%s}", projectPath)
	}

	// convert the files data to bytes
	dataToCompress := util.ObjectToBytes(filesData)

	// Compress the file data
	var buf = bytes.NewBuffer(nil)
	w := zlib.NewWriter(buf)
	w.Write(dataToCompress)
	w.Flush()

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
