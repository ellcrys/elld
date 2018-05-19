package testutil

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	mrand "math/rand"
	"os"
	"os/exec"
	"time"

	"github.com/apex/log"
	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-host"
	net "github.com/libp2p/go-libp2p-net"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"
)

// NoOpStreamHandler accepts a stream and does nothing
var NoOpStreamHandler = func(s net.Stream) {}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func init() {
	mrand.Seed(time.Now().UnixNano())
}

// GenerateKeyPair generates private and public keys
func GenerateKeyPair(r io.Reader) (crypto.PrivKey, crypto.PubKey, error) {
	return crypto.GenerateEd25519Key(r)
}

// RandomHost creates a host with random identity
func RandomHost(seed int64, port int) (host.Host, error) {

	priv, _, err := GenerateKeyPair(mrand.New(mrand.NewSource(seed)))
	if seed == 0 {
		priv, _, _ = GenerateKeyPair(rand.Reader)
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port)),
		libp2p.Identity(priv),
	}

	host, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create host")
	}

	return host, nil
}

// RandString is like RandBytes but returns string
func RandString(n int) string {
	return string(RandBytes(n))
}

// RandBytes gets random string of fixed length
func RandBytes(n int) []byte {
	b := make([]byte, n)
	for i, cache, remain := n-1, mrand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = mrand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return b
}

func fileExists(path string) (bool, error) {
	fs := afero.NewOsFs()
	_, err := fs.Stat(path)
	if os.IsNotExist(err) {
		return false, errors.New(path + " does not exist")
	}

	return true, nil
}

func removeDir(path string) error {
	fs := afero.NewOsFs()

	err := fs.RemoveAll(path)

	if err != nil {
		log.Infof("could not remove directory %s", err)
		return err
	}

	return nil
}

func FetchTestContract() {
	var (
		contractURL     string
		gitCmd          *exec.Cmd
		testContractDir string
		ellcrysDir      string
	)

	contractURL = "https://github.com/ellcrys/smartcontracts-template-ts.git"
	homeDir, _ := homedir.Dir()
	ellcrysDir = fmt.Sprintf("%s/.ellcrys", homeDir)
	testContractDir = fmt.Sprintf("%s/test/test-contract", ellcrysDir)

	isFileExist, _ := fileExists(testContractDir)

	if isFileExist {
		removeDir(testContractDir)
	}

	gitCmd = exec.Command("git", "clone", contractURL, testContractDir)

	gitCmd.Start()

	err := gitCmd.Wait()
	if err != nil {
		panic(err)
	}

	npmCmd := exec.Command("npm", "install", "--prefix", testContractDir)

	npmCmd.Start()
	err = npmCmd.Wait()
	if err != nil {
		panic(err)
	}

}

func RemoveTestContractDir() {
	homeDir, _ := homedir.Dir()
	ellcrysDir := fmt.Sprintf("%s/.ellcrys", homeDir)
	testContractDir := fmt.Sprintf("%s/test/test-contract", ellcrysDir)

	isFileExist, _ := fileExists(testContractDir)

	if isFileExist {
		removeDir(testContractDir)
	}
}

//capture commandline outputs
func capture(w io.Writer, r io.Reader) ([]byte, error) {
	var out []byte
	buf := make([]byte, 1024, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			_, err := w.Write(d)
			if err != nil {
				return out, err
			}
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return out, err
		}
	}
}
