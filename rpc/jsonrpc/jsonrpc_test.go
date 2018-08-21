package jsonrpc

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Jsonrpc", func() {

	var rpc *JSONRPC

	BeforeEach(func() {
		rpc = New("")
	})

	Describe(".handle", func() {

		It("should return 'Parse error' when json is invalid", func() {

			data := []byte("{,}")
			req, _ := http.NewRequest("POST", "/rpc", bytes.NewReader(data))

			rr := httptest.NewRecorder()
			rr.Header().Set("Content-Type", "application/json")

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				resp := rpc.handle(w, r)
				Expect(resp.Err).ToNot(BeNil())
				Expect(resp.Err.Code).To(Equal(-32700))
				Expect(resp.Err.Message).To(Equal("Parse error"))
				Expect(resp.Result).To(BeNil())
			})

			handler.ServeHTTP(rr, req)
		})

		It("should return 'Invalid Request' when json rpc version is invalid", func() {

			data, _ := json.Marshal(Request{})
			req, _ := http.NewRequest("POST", "/rpc", bytes.NewReader(data))

			rr := httptest.NewRecorder()
			rr.Header().Set("Content-Type", "application/json")

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				resp := rpc.handle(w, r)
				Expect(resp.Err).ToNot(BeNil())
				Expect(resp.Err.Code).To(Equal(-32600))
				Expect(resp.Err.Message).To(Equal("Invalid Request"))
				Expect(resp.Result).To(BeNil())
			})

			handler.ServeHTTP(rr, req)
		})

		It("should return Method not found' when json rpc method is unknown", func() {

			data, _ := json.Marshal(Request{
				JSONRPCVersion: "2.0",
				Method:         "unknown",
			})
			req, _ := http.NewRequest("POST", "/rpc", bytes.NewReader(data))

			rr := httptest.NewRecorder()
			rr.Header().Set("Content-Type", "application/json")

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				resp := rpc.handle(w, r)
				Expect(resp.Err).ToNot(BeNil())
				Expect(resp.Err.Code).To(Equal(-32601))
				Expect(resp.Err.Message).To(Equal("Method not found"))
				Expect(resp.Result).To(BeNil())
			})

			handler.ServeHTTP(rr, req)
		})

		It("should return Method not found' when json rpc method is not provided", func() {

			data, _ := json.Marshal(Request{
				JSONRPCVersion: "2.0",
				Method:         "",
			})
			req, _ := http.NewRequest("POST", "/rpc", bytes.NewReader(data))

			rr := httptest.NewRecorder()
			rr.Header().Set("Content-Type", "application/json")

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				resp := rpc.handle(w, r)
				Expect(resp.Err).ToNot(BeNil())
				Expect(resp.Err.Code).To(Equal(-32601))
				Expect(resp.Err.Message).To(Equal("Method not found"))
				Expect(resp.Result).To(BeNil())
			})

			handler.ServeHTTP(rr, req)
		})

		Context("Successfully call method", func() {
			When("ID is added to the request body", func() {
				It("should return result", func() {
					rpc.apiSet["add"] = APIInfo{
						Func: func(params Params) *Response {
							return Success(params["x"].(float64) + params["y"].(float64))
						},
					}

					data, _ := json.Marshal(Request{
						JSONRPCVersion: "2.0",
						Method:         "add",
						Params: map[string]interface{}{
							"x": 2, "y": 2,
						},
						ID: 1,
					})

					req, _ := http.NewRequest("POST", "/rpc", bytes.NewReader(data))

					rr := httptest.NewRecorder()
					rr.Header().Set("Content-Type", "application/json")

					handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						resp := rpc.handle(w, r)
						Expect(resp.Err).To(BeNil())
						Expect(resp.Result).To(Equal(float64(4)))
						Expect(resp.ID).To(Equal(1))
					})

					handler.ServeHTTP(rr, req)
				})
			})

			When("ID is not added to the request body", func() {
				It("should not return result", func() {
					rpc.apiSet["add"] = APIInfo{
						Func: func(params Params) *Response {
							return Success(params["x"].(float64) + params["y"].(float64))
						},
					}

					data, _ := json.Marshal(Request{
						JSONRPCVersion: "2.0",
						Method:         "add",
						Params: map[string]interface{}{
							"x": 2, "y": 2,
						},
					})

					req, _ := http.NewRequest("POST", "/rpc", bytes.NewReader(data))

					rr := httptest.NewRecorder()
					rr.Header().Set("Content-Type", "application/json")

					handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						resp := rpc.handle(w, r)
						Expect(resp.Err).To(BeNil())
						Expect(resp.Result).To(BeNil())
						Expect(resp.ID).To(BeZero())
					})

					handler.ServeHTTP(rr, req)
				})
			})
		})
	})

	Describe(".AddAPI", func() {
		It("should add API", func() {
			rpc.AddAPI("add", APIInfo{
				Func: func(params Params) *Response {
					return Success(params["x"].(float64) + params["y"].(float64))
				},
			})
			Expect(rpc.apiSet).To(HaveLen(1))
		})
	})

	Describe(".AddAPI", func() {
		It("should add API", func() {
			apiSet1 := APISet(map[string]APIInfo{
				"add": APIInfo{
					Func: func(params Params) *Response {
						return Success(params["x"].(float64) + params["y"].(float64))
					},
				},
			})
			apiSet2 := APISet(map[string]APIInfo{
				"add": APIInfo{
					Func: func(params Params) *Response {
						return Success(params["x"].(float64) + params["y"].(float64))
					},
				},
				"div": APIInfo{
					Func: func(params Params) *Response {
						return Success(params["x"].(float64) / params["y"].(float64))
					},
				},
			})
			rpc.MergeAPISet(apiSet1, apiSet2)
			Expect(rpc.apiSet).To(HaveLen(2))
		})
	})
})
