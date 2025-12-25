package com.example.openapi;

import com.fasterxml.jackson.core.JsonProcessingException;
import java.util.Base64;
import java.util.HashMap;
import java.util.Map;

/**
 * OpenAPI Token认证客户端
 */
public class OpenAPITokenClient implements AutoCloseable {
    protected String token;
    protected String baseUrl;
    protected HttpClient client;
    
    public OpenAPITokenClient(String baseUrl, String token) {
        this.token = token;
        this.baseUrl = baseUrl.endsWith("/") ? baseUrl : baseUrl + "/";
        
        Map<String, String> headers = new HashMap<>();
        headers.put("token", token);
        
        HttpClientOption option = new HttpClientOption();
        option.setHeader(headers);
        this.client = new HttpClient(option);
    }
    
    /**
     * 发送GET请求
     */
    public Object get(String url) throws Exception {
        String fullUrl = baseUrl + url.replaceFirst("^/", "");
        
        HttpResponse response = client.get(fullUrl, null);
        
        // 检查响应状态
        if (response.getStatusCode() >= 400) {
            throw new Exception(String.format("HTTP %d: %s", response.getStatusCode(), response.getText()));
        }
        
        // 尝试解析JSON响应
        try {
            return response.getJson();
        } catch (JsonProcessingException e) {
            return response.getText();
        }
    }
    
    /**
     * 发送POST请求
     */
    public Object post(String url, Map<String, Object> data) throws Exception {
        String fullUrl = baseUrl + url.replaceFirst("^/", "");
        
        HttpResponse response = client.postJson(fullUrl, data, null);
        
        // 检查响应状态
        if (response.getStatusCode() >= 400) {
            throw new Exception(String.format("HTTP %d: %s", response.getStatusCode(), response.getText()));
        }
        
        // 尝试解析JSON响应
        try {
            return response.getJson();
        } catch (JsonProcessingException e) {
            return response.getText();
        }
    }
    
    /**
     * 发送POST表单请求
     */
    public Object postForm(String url, Map<String, String> data) throws Exception {
        String fullUrl = baseUrl + url.replaceFirst("^/", "");
        
        HttpResponse response = client.postForm(fullUrl, data, null);
        
        // 检查响应状态
        if (response.getStatusCode() >= 400) {
            throw new Exception(String.format("HTTP %d: %s", response.getStatusCode(), response.getText()));
        }
        
        // 尝试解析JSON响应
        try {
            return response.getJson();
        } catch (JsonProcessingException e) {
            return response.getText();
        }
    }
    
    /**
     * 发送PUT请求
     */
    public Object put(String url, Map<String, Object> data) throws Exception {
        String fullUrl = baseUrl + url.replaceFirst("^/", "");
        
        Map<String, String> headers = new HashMap<>();
        headers.put("X-HTTP-Method-Override", "PUT");
        
        HttpResponse response = client.postJson(fullUrl, data, headers);
        
        // 检查响应状态
        if (response.getStatusCode() >= 400) {
            throw new Exception(String.format("HTTP %d: %s", response.getStatusCode(), response.getText()));
        }
        
        // 尝试解析JSON响应
        try {
            return response.getJson();
        } catch (JsonProcessingException e) {
            return response.getText();
        }
    }
    
    /**
     * 发送DELETE请求
     */
    public Object delete(String url) throws Exception {
        String fullUrl = baseUrl + url.replaceFirst("^/", "");
        
        Map<String, String> headers = new HashMap<>();
        headers.put("X-HTTP-Method-Override", "DELETE");
        
        HttpResponse response = client.get(fullUrl, headers);
        
        // 检查响应状态
        if (response.getStatusCode() >= 400) {
            throw new Exception(String.format("HTTP %d: %s", response.getStatusCode(), response.getText()));
        }
        
        // 尝试解析JSON响应
        try {
            return response.getJson();
        } catch (JsonProcessingException e) {
            return response.getText();
        }
    }
    
    /**
     * 关闭连接 - 实现AutoCloseable接口
     */
    @Override
    public void close() {
        if (client != null) {
            client.close();
        }
    }
}

/**
 * OpenAPI Token认证客户端（增强版）
 */
class OpenAPITokenClientV2 extends OpenAPITokenClient {
    private String authHeaderName;
    private String authHeaderFormat;
    
    public OpenAPITokenClientV2(String baseUrl, String token,
                               Integer timeout, Boolean verifySsl,
                               String proxy, String authHeaderName,
                               String authHeaderFormat) {
        super(baseUrl, token);
        this.authHeaderName = authHeaderName;
        this.authHeaderFormat = authHeaderFormat;
        
        // 创建HTTP客户端配置
        String authHeaderValue = String.format(authHeaderFormat, token);
        
        Map<String, String> headers = new HashMap<>();
        headers.put(authHeaderName, authHeaderValue);
        headers.put("User-Agent", "OpenAPI-Token-Client/1.0");
        headers.put("Accept", "application/json");
        
        HttpClientOption option = new HttpClientOption();
        option.setHeader(headers);
        
        if (timeout != null) {
            option.setSocketTimeout(timeout * 1000);
            option.setConnectTimeout(timeout * 1000);
        }
        
        option.setIgnoreSsl(!verifySsl);
        option.setProxyAddress(proxy);
        
        this.client = new HttpClient(option);
    }
    
    /**
     * 刷新认证令牌
     */
    public void refreshToken(String newToken) {
        this.token = newToken;
        
        // 更新认证头
        String authHeaderValue = String.format(authHeaderFormat, newToken);
        // 这里需要更新客户端的headers，但OkHttpClient的headers是immutable的
        // 在实际应用中，可能需要重新创建client或者使用interceptor
    }
    
    /**
     * 发送GET请求（支持多种认证类型）
     */
    public Object getWithAuthType(String url, String authType) throws Exception {
        Map<String, String> headers = new HashMap<>();
        
        // 根据认证类型设置不同的头
        switch (authType) {
            case "bearer":
                headers.put("Authorization", "Bearer " + token);
                break;
            case "basic":
                String encodedToken = Base64.getEncoder().encodeToString(token.getBytes());
                headers.put("Authorization", "Basic " + encodedToken);
                break;
            default: // 默认使用token
                headers.put("token", token);
                break;
        }
        
        String fullUrl = baseUrl + url.replaceFirst("^/", "");
        
        HttpResponse response = client.get(fullUrl, headers);
        
        // 检查响应状态
        if (response.getStatusCode() >= 400) {
            throw new Exception(String.format("HTTP %d: %s", response.getStatusCode(), response.getText()));
        }
        
        // 尝试解析JSON响应
        try {
            return response.getJson();
        } catch (JsonProcessingException e) {
            return response.getText();
        }
    }
    
    /**
     * 发送通用请求
     */
    public Object request(String method, String url,
                         Map<String, Object> data,
                         Map<String, String> params) throws Exception {
        String fullUrl = baseUrl + url.replaceFirst("^/", "");
        
        HttpResponse response;
        String methodUpper = method.toUpperCase();
        
        switch (methodUpper) {
            case "GET":
                response = client.get(fullUrl, null);
                break;
            case "POST":
                response = client.postJson(fullUrl, data != null ? data : new HashMap<>(), null);
                break;
            case "PUT":
                Map<String, String> putHeaders = new HashMap<>();
                putHeaders.put("X-HTTP-Method-Override", "PUT");
                response = client.postJson(fullUrl, data != null ? data : new HashMap<>(), putHeaders);
                break;
            case "DELETE":
                Map<String, String> deleteHeaders = new HashMap<>();
                deleteHeaders.put("X-HTTP-Method-Override", "DELETE");
                response = client.get(fullUrl, deleteHeaders);
                break;
            default:
                throw new IllegalArgumentException("不支持的HTTP方法: " + method);
        }
        
        // 检查响应状态
        if (response.getStatusCode() >= 400) {
            throw new Exception(String.format("HTTP %d: %s", response.getStatusCode(), response.getText()));
        }
        
        // 尝试解析JSON响应
        try {
            return response.getJson();
        } catch (JsonProcessingException e) {
            return response.getText();
        }
    }
}