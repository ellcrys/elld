package blockcode

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Blockcode", func() {

	Describe(".validateManifest", func() {
		It("should validate correctly", func() {
			cases := map[*Manifest]error{
				&Manifest{}:                                                             fmt.Errorf("manifest error: language is missing"),
				&Manifest{Lang: "c++"}:                                                  fmt.Errorf("manifest error: language {c++} is not supported"),
				&Manifest{Lang: "go", LangVersion: ""}:                                  fmt.Errorf("manifest error: language version is required"),
				&Manifest{Lang: "go", LangVersion: "1.10.2"}:                            fmt.Errorf("manifest error: at least one public function is required"),
				&Manifest{Lang: "go", LangVersion: "1.10.2", PublicFuncs: []string{""}}: fmt.Errorf("manifest error: at least one public function is required"),
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

		It("should return error = 'manifest is malformed. invalid character ':' after top-level value' if manifest is not valid JSON", func() {
			_, err := FromDir("./testdata/invalid_manifest")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("manifest is malformed. invalid character ':' after top-level value"))
		})

		It("should return error = ''package.json' file not found in {./testdata/missing_manifest}' if package.json is missing", func() {
			_, err := FromDir("./testdata/missing_manifest")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("'package.json' file not found in {./testdata/missing_manifest}"))
		})

		It("should successfully create a Blockcode", func() {
			bc, err := FromDir("./testdata/blockcode_example")
			Expect(err).To(BeNil())
			Expect(bc.Code).ToNot(BeEmpty())
			Expect(bc.Manifest).ToNot(BeNil())
			Expect(bc.Manifest.Lang).ToNot(BeEmpty())
			Expect(bc.Manifest.LangVersion).ToNot(BeEmpty())
			Expect(bc.Manifest.PublicFuncs).ToNot(BeEmpty())
		})
	})

	Describe(".Len", func() {
		It("should return 3584", func() {
			bc, err := FromDir("./testdata/blockcode_example")
			Expect(err).To(BeNil())
			Expect(bc.Len()).To(Equal(3584))
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
			Expect(bc.Hash()).To(Equal([]byte{55, 94, 9, 239, 144, 166, 50, 60, 223, 17, 196, 34, 96, 176, 237, 216, 81, 77, 187, 244, 193, 245, 78, 143, 131, 224, 96, 15, 44, 243, 233, 110}))
		})
	})

	Describe(".ID", func() {
		It("should return ID", func() {
			bc, err := FromDir("./testdata/blockcode_example")
			Expect(err).To(BeNil())
			Expect(bc.ID()).To(Equal("375e09ef90a6323cdf11c42260b0edd8514dbbf4c1f54e8f83e0600f2cf3e96e"))
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

	Describe(".Read", func() {
		It("should return err = 'destination path does not exist' if destination path does not exist", func() {
			bc, err := FromDir("./testdata/blockcode_example")
			Expect(err).To(BeNil())
			err = bc.Read("./unknown/path")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("destination path does not exist"))
		})

		It("should successfully un-tar to destination", func() {

			destination := "/Users/ncodes/go/src/github.com/ellcrys/druid/blockcode/testdata/blockcode_example_untar"
			err := os.Mkdir(destination, 0700)
			Expect(err).To(BeNil())
			defer os.RemoveAll(destination)

			bc, err := FromDir("./testdata/blockcode_example")
			Expect(err).To(BeNil())

			err = bc.Read(destination)
			Expect(err).To(BeNil())

		})
	})
})
