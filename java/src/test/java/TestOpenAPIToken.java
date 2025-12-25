package com.example.openapi;
import com.fasterxml.jackson.databind.ObjectMapper;
public class TestOpenAPIToken {
    public static void main(String[] args) {
        String TOKEN = "eyJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJhaXJ0b20iLCJpZCI6MTAsImFjY291bnQiOiLotbXno4oiLCJzdWIiOiLotbXno4oiLCJpYXQiOjE3NjQ1NzA2NDIsImV4cCI6MzMzMDA1NzA2NDJ9.cDnyVC_FiAy96p1yg22mTiEaLA_oHeVRBVMdjh_qlsA";
        String BASE_URL = "http://10.0.0.132";
        String RID = "tushuguan";
        String URL = "/openapi/asset/connection/getData?rid=" + RID;       
        try (OpenAPITokenClient client = new OpenAPITokenClient(BASE_URL, TOKEN)) {
            Object result = client.get(URL);           
            ObjectMapper objectMapper = new ObjectMapper();
            String jsonResult = objectMapper.writerWithDefaultPrettyPrinter().writeValueAsString(result);
            System.out.println(jsonResult);           
        } catch (Exception e) {
            e.printStackTrace();
        }
    }
}