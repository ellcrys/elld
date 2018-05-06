package configdir

import (
	"fmt"
	"io/ioutil"
	"os"
	path "path/filepath"

	"github.com/mitchellh/go-homedir"

	"github.com/ellcrys/druid/testutil"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Configdir", func() {
	var homeDir string

	BeforeEach(func() {
		homeDir, _ = homedir.Dir()
	})

	Describe(".NewHomeDir", func() {
		It("should return error when the passed in directory does not exist", func() {
			_, err := NewConfigDir("~/not_existing")
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal("config directory is not ok; may not exist or we don't have enough permission"))
		})

		It("should return nil if the passed in directory exists", func() {
			dirName := fmt.Sprintf("%s/%s", homeDir, testutil.RandString(10))
			os.RemoveAll(dirName)
			err := os.Mkdir(dirName, 0700)
			Expect(err).To(BeNil())
			defer os.RemoveAll(dirName)

			_, err = NewConfigDir(dirName)
			Expect(err).To(BeNil())
		})

		It("should return the default directory path", func() {
			dirName := fmt.Sprintf("%s/.ellcrys", homeDir)
			cfgDir, err := NewConfigDir("")
			Expect(err).To(BeNil())
			Expect(cfgDir.path).To(Equal(dirName))
		})
	})

	Describe(".createConfigFileInNotExist", func() {
		When("no config file is found", func() {
			It("should create a new one with default content", func() {
				dirName := fmt.Sprintf("%s/%s", homeDir, testutil.RandString(10))
				os.RemoveAll(dirName)
				err := os.Mkdir(dirName, 0700)
				Expect(err).To(BeNil())
				defer os.RemoveAll(dirName)

				cfg, err := NewConfigDir(dirName)
				Expect(err).To(BeNil())
				existed, err := cfg.createConfigFileInNotExist()
				Expect(err).To(BeNil())
				Expect(existed).To(BeFalse())

				cfgDirPath := path.Join(cfg.path, "ellcrys.json")
				d, err := ioutil.ReadFile(cfgDirPath)
				Expect(err).To(BeNil())
				Expect(len(d)).NotTo(BeZero())
			})
		})

		When("config file already exist", func() {
			It("should return true and nil if config file already exist", func() {
				dirName := fmt.Sprintf("%s/%s", homeDir, testutil.RandString(10))
				os.RemoveAll(dirName)
				err := os.Mkdir(dirName, 0700)
				Expect(err).To(BeNil())
				defer os.RemoveAll(dirName)

				cfg, err := NewConfigDir(dirName)
				Expect(err).To(BeNil())
				existed, err := cfg.createConfigFileInNotExist()
				Expect(err).To(BeNil())
				Expect(existed).To(BeFalse())

				existed, err = cfg.createConfigFileInNotExist()
				Expect(err).To(BeNil())
				Expect(existed).To(BeTrue())
			})
		})
	})

	Describe(".Load", func() {
		When("config file content is not malformed", func() {
			It("should scan configuration into Config object", func() {
				dirName := fmt.Sprintf("%s/%s", homeDir, testutil.RandString(10))
				os.RemoveAll(dirName)
				err := os.Mkdir(dirName, 0700)
				Expect(err).To(BeNil())
				defer os.RemoveAll(dirName)

				defaultConfig = Config{
					Node: &PeerConfig{
						BootstrapNodes: []string{"127.0.0.1:4000"},
					},
				}

				cfgDir, err := NewConfigDir(dirName)
				Expect(err).To(BeNil())
				existed, err := cfgDir.createConfigFileInNotExist()
				Expect(err).To(BeNil())
				Expect(existed).To(BeFalse())

				cfg, err := cfgDir.Load()
				Expect(err).To(BeNil())
				Expect(cfg).To(Equal(&Config{
					Node: &PeerConfig{
						BootstrapNodes: []string{"127.0.0.1:4000"},
					},
				}))

				defaultConfig.Node = &PeerConfig{}
			})
		})

		When("config file content is malformed", func() {

			It("should return error", func() {
				dirName := fmt.Sprintf("%s/%s", homeDir, testutil.RandString(10))
				os.RemoveAll(dirName)
				err := os.Mkdir(dirName, 0700)
				Expect(err).To(BeNil())
				defer os.RemoveAll(dirName)

				filePath := fmt.Sprintf("%s/ellcrys.json", dirName)
				f, err := os.Create(filePath)
				Expect(err).To(BeNil())
				defer f.Close()
				err = ioutil.WriteFile(filePath, []byte("invalid content"), 0600)
				Expect(err).To(BeNil())

				cfgDir, err := NewConfigDir(dirName)
				Expect(err).To(BeNil())
				_, err = cfgDir.Load()
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(Equal("failed to parse config file -> invalid character 'i' looking for beginning of value"))
			})
		})
	})
})
