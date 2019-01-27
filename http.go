package nono

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

//M 一个map[string]interface{}别名
type M map[string]interface{}

//HTTPGet 通过url地址和字典发送URL,并且使用result解析出返回值
func HTTPGet(url string, p map[string]interface{}, result interface{}) (string, error) {
	q := ""
	if p != nil && len(p) > 0 {
		q = "?"
		if p != nil {
			for k, v := range p {
				if q == "?" {
					q = q + k + "=" + fmt.Sprint(v)
				} else {
					q = q + "&" + k + "=" + fmt.Sprint(v)
				}
			}
		}
	}

	resp, err := http.Get(url + q)
	if err != nil {
		return "", err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	json.Unmarshal(b, &result)
	return string(b), err
}

//HTTPStr 1
type HTTPStr struct {
	req    *http.Request
	client http.Client
	err    error
}

//HTTP 1
func HTTP(u string, p map[string]interface{}) *HTTPStr {
	c := http.Client{}
	params := url.Values{}

	if p != nil && len(p) > 0 {
		for k, v := range p {
			params.Set(k, fmt.Sprint(v))
		}
	}
	req, err := http.NewRequest("GET", u+"?"+params.Encode(), nil)

	return &HTTPStr{req: req, client: c, err: err}
}

//Proxy 给请求加上代理，接受https://xxxx.com:1111的格式,其他的接受不了
func (t *HTTPStr) Proxy(p string) *HTTPStr {
	ii := url.URL{}
	proxy, err := ii.Parse(p)
	t.err = err
	trans := &http.Transport{
		Proxy: http.ProxyURL(proxy),
	}
	t.client.Transport = trans
	return t
}

//Header 写入header
func (t *HTTPStr) Header(hd map[string]string) *HTTPStr {
	for k, v := range hd {
		t.req.Header.Set(k, v)
	}
	return t
}

//Auth SetBasicAuth
func (t *HTTPStr) Auth(k, v string) *HTTPStr {
	t.req.SetBasicAuth(k, v)
	return t
}

//Get 结构
func (t *HTTPStr) Get() (*HTTPRespone, error) {
	if t.err != nil {
		return nil, t.err
	}
	rsp, err := t.client.Do(t.req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	return &HTTPRespone{URL: t.req.URL.String(), Byte: body}, err
}

//Post POST请求，参数是body内容
func (t *HTTPStr) Post(date []byte) (*HTTPRespone, error) {
	if t.err != nil {
		return nil, t.err
	}
	t.req.Method = "POST"
	t.req.Body = ioutil.NopCloser(bytes.NewReader(date))
	rsp, err := t.client.Do(t.req)
	if err != nil {
		return nil, err
	}

	defer rsp.Body.Close()
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	return &HTTPRespone{URL: t.req.URL.String(), Byte: body}, err
}

//HTTPRespone HTTP的返回值结构
type HTTPRespone struct {
	URL  string
	Byte []byte
}

//JSON 输入一个结构体，unmarshal这个结构并返回unmausharl的错误
func (t *HTTPRespone) JSON(result interface{}) error {
	return json.Unmarshal(t.Byte, &result)
}

//String 返回值的string格式
func (t *HTTPRespone) String() string {
	return string(t.Byte)
}
