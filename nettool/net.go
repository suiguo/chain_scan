package nettool

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

func GetMethod(url string, httpMethod string, method string, data []byte, head ...string) ([]byte, error) {
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
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
