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
		rpc = New("", "abc", false)
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
				Expect(rr.Code).To(Equal(400))
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
				Expect(rr.Code).To(Equal(400))
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
				Expect(rr.Code).To(Equal(404))
			})

			handler.ServeHTTP(rr, req)
		})

		It("should return 'Method not found' when json rpc method is not provided", func() {

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
				Expect(rr.Code).To(Equal(404))
			})

			handler.ServeHTTP(rr, req)
		})

		It("should return 'Method not found' error", func() {
			rpc.apiSet["add"] = APIInfo{
				Func: func(params interface{}) *Response {
					m := params.(map[string]interface{})
					return Success(m["x"].(float64) + m["y"].(float64))
				},
			}

			data, _ := json.Marshal(Request{
				JSONRPCVersion: "2.0",
				Method:         "plus",
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
				Expect(resp.Err).ToNot(BeNil())
				Expect(resp.Err.Code).To(Equal(-32601))
				Expect(resp.Err.Message).To(Equal("Method not found"))
				Expect(resp.Result).To(BeNil())
				Expect(rr.Code).To(Equal(404))
			})

			handler.ServeHTTP(rr, req)
		})

		Context("Successfully call method", func() {
			When("ID is added to the request body", func() {
				It("should return result", func() {
					rpc.apiSet["add"] = APIInfo{
						Namespace: "math",
						Func: func(params interface{}) *Response {
							m := params.(map[string]interface{})
							return Success(m["x"].(float64) + m["y"].(float64))
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
						Expect(resp.ID).To(Equal(uint64(1)))
						Expect(rr.Code).To(Equal(200))
					})

					handler.ServeHTTP(rr, req)
				})
			})

			When("ID is not added to the request body", func() {
				It("should not return result", func() {
					rpc.apiSet["add"] = APIInfo{
						Namespace: "math",
						Func: func(params interface{}) *Response {
							m := params.(map[string]interface{})
							return Success(m["x"].(float64) + m["y"].(float64))
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
						Expect(rr.Code).To(Equal(200))
					})

					handler.ServeHTTP(rr, req)
				})
			})
		})
	})

	Context("Call private method", func() {
		When("authorization is not set", func() {
			It("should return error response", func() {
				rpc.apiSet["echo"] = APIInfo{
					Private:   true,
					Namespace: "test",
					Func: func(params interface{}) *Response {
						return Success(params)
					},
				}

				data, _ := json.Marshal(Request{
					JSONRPCVersion: "2.0",
					Method:         "echo",
					Params:         map[string]interface{}{},
				})

				req, _ := http.NewRequest("POST", "/rpc", bytes.NewReader(data))

				rr := httptest.NewRecorder()
				rr.Header().Set("Content-Type", "application/json")

				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					resp := rpc.handle(w, r)
					Expect(resp.Err).ToNot(BeNil())
					Expect(resp.Err.Message).To(Equal("Invalid Request: Authorization header required"))
					Expect(resp.Err.Code).To(Equal(-32600))
					Expect(rr.Code).To(Equal(401))
				})

				handler.ServeHTTP(rr, req)
			})
		})

		When("authorization format invalid", func() {
			It("should return error response", func() {
				rpc.apiSet["echo"] = APIInfo{
					Private:   true,
					Namespace: "test",
					Func: func(params interface{}) *Response {
						return Success(params)
					},
				}

				data, _ := json.Marshal(Request{
					JSONRPCVersion: "2.0",
					Method:         "echo",
					Params:         map[string]interface{}{},
				})

				req, _ := http.NewRequest("POST", "/rpc", bytes.NewReader(data))
				req.Header.Set("Authorization", "bea")

				rr := httptest.NewRecorder()
				rr.Header().Set("Content-Type", "application/json")

				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					resp := rpc.handle(w, r)
					Expect(resp.Err).ToNot(BeNil())
					Expect(resp.Err.Message).To(Equal("Invalid Request: Authorization requires Bearer scheme"))
					Expect(resp.Err.Code).To(Equal(-32600))
					Expect(rr.Code).To(Equal(401))
				})

				handler.ServeHTTP(rr, req)
			})
		})

		When("authorization format is valid", func() {
			It("should return error response when bearer token is invalid", func() {
				rpc.apiSet["echo"] = APIInfo{
					Private:   true,
					Namespace: "test",
					Func: func(params interface{}) *Response {
						return Success(params)
					},
				}

				data, _ := json.Marshal(Request{
					JSONRPCVersion: "2.0",
					Method:         "echo",
					Params:         map[string]interface{}{},
				})

				req, _ := http.NewRequest("POST", "/rpc", bytes.NewReader(data))
				req.Header.Set("Authorization", "Bearer abcxyz")

				rr := httptest.NewRecorder()
				rr.Header().Set("Content-Type", "application/json")

				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					resp := rpc.handle(w, r)
					Expect(resp.Err).ToNot(BeNil())
					Expect(resp.Err.Message).To(Equal("Authorization Error: session token is not valid"))
					Expect(resp.Err.Code).To(Equal(-32600))
					Expect(rr.Code).To(Equal(401))
				})

				handler.ServeHTTP(rr, req)
			})
		})

		When("authorization format is valid", func() {
			It("should be successful when bearer token is valid", func() {
				rpc.apiSet["echo"] = APIInfo{
					Private:   true,
					Namespace: "test",
					Func: func(params interface{}) *Response {
						return Success(params)
					},
				}

				data, _ := json.Marshal(Request{
					JSONRPCVersion: "2.0",
					Method:         "echo",
					ID:             1,
					Params: map[string]interface{}{
						"age": 100,
					},
				})

				req, _ := http.NewRequest("POST", "/rpc", bytes.NewReader(data))
				req.Header.Set("Authorization", "Bearer "+MakeSessionToken("user1", rpc.sessionKey))

				rr := httptest.NewRecorder()
				rr.Header().Set("Content-Type", "application/json")

				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					resp := rpc.handle(w, r)
					Expect(resp.Err).To(BeNil())
					Expect(rr.Code).To(Equal(200))
				})

				handler.ServeHTTP(rr, req)
			})
		})

		When("authorization format is valid and bearer token is not valid", func() {
			It("should be successful when disableAuth is true", func() {
				rpc.disableAuth = true
				rpc.apiSet["echo"] = APIInfo{
					Private:   true,
					Namespace: "test",
					Func: func(params interface{}) *Response {
						return Success(params)
					},
				}

				data, _ := json.Marshal(Request{
					JSONRPCVersion: "2.0",
					Method:         "echo",
					ID:             1,
					Params: map[string]interface{}{
						"age": 100,
					},
				})

				req, _ := http.NewRequest("POST", "/rpc", bytes.NewReader(data))
				req.Header.Set("Authorization", "Bearer abcxyz")

				rr := httptest.NewRecorder()
				rr.Header().Set("Content-Type", "application/json")

				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					resp := rpc.handle(w, r)
					Expect(resp.Err).To(BeNil())
					Expect(rr.Code).To(Equal(200))
				})

				handler.ServeHTTP(rr, req)
			})
		})
	})

	Describe(".AddAPI", func() {
		Context("with no namespace provided", func() {
			It("should add API", func() {
				rpc.AddAPI("add", APIInfo{
					Func: func(params interface{}) *Response {
						m := params.(map[string]interface{})
						return Success(m["x"].(float64) + m["y"].(float64))
					},
				})
				Expect(rpc.apiSet).To(HaveLen(2))
				Expect(rpc.apiSet).To(HaveKey("_add"))
			})
		})

		Context("with a namespace provided", func() {
			It("should add API", func() {
				rpc.AddAPI("add", APIInfo{
					Namespace: "math",
					Func: func(params interface{}) *Response {
						m := params.(map[string]interface{})
						return Success(m["x"].(float64) + m["y"].(float64))
					},
				})
				Expect(rpc.apiSet).To(HaveLen(2))
				Expect(rpc.apiSet).To(HaveKey("math_add"))
			})
		})
	})

	Describe(".AddAPI", func() {
		It("should add API", func() {
			apiSet1 := APISet(map[string]APIInfo{
				"add": {
					Func: func(params interface{}) *Response {
						m := params.(map[string]interface{})
						return Success(m["x"].(float64) + m["y"].(float64))
					},
				},
			})
			apiSet2 := APISet(map[string]APIInfo{
				"add": {
					Func: func(params interface{}) *Response {
						m := params.(map[string]interface{})
						return Success(m["x"].(float64) + m["y"].(float64))
					},
				},
				"div": {
					Func: func(params interface{}) *Response {
						m := params.(map[string]interface{})
						return Success(m["x"].(float64) / m["y"].(float64))
					},
				},
			})
			rpc.MergeAPISet(apiSet1, apiSet2)
			Expect(rpc.apiSet).To(HaveLen(3))
		})
	})

	Describe(".Methods", func() {
		It("should return all methods name", func() {
			apiSet1 := APISet(map[string]APIInfo{
				"add": {
					Namespace: "math",
					Func: func(params interface{}) *Response {
						m := params.(map[string]interface{})
						return Success(m["x"].(float64) + m["y"].(float64))
					},
				},
			})
			apiSet2 := APISet(map[string]APIInfo{
				"add": {
					Namespace: "math",
					Func: func(params interface{}) *Response {
						m := params.(map[string]interface{})
						return Success(m["x"].(float64) + m["y"].(float64))
					},
				},
				"div": {
					Namespace: "math",
					Func: func(params interface{}) *Response {
						m := params.(map[string]interface{})
						return Success(m["x"].(float64) / m["y"].(float64))
					},
				},
			})
			rpc.MergeAPISet(apiSet1, apiSet2)
			m := rpc.Methods()
			Expect(m).To(HaveLen(3))
			expectedMethods := []string{"rpc_methods", "math_add", "math_div"}
			Expect(expectedMethods).To(ContainElement(m[0].Name))
			Expect(expectedMethods).To(ContainElement(m[1].Name))
			Expect(expectedMethods).To(ContainElement(m[2].Name))
		})
	})
})
