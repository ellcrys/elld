package mint

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize"

	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"github.com/tensorflow/tensorflow/tensorflow/go/op"
)

var (
	homeDir             = os.Getenv("HOME")
	goPath              = os.Getenv("GOPATH")
	elldConfigDir       = os.Getenv("HOME") + "/elld_config"
	mintDir             = elldConfigDir + "/trainDir"
	kerasGoPath         = mintDir + "/forGo2"
	ellPathConfig       = elldConfigDir + "/train_file.ell"
	ellRemoteURL        = "http://192.168.4.103/train_file.ell"
	downloadELLFilename = "train_file.ell"
)

var publicNote BankNote

type BankNote struct {
	CurrencyName        string `json:"currencyName"`
	Country             string `json:"Country"`
	DenominationFigures string `json:"DenominationFigures"`
	DenominationText    string `json:"denominationText"`
	ElliesConversion    string `json:"elliesConversion"`
	DollarConversion    string `json:"dollarConversion"`
	ShortName           string `json:"shortName"`
}

func (b *BankNote) name() string {
	return b.CurrencyName
}
func (b *BankNote) country() string {
	return b.Country
}
func (b *BankNote) figure() string {
	return b.DenominationFigures
}
func (b *BankNote) text() string {
	return b.DenominationText
}
func (b *BankNote) ellies() string {
	return b.ElliesConversion
}
func (b *BankNote) dollar() string {
	return b.DollarConversion
}
func (b *BankNote) shortname() string {
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

func (wc WriteCounter) PrintProgress() {
	// Clear the line by using a character return to go back to the start and remove
	// the remaining characters by filling it with spaces
	fmt.Printf("\r%s", strings.Repeat(" ", 35))

	// Return again and print current status of download
	// We use the humanize package to print the bytes in a meaningful way (e.g. 10 MB)
	fmt.Printf("\rDownloading... %s complete", humanize.Bytes(wc.Total))
}

func Spec() {

	arguement := "Yes"
	ellFileArguement := "/Users/princesegzy01/Documents/sample_ell/train_file.ell"

	var ellFilePath string

	if arguement == "Yes" {

		ellFilePath = ellFileArguement

		response := checkValidELL(ellFilePath)
		if response != nil {
			fmt.Printf("Error  : %s ", response)
		}

		// delete existing resource
		deleteResources()

	} else {
		ellFilePath = ellPathConfig
	}

	//Prepare the .ell to be used
	prepare(arguement, ellFilePath)

	//Image to predict, This can come from commandline or any location
	imagePath := "mint/image128.png"

	_, err := mintLoader(imagePath)
	if err != nil {
		fmt.Errorf("Error from Mintloader ", err)
		os.Exit(0)
	}

	fmt.Println("<<<<<<<<<<<<<<<<<<<", publicNote.shortname())
	fmt.Println("<<<<<<<<<<<<<<<<<<<", publicNote.country())
	fmt.Println("<<<<<<<<<<<<<<<<<<<", publicNote.dollar())
	fmt.Println("<<<<<<<<<<<<<<<<<<<", publicNote.ellies())
	fmt.Println("<<<<<<<<<<<<<<<<<<<", publicNote.figure())
	fmt.Println("<<<<<<<<<<<<<<<<<<<", publicNote.name())
	fmt.Println("<<<<<<<<<<<<<<<<<<<", publicNote.text())
}

// prepare ell to be used
func prepare(argument string, ellPath string) {
	// check if ell path is supplied from argument passed
	//check if ell is available from config directory
	// if not available then download from source

	//argument := "No"
	// check if training file is supplied from flag, if yes.. use it
	// If no, use existing ell file in the config dir
	// if none is available in the config dir, them download new one and use it
	if argument == "Yes" {

		//get the path of the ell file to be used and extract it to the config_dir
		//ellPath := ellPathConfig
		er := unzip(ellPath, mintDir)
		if er != nil {
			panic(er)
		}

	} else {

		_, err := os.Stat(ellPathConfig)
		if err != nil {

			if os.IsNotExist(err) {
				//file does not exist
				fmt.Println("File not found in elld config directory, Downloading fresh copy")

				//download the ell file and save it to the config dir
				err := DownloadEllToPath(ellPath, ellRemoteURL)
				if err != nil {
					panic(err)
				}

				fmt.Println("Download Finished")

				er := unzip(ellPath, mintDir)
				if er != nil {
					panic(er)
				}

				fmt.Println("Extraction Done")

			} else {
				// other error encountered
				fmt.Println("Error reading training file from the config")
			}
		}
	}
}

func predictNote() {}

// take in imagepatha and convert it to tensor
func imageToTensor(imagePath string) (*tf.Tensor, error) {
	//predict image note from .png and .jpg

	imageFile, err := os.Open(imagePath)
	if err != nil {
		log.Fatal(err)
	}
	var imgBuffer bytes.Buffer
	io.Copy(&imgBuffer, imageFile)
	img, err := readImage(&imgBuffer, "png")
	if err != nil {
		log.Fatal("error making tensor: ", err)
	}

	return img, nil
}

// byteToTensor take in byte and convert it to a tensor
func byteToTensor(imgByte []byte) {

	//imgByte := code.PNG()

	// convert []byte to image for saving to file
	img, _, _ := image.Decode(bytes.NewReader(imgByte))

	//save the imgByte to file
	out, err := os.Create("./QRImg.png")

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = png.Encode(out, img)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func mintLoader(imagePathString string) (interface{}, error) {

	// fmt.Println("This is awesome")

	var errOutput error
	var resultOutput interface{}

	//imgName := os.Args[1]
	model, err := tf.LoadSavedModel(kerasGoPath, []string{"tags"}, nil)
	if err != nil {
		log.Fatal(err)
	}

	// get the tensor from input image
	img, _ := imageToTensor(imagePathString)

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
		log.Fatal(err)
		// fmt.Println("Error flowing left and right")
		os.Exit(0)
	}

	// get the one hot encoding result
	if preds, ok := result[0].Value().([][]float32); ok {
		// fmt.Println(">>>>>>>>>>>>> : ", preds[0])
		//fmt.Println(reflect.TypeOf(preds[0]))
		resultData := preds[0]

		//merge one hot encoder result to string
		stringData := ""
		for _, element := range resultData {
			s := fmt.Sprintf("%v", element)
			stringData = stringData + s
		}

		// fmt.Println(stringData)

		//load the result json file
		file, e := ioutil.ReadFile(mintDir + "/result.json")
		if e != nil {
			fmt.Printf("File error: %v\n", e)
			os.Exit(1)
		}

		// get the top level map from json
		var dataSource map[string]interface{}
		err := json.Unmarshal(file, &dataSource)

		if err != nil {
			fmt.Println("Unnable to marshall top level tree")
			os.Exit(0)
		}

		//treeData := dataSource["001"]
		treeData := dataSource[stringData]
		// fmt.Println(treeData)

		//GET The currency details for the top level tree
		bytex, _ := json.Marshal(treeData)

		err = json.Unmarshal(bytex, &publicNote)

		resultOutput = treeData
		errOutput = nil

		// if preds[0][0] > preds[0][1] {
		// 	fmt.Println("male")
		// } else {
		// 	fmt.Println("female")
		// }
	} else {
		fmt.Println("Issue with Predictions")

		//resultOutput = treeData
		errOutput = fmt.Errorf("Error performing Mint operation")
	}

	return resultOutput, errOutput
}

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

func transformGraph(imageFormat string) (graph *tf.Graph, input,
	output tf.Output, err error) {
	const (
		//H, W  = 224, 224
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

func deleteResources() {

	err := os.Remove(ellPathConfig)
	if err != nil {
		fmt.Println(err)
	}

	err1 := os.RemoveAll(mintDir)
	if err1 != nil {
		fmt.Println(err1)
	}
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

func checkValidELL(ellFile string) error {
	//zipFile := "mint/train_file.ell"
	ext := ellFile[len(ellFile)-4:]
	if ext != ".ell" {
		return fmt.Errorf("File must have an extension of .ell")
	}

	zf, err := os.Open(ellFile)
	if err != nil {
		return fmt.Errorf("error opeining supplied .ell file %s", err)
	}
	fileInfo, errorStat := zf.Stat()
	if errorStat != nil {
		return fmt.Errorf("Cannot get .ell info supplied %s", errorStat)
	}

	if fileInfo.Size() == 0 {
		return fmt.Errorf("Ell file truncated during downloaded, kindly re-download")
	}

	fileType, errType := GetFileContentType(zf)
	if errType != nil {
		return fmt.Errorf("Error :  %s", errType)
	}

	if fileType != "application/zip" {
		return fmt.Errorf("Invalid .ell file type, kindly redownload")
	}

	defer zf.Close()

	//zf, err := zip.OpenReader(zipFile)
	return nil
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory. We pass an io.TeeReader
// into Copy() to report progress on the download.
func DownloadEllToPath(filepath string, url string) error {

	// Create the file, but give it a tmp file extension, this means we won't overwrite a
	// file until it's downloaded, but we'll remove the tmp extension once downloaded.
	out, err := os.Create(filepath + ".tmp")
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
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

	err = os.Rename(filepath+".tmp", filepath)
	if err != nil {
		return err
	}

	return nil
}

// DownloadEllToPath download the ell trainer file to the config directory
func DownloadEllToPath2(filePath string, urlPath string) error {

	//create the file
	//out, err := os.Create(elldConfigDir + "/" + filePath)
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}

	// close the created file when done
	defer out.Close()

	//get the .ell path from the url supplied
	resp, err := http.Get(urlPath)
	if err != nil {
		return err
	}

	//close the http downloader when done
	defer resp.Body.Close()

	//write body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

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
