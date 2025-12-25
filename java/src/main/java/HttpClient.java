package com.example.openapi;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import okhttp3.*;
import javax.net.ssl.*;
import java.io.IOException;
import java.security.cert.CertificateException;
import java.util.Map;
import java.util.concurrent.TimeUnit;

/**
 * HTTP客户端类
 */
public class HttpClient implements AutoCloseable {
    private OkHttpClient client;
    private ObjectMapper objectMapper;
    private Map<String, String> defaultHeaders;
    
    public HttpClient(HttpClientOption option) {
        this.objectMapper = new ObjectMapper();
        this.defaultHeaders = option.getHeader();
        
        OkHttpClient.Builder builder = new OkHttpClient.Builder()
                .connectTimeout(option.getConnectTimeout() != null ? option.getConnectTimeout() : 60000, TimeUnit.MILLISECONDS)
                .readTimeout(option.getSocketTimeout() != null ? option.getSocketTimeout() : 60000, TimeUnit.MILLISECONDS)
                .writeTimeout(option.getSocketTimeout() != null ? option.getSocketTimeout() : 60000, TimeUnit.MILLISECONDS);
        
        // 配置SSL
        if (option.isIgnoreSsl()) {
            builder.sslSocketFactory(createInsecureSslSocketFactory(), createInsecureTrustManager())
                   .hostnameVerifier((hostname, session) -> true);
        }
        
        // 配置代理
        if (option.getProxyAddress() != null && !option.getProxyAddress().isEmpty()) {
            String[] proxyParts = option.getProxyAddress().split(":");
            if (proxyParts.length == 2) {
                String host = proxyParts[0];
                int port = Integer.parseInt(proxyParts[1]);
                builder.proxy(new java.net.Proxy(java.net.Proxy.Type.HTTP, 
                        new java.net.InetSocketAddress(host, port)));
            }
        }
        
        this.client = builder.build();
    }
    
    private SSLSocketFactory createInsecureSslSocketFactory() {
        try {
            SSLContext sslContext = SSLContext.getInstance("TLS");
            sslContext.init(null, new TrustManager[]{createInsecureTrustManager()}, new java.security.SecureRandom());
            return sslContext.getSocketFactory();
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }
    
    private X509TrustManager createInsecureTrustManager() {
        return new X509TrustManager() {
            @Override
            public void checkClientTrusted(java.security.cert.X509Certificate[] chain, String authType) throws CertificateException {
            }
            
            @Override
            public void checkServerTrusted(java.security.cert.X509Certificate[] chain, String authType) throws CertificateException {
            }
            
            @Override
            public java.security.cert.X509Certificate[] getAcceptedIssuers() {
                return new java.security.cert.X509Certificate[]{};
            }
        };
    }
    
    /**
     * GET请求
     */
    public HttpResponse get(String url, Map<String, String> headers) throws IOException {
        Request.Builder requestBuilder = new Request.Builder().url(url);
        
        // 添加默认headers
        if (defaultHeaders != null) {
            for (Map.Entry<String, String> entry : defaultHeaders.entrySet()) {
                requestBuilder.addHeader(entry.getKey(), entry.getValue());
            }
        }
        
        // 添加请求特定的headers
        if (headers != null) {
            for (Map.Entry<String, String> entry : headers.entrySet()) {
                requestBuilder.addHeader(entry.getKey(), entry.getValue());
            }
        }
        
        long startTime = System.currentTimeMillis();
        Response response = client.newCall(requestBuilder.build()).execute();
        long elapsedTime = System.currentTimeMillis() - startTime;
        
        return new HttpResponse(
                "GET",
                url,
                response.code(),
                response.headers().toMultimap(),
                response.body().bytes(),
                elapsedTime
        );
    }
    
    /**
     * POST JSON请求
     */
    public HttpResponse postJson(String url, Object data, Map<String, String> headers) throws IOException {
        String jsonData;
        if (data instanceof String) {
            jsonData = (String) data;
        } else {
            jsonData = objectMapper.writeValueAsString(data);
        }
        
        RequestBody body = RequestBody.create(
                jsonData,
                MediaType.parse("application/json; charset=utf-8")
        );
        
        Request.Builder requestBuilder = new Request.Builder()
                .url(url)
                .post(body);
        
        // 添加默认headers
        if (defaultHeaders != null) {
            for (Map.Entry<String, String> entry : defaultHeaders.entrySet()) {
                requestBuilder.addHeader(entry.getKey(), entry.getValue());
            }
        }
        
        // 添加请求特定的headers
        if (headers != null) {
            for (Map.Entry<String, String> entry : headers.entrySet()) {
                requestBuilder.addHeader(entry.getKey(), entry.getValue());
            }
        }
        
        long startTime = System.currentTimeMillis();
        Response response = client.newCall(requestBuilder.build()).execute();
        long elapsedTime = System.currentTimeMillis() - startTime;
        
        return new HttpResponse(
                "POST",
                url,
                response.code(),
                response.headers().toMultimap(),
                response.body().bytes(),
                elapsedTime
        );
    }
    
    /**
     * POST表单请求
     */
    public HttpResponse postForm(String url, Map<String, String> data, Map<String, String> headers) throws IOException {
        FormBody.Builder formBuilder = new FormBody.Builder();
        
        if (data != null) {
            for (Map.Entry<String, String> entry : data.entrySet()) {
                formBuilder.add(entry.getKey(), entry.getValue());
            }
        }
        
        Request.Builder requestBuilder = new Request.Builder()
                .url(url)
                .post(formBuilder.build());
        
        // 添加默认headers
        if (defaultHeaders != null) {
            for (Map.Entry<String, String> entry : defaultHeaders.entrySet()) {
                requestBuilder.addHeader(entry.getKey(), entry.getValue());
            }
        }
        
        // 添加请求特定的headers
        if (headers != null) {
            for (Map.Entry<String, String> entry : headers.entrySet()) {
                requestBuilder.addHeader(entry.getKey(), entry.getValue());
            }
        }
        
        long startTime = System.currentTimeMillis();
        Response response = client.newCall(requestBuilder.build()).execute();
        long elapsedTime = System.currentTimeMillis() - startTime;
        
        return new HttpResponse(
                "POST",
                url,
                response.code(),
                response.headers().toMultimap(),
                response.body().bytes(),
                elapsedTime
        );
    }
    
    /**
     * 关闭客户端 - 实现AutoCloseable接口
     */
    @Override
    public void close() {
        if (client != null) {
            client.dispatcher().executorService().shutdown();
            client.connectionPool().evictAll();
        }
    }
}

/**
 * HTTP响应类
 */
class HttpResponse {
    private String requestMethod;
    private String requestUrl;
    private int statusCode;
    private Map<String, java.util.List<String>> headers;
    private byte[] content;
    private long elapsedTime;
    
    public HttpResponse(String requestMethod, String requestUrl, int statusCode,
                       Map<String, java.util.List<String>> headers, byte[] content, long elapsedTime) {
        this.requestMethod = requestMethod;
        this.requestUrl = requestUrl;
        this.statusCode = statusCode;
        this.headers = headers;
        this.content = content;
        this.elapsedTime = elapsedTime;
    }
    
    public String getRequestMethod() {
        return requestMethod;
    }
    
    public String getRequestUrl() {
        return requestUrl;
    }
    
    public int getStatusCode() {
        return statusCode;
    }
    
    public Map<String, java.util.List<String>> getHeaders() {
        return headers;
    }
    
    public byte[] getContent() {
        return content;
    }
    
    public String getText() {
        return new String(content, java.nio.charset.StandardCharsets.UTF_8);
    }
    
    public <T> T getJson(Class<T> clazz) throws JsonProcessingException {
        ObjectMapper objectMapper = new ObjectMapper();
        return objectMapper.readValue(getText(), clazz);
    }
    
    public Object getJson() throws JsonProcessingException {
        ObjectMapper objectMapper = new ObjectMapper();
        return objectMapper.readValue(getText(), Object.class);
    }
    
    public long getElapsedTime() {
        return elapsedTime;
    }
    
    @Override
    public String toString() {
        return String.format("HTTPResponse(status_code=%d, elapsed_time=%dms, content_length=%d)",
                statusCode, elapsedTime, content.length);
    }
}

/**
 * HTTP客户端配置选项
 */
class HttpClientOption {
    private Map<String, String> header;
    private Map<String, String> cookie;
    private String proxyAddress;
    private Integer socketTimeout;
    private Integer connectTimeout;
    private boolean ignoreSsl = true;
    
    public HttpClientOption() {
    }
    
    public HttpClientOption(Map<String, String> header, Map<String, String> cookie,
                           String proxyAddress, Integer socketTimeout,
                           Integer connectTimeout, boolean ignoreSsl) {
        this.header = header;
        this.cookie = cookie;
        this.proxyAddress = proxyAddress;
        this.socketTimeout = socketTimeout;
        this.connectTimeout = connectTimeout;
        this.ignoreSsl = ignoreSsl;
    }
    
    // Getters and Setters
    public Map<String, String> getHeader() {
        return header;
    }
    
    public void setHeader(Map<String, String> header) {
        this.header = header;
    }
    
    public Map<String, String> getCookie() {
        return cookie;
    }
    
    public void setCookie(Map<String, String> cookie) {
        this.cookie = cookie;
    }
    
    public String getProxyAddress() {
        return proxyAddress;
    }
    
    public void setProxyAddress(String proxyAddress) {
        this.proxyAddress = proxyAddress;
    }
    
    public Integer getSocketTimeout() {
        return socketTimeout;
    }
    
    public void setSocketTimeout(Integer socketTimeout) {
        this.socketTimeout = socketTimeout;
    }
    
    public Integer getConnectTimeout() {
        return connectTimeout;
    }
    
    public void setConnectTimeout(Integer connectTimeout) {
        this.connectTimeout = connectTimeout;
    }
    
    public boolean isIgnoreSsl() {
        return ignoreSsl;
    }
    
    public void setIgnoreSsl(boolean ignoreSsl) {
        this.ignoreSsl = ignoreSsl;
    }
}