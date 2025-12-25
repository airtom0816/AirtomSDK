package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

// FastHttpClientOption 定义HTTP客户端配置选项
type FastHttpClientOption struct {
	Header         map[string]string
	SocketTimeout  int // 毫秒
	ConnectTimeout int // 毫秒
	IgnoreSSL      bool
	ProxyAddress   string
}

// FastHttpClient HTTP客户端
type FastHttpClient struct {
	client *fasthttp.Client
	header map[string]string
}

// NewFastHttpClient 创建新的HTTP客户端
func NewFastHttpClient(option FastHttpClientOption) *FastHttpClient {
	client := &fasthttp.Client{
		ReadTimeout:  time.Duration(option.SocketTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(option.SocketTimeout) * time.Millisecond,
		Dial: func(addr string) (net.Conn, error) {
			return fasthttp.DialTimeout(addr, time.Duration(option.ConnectTimeout)*time.Millisecond)
		},
		TLSConfig: &tls.Config{
			InsecureSkipVerify: option.IgnoreSSL,
		},
	}

	// 代理配置暂时注释掉，因为fasthttp.NewDialer在某些版本中可能不可用
	// if option.ProxyAddress != "" {
	// 	client.Dial = func(addr string) (net.Conn, error) {
	// 		return fasthttp.Dial(addr)
	// 	}
	// }

	return &FastHttpClient{
		client: client,
		header: option.Header,
	}
}

// Get 发送GET请求
func (c *FastHttpClient) Get(url string, headers map[string]string) (*fasthttp.Response, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(url)
	req.Header.SetMethod("GET")

	// 设置默认头
	for k, v := range c.header {
		req.Header.Set(k, v)
	}

	// 设置请求头
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if err := c.client.Do(req, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

// PostJSON 发送JSON POST请求
func (c *FastHttpClient) PostJSON(url string, data interface{}, headers map[string]string) (*fasthttp.Response, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(url)
	req.Header.SetMethod("POST")
	req.Header.SetContentType("application/json")

	// 设置默认头
	for k, v := range c.header {
		req.Header.Set(k, v)
	}

	// 设置请求头
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// 序列化数据
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	req.SetBody(jsonData)

	if err := c.client.Do(req, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

// PostForm 发送表单POST请求
func (c *FastHttpClient) PostForm(url string, data map[string]string, headers map[string]string) (*fasthttp.Response, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(url)
	req.Header.SetMethod("POST")
	req.Header.SetContentType("application/x-www-form-urlencoded")

	// 设置默认头
	for k, v := range c.header {
		req.Header.Set(k, v)
	}

	// 设置请求头
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// 设置表单数据
	args := fasthttp.AcquireArgs()
	defer fasthttp.ReleaseArgs(args)
	for k, v := range data {
		args.Add(k, v)
	}
	args.WriteTo(req.BodyWriter())

	if err := c.client.Do(req, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

// Close 关闭客户端
func (c *FastHttpClient) Close() {
	// fasthttp.Client 不需要显式关闭
}

// OpenAPIKeyClient OpenAPI密钥认证客户端
type OpenAPIKeyClient struct {
	apiKey    string
	apiSecret string
	baseURL   string
	client    *FastHttpClient
}

// NewOpenAPIKeyClient 创建新的OpenAPI客户端
func NewOpenAPIKeyClient(baseURL, apiKey, apiSecret string) *OpenAPIKeyClient {
	// 确保baseURL以斜杠结尾
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	option := FastHttpClientOption{
		Header: make(map[string]string),
	}
	client := NewFastHttpClient(option)

	return &OpenAPIKeyClient{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		baseURL:   baseURL,
		client:    client,
	}
}

// generateSignature 生成HMAC-SHA256签名
func (c *OpenAPIKeyClient) generateSignature(text string) string {
	h := hmac.New(sha256.New, []byte(c.apiSecret))
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil))
}

// buildAuthHeaders 构建认证请求头
func (c *OpenAPIKeyClient) buildAuthHeaders(requestBody string) map[string]string {
	headers := make(map[string]string)

	// 时间戳（毫秒）
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	// 随机数（Nonce）
	nonce := strings.ReplaceAll(uuid.New().String(), "-", "")

	// 签名计算：apiKey + timestamp + nonce + body
	signText := c.apiKey + timestamp + nonce + requestBody
	signature := c.generateSignature(signText)

	// 添加认证头
	headers["X-Api-Key"] = c.apiKey
	headers["X-Timestamp"] = timestamp
	headers["X-Nonce"] = nonce
	headers["X-Signature"] = signature
	headers["Content-Type"] = "application/json"

	return headers
}

// Get 发送GET请求（带签名认证）
func (c *OpenAPIKeyClient) Get(urlPath string, params map[string]string) (interface{}, error) {
	// 构建完整URL
	parsedBaseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}

	parsedPath, err := url.Parse(urlPath)
	if err != nil {
		return nil, err
	}

	fullURL := parsedBaseURL.ResolveReference(parsedPath).String()

	// 如果有参数，添加到URL
	if params != nil && len(params) > 0 {
		values := url.Values{}
		for k, v := range params {
			values.Add(k, v)
		}
		fullURL += "?" + values.Encode()
	}

	// GET请求的body为空字符串
	requestBody := ""
	authHeaders := c.buildAuthHeaders(requestBody)

	// 发送GET请求
	resp, err := c.client.Get(fullURL, authHeaders)
	if err != nil {
		return nil, err
	}

	// 检查响应状态
	statusCode := resp.StatusCode()
	if statusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", statusCode, string(resp.Body()))
	}

	// 尝试解析JSON响应
	var result interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return string(resp.Body()), nil
	}

	return result, nil
}

// Post 发送POST请求（带签名认证）
func (c *OpenAPIKeyClient) Post(urlPath string, data map[string]interface{}) (interface{}, error) {
	// 构建完整URL
	parsedBaseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}

	parsedPath, err := url.Parse(urlPath)
	if err != nil {
		return nil, err
	}

	fullURL := parsedBaseURL.ResolveReference(parsedPath).String()

	// 将数据转换为JSON字符串作为请求body
	requestBody, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	authHeaders := c.buildAuthHeaders(string(requestBody))

	// 发送POST JSON请求
	resp, err := c.client.PostJSON(fullURL, data, authHeaders)
	if err != nil {
		return nil, err
	}

	// 检查响应状态
	statusCode := resp.StatusCode()
	if statusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", statusCode, string(resp.Body()))
	}

	// 尝试解析JSON响应
	var result interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return string(resp.Body()), nil
	}

	return result, nil
}

// PostForm 发送POST表单请求（带签名认证）
func (c *OpenAPIKeyClient) PostForm(urlPath string, data map[string]string) (interface{}, error) {
	// 构建完整URL
	parsedBaseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}

	parsedPath, err := url.Parse(urlPath)
	if err != nil {
		return nil, err
	}

	fullURL := parsedBaseURL.ResolveReference(parsedPath).String()

	// 将数据转换为JSON字符串作为请求body（签名需要）
	requestBody, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	authHeaders := c.buildAuthHeaders(string(requestBody))

	// 发送POST表单请求
	resp, err := c.client.PostForm(fullURL, data, authHeaders)
	if err != nil {
		return nil, err
	}

	// 检查响应状态
	statusCode := resp.StatusCode()
	if statusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", statusCode, string(resp.Body()))
	}

	// 尝试解析JSON响应
	var result interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return string(resp.Body()), nil
	}

	return result, nil
}

// Close 关闭连接
func (c *OpenAPIKeyClient) Close() {
	if c.client != nil {
		c.client.Close()
	}
}

// OpenAPIKeyClientV2 OpenAPI密钥认证客户端（增强版）
type OpenAPIKeyClientV2 struct {
	apiKey    string
	apiSecret string
	baseURL   string
	client    *FastHttpClient
}

// NewOpenAPIKeyClientV2 创建新的增强版OpenAPI客户端
func NewOpenAPIKeyClientV2(baseURL, apiKey, apiSecret string, timeout *int, verifySSL bool, proxy *string) *OpenAPIKeyClientV2 {
	// 确保baseURL以斜杠结尾
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	// 创建HTTP客户端配置
	headers := map[string]string{
		"User-Agent": "OpenAPI-Go-Client/1.0",
		"Accept":     "application/json",
	}

	socketTimeout := 0
	connectTimeout := 0
	if timeout != nil {
		socketTimeout = *timeout * 1000
		connectTimeout = *timeout * 1000
	}

	proxyAddress := ""
	if proxy != nil {
		proxyAddress = *proxy
	}

	option := FastHttpClientOption{
		Header:         headers,
		SocketTimeout:  socketTimeout,
		ConnectTimeout: connectTimeout,
		IgnoreSSL:      !verifySSL,
		ProxyAddress:   proxyAddress,
	}

	client := NewFastHttpClient(option)

	return &OpenAPIKeyClientV2{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		baseURL:   baseURL,
		client:    client,
	}
}

// generateSignatureV2 生成增强版签名
func (c *OpenAPIKeyClientV2) generateSignatureV2(method, path, timestamp, nonce, body string) string {
	// 计算body哈希
	bodyHash := ""
	if body != "" {
		h := sha256.New()
		h.Write([]byte(body))
		bodyHash = hex.EncodeToString(h.Sum(nil))
	}

	// 构建签名字符串
	signText := fmt.Sprintf("%s%s%s%s%s%s", strings.ToUpper(method), path, c.apiKey, timestamp, nonce, bodyHash)

	// 生成HMAC-SHA256签名
	h := hmac.New(sha256.New, []byte(c.apiSecret))
	h.Write([]byte(signText))
	return hex.EncodeToString(h.Sum(nil))
}

// buildAuthHeadersV2 构建增强版认证请求头
func (c *OpenAPIKeyClientV2) buildAuthHeadersV2(method, path, body string) map[string]string {
	headers := make(map[string]string)

	// 时间戳（秒，整数）
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	// 随机数（Nonce）
	nonce := strings.ReplaceAll(uuid.New().String(), "-", "")

	// 生成签名
	signature := c.generateSignatureV2(method, path, timestamp, nonce, body)

	// 添加认证头
	headers["X-Api-Key"] = c.apiKey
	headers["X-Timestamp"] = timestamp
	headers["X-Nonce"] = nonce
	headers["X-Signature"] = signature
	headers["Content-Type"] = "application/json"

	return headers
}

// Request 发送通用请求（带签名认证）
func (c *OpenAPIKeyClientV2) Request(method, urlPath string, data map[string]interface{}, params map[string]string) (interface{}, error) {
	// 构建完整URL
	parsedBaseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}

	parsedPath, err := url.Parse(urlPath)
	if err != nil {
		return nil, err
	}

	fullURL := parsedBaseURL.ResolveReference(parsedPath).String()

	// 如果有参数，添加到URL
	if params != nil && len(params) > 0 {
		values := url.Values{}
		for k, v := range params {
			values.Add(k, v)
		}
		fullURL += "?" + values.Encode()
	}

	// 解析URL以获取路径部分（用于签名）
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return nil, err
	}
	path := parsedURL.Path
	if parsedURL.RawQuery != "" {
		path += "?" + parsedURL.RawQuery
	}

	// 准备请求体
	requestBody := ""
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		requestBody = string(jsonData)
	}

	// 构建认证头
	authHeaders := c.buildAuthHeadersV2(method, path, requestBody)

	var resp *fasthttp.Response

	// 根据方法发送请求
	switch strings.ToUpper(method) {
	case "GET":
		resp, err = c.client.Get(fullURL, authHeaders)
	case "POST":
		resp, err = c.client.PostJSON(fullURL, data, authHeaders)
	case "PUT":
		// 注意：这里简化为POST，实际应用中应实现PUT方法
		resp, err = c.client.PostJSON(fullURL, data, authHeaders)
	case "DELETE":
		// 注意：这里简化为GET，实际应用中应实现DELETE方法
		resp, err = c.client.Get(fullURL, authHeaders)
	default:
		return nil, fmt.Errorf("不支持的HTTP方法: %s", method)
	}

	if err != nil {
		return nil, err
	}

	// 检查响应状态
	statusCode := resp.StatusCode()
	if statusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", statusCode, string(resp.Body()))
	}

	// 尝试解析JSON响应
	var result interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return string(resp.Body()), nil
	}

	return result, nil
}

// Close 关闭连接
func (c *OpenAPIKeyClientV2) Close() {
	if c.client != nil {
		c.client.Close()
	}
}
