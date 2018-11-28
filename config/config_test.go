package config

import (
	"fmt"
	"io/ioutil"
	"os"
	path "path/filepath"
	"testing"

	"github.com/mitchellh/go-homedir"

	"github.com/ellcrys/elld/util"
	. "github.com/ncodes/goblin"
	. "github.com/onsi/gomega"
)

func TestConfigDir(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("ConfigDir", func() {
		var homeDir string

		g.BeforeEach(func() {
			homeDir, _ = homedir.Dir()
		})

		g.Describe(".NewDataDir", func() {
			g.It("should return error when the passed in directory does not exist", func() {
				_, err := NewDataDir("~/not_existing", "")
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(Equal("config directory is not ok; may not exist or we don't have enough permission"))
			})

			g.It("should return nil if the passed in directory exists", func() {
				dirName := fmt.Sprintf("%s/%s", homeDir, util.RandString(10))
				os.RemoveAll(dirName)
				err := os.Mkdir(dirName, 0700)
				Expect(err).To(BeNil())
				defer os.RemoveAll(dirName)

				_, err = NewDataDir(dirName, "")
				Expect(err).To(BeNil())
			})

			g.It("should return the default directory path", func() {
				dirName := fmt.Sprintf("%s/.ellcrys", homeDir)
				dataDir, err := NewDataDir("", "")
				Expect(err).To(BeNil())
				Expect(dataDir.path).To(Equal(dirName))
			})
		})

		g.Describe(".createConfigFileInNotExist", func() {
			g.When("no config file is found", func() {
				g.It("should create a new one with default content", func() {
					dirName := fmt.Sprintf("%s/%s", homeDir, util.RandString(10))
					os.RemoveAll(dirName)
					err := os.Mkdir(dirName, 0700)
					Expect(err).To(BeNil())
					defer os.RemoveAll(dirName)

					dd, err := NewDataDir(dirName, "")
					Expect(err).To(BeNil())
					existed, err := dd.createConfigFileInNotExist()
					Expect(err).To(BeNil())
					Expect(existed).To(BeFalse())

					dataDirPath := path.Join(dd.path, "ellcrys.json")
					d, err := ioutil.ReadFile(dataDirPath)
					Expect(err).To(BeNil())
					Expect(len(d)).NotTo(BeZero())
				})
			})

			g.When("config file already exist", func() {
				g.It("should return true and nil if config file already exist", func() {
					dirName := fmt.Sprintf("%s/%s", homeDir, util.RandString(10))
					os.RemoveAll(dirName)
					err := os.Mkdir(dirName, 0700)
					Expect(err).To(BeNil())
					defer os.RemoveAll(dirName)

					dd, err := NewDataDir(dirName, "")
					Expect(err).To(BeNil())
					existed, err := dd.createConfigFileInNotExist()
					Expect(err).To(BeNil())
					Expect(existed).To(BeFalse())

					existed, err = dd.createConfigFileInNotExist()
					Expect(err).To(BeNil())
					Expect(existed).To(BeTrue())
				})
			})
		})

		g.Describe(".Load", func() {
			g.When("config file content is not malformed", func() {
				g.It("should scan configuration into Config object", func() {
					dirName := fmt.Sprintf("%s/%s", homeDir, util.RandString(10))
					os.RemoveAll(dirName)
					err := os.Mkdir(dirName, 0700)
					Expect(err).To(BeNil())
					defer os.RemoveAll(dirName)

					defaultConfig = EngineConfig{
						Node: &PeerConfig{
							BootstrapAddresses: []string{"127.0.0.1:4000"},
						},
					}

					dataDir, err := NewDataDir(dirName, "")
					Expect(err).To(BeNil())
					existed, err := dataDir.createConfigFileInNotExist()
					Expect(err).To(BeNil())
					Expect(existed).To(BeFalse())

					dd, err := dataDir.Load()
					Expect(err).To(BeNil())
					Expect(dd).To(Equal(&EngineConfig{
						Node: &PeerConfig{
							BootstrapAddresses: []string{"127.0.0.1:4000"},
						},
					}))

					defaultConfig.Node = &PeerConfig{}
				})
			})

			g.When("config file content is malformed", func() {

				g.It("should return error", func() {
					dirName := fmt.Sprintf("%s/%s", homeDir, util.RandString(10))
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

					dataDir, err := NewDataDir(dirName, "")
					Expect(err).To(BeNil())
					_, err = dataDir.Load()
					Expect(err).NotTo(BeNil())
					Expect(err.Error()).To(Equal("failed to parse config file -> invalid character 'i' looking for beginning of value"))
				})
			})
		})
	})
}
