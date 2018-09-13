package miner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	// "github.com/k0kubun/pp"

	"gopkg.in/asaskevich/govalidator.v4"

	"github.com/ellcrys/elld/util"

	"github.com/shopspring/decimal"

	"github.com/franela/goreq"

	"github.com/ellcrys/elld/config"

	"image"
	_ "image/jpeg"
	_ "image/png"

	"github.com/ellcrys/elld/util/logger"
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"github.com/tensorflow/tensorflow/tensorflow/go/op"
)

const (
	// TrainDataDirName refers to the name of the directory
	// where training models and other data are stored
	TrainDataDirName = "traindata"

	// TrainModelDirName is the actual directory
	// within the training data directory containing the
	// tensorflow model for predicting currency
	TrainModelDirName = "models"

	// ValidatorModelDirName is the directory
	// within the training data directory containing the
	// tensorflow model for validating currency
	ValidatorModelDirName = "validators"
)

var (
	// DefaultTrainModelFetchURI is the remote location
	// to download the banknote training model
	DefaultTrainModelFetchURI = "https://storage.googleapis.com/krogan/sample_train_file.model"
)

// argMax returns the maximum number  and its index in a slice
func argMax(result []float32) (int, float32) {

	index := 0
	max := float32(0)

	for idx, arg := range result {
		if arg > max {
			max = arg
			index = idx
		}
	}
	return index, max
}

// subImagePosition provides interface to
// get the sliced sub section of the image
type subImagePosition interface {
	subImage(r image.Rectangle) image.Image
}

// BankNote hold informations about the banknote Scanned
type BankNote struct {
	currencyCode string
	denomination string
}

// Denomination gets the denomination
func (b *BankNote) Denomination() string {
	return b.denomination
}

// Code gets the currency code
func (b *BankNote) Code() string {
	return b.currencyCode
}

// ProgressLogger calculates download
// progress and logs it
type ProgressLogger struct {
	log             logger.Logger
	ContentSize     int64
	TotalDownloaded int64
	reads           int64
}

// newProgressLogger creates an instance of ProgressLogger
func newProgressLogger(log logger.Logger, contentSize int64) *ProgressLogger {
	return &ProgressLogger{log: log, ContentSize: contentSize}
}

// Write is where the calculation is done and logged
func (pc *ProgressLogger) Write(p []byte) (int, error) {
	n := len(p)
	pc.TotalDownloaded += int64(n)
	pc.reads++

	// We shouldn't do this all the time. Print long every 2^X reads
	if (pc.reads % (1 << 10)) == 0 {
		pct := decimal.NewFromFloat(float64(100) * (float64(pc.TotalDownloaded) / float64(pc.ContentSize)))
		pc.log.Info("Training model download progress", "Percent", pct.Round(1))
		fmt.Println(pct)
	} else {
		if pc.TotalDownloaded == pc.ContentSize {
			pc.log.Info("Training model download progress", "Percent", 100)
		}
	}

	return n, nil
}

// BanknoteAnalyzer defines functionalities for
// performing operations to detect and determine
// features of national currencies.
type BanknoteAnalyzer struct {

	// log is the default logger
	log logger.Logger

	// cfg is engine configuration
	cfg *config.EngineConfig

	// trainingModelURI is the user-defined file path or
	// url of the training model to fetch.
	trainingModelURI string

	// forceFetch when set to true forces
	// the analyzer to fetch the training model.
	forceFetch bool

	// fetched indicates the model was successfully
	// fetched.
	fetched bool
}

// NewBanknoteAnalyzer creates an instance of BanknoteAnalyzer
func NewBanknoteAnalyzer(cfg *config.EngineConfig, log logger.Logger) *BanknoteAnalyzer {
	return &BanknoteAnalyzer{
		log: log,
		cfg: cfg,
	}
}

// ForceFetch sets forces the training
// model to be re-fetched
func (a *BanknoteAnalyzer) ForceFetch() {
	a.forceFetch = true
}

// SetTrainingModelURI sets the URI to fetch the training model from.
func (a *BanknoteAnalyzer) SetTrainingModelURI(uri string) error {
	isFilePath, _ := govalidator.IsFilePath(uri)
	if !isFilePath && !govalidator.IsURL(uri) {
		return fmt.Errorf("invalid uri: expected a path or a URL")
	}
	if isFilePath {
		if _, err := os.Stat(uri); os.IsNotExist(err) {
			return fmt.Errorf("invalid uri: file path does not exist")
		}
	}
	a.trainingModelURI = uri
	return nil
}

// Prepare prepares the training model for use.
// If the training model does not exists on the
// local machine, it will fetch it.
func (a *BanknoteAnalyzer) Prepare() error {

	var notFound = false
	var fetchURI = DefaultTrainModelFetchURI
	a.fetched = false

	// If the user specified a training model fetch
	// URI, then we set fetchURI to it.
	if len(a.trainingModelURI) != 0 {
		fetchURI = a.trainingModelURI
	}

	// Check whether the training data directory exists
	// at the expected path. If it doesn't, we must fetch it.
	trainingDataPath := filepath.Join(a.cfg.ConfigDir(), TrainDataDirName)
	if _, err := os.Stat(trainingDataPath); os.IsNotExist(err) {
		notFound = true
	}

	// If the training model exists and
	// force fetch is not enabled, we return
	if !notFound && !a.forceFetch {
		return nil
	}

	// If URI is a file path, fetch from file
	if ok, _ := govalidator.IsFilePath(fetchURI); ok {
		return a.fetchLocalModel(fetchURI)
	}

	// if URI is a remote URL, fetch from URL
	if err := a.fetchRemoteModel(fetchURI); err != nil {
		return err
	}

	a.fetched = true

	return nil
}

// Validator validates the supplied image and output confidence level of it being a currency
func (a *BanknoteAnalyzer) Validator(imgName string) (float32, error) {

	// get the image extension for the proposed validated image
	// we are dealing with .png, .jpg and .jpeg file format
	imgExtension := filepath.Ext(imgName)
	imgExtension = imgExtension[1:]

	// open the image to be validated
	imageFile, err := os.Open(imgName)
	if err != nil {
		return 0, fmt.Errorf("failed to open image: %s", err)
	}

	var decodedImage image.Image

	if imgExtension == "png" {
		decodedImage, err = png.Decode(imageFile)
		if err != nil {
			return 0, fmt.Errorf("failed to decode png image: %s", err)
		}
	}

	if imgExtension == "jpg" || imgExtension == "jpeg" {
		decodedImage, err = jpeg.Decode(imageFile)
		if err != nil {
			return 0, fmt.Errorf("failed to decode jpg image: %s", err)
		}
	}

	//load the model for Note validator
	validatorRootDir := filepath.Join(a.cfg.ConfigDir(), TrainDataDirName, ValidatorModelDirName)
	model, err := tf.LoadSavedModel(validatorRootDir, []string{"tags"}, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to load validator model : %s", err)
	}

	// get the total of all average of the slices
	var cummulativeTotal float32

	// step size for moving the sliding window around the image
	stepSize := 10

	// get the height and width of the image to validate
	imageSize := decodedImage.Bounds().Size()
	imageWidth := imageSize.X
	imageHeight := imageSize.Y

	// get the size of the sliding window
	window_width := imageWidth / 4
	window_height := imageHeight

	y := 0
	sn := 0

	// loop through one-way sliding window by shifting only the x cordinates
	for x := 0; x <= imageWidth; x += stepSize {

		// continue slicing if the remains is greater than the window witdth
		if (imageWidth - x) > window_width {

			sn = sn + 1

			// slice the image at cordiante x and y with the window width and height
			slicedImage := decodedImage.(interface {
				SubImage(r image.Rectangle) image.Image
			}).SubImage(image.Rect(x, y, window_width+x, window_height))

			// create a buffer to hold the sliced image
			var imgBuffer bytes.Buffer

			if imgExtension == "png" {
				err = png.Encode(&imgBuffer, slicedImage)

				if err != nil {
					return 0, fmt.Errorf("failed to encode png image to byte : %s", err)
				}
			}

			if imgExtension == "jpg" || imgExtension == "jpeg" {
				err = jpeg.Encode(&imgBuffer, slicedImage, nil)

				if err != nil {
					return 0, fmt.Errorf("failed to encode jpg image to byte : %s", err)
				}
			}

			// create a tensor from the image buffer
			img, err := imageToTensor(&imgBuffer, imgExtension)
			if err != nil {
				return 0, fmt.Errorf("failed to make tensor from image : %s", err)
			}

			// pass the generated tensor to the computational graph
			result, err := model.Session.Run(
				map[tf.Output]*tf.Tensor{
					model.Graph.Operation("inputNode_input").Output(0): img,
				},
				[]tf.Output{
					model.Graph.Operation("inferNode/Softmax").Output(0),
				},
				nil,
			)

			if err != nil {
				return 0, err
			}

			// predictions is the array of result from  the predictions on all the features
			prediction := result[0].Value().([][]float32)

			//fmt.Println(">>>>> : ", sn, " -- ", prediction)

			// resultData is the array of result
			resultData := prediction[0]

			var sliceSumResult float32

			totalCountSet := float32(len(resultData))

			for _, singleResult := range resultData {
				sliceSumResult = (sliceSumResult + singleResult)
			}

			sliceAverageResult := sliceSumResult / totalCountSet
			cummulativeTotal = cummulativeTotal + sliceAverageResult

		}
	}

	averageTotal := cummulativeTotal / float32(sn)
	return averageTotal, nil
}

// Predict is like PredictBytes except it accepts an image path
func (a *BanknoteAnalyzer) Predict(imagePath string) (*BankNote, error) {

	img, err := os.Open(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %s", err)
	}
	defer img.Close()

	var buf bytes.Buffer
	io.Copy(&buf, img)

	return a.PredictBytes(&buf)
}

// PredictBytes accept a given buffer representing
// an image and attempts to perform banknote predictions
func (a *BanknoteAnalyzer) PredictBytes(img *bytes.Buffer) (*BankNote, error) {

	_, format, err := image.DecodeConfig(bytes.NewReader(img.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %s", err)
	}

	tfImg, err := imageToTensor(img, format)
	if err != nil {
		return nil, err
	}

	res, err := a.predict(tfImg)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// predict accepts a tensor, predicts and outputs
// a Banknote object.
func (a *BanknoteAnalyzer) predict(img *tf.Tensor) (*BankNote, error) {

	// Load the model
	modelRootDir := filepath.Join(a.cfg.ConfigDir(), TrainDataDirName, TrainModelDirName)
	model, err := tf.LoadSavedModel(modelRootDir, []string{"tags"}, nil)
	if err != nil {
		return nil, err
	}

	// Run the graph with the associated session
	tensors, err := model.Session.Run(
		map[tf.Output]*tf.Tensor{
			model.Graph.Operation("inputNode_input").Output(0): img,
		},
		[]tf.Output{
			model.Graph.Operation("inferNode/Softmax").Output(0),
		},
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to run graph: %s", err)
	}

	if tensors == nil || len(tensors) == 0 {
		return nil, nil
	}
	predictions := tensors[0].Value().([][]float32)
	resultData := predictions[0]

	position, _ := argMax(resultData)

	// Merge one hot encoder to construct
	// a key that we can use to extract the result object
	// from the manifest.
	var resultKey string
	for _, element := range resultData {
		s := fmt.Sprintf("%d", int(element))
		resultKey = resultKey + s
	}

	// load the result json file
	manifestPath := filepath.Join(a.cfg.ConfigDir(), TrainDataDirName, "manifest.json")
	file, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load manifest file: %s", err)
	}

	// Decode the manifest content
	var manifest map[string]interface{}
	err = json.Unmarshal(file, &manifest)
	if err != nil {
		return nil, fmt.Errorf("failed to decode manifest file: %s", err)
	}

	var cMap = manifest["currencies"].(map[string]interface{})
	if result, ok := cMap[string(position)]; ok {
		// if result, ok := cMap[resultKey]; ok {
		return &BankNote{
			currencyCode: result.(map[string]interface{})["currencyCode"].(string),
			denomination: result.(map[string]interface{})["denomination"].(string),
		}, nil
	}

	return nil, nil
}

// fetchLocalModel fetches the training model from a
// local path and saves it to a target directory.
// It will overwrite existing train model.
//
// Expects the caller to have validated the path.
func (a *BanknoteAnalyzer) fetchLocalModel(path string) error {

	// decompress the model
	mr, _ := os.Open(path)
	defer mr.Close()
	root, err := util.Untar(a.cfg.ConfigDir(), mr)
	if err != nil {
		return fmt.Errorf("failed to decompress model: %s", err)
	}

	// Rename the root directory after decompression to
	// something we will recognize. Remove any previous
	// directory with matching name.
	newRootPath := filepath.Join(a.cfg.ConfigDir(), TrainDataDirName)
	if _, err := os.Stat(newRootPath); err == nil {
		if err := os.RemoveAll(newRootPath); err != nil {
			return fmt.Errorf("failed to remove train data directory: %s", err)
		}
	}
	if err = os.Rename(root, newRootPath); err != nil {
		return fmt.Errorf("failed to rename decompress root dir: %s", err)
	}

	return nil
}

// fetchRemoteModel fetches the training model from a
// remote URL and saves it to a target directory.
// It will overwrite existing train model.
func (a *BanknoteAnalyzer) fetchRemoteModel(url string) error {

	resp, err := goreq.Request{Uri: url}.Do()
	if err != nil {
		return fmt.Errorf("failed to fetch training model: %s", err)
	}

	if sc := resp.StatusCode; sc != 200 {
		return fmt.Errorf("failed to fetch training model. HttpStatus: %d", sc)
	}

	tmpFile, err := ioutil.TempFile("", util.RandString(10))
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %s", err)
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// Read the response content into the tmp file.
	// Create our progress reporter and pass it to
	// be used alongside our writer
	progressChecker := newProgressLogger(a.log, resp.ContentLength)
	_, err = io.Copy(tmpFile, io.TeeReader(resp.Body, progressChecker))
	if err != nil {
		return fmt.Errorf("failed to read response to tmp file: %s", err)
	}

	// Uncompress the archive and return the root
	tmpFile.Seek(0, 0)
	root, err := util.Untar(a.cfg.ConfigDir(), tmpFile)
	if err != nil {
		return fmt.Errorf("failed to decompress model: %s", err)
	}

	// Rename the root directory after decompression to
	// something we will recognize. Remove any previous
	// directory with matching name.
	newRootPath := filepath.Join(a.cfg.ConfigDir(), TrainDataDirName)
	if _, err := os.Stat(newRootPath); err == nil {
		if err := os.RemoveAll(newRootPath); err != nil {
			return fmt.Errorf("failed to remove train data directory: %s", err)
		}
	}
	if err = os.Rename(root, newRootPath); err != nil {
		return fmt.Errorf("failed to rename decompress root dir: %s", err)
	}

	return nil
}

// imageToTensor takes an image and converts to a tensor.
// Expect imageBuffer to include a valid image.
func imageToTensor(imageBuffer *bytes.Buffer, imageFormat string) (*tf.Tensor, error) {
	tensor, err := tf.NewTensor(imageBuffer.String())
	if err != nil {
		return nil, err
	}
	graph, input, output, err := transformGraph(imageFormat)
	if err != nil {
		return nil, err
	}
	session, err := tf.NewSession(graph, nil)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	normalized, err := session.Run(
		map[tf.Output]*tf.Tensor{input: tensor},
		[]tf.Output{output},
		nil)
	if err != nil {
		return nil, err
	}
	return normalized[0], nil
}

// transformGraph takes in an image format then return a graph
func transformGraph(imageFormat string) (graph *tf.Graph, input, output tf.Output, err error) {
	const (
		H, W  = 128, 128
		Mean  = float32(117)
		Scale = float32(1)
	)
	s := op.NewScope()
	input = op.Placeholder(s, tf.String)

	var decode tf.Output
	switch imageFormat {
	case "png":
		decode = op.DecodePng(s, input, op.DecodePngChannels(3))
	case "jpg",
		"jpeg":
		decode = op.DecodeJpeg(s, input, op.DecodeJpegChannels(3))
	default:
		return nil, tf.Output{}, tf.Output{},
			fmt.Errorf("imageFormat not supported: %s", imageFormat)
	}

	output = op.Div(s,
		op.Sub(s,
			op.ResizeBilinear(s,
				op.ExpandDims(s,
					op.Cast(s, decode, tf.Float),
					op.Const(s.SubScope("make_batch"), int32(0))),
				op.Const(s.SubScope("size"), []int32{H, W})),
			op.Const(s.SubScope("mean"), Mean)),
		op.Const(s.SubScope("scale"), Scale))
	graph, err = s.Finalize()
	return graph, input, output, err
}
