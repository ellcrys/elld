package miner

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/franela/goreq"

	// "bytes"
	// "io"
	// "io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var BanknoteAnalyzerTest = func() bool {

	return Describe("BanknoteAnalyzer", func() {

		var ba *BanknoteAnalyzer

		BeforeEach(func() {
			ba = NewBanknoteAnalyzer(cfg, log)
			goreq.SetConnectTimeout(5 * time.Second)
		})

		Describe(".fetchModel", func() {

			It("should return err when url is invalid", func() {
				err = ba.fetchRemoteModel("ht://storage.googleapis.com/somewhere/unkwn")
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal(`failed to fetch training model: Get ht://storage.googleapis.com/somewhere/unkwn: unsupported protocol scheme "ht"`))
			})

			It("should return err when unable to reach url", func() {
				err = ba.fetchRemoteModel("https://storage.googleapis.com/somewhere/unkwn")
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("failed to fetch training model. HttpStatus: 404"))
			})

			It("should successfully fetch and save train data to expected directory", func() {
				err = ba.fetchRemoteModel("http://" + server.address() + "/models.tar")
				Expect(err).To(BeNil())

				dirPath := filepath.Join(cfg.ConfigDir(), TrainDataDirName)
				_, err = os.Stat(dirPath)
				Expect(err).To(BeNil())
			})

			Context("model already exist in training directory", func() {
				It("should successfully fetch and replace it", func() {
					err = ba.fetchRemoteModel("http://" + server.address() + "/models.tar")
					Expect(err).To(BeNil())
					err = ba.fetchRemoteModel("http://" + server.address() + "/models.tar")
					Expect(err).To(BeNil())
				})
			})

			It("should return error when model is not a valid tar file", func() {
				err = ba.fetchRemoteModel("http://" + server.address() + "/models.zip")
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("failed to decompress model: gzip: invalid header"))
			})
		})

		Describe(".SetTrainingModelURI", func() {
			It("should return error when uri is not a valid path or url", func() {
				err = ba.SetTrainingModelURI("invalid")
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("invalid uri: expected a path or a URL"))
			})

			It("should return error when path does not exist", func() {
				err = ba.SetTrainingModelURI("/something/unknown")
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("invalid uri: file path does not exist"))
			})
		})

		Describe(".fetchLocalModel", func() {

			var path string

			BeforeEach(func() {
				path, err = filepath.Abs("./testdata/models.tar")
				Expect(err).To(BeNil())
			})

			It("should fetch and copy local file to target model path", func() {
				err = ba.fetchLocalModel(path)
				Expect(err).To(BeNil())

				dirPath := filepath.Join(cfg.ConfigDir(), TrainDataDirName)
				_, err = os.Stat(dirPath)
				Expect(err).To(BeNil())
			})

			Context("model already exist in training directory", func() {
				It("should successfully fetch and replace it", func() {
					err = ba.fetchLocalModel(path)
					Expect(err).To(BeNil())
					err = ba.fetchLocalModel(path)
					Expect(err).To(BeNil())

					dirPath := filepath.Join(cfg.ConfigDir(), TrainDataDirName)
					_, err = os.Stat(dirPath)
					Expect(err).To(BeNil())
				})
			})

			It("should return error when model is not a valid tar file", func() {
				path, err := filepath.Abs("./testdata/models.zip")
				Expect(err).To(BeNil())
				err = ba.fetchLocalModel(path)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("failed to decompress model: gzip: invalid header"))
			})

		})

		Describe(".Prepare", func() {

			BeforeEach(func() {
				filePath := filepath.Join(cfg.ConfigDir(), TrainDataDirName)
				_, err = os.Stat(filePath)
				Expect(err).To(BeAssignableToTypeOf(&os.PathError{}))
			})

			It("should fetch the model from the remote url when no model exists", func() {
				DefaultTrainModelFetchURI = "http://" + server.address() + "/models.tar"
				Expect(ba.trainingModelURI).To(BeEmpty())
				err := ba.Prepare()
				Expect(err).To(BeNil())
			})

			It("should use and fetch remote model explicitly set", func() {
				err := ba.SetTrainingModelURI("http://" + server.address() + "/models.tar")
				Expect(err).To(BeNil())
				Expect(ba.trainingModelURI).ToNot(BeEmpty())
				err = ba.Prepare()
				Expect(err).To(BeNil())
			})

			It("should use and fetch local mode explicitly", func() {
				path, _ := filepath.Abs("./testdata/models.tar")
				err := ba.SetTrainingModelURI(path)
				Expect(err).To(BeNil())
				Expect(ba.trainingModelURI).ToNot(BeEmpty())
				err = ba.Prepare()
				Expect(err).To(BeNil())
			})

			When("model had already been fetched", func() {

				BeforeEach(func() {
					err := ba.SetTrainingModelURI("http://" + server.address() + "/models.tar")
					Expect(err).To(BeNil())
					err = ba.Prepare()
					Expect(err).To(BeNil())
					ba.fetched = false
				})

				It("should re-fetch the model if forced fetch is enabled", func() {
					ba.ForceFetch()
					Expect(ba.forceFetch).To(BeTrue())
					Expect(ba.fetched).To(BeFalse())

					err := ba.SetTrainingModelURI("http://" + server.address() + "/models.tar")
					Expect(err).To(BeNil())
					err = ba.Prepare()
					Expect(err).To(BeNil())

					Expect(ba.fetched).To(BeTrue())
				})
			})
		})

		Describe(".PredictBytes", func() {

			Context("with valid model archive", func() {
				BeforeEach(func() {
					path, err := filepath.Abs("./testdata/models.tar")
					Expect(err).To(BeNil())
					err = ba.fetchLocalModel(path)
					Expect(err).To(BeNil())
				})

				It("should return err when image could not be decoded", func() {
					var img bytes.Buffer
					img.Write([]byte("garbage"))
					res, err := ba.PredictBytes(&img)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(Equal("failed to decode image: image: unknown format"))
					Expect(res).To(BeNil())
				})

				It("should return err when image is not png or jpg", func() {
					f, _ := os.Open("./testdata/image.tiff")
					defer f.Close()
					var img bytes.Buffer
					io.Copy(&img, f)
					res, err := ba.PredictBytes(&img)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(Equal("failed to decode image: image: unknown format"))
					Expect(res).To(BeNil())
				})

				It("should return err when image is a gif", func() {
					f, _ := os.Open("./testdata/marshmellow.gif")
					defer f.Close()
					var img bytes.Buffer
					io.Copy(&img, f)
					res, err := ba.PredictBytes(&img)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(Equal("failed to decode image: image: unknown format"))
					Expect(res).To(BeNil())
				})
			})

			Context("with invalid model archive", func() {

				BeforeEach(func() {
					path, err := filepath.Abs("./testdata/models_invalid.tar")
					Expect(err).To(BeNil())
					err = ba.fetchLocalModel(path)
					Expect(err).To(BeNil())
				})

				It("should fail to predict", func() {
					f, _ := os.Open("./testdata/image128.png")
					defer f.Close()
					var img bytes.Buffer
					io.Copy(&img, f)
					res, err := ba.PredictBytes(&img)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(MatchRegexp("Can't parse.*as binary proto"))
					Expect(res).To(BeNil())
				})
			})

			Context("with missing manifest in model archive", func() {

				BeforeEach(func() {
					path, err := filepath.Abs("./testdata/models_missing_manifest.tar")
					Expect(err).To(BeNil())
					err = ba.fetchLocalModel(path)
					Expect(err).To(BeNil())
				})

				It("should fail to predict", func() {
					f, _ := os.Open("./testdata/image128.png")
					defer f.Close()
					var img bytes.Buffer
					io.Copy(&img, f)
					res, err := ba.PredictBytes(&img)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(MatchRegexp("failed to load manifest file: open .* no such file or directory"))
					Expect(res).To(BeNil())
				})
			})

			Context("with malformed manifest in model archive", func() {

				BeforeEach(func() {
					path, err := filepath.Abs("./testdata/models_malformed_manifest.tar")
					Expect(err).To(BeNil())
					err = ba.fetchLocalModel(path)
					Expect(err).To(BeNil())
				})

				It("should fail to predict", func() {
					f, _ := os.Open("./testdata/image128.png")
					defer f.Close()
					var img bytes.Buffer
					io.Copy(&img, f)
					res, err := ba.PredictBytes(&img)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(MatchRegexp("failed to decode manifest file: invalid character.*"))
					Expect(res).To(BeNil())
				})
			})

			Context("with valid model", func() {

				BeforeEach(func() {
					path, err := filepath.Abs("./testdata/models.tar")
					Expect(err).To(BeNil())
					err = ba.fetchLocalModel(path)
					Expect(err).To(BeNil())
				})

				// It("should successfully return a predicted banknote", func() {
				// 	f, _ := os.Open("./testdata/image128.png")
				// 	defer f.Close()
				// 	var img bytes.Buffer
				// 	io.Copy(&img, f)
				// 	res, err := ba.PredictBytes(&img)
				// 	Expect(err).To(BeNil())
				// 	Expect(res).To(Equal(&BankNote{
				// 		currencyCode: "NGN",
				// 		denomination: "500",
				// 	}))
				// })

				It("should return not nil result and nil error when a prediction cannot be made", func() {
					f, _ := os.Open("./testdata/ball.jpg")
					defer f.Close()
					var img bytes.Buffer
					io.Copy(&img, f)
					res, err := ba.PredictBytes(&img)
					Expect(err).To(BeNil())
					Expect(res).ToNot(BeNil())
				})
			})
		})

		Describe(".Predict", func() {
			It("should fail to predict when model is invalid", func() {
				path, err := filepath.Abs("./testdata/unknown.tar")
				Expect(err).To(BeNil())
				err = ba.fetchLocalModel(path)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("failed to decompress model: invalid argument"))
			})
		})

		// Describe(".Validator", func() {

		// })
	})
}
