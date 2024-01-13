package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

func GetHttpRequest(url string, header map[string]string) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("accept", "application/json")
	if header != nil {
		for k, v := range header {
			req.Header.Add(k, v)
		}
	}
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	response, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	return io.ReadAll(response.Body)
}

// PostHttpRequest 发送Post请求，json格式数据
func PostHttpRequest(url string, header map[string]string, payload []byte) ([]byte, error) {
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if nil != err {
		return nil, err
	}

	if header != nil {
		for k, v := range header {
			request.Header.Add(k, v)
		}
	}
	request.Header.Set("Content-Type", "application/json")
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	if response != nil {
		defer func(Body io.ReadCloser) {
			err = Body.Close()
			if err != nil {
				fmt.Errorf("close body error: %v", err)
			}
		}(response.Body)
	}
	return io.ReadAll(response.Body)
}

// PutHttpRequest 发送Put请求，json格式数据
func PutHttpRequest(url string, header map[string]string, payload []byte) ([]byte, error) {
	request, err := http.NewRequest("PUT", url, bytes.NewBuffer(payload))
	if nil != err {
		return nil, err
	}

	if header != nil {
		for k, v := range header {
			request.Header.Add(k, v)
		}
	}
	request.Header.Set("Content-Type", "application/json")
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	if response != nil {
		defer func(Body io.ReadCloser) {
			err = Body.Close()
			if err != nil {
				fmt.Errorf("close body error: %v", err)
			}
		}(response.Body)
	}
	return io.ReadAll(response.Body)
}
