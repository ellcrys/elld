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

	"github.com/dustin/go-humanize"

	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"github.com/tensorflow/tensorflow/tensorflow/go/op"

	"github.com/ellcrys/elld/util/logger"
)

var (
	homeDir            = os.Getenv("HOME")
	goPath             = os.Getenv("GOPATH")
	elldConfigDir      = os.Getenv("HOME") + "/elld_config"
	mintDir            = elldConfigDir + "/trainDir"
	kerasGoPath        = mintDir + "/forGo2"
	trainDataName      = "trainfile"
	trainDataExtension = ".ell"
	ellPathConfig      = elldConfigDir + "/" + trainDataName + trainDataExtension
	ellRemoteURL       = "http://192.168.4.103/train_file.ell"
)

//log is the logger we want to use
var log = logger.NewLogrus()

//BankNote hold informations about the banknote scaNNED
type BankNote struct {
	CurrencyName        string `json:"currencyName"`
	Country             string `json:"country"`
	DenominationFigures string `json:"denominationFigures"`
	DenominationText    string `json:"denominationText"`
	ElliesConversion    string `json:"elliesConversion"`
	DollarConversion    string `json:"dollarConversion"`
	ShortName           string `json:"shortName"`
}

//NoteName returns the currency name in string
func (b *BankNote) NoteName() string {
	return b.CurrencyName
}

//NoteCountry return the country of the currency
func (b *BankNote) NoteCountry() string {
	return b.Country
}

//NoteCountry retrn the integer value of the currency scanned
func (b *BankNote) NoteFigure() string {
	return b.DenominationFigures
}

//NoteText return the currency in word
func (b *BankNote) NoteText() string {
	return b.DenominationText
}

//NoteEllies returns the ellies equivalent
func (b *BankNote) NoteEllies() string {
	return b.ElliesConversion
}

//NoteDollar returns the Dollar equivalent
func (b *BankNote) NoteDollar() string {
	return b.DollarConversion
}

//NoteShortname returns the shortname of the currency
func (b *BankNote) NoteShortname() string {
	return b.ShortName
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

//PrintProgress of the downloader
func (wc WriteCounter) PrintProgress() {
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
	trainFilePath string
}

//NewAnalyzer initialized the Analazer struct and set default
//train file to use
func NewAnalyzer(trainFilePath string) *Analyzer {
	return &Analyzer{
		trainFilePath: trainFilePath,
	}
}

// Prepare prepares the training ell to be used
func (a *Analyzer) Prepare() error {

	_, err := os.Stat(ellPathConfig)
	if err != nil {

		if os.IsNotExist(err) {
			//file does not exist in the ellParthDonfigDir
			log.Error("File not found in elld config Dir, Checking if user supplied one via flag")

			//check if user supplied it as flag
			if a.trainFilePath != "" {

				log.Info("User supplied train file to be used")

				er := unzip(a.trainFilePath, mintDir)
				if er != nil {
					log.Error("Cannot unzip supplied Train file", "Error", err)
					return err
				}

			} else {

				//If user did not supply ithe training file, then
				//download the ell file and save it to the config dir
				err := DownloadEllToPath(elldConfigDir, ellRemoteURL)
				if err != nil {
					log.Error("Error Downloading TrainData from remote URL", "Error", err)
					return (err)
				}

				log.Info("Download Finished")

				er := unzip(ellPathConfig, mintDir)
				if er != nil {
					log.Error("Error unzipping TrainData to configDir", "Error", err)
					return (err)
				}

				log.Info("Extraction Done")
			}
		} else {
			// other error encountered
			log.Error("Error reading training file from the config")
		}
	}

	return nil
}

// PredictImage accept imagepath as string
// then return  BankNote, *tf.Tensor, error
func (a *Analyzer) PredictImage(imagePath string) (*BankNote, *tf.Tensor, error) {

	imageFile, err := os.Open(imagePath)
	if err != nil {
		log.Error("Unable to open image from path", "Error", err)
		return nil, nil, err
	}
	var imgBuffer bytes.Buffer
	io.Copy(&imgBuffer, imageFile)
	img, err := readImage(&imgBuffer, "png")
	if err != nil {
		log.Error("Error making a tensor from image", "Error", err)
		return nil, nil, err
	}

	res, er := mintLoader(img)
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
		log.Fatal("error making tensor from bytes : ", err)
		return nil, nil, err
	}

	res, er := mintLoader(img)
	if er != nil {
		return nil, nil, err
	}

	return res, img, nil
}

//mintLoader accepts a tensor and output Banknote
func mintLoader(img *tf.Tensor) (*BankNote, error) {

	model, err := tf.LoadSavedModel(kerasGoPath, []string{"tags"}, nil)
	if err != nil {
		log.Error("Unable to load tensorflow model", "Error", err)
		return nil, (err)
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
		log.Error("Unable to predict", "Error", err)
		return nil, (err)
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
		file, e := ioutil.ReadFile(mintDir + "/result.json")
		if e != nil {
			log.Error("File error", "Error", e)
			return nil, e
		}

		// get the top level map from json
		var dataSource map[string]interface{}
		err := json.Unmarshal(file, &dataSource)

		if err != nil {
			log.Error("File error", "Error", err)
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

func GetFileContentType(out *os.File) (string, error) {

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err := out.Read(buffer)
	if err != nil {
		return "", err
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	return contentType, nil
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory. We pass an io.TeeReader
// into Copy() to report progress on the download.
func DownloadEllToPath(filepath string, url string) error {

	// Create the file, but give it a tmp file extension, this means we won't overwrite a
	// file until it's downloaded, but we'll remove the tmp extension once downloaded.
	out, err := os.Create(filepath + "/" + trainDataName + ".tmp")
	if err != nil {
		log.Error("Unable to create a temporary file for downloader", "Error", err)
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		log.Error("Unable to get response from the http(ellUrl)", "Error", err)
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

	err = os.Rename(filepath+"/"+trainDataName+".tmp", filepath+"/"+trainDataName+trainDataExtension)
	if err != nil {
		log.Error("Unable to rename downloaded file to actual name", "Error", err)
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
