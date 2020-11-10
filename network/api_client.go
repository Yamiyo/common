package network

import (
	"encoding/json"
	"errors"
	"fmt"

	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

type ApiClient struct {
	client *http.Client
	URL    string
	header map[string]string
}

func NewClient(url string, opts ...OptionFunc) *ApiClient {
	return newApiClient(url, opts...)
}

func newApiClient(url string, opts ...OptionFunc) *ApiClient {
	// 先取得預設的設定
	opt := defaultOption()

	// 如果需要特別設定時，複寫所需要的設定值
	for _, of := range opts {
		of(opt)
	}

	// 將設定帶入
	transport := http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   opt.timeout,
			KeepAlive: opt.keepAlive,
		}).DialContext,
		TLSHandshakeTimeout:   opt.tlsHandshakeTimeout,
		ExpectContinueTimeout: opt.expectContinueTimeout,
	}

	client := &http.Client{
		Transport: &transport,
	}

	c := &ApiClient{
		client: client,
		URL:    url,
		header: make(map[string]string),
	}

	c.client = client

	return c
}

// GetDataFromResponse returns nil if
// 1. json.Unmarshal fails OR
// 2. datakey is not presented in the response OR
// 3. data itself is null
func GetDataFromResponse(response []byte, dataKey string) (interface{}, error) {
	m := map[string]interface{}{}

	err := json.Unmarshal(response, &m)
	if err != nil {
		return nil, err
	}

	data, ok := m[dataKey]
	if !ok {
		return nil, fmt.Errorf("datakey is not presented in the response, dataKey: %s", dataKey)
	}

	if data == nil {
		return nil, fmt.Errorf("data itself is null")
	}

	return data, nil
}

func (c *ApiClient) AddHeader(key, value string) {
	c.header[key] = value
}

func (c *ApiClient) Get(data map[string]string) ([]byte, error) {
	form := url.Values{}
	addr := c.URL

	if data == nil {
		data = make(map[string]string)
	}

	for k, v := range data {
		form.Add(k, v)
	}

	if len(form) > 0 {
		if strings.IndexAny(addr, "?") > -1 {
			addr += "&" + form.Encode()
		} else {
			addr += "?" + form.Encode()
		}
	}
	return c.doRequest("GET", addr, nil, c.header)
}

func (c *ApiClient) doRequest(method, addr string, data map[string]string, header map[string]string) ([]byte, error) {
	form := url.Values{}

	if data != nil {
		for k, v := range data {
			log.Debugln(k, "->", v)
			form.Add(k, v)
		}
	}

	req, err := http.NewRequest(method, addr, strings.NewReader(form.Encode()))
	if err != nil {
		log.Errorln(err.Error())
		return nil, err
	}
	defer func() { req.Close = true }()

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/43.0.2357.81 Safari/537.36")
	req.Header.Set("Connection", "close")

	if header != nil {
		for k, v := range header {
			req.Header.Set(k, v)
		}
	}

	resp, err := c.client.Do(req)

	if err != nil {
		log.Errorln(method, addr, "->", err.Error())
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorln(method, addr, "->", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	log.Debugln(method, " ", addr, "->", string(b))

	if resp.StatusCode == http.StatusOK {
		return b, nil
	}

	return nil, errors.New(string(b))
}

func (c *ApiClient) Put(data map[string]string) ([]byte, error) {

	form := url.Values{}
	addr := c.URL

	if data == nil {
		data = make(map[string]string)
	}

	for k, v := range data {
		form.Add(k, v)
	}

	if len(form) > 0 {
		if strings.IndexAny(addr, "?") > -1 {
			addr += "&" + form.Encode()
		} else {
			addr += "?" + form.Encode()
		}
	}

	return c.doRequest("PUT", addr, data, c.header)
}

func (c *ApiClient) Post(data map[string]string) ([]byte, error) {

	form := url.Values{}
	addr := c.URL

	if data == nil {
		data = make(map[string]string)
	}

	for k, v := range data {
		form.Add(k, v)
	}

	if len(form) > 0 {
		if strings.IndexAny(addr, "?") > -1 {
			addr += "&" + form.Encode()
		} else {
			addr += "?" + form.Encode()
		}
	}

	return c.doRequest("POST", addr, data, c.header)
}

func (c *ApiClient) Del(data map[string]string) ([]byte, error) {
	form := url.Values{}
	addr := c.URL

	if data == nil {
		data = make(map[string]string)
	}

	for k, v := range data {
		form.Add(k, v)
	}

	if len(form) > 0 {
		if strings.IndexAny(addr, "?") > -1 {
			addr += "&" + form.Encode()
		} else {
			addr += "?" + form.Encode()
		}
	}

	return c.doRequest("DELETE", addr, nil, nil)
}
