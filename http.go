package nono

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

//HTTPGet 通过url地址和字典发送URL,并且使用result解析出返回值
func HTTPGet(url string, p map[string]interface{}, result interface{}) error {
	q := "?"
	if p != nil {
		for k, v := range p {
			if q == "?" {
				q = q + k + "=" + fmt.Sprint(v)
			} else {
				q = q + "&" + k + "=" + fmt.Sprint(v)
			}
		}
	}
	resp, err := http.Get(url + q)
	if err != nil {
		return err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &result)
	return err
}
