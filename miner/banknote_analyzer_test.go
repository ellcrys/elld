package miner

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"

	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/util/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func delResources(bk *Analyzer) {
	os.Remove(bk.ellFilePath)
	os.RemoveAll(bk.mintDir)
}

var BanknoteAnalyzerTest = func() bool {
	return Describe("BanknoteAnalyzer", func() {

		BeforeEach(func() {
			var err error
			cfg, err = testutil.SetTestCfg()
			Expect(err).To(BeNil())
		})

		cfg, err = testutil.SetTestCfg()

		log := logger.NewLogrus()

		Describe(".NewAnalyzer", func() {

			Describe("NewAnalyzer with empty train path", func() {

				log := logger.NewLogrus()
				bk := NewAnalyzer("", cfg, log)

				delResources(bk)

				It("bankNote struct must not be  nil", func() {
					Expect(bk).ToNot(BeNil())
				})

			})

			Describe("NewAnalyzer with file path to train data", func() {

				bk := NewAnalyzer("testDirResources/sample_train_file.ell", cfg, log)

				delResources(bk)

				It("bankNote struct must not be  nil", func() {
					Expect(bk).ToNot(BeNil())
				})
			})

			Describe("NewAnalyzer with url path to train data", func() {

				bk := NewAnalyzer("http://192.168.4.103/train_file.ell", cfg, log)

				delResources(bk)

				It("bankNote struct must not be  nil", func() {
					Expect(bk).ToNot(BeNil())
				})

			})

		})

		Describe(".Prepare", func() {

			Context("When a 404 url path is supplied for train file", func() {

				bk := NewAnalyzer("http://192.168.4.103/x/image128.png", cfg, log)

				delResources(bk)

				It("Error must not be nil", func() {
					err = bk.Prepare()
					Expect(err).ToNot(BeNil())
					delResources(bk)
				})

			})

			Context("When empty train file is passed", func() {

				bk := NewAnalyzer("", cfg, log)

				delResources(bk)

				err = bk.Prepare()

				It("Error must be nil, becuase it uses the system train file", func() {
					Expect(err).To(BeNil())
					delResources(bk)
				})

			})

			Context("When valid path is supplied for train file", func() {

				bk := NewAnalyzer("miner/testDirResources/sample_train_file.ell", cfg, log)

				delResources(bk)
				err = bk.Prepare()

				It("Error must be nil", func() {
					Expect(err).To(BeNil())
					delResources(bk)
				})

			})

			Context("When Invalid path is supplied for train file", func() {

				bk := NewAnalyzer("miner/testDirResources/x/image128.png", cfg, log)

				delResources(bk)

				It("Error must not be nil", func() {
					err = bk.Prepare()
					Expect(err).ToNot(BeNil())
					delResources(bk)
				})

			})

			Context("When Invalid url path is supplied for train file", func() {

				bk := NewAnalyzer("http://192.168.4.103/image128.png", cfg, log)

				delResources(bk)

				It("Error must not be nil", func() {
					err = bk.Prepare()
					Expect(err).ToNot(BeNil())
					delResources(bk)
				})

			})

		})

		Describe(".downloadEllToPath", func() {

			bk := NewAnalyzer("", cfg, log)

			It("When url is not valid, Error Must not be nil", func() {
				err = bk.downloadEllToPath(bk.elldConfigDir, "xyzcommm")
				Expect(err).ToNot(BeNil())
			})

			It("When url is empty, Error Must not be nil", func() {
				err = bk.downloadEllToPath(bk.elldConfigDir, "")
				Expect(err).ToNot(BeNil())
			})

			It("When url is valid, Error Must be nil", func() {
				err = bk.downloadEllToPath(bk.elldConfigDir, bk.ellRemoteURL)
				Expect(err).To(BeNil())
				delResources(bk)
			})

		})

		Describe(".Predict", func() {

			Context("When a valid image is supplied to predict", func() {

				bk := NewAnalyzer("", cfg, log)
				imagePath := "testDirResources/image128.png"

				bk.Prepare()

				note, err := bk.Predict(imagePath)

				It("Expect only error to be nil", func() {

					Expect(note).ToNot(BeNil())
					Expect(err).To(BeNil())

					Expect(note.Name()).ToNot(BeNil())
					Expect(note.Country()).ToNot(BeNil())
					Expect(note.Figure()).ToNot(BeNil())
					Expect(note.Shortname()).ToNot(BeNil())
					Expect(note.Text()).ToNot(BeNil())

					Expect(note.Name()).ShouldNot(BeEmpty())
					Expect(note.Country()).ShouldNot(BeEmpty())
					Expect(note.Figure()).ShouldNot(BeEmpty())
					Expect(note.Shortname()).ShouldNot(BeEmpty())
					Expect(note.Text()).ShouldNot(BeEmpty())

				})

			})

			Context("When an invalid path to image is supplied ", func() {

				bk := NewAnalyzer("", cfg, log)
				imagePath := "testDirResources/x/image128.png"

				bk.Prepare()

				note, err := bk.Predict(imagePath)

				It("Expect error to not be nil", func() {
					Expect(note).To(BeNil())
					Expect(err).ToNot(BeNil())
				})

			})

			Context("When an invalid image is supplied ", func() {

				bk := NewAnalyzer("", cfg, log)
				imagePath := "testDirResources/sample_train_file.ell"

				bk.Prepare()

				note, err := bk.Predict(imagePath)

				It("Expect error to not be nil", func() {
					Expect(note).To(BeNil())
					Expect(err).ToNot(BeNil())
				})

			})

			Context("When the path to the train file is invalid ", func() {

				bk := NewAnalyzer("", cfg, log)
				imagePath := "testDirResources/sample_train_file.ell"

				bk.Prepare()

				bk.kerasGoPath = ""
				note, err := bk.Predict(imagePath)

				It("Expect error to not be nil", func() {
					Expect(note).To(BeNil())
					Expect(err).ToNot(BeNil())
				})

			})

		})

		Describe(".PredictBytes", func() {

			Context("When a valid byte is pass as input parameter ", func() {

				bk := NewAnalyzer("", cfg, log)

				It("Error Must be nil", func() {

					imagePath := "testDirResources/image128.png"
					byteImage, _ := ioutil.ReadFile(imagePath)
					note, err := bk.PredictBytes(byteImage)

					Expect(note).ToNot(BeNil())
					Expect(err).To(BeNil())

					Expect(note.Name()).ToNot(BeNil())
					Expect(note.Country()).ToNot(BeNil())
					Expect(note.Figure()).ToNot(BeNil())
					Expect(note.Shortname()).ToNot(BeNil())
					Expect(note.Text()).ToNot(BeNil())

					Expect(note.Name()).ShouldNot(BeEmpty())
					Expect(note.Country()).ShouldNot(BeEmpty())
					Expect(note.Figure()).ShouldNot(BeEmpty())
					Expect(note.Shortname()).ShouldNot(BeEmpty())
					Expect(note.Text()).ShouldNot(BeEmpty())

				})

			})

			Context("When an invalid byte is pass as input parameter ", func() {

				bk := NewAnalyzer("", cfg, log)

				It("Error Must not be nil", func() {

					bytedata := []byte("Here is a string image")

					note, err := bk.PredictBytes(bytedata)
					Expect(note).To(BeNil())
					Expect(err).ToNot(BeNil())

				})

			})

		})

		Describe(".mintLoader", func() {

			Context("When a valid data is pass as input parameter ", func() {

				It("Error Must be nil", func() {

					bk := NewAnalyzer("", cfg, log)

					imageFile, _ := os.Open("testDirResources/image128.png")

					var imgBuffer bytes.Buffer
					io.Copy(&imgBuffer, imageFile)
					tensor, _ := readImage(&imgBuffer, "png")

					bnote, err := bk.mintLoader(tensor)

					Expect(bnote).ToNot(BeNil())
					Expect(err).To(BeNil())

				})

			})

			Context("When an invalid train path is supplied", func() {

				It("Error Must not be nil", func() {

					bk := NewAnalyzer("", cfg, log)

					bk.kerasGoPath = ""

					imageFile, _ := os.Open("testDirResources/image128.png")

					var imgBuffer bytes.Buffer
					io.Copy(&imgBuffer, imageFile)
					tensor, _ := readImage(&imgBuffer, "png")

					bnote, err := bk.mintLoader(tensor)

					Expect(bnote).To(BeNil())
					Expect(err).ToNot(BeNil())

				})

			})

			Context("When an invalid mint path is supplied", func() {

				It("Error Must not be nil", func() {

					bk := NewAnalyzer("", cfg, log)

					bk.mintDir = ""

					imageFile, _ := os.Open("testDirResources/image128.png")

					var imgBuffer bytes.Buffer
					io.Copy(&imgBuffer, imageFile)
					tensor, _ := readImage(&imgBuffer, "png")

					bnote, err := bk.mintLoader(tensor)

					Expect(bnote).To(BeNil())
					Expect(err).ToNot(BeNil())

				})

			})

		})

		Describe(".unzip", func() {

			It("Error Must not be nil", func() {
				res := unzip("testDirResources/image128.png", cfg.ConfigDir())
				Expect(res).ToNot(BeNil())
			})

			It("Error Must not be nil", func() {
				res := unzip("testDirResources/sample_train_file.ell", "/Users/princesegzy01/elld_config_locked")
				Expect(res).ToNot(BeNil())
			})

			It("Error Must not be nil", func() {
				res := unzip("testDirResources/fake_train_file.ell", cfg.ConfigDir())
				Expect(res).ToNot(BeNil())
			})

			It("Error Must be nil", func() {
				res := unzip("testDirResources/sample_train_file.ell", cfg.ConfigDir())
				Expect(res).To(BeNil())
			})

		})

		Describe(".isValidUrl", func() {

			It("Must return true", func() {
				res := isValidUrl("http://www.example.com")
				Expect(res).To(BeTrue())
			})

			It("Must return false", func() {
				res := isValidUrl("document/golangcode/data")
				Expect(res).To(BeFalse())
			})

		})

		Describe(".readImage", func() {

			Context("When Invalid buffer image and png format is supplied", func() {

				var b bytes.Buffer
				b.Write([]byte("Hello"))

				res, err := readImage(&b, "png")

				It("should reject invalid buffer & png as input ", func() {
					Expect(res).To(BeNil())
					Expect(err).ToNot(BeNil())
				})
			})

			Context("When Invalid buffer image and ttif format is supplied", func() {

				var b bytes.Buffer
				b.Write([]byte("Hello"))

				res, err := readImage(&b, "ttif")

				It("should reject invalid buffer & ttif as input ", func() {
					Expect(res).To(BeNil())
					Expect(err).ToNot(BeNil())
				})
			})

			Context("When empty buffer and invalid format is supplied", func() {

				var b bytes.Buffer
				b.Write([]byte(""))

				res, err := readImage(&b, "ttif")

				It("should reject empty buffer & invalid format as input ", func() {
					Expect(res).To(BeNil())
					Expect(err).ToNot(BeNil())
				})
			})

			Context("When empty buffer and valid format is supplied", func() {

				var b bytes.Buffer
				b.Write([]byte(""))

				res, err := readImage(&b, "png")

				It("should reject empty buffer & valid format as input ", func() {
					Expect(res).To(BeNil())
					Expect(err).ToNot(BeNil())
				})
			})

			Context("When valid buffer and invalid format is supplied", func() {

				imByte, _ := ioutil.ReadFile("testDirResources/image128.png")

				var b bytes.Buffer
				b.Write([]byte(imByte))

				res, err := readImage(&b, "ttif")

				It("should reject valid buffer & invalid format as input ", func() {
					Expect(res).To(BeNil())
					Expect(err).ToNot(BeNil())
				})
			})

			Context("When valid buffer and valid format is supplied", func() {

				imByte, _ := ioutil.ReadFile("testDirResources/image128.png")

				var b bytes.Buffer
				b.Write([]byte(imByte))

				res, err := readImage(&b, "png")

				It("should accept valid buffer & valid format as input ", func() {
					Expect(res).ToNot(BeNil())
					Expect(err).To(BeNil())
				})
			})
		})

		Describe(".transformGraph", func() {

			Context("When png file is supplied", func() {

				graph, input, output, err := transformGraph("png")

				It("should accepts png as input ", func() {
					Expect(graph).ToNot(BeNil())
					Expect(input).ToNot(BeNil())
					Expect(output).ToNot(BeNil())
					Expect(err).To(BeNil())
				})
			})

			Context("When jpg file is supplied", func() {

				graph, input, output, err := transformGraph("jpg")

				It("should accepts jpg as input ", func() {
					Expect(graph).ToNot(BeNil())
					Expect(input).ToNot(BeNil())
					Expect(output).ToNot(BeNil())
					Expect(err).To(BeNil())
				})
			})

			Context("When jpeg file is supplied", func() {

				graph, input, output, err := transformGraph("jpeg")

				It("should accepts jpeg as input ", func() {
					Expect(graph).ToNot(BeNil())
					Expect(input).ToNot(BeNil())
					Expect(output).ToNot(BeNil())
					Expect(err).To(BeNil())
				})
			})

			Context("When gif file is supplied", func() {

				graph, input, output, err := transformGraph("gif")
				It("should Not accepts gif as input ", func() {
					Expect(graph).To(BeNil())
					Expect(input).ToNot(BeNil())
					Expect(output).ToNot(BeNil())
					Expect(err).ToNot(BeNil())
				})
			})

		})

	})
}
