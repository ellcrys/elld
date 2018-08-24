package mint

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ellcrys/elld/config"

	"github.com/dustin/go-humanize"

	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"github.com/tensorflow/tensorflow/tensorflow/go/op"

	"github.com/ellcrys/elld/util/logger"
)

// BankNote hold informations about the banknote scaNNED
type BankNote struct {
	currencyName        string `json:"currencyName"`
	country             string `json:"country"`
	denominationFigures string `json:"denominationFigures"`
	denominationText    string `json:"denominationText"`
	shortName           string `json:"shortName"`
}

// Name returns the currency name in string
func (b *BankNote) Name() string {
	return b.currencyName
}

// Country return the country of the currency
func (b *BankNote) Country() string {
	return b.country
}

// Figure retrn the integer value of the currency scanned
func (b *BankNote) Figure() string {
	return b.denominationFigures
}

// Text return the currency in word
func (b *BankNote) Text() string {
	return b.denominationText
}

// Shortname returns the shortname of the currency
func (b *BankNote) Shortname() string {
	return b.shortName
}

// WriteCounter counts the number of bytes written to it. It implements to the io.Writer
// interface and we can pass this into io.TeeReader() which will report progress on each
// write cycle.
type WriteCounter struct {
	Total uint64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

// PrintProgress of the downloader
func (wc *WriteCounter) PrintProgress() {
	// Clear the line by using a character return to go back to the start and remove
	// the remaining characters by filling it with spaces
	fmt.Printf("\r%s", strings.Repeat(" ", 35))

	// Return again and print current status of download
	// We use the humanize package to print the bytes in a meaningful way (e.g. 10 MB)
	fmt.Printf("\rDownloading... %s complete", humanize.Bytes(wc.Total))
}

//Analyzer contains methods that
// predict notes from image and bytes
type Analyzer struct {
	trainFilePath      string
	elldConfigDir      string
	log                logger.Logger
	trainDataDirName   string
	mintDir            string
	kerasGoPath        string
	trainDataName      string
	trainDataExtension string
	ellFilePath        string
	ellRemoteURL       string
}

//NewAnalyzer initialized the Analazer struct and set default
//train file to use
func NewAnalyzer(trainFilePath string, cfg *config.EngineConfig, log logger.Logger) *Analyzer {

	var (
		trainDataDirName   = "train_data"
		elldConfigDir      = cfg.ConfigDir()
		mintDir            = elldConfigDir + "/trainDir"
		kerasGoPath        = mintDir + "/forGo2"
		trainDataName      = "trainfile"
		trainDataExtension = ".ell"
		ellFilePath        = elldConfigDir + "/" + trainDataName + trainDataExtension
		ellRemoteURL       = "http://192.168.4.103/train_file.ell"
	)

	return &Analyzer{
		trainFilePath:      trainFilePath,
		elldConfigDir:      elldConfigDir,
		log:                log,
		trainDataDirName:   trainDataDirName,
		mintDir:            mintDir,
		kerasGoPath:        kerasGoPath,
		trainDataName:      trainDataName,
		trainDataExtension: trainDataExtension,
		ellFilePath:        ellFilePath,
		ellRemoteURL:       ellRemoteURL,
	}
}

// Prepare prepares the training ell to be used
func (a *Analyzer) Prepare() error {

	//check if elld_directy exist, if not then create it
	if _, err := os.Stat(a.elldConfigDir); os.IsNotExist(err) {
		if err = os.Mkdir(a.elldConfigDir, 0777); err != nil {
			return fmt.Errorf("failed ")
		}
	}

	_, err := os.Stat(a.ellFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			a.log.Error("File not found in elld config Dir, Checking if user supplied one via flag")
		} else {
			return fmt.Errorf("failed to read train dataset: %s", err)
		}
	} else {
		return nil
	}

	//check if user supplied it during initialization
	if a.trainFilePath != "" {

		a.log.Info("User supplied train file to be used")

		err := unzip(a.trainFilePath, a.mintDir)
		if err != nil {
			a.log.Error("Cannot unzip supplied Train file", "Error", err)
			return err
		}
		return nil
	}

	//If user did not supply ithe training file, then
	//download the ell file and save it to the config dir
	err = a.downloadEllToPath(a.elldConfigDir, a.ellRemoteURL)
	if err != nil {
		a.log.Error("Error Downloading TrainData from remote URL", "Error", err)
		return (err)
	}

	a.log.Info("Download Finished")

	err = unzip(a.ellFilePath, a.mintDir)
	if err != nil {
		a.log.Error("Error unzipping TrainData to configDir", "Error", err)
		return (err)
	}

	a.log.Info("Extraction Done")

	return nil
}

// Predict accept imagepath as string
// then return  BankNote, *tf.Tensor, error
func (a *Analyzer) Predict(imagePath string) (*BankNote, *tf.Tensor, error) {

	imageFile, err := os.Open(imagePath)
	if err != nil {
		a.log.Error("Unable to open image from path", "Error", err)
		return nil, nil, err
	}
	var imgBuffer bytes.Buffer
	io.Copy(&imgBuffer, imageFile)
	img, err := readImage(&imgBuffer, "png")
	if err != nil {
		a.log.Error("Error making a tensor from image", "Error", err)
		return nil, nil, err
	}

	res, er := a.mintLoader(img)
	if er != nil {
		return nil, nil, err
	}

	return res, img, nil
}

// PredictBytes accept byte as []byte
// then return  BankNote, *tf.Tensor, error
func (a *Analyzer) PredictBytes(imgByte []byte) (*BankNote, *tf.Tensor, error) {
	//predict image note from .png and .jpg

	var imgBuffer bytes.Buffer
	imgBuffer.Write(imgByte)
	img, err := readImage(&imgBuffer, "png")
	if err != nil {
		a.log.Fatal("error making tensor from bytes : ", err)
		return nil, nil, err
	}

	res, er := a.mintLoader(img)
	if er != nil {
		return nil, nil, err
	}

	return res, img, nil
}

//mintLoader accepts a tensor and output Banknote
func (a *Analyzer) mintLoader(img *tf.Tensor) (*BankNote, error) {

	model, err := tf.LoadSavedModel(a.kerasGoPath, []string{"tags"}, nil)
	if err != nil {
		a.log.Error("unable to load tensorflow model", "Error", err)
		return nil, err
	}

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
		a.log.Error("unable to predict", "Error", err)
		return nil, err
	}

	// get the one hot encoding result
	if preds, ok := result[0].Value().([][]float32); ok {

		resultData := preds[0]

		//merge one hot encoder result to string
		stringData := ""
		for _, element := range resultData {
			s := fmt.Sprintf("%v", element)
			stringData = stringData + s
		}

		//load the result json file
		file, e := ioutil.ReadFile(a.mintDir + "/result.json")
		if e != nil {
			a.log.Error("file error", "Error", e)
			return nil, e
		}

		// get the top level map from json
		var dataSource map[string]interface{}
		err := json.Unmarshal(file, &dataSource)
		if err != nil {
			a.log.Error("file error", "Error", err)
			return nil, err
		}

		treeData := dataSource[stringData]

		//GET The currency details for the top level tree
		bytex, _ := json.Marshal(treeData)

		var note BankNote
		err = json.Unmarshal(bytex, &note)

		return &note, nil
	}

	return nil, nil
}

// readImage takes a buffer as input then output a tensor
func readImage(imageBuffer *bytes.Buffer, imageFormat string) (*tf.Tensor, error) {
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

//transformGraph takes in an image format then return a graph
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

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory. We pass an io.TeeReader
// into Copy() to report progress on the download.
func (a *Analyzer) downloadEllToPath(filepath string, url string) error {

	// Create the file, but give it a tmp file extension, this means we won't overwrite a
	// file until it's downloaded, but we'll remove the tmp extension once downloaded.
	out, err := os.Create(filepath + "/" + a.trainDataName + ".tmp")
	if err != nil {
		a.log.Error("Unable to create a temporary file for downloader", "Error", err)
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		a.log.Error("Unable to get response from the http(ellUrl)", "Error", err)
		return err
	}
	defer resp.Body.Close()

	// Create our progress reporter and pass it to be used alongside our writer
	counter := &WriteCounter{}
	_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	if err != nil {
		return err
	}

	// The progress use the same line so print a new line once it's finished downloading
	fmt.Print("\n")

	err = os.Rename(filepath+"/"+a.trainDataName+".tmp", filepath+"/"+a.trainDataName+a.trainDataExtension)
	if err != nil {
		a.log.Error("Unable to rename downloaded file to actual name", "Error", err)
		return err
	}

	return nil
}

//unzip extract zip file to a target location
func unzip(archive, target string) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(target, 0755); err != nil {
		return err
	}

	for _, file := range reader.File {
		path := filepath.Join(target, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return err
		}
	}

	return nil
}
