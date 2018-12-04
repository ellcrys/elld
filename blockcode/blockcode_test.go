package blockcode

import (
	"fmt"

	"github.com/ellcrys/elld/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Blockcode", func() {

	Describe(".validateManifest", func() {
		It("should validate correctly", func() {
			cases := map[*Manifest]error{
				{}:                                                             fmt.Errorf("manifest error: language is missing"),
				{Lang: "c++"}:                                                  fmt.Errorf("manifest error: language {c++} is not supported"),
				{Lang: "go", LangVersion: ""}:                                  fmt.Errorf("manifest error: language version is required"),
				{Lang: "go", LangVersion: "1.10.2"}:                            fmt.Errorf("manifest error: at least one public function is required"),
				{Lang: "go", LangVersion: "1.10.2", PublicFuncs: []string{""}}: fmt.Errorf("manifest error: at least one public function is required"),
			}

			for manifest, err := range cases {
				Expect(validateManifest(manifest)).To(Equal(err))
			}

			validManifest := &Manifest{Lang: "go", LangVersion: "1.10.2", PublicFuncs: []string{"some_func"}}
			Expect(validateManifest(validManifest)).To(BeNil())
		})
	})

	Describe(".FromDir", func() {

		It("should return error = 'project path does not exist' if project path is not found", func() {
			_, err := FromDir("./testdata/unknown")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("project path does not exist"))
		})

		It("should return error = 'failed to decode manifest: json: cannot unmarshal string into Go value of type blockcode.Manifest' if manifest is not valid JSON", func() {
			_, err := FromDir("./testdata/invalid_manifest")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("failed to decode manifest: json: cannot unmarshal string into Go value of type blockcode.Manifest"))
		})

		It("should return error = ''package.json' file not found in {./testdata/missing_manifest}' if package.json is missing", func() {
			_, err := FromDir("./testdata/missing_manifest")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("'package.json' file not found in {./testdata/missing_manifest}"))
		})

		It("should successfully create a Blockcode", func() {
			bc, err := FromDir("./testdata/blockcode_example")
			Expect(err).To(BeNil())
			Expect(bc.code).ToNot(BeEmpty())
			Expect(bc.Manifest).ToNot(BeNil())
			Expect(bc.Manifest.Lang).To(Equal(Lang("go")))
			Expect(bc.Manifest.LangVersion).To(Equal("1.10.2"))
			Expect(bc.Manifest.PublicFuncs).To(Equal([]string{"some_func"}))
		})
	})

	Describe(".Bytes", func() {
		It("should return 305", func() {
			bc, err := FromDir("./testdata/blockcode_example")
			Expect(err).To(BeNil())
			Expect(bc.Size()).To(Equal(305))
		})
	})

	Describe(".Bytes", func() {
		It("should return bytes", func() {
			bc, err := FromDir("./testdata/blockcode_example")
			Expect(err).To(BeNil())
			Expect(bc.Bytes()).ToNot(BeEmpty())
		})
	})

	Describe(".Hash", func() {
		It("should return Hash", func() {
			bc, err := FromDir("./testdata/blockcode_example")
			Expect(err).To(BeNil())
			Expect(bc.Hash()).To(Equal(util.Hash{56, 35, 70, 128, 167, 183, 163, 31, 92, 65, 159, 242, 111, 193, 88, 182, 139, 48, 93, 65, 34, 12, 71, 8, 41, 192, 22, 89, 81, 52, 72, 77}))
		})
	})

	Describe(".ID", func() {
		It("should return ID", func() {
			bc, err := FromDir("./testdata/blockcode_example")
			Expect(err).To(BeNil())
			Expect(bc.ID()).To(Equal("0x38234680a7b7a31f5c419ff26fc158b68b305d41220c470829c016595134484d"))
		})
	})

	Describe(".FromBytes", func() {
		It("should return ID", func() {
			bc, err := FromDir("./testdata/blockcode_example")
			Expect(err).To(BeNil())
			bs := bc.Bytes()
			blockcode, err := FromBytes(bs)
			Expect(err).To(BeNil())
			Expect(blockcode).To(Equal(bc))
		})
	})

	// Describe(".Read", func() {
	// 	It("should return err = 'destination path does not exist' if destination path does not exist", func() {
	// 		bc, err := FromDir("./testdata/blockcode_example")
	// 		Expect(err).To(BeNil())
	// 		err = bc.Read("./unknown/path")
	// 		Expect(err).ToNot(BeNil())
	// 		Expect(err.Error()).To(Equal("destination path does not exist"))
	// 	})

	// 	It("should successfully un-tar to destination", func() {

	// 		destination := "/tmp/blockcode_example_untar"
	// 		err := os.Mkdir(destination, 0700)
	// 		Expect(err).To(BeNil())
	// 		defer os.RemoveAll(destination)

	// 		bc, err := FromDir("./testdata/blockcode_example")
	// 		Expect(err).To(BeNil())

	// 		err = bc.Read(destination)
	// 		Expect(err).To(BeNil())
	// 	})
	// })

})
