package nettool

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

var ErrorStatus error = fmt.Errorf("status error")

func realRequest(url string, httpMethod string, method string, data []byte, head ...string) ([]byte, error) {
	client := &http.Client{}
	req, _ := http.NewRequest(httpMethod, url+method, bytes.NewBuffer(data))
	req.Header.Add("Accept", "application/json")
	tmp := make([]string, 2)
	for idx, h := range head {
		tmp[idx%2] = h
		if idx%2 == 1 {
			req.Header.Add(tmp[0], tmp[1])
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, ErrorStatus
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func GetMethod(url string, httpMethod string, method string, data []byte, head ...string) ([]byte, error) {
	for i := 0; i < 5; i++ {
		data, err := realRequest(url, httpMethod, method, data, head...)
		if err == nil {
			return data, err
		}
	}
	return nil, fmt.Errorf("try 5 times")
}
