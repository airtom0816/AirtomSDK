package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TokenHttpClientOption 定义HTTP客户端配置选项
type TokenHttpClientOption struct {
	Header         map[string]string
	SocketTimeout  int
	ConnectTimeout int
	IgnoreSSL      bool
	ProxyAddress   string
}

// TokenHttpClient HTTP客户端
type TokenHttpClient struct {
	client *http.Client
	header map[string]string
}

// NewTokenHttpClient 创建新的HTTP客户端
func NewTokenHttpClient(option TokenHttpClientOption) *TokenHttpClient {
	transport := &http.Transport{}

	// 配置代理
	if option.ProxyAddress != "" {
		proxyURL, err := url.Parse(option.ProxyAddress)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	// 配置SSL验证
	if option.IgnoreSSL {
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}
		transport.TLSClientConfig.InsecureSkipVerify = true
	}

	// 配置超时
	client := &http.Client{
		Transport: transport,
	}

	if option.ConnectTimeout > 0 {
		client.Timeout = time.Duration(option.ConnectTimeout) * time.Millisecond
	}

	return &TokenHttpClient{
		client: client,
		header: option.Header,
	}
}

// Get 发送GET请求
func (c *TokenHttpClient) Get(url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	c.setHeaders(req, headers)

	return c.client.Do(req)
}

// PostJSON 发送JSON POST请求
func (c *TokenHttpClient) PostJSON(url string, data interface{}, headers map[string]string) (*http.Response, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	// 设置默认Content-Type
	req.Header.Set("Content-Type", "application/json")

	// 设置请求头
	c.setHeaders(req, headers)

	return c.client.Do(req)
}

// PostForm 发送表单POST请求
func (c *TokenHttpClient) PostForm(urlPath string, data map[string]string, headers map[string]string) (*http.Response, error) {
	// 创建表单数据
	var formBuf bytes.Buffer
	for k, v := range data {
		if formBuf.Len() > 0 {
			formBuf.WriteString("&")
		}
		formBuf.WriteString(url.QueryEscape(k))
		formBuf.WriteString("=")
		formBuf.WriteString(url.QueryEscape(v))
	}

	req, err := http.NewRequest("POST", urlPath, &formBuf)
	if err != nil {
		return nil, err
	}

	// 设置默认Content-Type
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 设置请求头
	c.setHeaders(req, headers)

	return c.client.Do(req)
}

// Close 关闭客户端（Go的http.Client不需要显式关闭，这里留空作为兼容）
func (c *TokenHttpClient) Close() {
	// 无需操作，Go的http.Client会自动管理连接
}

// 设置请求头
func (c *TokenHttpClient) setHeaders(req *http.Request, headers map[string]string) {
	// 设置客户端默认头
	for k, v := range c.header {
		req.Header.Set(k, v)
	}

	// 设置请求特定头（会覆盖默认头）
	for k, v := range headers {
		req.Header.Set(k, v)
	}
}

// OpenAPITokenClient OpenAPI Token认证客户端
type OpenAPITokenClient struct {
	token   string
	baseURL string
	client  *TokenHttpClient
}

// NewOpenAPITokenClient 创建新的OpenAPITokenClient
func NewOpenAPITokenClient(baseURL, token string) *OpenAPITokenClient {
	headers := map[string]string{
		"token": token,
	}

	option := TokenHttpClientOption{
		Header: headers,
	}

	client := NewTokenHttpClient(option)

	// 确保baseURL以斜杠结尾
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	return &OpenAPITokenClient{
		token:   token,
		baseURL: baseURL,
		client:  client,
	}
}

// Get 发送GET请求
func (c *OpenAPITokenClient) Get(urlPath string, params map[string]string) (interface{}, error) {
	fullURL, err := c.buildURL(urlPath, params)
	if err != nil {
		return nil, err
	}

	response, err := c.client.Get(fullURL, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return c.handleResponse(response)
}

// Post 发送POST请求
func (c *OpenAPITokenClient) Post(urlPath string, data map[string]interface{}) (interface{}, error) {
	fullURL, err := c.buildURL(urlPath, nil)
	if err != nil {
		return nil, err
	}

	response, err := c.client.PostJSON(fullURL, data, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return c.handleResponse(response)
}

// PostForm 发送POST表单请求
func (c *OpenAPITokenClient) PostForm(urlPath string, data map[string]string) (interface{}, error) {
	fullURL, err := c.buildURL(urlPath, nil)
	if err != nil {
		return nil, err
	}

	response, err := c.client.PostForm(fullURL, data, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return c.handleResponse(response)
}

// Put 发送PUT请求
func (c *OpenAPITokenClient) Put(urlPath string, data map[string]interface{}) (interface{}, error) {
	fullURL, err := c.buildURL(urlPath, nil)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		"X-HTTP-Method-Override": "PUT",
	}

	response, err := c.client.PostJSON(fullURL, data, headers)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return c.handleResponse(response)
}

// Delete 发送DELETE请求
func (c *OpenAPITokenClient) Delete(urlPath string) (interface{}, error) {
	fullURL, err := c.buildURL(urlPath, nil)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		"X-HTTP-Method-Override": "DELETE",
	}

	response, err := c.client.Get(fullURL, headers)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return c.handleResponse(response)
}

// Close 关闭客户端
func (c *OpenAPITokenClient) Close() {
	c.client.Close()
}

// 构建完整URL
func (c *OpenAPITokenClient) buildURL(urlPath string, params map[string]string) (string, error) {
	// 处理相对路径
	parsedBaseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return "", err
	}

	parsedPath, err := url.Parse(urlPath)
	if err != nil {
		return "", err
	}

	fullURL := parsedBaseURL.ResolveReference(parsedPath).String()

	// 添加查询参数
	if params != nil && len(params) > 0 {
		parsedURL, err := url.Parse(fullURL)
		if err != nil {
			return "", err
		}

		q := parsedURL.Query()
		for k, v := range params {
			q.Add(k, v)
		}
		parsedURL.RawQuery = q.Encode()
		fullURL = parsedURL.String()
	}

	return fullURL, nil
}

// 处理响应
func (c *OpenAPITokenClient) handleResponse(response *http.Response) (interface{}, error) {
	if response.StatusCode >= 400 {
		body, _ := io.ReadAll(response.Body)
		return nil, fmt.Errorf("HTTP %d: %s", response.StatusCode, string(body))
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// 尝试解析JSON
	var jsonData interface{}
	if err := json.Unmarshal(body, &jsonData); err == nil {
		return jsonData, nil
	}

	// 不是JSON则返回原始文本
	return string(body), nil
}

// OpenAPITokenClientV2 OpenAPI Token认证客户端（增强版）
type OpenAPITokenClientV2 struct {
	OpenAPITokenClient
	authHeaderName   string
	authHeaderFormat string
}

// NewOpenAPITokenClientV2 创建新的OpenAPITokenClientV2
func NewOpenAPITokenClientV2(baseURL, token string, timeout *int, verifySSL bool, proxy string, authHeaderName, authHeaderFormat string) *OpenAPITokenClientV2 {
	// 格式化认证头
	authHeaderValue := fmt.Sprintf(authHeaderFormat, token)

	headers := map[string]string{
		authHeaderName: authHeaderValue,
		"User-Agent":   "OpenAPI-Token-Client/1.0",
		"Accept":       "application/json",
	}

	socketTimeout := 0
	connectTimeout := 0
	if timeout != nil {
		socketTimeout = *timeout * 1000
		connectTimeout = *timeout * 1000
	}

	option := TokenHttpClientOption{
		Header:         headers,
		SocketTimeout:  socketTimeout,
		ConnectTimeout: connectTimeout,
		IgnoreSSL:      !verifySSL,
		ProxyAddress:   proxy,
	}

	client := NewTokenHttpClient(option)

	// 确保baseURL以斜杠结尾
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	return &OpenAPITokenClientV2{
		OpenAPITokenClient: OpenAPITokenClient{
			token:   token,
			baseURL: baseURL,
			client:  client,
		},
		authHeaderName:   authHeaderName,
		authHeaderFormat: authHeaderFormat,
	}
}

// RefreshToken 刷新认证令牌
func (c *OpenAPITokenClientV2) RefreshToken(newToken string) {
	c.token = newToken
	authHeaderValue := fmt.Sprintf(c.authHeaderFormat, newToken)
	c.client.header[c.authHeaderName] = authHeaderValue
}

// GetWithAuthType 发送GET请求（支持多种认证类型）
func (c *OpenAPITokenClientV2) GetWithAuthType(urlPath string, authType string) (interface{}, error) {
	var headers map[string]string

	switch authType {
	case "bearer":
		headers = map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", c.token),
		}
	case "basic":
		encodedToken := base64.StdEncoding.EncodeToString([]byte(c.token))
		headers = map[string]string{
			"Authorization": fmt.Sprintf("Basic %s", encodedToken),
		}
	default:
		headers = map[string]string{
			"token": c.token,
		}
	}

	fullURL, err := c.buildURL(urlPath, nil)
	if err != nil {
		return nil, err
	}

	response, err := c.client.Get(fullURL, headers)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return c.handleResponse(response)
}

// Request 发送通用请求
func (c *OpenAPITokenClientV2) Request(method, urlPath string, data map[string]interface{}, params map[string]string) (interface{}, error) {
	fullURL, err := c.buildURL(urlPath, params)
	if err != nil {
		return nil, err
	}

	var response *http.Response

	switch strings.ToUpper(method) {
	case "GET":
		response, err = c.client.Get(fullURL, nil)
	case "POST":
		response, err = c.client.PostJSON(fullURL, data, nil)
	case "PUT":
		headers := map[string]string{
			"X-HTTP-Method-Override": "PUT",
		}
		response, err = c.client.PostJSON(fullURL, data, headers)
	case "DELETE":
		headers := map[string]string{
			"X-HTTP-Method-Override": "DELETE",
		}
		response, err = c.client.Get(fullURL, headers)
	default:
		return nil, fmt.Errorf("不支持的HTTP方法: %s", method)
	}

	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return c.handleResponse(response)
}

// TokenManager Token管理器，支持自动刷新
type TokenManager struct {
	baseURL         string
	currentToken    string
	refreshURL      string
	refreshInterval int
	lastRefreshTime int64
	client          *OpenAPITokenClientV2
}

// NewTokenManager 创建新的TokenManager
func NewTokenManager(baseURL, token, refreshURL string, refreshInterval int) *TokenManager {
	if refreshInterval <= 0 {
		refreshInterval = 3600 // 默认1小时
	}

	client := NewOpenAPITokenClientV2(
		baseURL,
		token,
		nil,
		true,
		"",
		"Authorization",
		"Bearer {}",
	)

	return &TokenManager{
		baseURL:         baseURL,
		currentToken:    token,
		refreshURL:      refreshURL,
		refreshInterval: refreshInterval,
		lastRefreshTime: 0,
		client:          client,
	}
}

// ShouldRefresh 检查是否应该刷新令牌
func (m *TokenManager) ShouldRefresh() bool {
	if m.refreshURL == "" {
		return false
	}

	currentTime := time.Now().Unix()
	return currentTime-m.lastRefreshTime > int64(m.refreshInterval)
}

// Refresh 刷新令牌
func (m *TokenManager) Refresh() bool {
	if m.refreshURL == "" {
		return false
	}

	data := map[string]interface{}{
		"token": m.currentToken,
	}

	response, err := m.client.Post(m.refreshURL, data)
	if err != nil {
		fmt.Printf("刷新令牌失败: %v\n", err)
		return false
	}

	// 尝试解析响应获取新令牌
	responseMap, ok := response.(map[string]interface{})
	if !ok {
		return false
	}

	newToken, ok := responseMap["access_token"].(string)
	if !ok {
		return false
	}

	m.currentToken = newToken
	m.client.RefreshToken(newToken)
	m.lastRefreshTime = time.Now().Unix()
	return true
}

// Get 发送GET请求（支持自动刷新令牌）
func (m *TokenManager) Get(urlPath string, autoRefresh bool) (interface{}, error) {
	if autoRefresh && m.ShouldRefresh() {
		m.Refresh()
	}

	return m.client.Get(urlPath, nil)
}

// Post 发送POST请求（支持自动刷新令牌）
func (m *TokenManager) Post(urlPath string, data map[string]interface{}, autoRefresh bool) (interface{}, error) {
	if autoRefresh && m.ShouldRefresh() {
		m.Refresh()
	}

	return m.client.Post(urlPath, data)
}

// Close 关闭连接
func (m *TokenManager) Close() {
	m.client.Close()
}
