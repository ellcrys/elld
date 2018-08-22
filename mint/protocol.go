package mint

import (
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

	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"github.com/tensorflow/tensorflow/tensorflow/go/op"
)

const (
	ConfigDir = "~/elld_config/"
)

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

func Spec() {
	fmt.Println("This is great")
}

// prepare ell to be used
func prepare(ellTrainPath string) {
	// check if ell path is supplied from argument passed
	//check if ell is available from config directory
	// if not available then download from source

	if _, err := os.Stat(ellTrainPath); err != nil {
		if os.IsNotExist(err) {
			//file does not exist
			fmt.Errorf("File not found in elld config directory, Downloading fresh copy")

			//download the ell file and save it to the config dir
			ellUrl := "http://gooogle.com"

			err := DownloadEllToPath("train_file.ell", ellUrl)
			if err != nil {
				panic(err)
			}

		} else {
			// other error encountered
			fmt.Errorf("Error reading training file from the config")
		}
	}

}

// predict image
func predictNote(imagePath string) {
	//predict image note from .png and .jpg
}

func predictNotefromByte(imageByte []byte) {
	//predict image note from byte

	// convert the image into byte the pipe it to MintLoader as file parameter

	//convert byte to image
	img, _, _ := image.Decode(bytes.NewReader(imageByte))

	//save the imageByte to a file
	out, err := os.Create("./tempimages.png")
	if err != nil {
		fmt.Errorf("Error creating the images in temp folder")
	}

	newImage, er := png.Decode(out, img)
	if err != nil {
		fmt.Errorf(er)
	}

}

func mintLoader() (interface{}, error) {

	fmt.Println("This is awesome")

	var errOutput error
	var resultOutput interface{}

	imgName := os.Args[1]
	model, err := tf.LoadSavedModel("forGo2", []string{"tags"}, nil)
	if err != nil {
		log.Fatal(err)
	}

	imageFile, err := os.Open(imgName)
	if err != nil {
		log.Fatal(err)
	}
	var imgBuffer bytes.Buffer
	io.Copy(&imgBuffer, imageFile)
	img, err := readImage(&imgBuffer, "png")
	if err != nil {
		log.Fatal("error making tensor: ", err)
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
		log.Fatal(err)
		// fmt.Println("Error flowing left and right")
		os.Exit(0)
	}

	// get the one hot encoding result
	if preds, ok := result[0].Value().([][]float32); ok {
		fmt.Println(">>>>>>>>>>>>> : ", preds[0])
		//fmt.Println(reflect.TypeOf(preds[0]))
		resultData := preds[0]

		//merge one hot encoder result to string
		stringData := ""
		for _, element := range resultData {
			s := fmt.Sprintf("%v", element)
			stringData = stringData + s
		}

		fmt.Println(stringData)

		//load the result json file
		file, e := ioutil.ReadFile("./result.json")
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
		fmt.Println(treeData)

		//GET The currency details for the top level tree
		bytex, _ := json.Marshal(treeData)
		var p BankNote
		err = json.Unmarshal(bytex, &p)

		fmt.Println(p)

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

// DownloadEllToPath download the ell trainer file to the config directory
func DownloadEllToPath(filePath string, urlPath string) error {

	//create the file
	out, err := os.Create(ConfigDir + filePath)
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
