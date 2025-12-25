package com.example.openapi;

import com.fasterxml.jackson.core.JsonProcessingException;
import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;
import java.nio.charset.StandardCharsets;
import java.util.HashMap;
import java.util.Map;
import java.util.UUID;

/**
 * OpenAPI 密钥认证客户端
 */
public class OpenAPIKeyClient implements AutoCloseable {
    private String apiKey;
    private String apiSecret;
    private String baseUrl;
    private HttpClient client;
    
    public OpenAPIKeyClient(String baseUrl, String apiKey, String apiSecret) {
        this.apiKey = apiKey;
        this.apiSecret = apiSecret;
        this.baseUrl = baseUrl.endsWith("/") ? baseUrl : baseUrl + "/";
        
        HttpClientOption option = new HttpClientOption();
        option.setHeader(new HashMap<>());
        this.client = new HttpClient(option);
    }
    
    /**
     * 生成HMAC-SHA256签名
     */
    private String generateSignature(String text) throws Exception {
        Mac sha256HMAC = Mac.getInstance("HmacSHA256");
        SecretKeySpec secretKey = new SecretKeySpec(apiSecret.getBytes(StandardCharsets.UTF_8), "HmacSHA256");
        sha256HMAC.init(secretKey);
        
        byte[] hash = sha256HMAC.doFinal(text.getBytes(StandardCharsets.UTF_8));
        
        StringBuilder hexString = new StringBuilder();
        for (byte b : hash) {
            String hex = Integer.toHexString(0xff & b);
            if (hex.length() == 1) {
                hexString.append('0');
            }
            hexString.append(hex);
        }
        return hexString.toString();
    }
    
    /**
     * 构建认证请求头
     */
    private Map<String, String> buildAuthHeaders(String requestBody) throws Exception {
        Map<String, String> headers = new HashMap<>();
        
        // 时间戳（毫秒）
        String timestamp = String.valueOf(System.currentTimeMillis());
        
        // 随机数（Nonce）
        String nonce = UUID.randomUUID().toString().replace("-", "");
        
        // 签名计算：apiKey + timestamp + nonce + body
        String signText = apiKey + timestamp + nonce + requestBody;
        String signature = generateSignature(signText);
        
        // 添加认证头
        headers.put("X-Api-Key", apiKey);
        headers.put("X-Timestamp", timestamp);
        headers.put("X-Nonce", nonce);
        headers.put("X-Signature", signature);
        headers.put("Content-Type", "application/json");
        
        return headers;
    }
    
    /**
     * 发送GET请求
     */
    public Object get(String url) throws Exception {
        String fullUrl = baseUrl + url.replaceFirst("^/", "");
        
        // GET请求的body为空字符串
        String requestBody = "";
        Map<String, String> authHeaders = buildAuthHeaders(requestBody);
        
        HttpResponse response = client.get(fullUrl, authHeaders);
        
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
        
        // 将数据转换为JSON字符串作为请求body
        com.fasterxml.jackson.databind.ObjectMapper objectMapper = new com.fasterxml.jackson.databind.ObjectMapper();
        String requestBody = objectMapper.writeValueAsString(data);
        Map<String, String> authHeaders = buildAuthHeaders(requestBody);
        
        HttpResponse response = client.postJson(fullUrl, data, authHeaders);
        
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
        
        // 将数据转换为JSON字符串作为请求body（签名需要）
        com.fasterxml.jackson.databind.ObjectMapper objectMapper = new com.fasterxml.jackson.databind.ObjectMapper();
        String requestBody = objectMapper.writeValueAsString(data);
        Map<String, String> authHeaders = buildAuthHeaders(requestBody);
        
        HttpResponse response = client.postForm(fullUrl, data, authHeaders);
        
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
 * OpenAPI密钥认证客户端（增强版）
 */
class OpenAPIKeyClientV2 implements AutoCloseable {
    private String apiKey;
    private String apiSecret;
    private String baseUrl;
    private HttpClient client;
    
    public OpenAPIKeyClientV2(String baseUrl, String apiKey, String apiSecret,
                              Integer timeout, Boolean verifySsl, String proxy) {
        this.apiKey = apiKey;
        this.apiSecret = apiSecret;
        this.baseUrl = baseUrl.endsWith("/") ? baseUrl : baseUrl + "/";
        
        // 创建HTTP客户端配置
        Map<String, String> headers = new HashMap<>();
        headers.put("User-Agent", "OpenAPI-Python-Client/1.0");
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
    
    // ... 其他方法保持不变 ...
    
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