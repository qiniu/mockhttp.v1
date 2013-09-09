package foo

import (
	"fmt"
	"testing"
	"github.com/qiniu/mockhttp"
)

func TestBasic(t *testing.T) {

	svr := new(Service)
	mockhttp.Bind("foo.com", svr)

	c := mockhttp.Client
	{
		var foo FooRet
		err := c.Call(nil, &foo, "http://foo.com/foo")
		if err != nil {
			t.Fatal("call foo failed:", err)
		}
		if foo.A != 1 || foo.B != "foo.com" || foo.C != "/foo" {
			t.Fatal("call foo: invalid ret")
		}
		fmt.Println(foo)
	}
	{
		var ret map[string]string
		err := c.Call(nil, &ret, "http://foo.com/bar")
		if err != nil {
			t.Fatal("call foo failed:", err)
		}
		if ret["foo"] != "1" || ret["bar"] != "2" {
			t.Fatal("call bar: invalid ret")
		}
		fmt.Println(ret)
	}
}

// --------------------------------------------------------------------

