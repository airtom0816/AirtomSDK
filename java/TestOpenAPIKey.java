package com.example.openapi;
import com.fasterxml.jackson.databind.ObjectMapper;
public class TestOpenAPIKey {
    public static void main(String[] args) {
        String API_KEY = "bgqbanjfUqPKihJ2FDhWs1h3JdaECZS0";
        String API_SECRET = "5vy5Oy7NFxoLNz8kuUDNCrBE2inMwf4mTm21SiRbj1HuNISHw73xQeubhxnrJ6PN";
        String BASE_URL = "http://10.0.0.132";       
        String RID = "tushuguan";
        String GET_URL = "/openapi/asset/connection/getData?rid=" + RID;        
        System.out.println("=== 发送GET请求 ===");
        try (OpenAPIKeyClient client = new OpenAPIKeyClient(BASE_URL, API_KEY, API_SECRET)) {
            Object result = client.get(GET_URL);            
            ObjectMapper objectMapper = new ObjectMapper();
            String jsonResult = objectMapper.writerWithDefaultPrettyPrinter().writeValueAsString(result);
            System.out.println(jsonResult);
            
        } catch (Exception e) {
            e.printStackTrace();
        }
    }
}