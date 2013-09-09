package mockhttp

import (
	"io"
	"strconv"
	"reflect"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"github.com/qiniu/log"
	"github.com/qiniu/api/auth/digest"
	"github.com/qiniu/rpc"
)

// --------------------------------------------------------------------

type mockServerRequestBody struct {
	reader io.Reader
	closeSignal bool
}

func (r *mockServerRequestBody) Read(p []byte) (int, error) {
	if r.closeSignal || r.reader == nil {
		return 0, io.EOF
	}
	return r.reader.Read(p)
}

func (r *mockServerRequestBody) Close() error {
	r.closeSignal = true
	if c, ok := r.reader.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// --------------------------------------------------------------------

var Route = map[string]http.Handler{}

// --------------------------------------------------------------------
// type Transport

type transportImpl struct {}

var Transport transportImpl

func (r transportImpl) RoundTrip(req *http.Request) (resp *http.Response, err error) {

	h := Route[req.Host]
	if h == nil {
		log.Fatal("Server not found:", req.Host)
	}

	cp := *req
	cp.Body = &mockServerRequestBody{req.Body, false}
	req = &cp

	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	ctlen := int64(-1)
	if v := rw.HeaderMap.Get("Content-Length"); v != "" {
		ctlen, _ = strconv.ParseInt(v, 10, 64)
	}

	return &http.Response{
		Status: "",
		StatusCode: rw.Code,
		Header: rw.HeaderMap,
		Body: ioutil.NopCloser(rw.Body),
		ContentLength: ctlen,
		TransferEncoding: nil,
		Close: false,
		Trailer: nil,
		Request: req,
	}, nil
}

var Client = rpc.Client{&http.Client{Transport: Transport}}

func ClientMac(key, secret string) rpc.Client {

	mac := &digest.Mac{key, []byte(secret)}
	tr := digest.NewTransport(mac, Transport)
	return rpc.Client{&http.Client{Transport: tr}}
}

// --------------------------------------------------------------------

func BindEx(host string, p interface{}, method string, args ...interface{}) {

	mux := http.NewServeMux()

	v := reflect.ValueOf(p)
	f := v.MethodByName(method)
	if !f.IsValid() {
		log.Fatal("mockhttp.Bind: method not found -", method)
	} else {
		in := make([]reflect.Value, len(args)+1)
		in[0] = reflect.ValueOf(mux)
		for i, arg := range args {
			in[i+1] = reflect.ValueOf(arg)
		}
		f.Call(in)
	}

	Route[host] = mux
}

func Bind(host string, p interface{}) {

	if p == nil {
		Route[host] = http.DefaultServeMux
		return
	}

	if h, ok := p.(http.Handler); ok {
		Route[host] = h
	} else {
		BindEx(host, p, "RegisterHandlers")
	}
}

// --------------------------------------------------------------------

