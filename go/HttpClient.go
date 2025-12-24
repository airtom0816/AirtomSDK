package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HttpClientOption HTTP客户端配置选项
type HttpClientOption struct {
	Header                   map[string]string
	Cookie                   map[string]string
	ProxyAddress             string
	SocketTimeout            int // 毫秒
	ConnectTimeout           int // 毫秒
	IgnoreSSL                bool
	defaultSocketTimeout     int // 秒
	defaultConnectTimeout    int // 秒
	connectionRequestTimeout int // 秒
}

// NewHttpClientOption 创建默认的HTTP客户端配置
func NewHttpClientOption() *HttpClientOption {
	return &HttpClientOption{
		Header:                   make(map[string]string),
		Cookie:                   make(map[string]string),
		IgnoreSSL:                true,
		defaultSocketTimeout:     60,
		defaultConnectTimeout:    60,
		connectionRequestTimeout: 5,
	}
}

// HTTPResponse HTTP响应封装
type HTTPResponse struct {
	RequestMethod string
	RequestURL    string
	StatusCode    int
	Headers       map[string][]string
	Content       []byte
	ElapsedTime   float64 // 毫秒
}

// Text 获取响应文本
func (r *HTTPResponse) Text() string {
	return string(r.Content)
}

// JSON 解析JSON响应
func (r *HTTPResponse) JSON(v interface{}) error {
	return json.Unmarshal(r.Content, v)
}

func (r *HTTPResponse) String() string {
	return fmt.Sprintf("HTTPResponse(status_code=%d, elapsed_time=%.2fms, content_length=%d)",
		r.StatusCode, r.ElapsedTime, len(r.Content))
}

// HttpClient HTTP客户端
type HttpClient struct {
	option *HttpClientOption
	client *http.Client
}

// NewHttpClient 创建HTTP客户端
func NewHttpClient(option *HttpClientOption) *HttpClient {
	if option == nil {
		option = NewHttpClientOption()
	}

	// 创建HTTP客户端
	client := &http.Client{}

	// 配置超时
	connectTimeout := time.Duration(option.ConnectTimeout) * time.Millisecond
	if option.ConnectTimeout == 0 {
		connectTimeout = time.Duration(option.defaultConnectTimeout) * time.Second
	}

	socketTimeout := time.Duration(option.SocketTimeout) * time.Millisecond
	if option.SocketTimeout == 0 {
		socketTimeout = time.Duration(option.defaultSocketTimeout) * time.Second
	}

	client.Timeout = connectTimeout + socketTimeout

	// 配置Transport
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   connectTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          200,
		MaxIdleConnsPerHost:   20,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxConnsPerHost:       20,
	}

	// 配置代理
	if option.ProxyAddress != "" {
		proxyURL, err := url.Parse("http://" + option.ProxyAddress)
		if err != nil {
			panic(fmt.Sprintf("无效的代理地址: %v", err))
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	// 配置SSL验证
	if option.IgnoreSSL {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	client.Transport = transport

	return &HttpClient{
		option: option,
		client: client,
	}
}

// buildRequest 构建基础请求
func (c *HttpClient) buildRequest(method, urlStr string, body io.Reader, headers map[string]string) (*http.Request, error) {
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}

	// 设置默认头
	for k, v := range c.option.Header {
		req.Header.Set(k, v)
	}

	// 设置额外头
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	// 设置Cookie
	for k, v := range c.option.Cookie {
		req.AddCookie(&http.Cookie{Name: k, Value: v})
	}

	return req, nil
}

// doRequest 执行请求并返回响应
func (c *HttpClient) doRequest(req *http.Request) (*HTTPResponse, error) {
	startTime := time.Now()

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应内容
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 计算耗时(毫秒)
	elapsedTime := float64(time.Since(startTime).Microseconds()) / 1000

	return &HTTPResponse{
		RequestMethod: req.Method,
		RequestURL:    req.URL.String(),
		StatusCode:    resp.StatusCode,
		Headers:       resp.Header,
		Content:       content,
		ElapsedTime:   elapsedTime,
	}, nil
}

// Get 发送GET请求
func (c *HttpClient) Get(urlStr string, headers map[string]string) (*HTTPResponse, error) {
	req, err := c.buildRequest("GET", urlStr, nil, headers)
	if err != nil {
		return nil, err
	}

	return c.doRequest(req)
}

// PostJSON 发送POST JSON请求
func (c *HttpClient) PostJSON(urlStr string, data interface{}, headers map[string]string) (*HTTPResponse, error) {
	// 序列化JSON数据
	var jsonData []byte
	var err error

	switch v := data.(type) {
	case string:
		jsonData = []byte(v)
	default:
		jsonData, err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
	}

	// 准备请求头
	reqHeaders := make(map[string]string)
	reqHeaders["Content-Type"] = "application/json; charset=utf-8"

	// 合并传入的头
	if headers != nil {
		for k, v := range headers {
			reqHeaders[k] = v
		}
	}

	req, err := c.buildRequest("POST", urlStr, bytes.NewBuffer(jsonData), reqHeaders)
	if err != nil {
		return nil, err
	}

	return c.doRequest(req)
}

// PostForm 发送POST表单请求
func (c *HttpClient) PostForm(urlStr string, data map[string]string, headers map[string]string) (*HTTPResponse, error) {
	// 准备表单数据
	formData := url.Values{}
	for k, v := range data {
		formData.Set(k, v)
	}

	// 准备请求头
	reqHeaders := make(map[string]string)
	reqHeaders["Content-Type"] = "application/x-www-form-urlencoded; charset=UTF-8"

	// 合并传入的头
	if headers != nil {
		for k, v := range headers {
			reqHeaders[k] = v
		}
	}

	req, err := c.buildRequest("POST", urlStr, strings.NewReader(formData.Encode()), reqHeaders)
	if err != nil {
		return nil, err
	}

	return c.doRequest(req)
}

// GetJSON 发送带JSON体的GET请求
func (c *HttpClient) GetJSON(urlStr string, data interface{}) (*HTTPResponse, error) {
	// 序列化JSON数据
	var jsonData []byte
	var err error

	switch v := data.(type) {
	case string:
		jsonData = []byte(v)
	default:
		jsonData, err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
	}

	headers := map[string]string{
		"Content-Type": "application/json; charset=utf-8",
	}

	req, err := c.buildRequest("GET", urlStr, bytes.NewBuffer(jsonData), headers)
	if err != nil {
		return nil, err
	}

	return c.doRequest(req)
}

// File 用于上传的文件结构
type File struct {
	Filename string
	Content  []byte
}

// UploadFiles 上传文件
func (c *HttpClient) UploadFiles(urlStr string, files []File) (*HTTPResponse, error) {
	// 构建multipart/form-data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for i, file := range files {
		part, err := writer.CreateFormFile(fmt.Sprintf("file_%d", i), file.Filename)
		if err != nil {
			return nil, err
		}
		_, err = part.Write(file.Content)
		if err != nil {
			return nil, err
		}
	}

	err := writer.Close()
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		"Content-Type": writer.FormDataContentType(),
	}

	req, err := c.buildRequest("POST", urlStr, body, headers)
	if err != nil {
		return nil, err
	}

	return c.doRequest(req)
}

// Close 关闭客户端（在Go中HTTP客户端通常不需要显式关闭）
func (c *HttpClient) Close() {
	// 对于http.Client，没有需要显式关闭的资源
	// 如果需要关闭连接，可以配置Transport的CloseIdleConnections
	if transport, ok := c.client.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
}
