package miner

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"

	config "github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/util/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Banknote", func() {

	var log = logger.NewLogrus()
	var cfg config.EngineConfig

	cfg.SetConfigDir("/Users/princesegzy01/cfg")

	Describe("NewAnalyzer", func() {

		Describe("NewAnalyzer with empty train path, elld should download it from ell remore url", func() {
			// var bankNote *Analyzer
			bankNote := NewAnalyzer("", &cfg, log)

			_ = os.Remove(bankNote.ellFilePath)
			_ = os.RemoveAll(bankNote.mintDir)

			It("Error must not be  nil", func() {
				Expect(bankNote).ToNot(BeNil())
			})
			It("Error must not  nil", func() {
				err := bankNote.Prepare()
				Expect(err).To(BeNil())
			})

		})

		Describe("NewAnalyzer with file path to train data", func() {

			// var bankNote *Analyzer
			bankNote := NewAnalyzer("testDirResources/sample_train_file.ell", &cfg, log)

			_ = os.Remove(bankNote.ellFilePath)
			_ = os.RemoveAll(bankNote.mintDir)

			It("Error must not be  nil", func() {
				Expect(bankNote).ToNot(BeNil())
			})
			It("Error must not  nil", func() {
				err := bankNote.Prepare()
				Expect(err).To(BeNil())
			})

		})

		Describe("NewAnalyzer with url path to train data", func() {

			bankNote := NewAnalyzer("http://192.168.4.103/train_file.ell", &cfg, log)

			_ = os.Remove(bankNote.ellFilePath)
			_ = os.RemoveAll(bankNote.mintDir)

			It("Error must not be  nil", func() {
				Expect(bankNote).ToNot(BeNil())
			})
			It("Error must not  nil", func() {
				err := bankNote.Prepare()
				Expect(err).To(BeNil())
			})

		})

		Describe("NewAnalyzer with empty train path, elld should use existing ell from config", func() {
			// var bankNote *Analyzer
			bankNote := NewAnalyzer("", &cfg, log)

			It("Error must not be  nil", func() {
				Expect(bankNote).ToNot(BeNil())
			})
			It("Error must not  nil", func() {
				err := bankNote.Prepare()
				Expect(err).To(BeNil())
			})

		})

	})
	Describe("downloadEllToPath", func() {
		bankNote := NewAnalyzer("", &cfg, log)
		bankNote.downloadEllToPath(bankNote.elldConfigDir, "xyz.commm")
	})

	Describe("Predict", func() {
		bankNote := NewAnalyzer("", &cfg, log)

		It("Error Must not be nil", func() {
			imagePath := "testDirResources/image128.png"
			notex, err := bankNote.Predict(imagePath)

			Expect(notex).ToNot(BeNil())
			Expect(err).To(BeNil())

			Expect(notex.Name()).ToNot(BeNil())
			Expect(notex.Country()).ToNot(BeNil())
			Expect(notex.Figure()).ToNot(BeNil())
			Expect(notex.Shortname()).ToNot(BeNil())
			Expect(notex.Text()).ToNot(BeNil())

			Expect(notex.Name()).ShouldNot(BeEmpty())
			Expect(notex.Country()).ShouldNot(BeEmpty())
			Expect(notex.Figure()).ShouldNot(BeEmpty())
			Expect(notex.Shortname()).ShouldNot(BeEmpty())
			Expect(notex.Text()).ShouldNot(BeEmpty())

		})

		It("When Invalid image is passed to predict method", func() {

			notex, err := bankNote.Predict("im")
			// fmt.Println("<<<<<<<<<<<<<<<<<<<", notex, err)
			Expect(notex).To(BeNil())
			Expect(err).ToNot(BeNil())

		})

	})

	Describe("PredictBytes", func() {
		bankNote := NewAnalyzer("", &cfg, log)

		It("Error Must not be nil", func() {
			imagePath := "testDirResources/image128.png"
			byteImage, _ := ioutil.ReadFile(imagePath)
			notex, err := bankNote.PredictBytes(byteImage)

			Expect(notex).ToNot(BeNil())
			Expect(err).To(BeNil())

			Expect(notex.Name()).ToNot(BeNil())
			Expect(notex.Country()).ToNot(BeNil())
			Expect(notex.Figure()).ToNot(BeNil())
			Expect(notex.Shortname()).ToNot(BeNil())
			Expect(notex.Text()).ToNot(BeNil())

			Expect(notex.Name()).ShouldNot(BeEmpty())
			Expect(notex.Country()).ShouldNot(BeEmpty())
			Expect(notex.Figure()).ShouldNot(BeEmpty())
			Expect(notex.Shortname()).ShouldNot(BeEmpty())
			Expect(notex.Text()).ShouldNot(BeEmpty())

		})

	})

	Describe("mintLoader", func() {

		It("Error Must not be nil", func() {
			cfg.SetConfigDir("")
			bankNote := NewAnalyzer("", &cfg, log)

			imageFile, _ := os.Open("testDirResources/image128.png")

			var imgBuffer bytes.Buffer
			io.Copy(&imgBuffer, imageFile)
			tensor, _ := readImage(&imgBuffer, "png")

			bn, erx := bankNote.mintLoader(tensor)

			Expect(bn).To(BeNil())
			Expect(erx).ToNot(BeNil())

		})

	})

	Describe("unzip", func() {

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

	Describe("isValidUrl", func() {

		It("Must return true", func() {
			res := isValidUrl("http://www.example.com")
			Expect(res).To(BeTrue())
		})

		It("Must return false", func() {
			res := isValidUrl("document/golangcode/data")
			Expect(res).To(BeFalse())
		})

	})

	Describe("readImage", func() {

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

	Describe("transformGraph", func() {

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
