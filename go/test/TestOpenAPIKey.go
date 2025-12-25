package main
import (
 "encoding/json"
 "fmt"
)
const (
 API_KEY1    = "bggbanjUqGPkIh2FDWslh3JdEaCZSO"
 API_SECRET1 = "5vySoyNFXoLnZ8kuJDNCRbE2inMwf4fmT21sIRbJ1hNISIwH73xQeubhxnJ6PN"
 BASE_URL1   = "http://10.0.0.132"
 RID1        = "tushuguan"
)
func main() {
 GET_URL := fmt.Sprintf(format: "/openapi/asset/connection/getData?rid=%s", a...: RID1)
 fmt.Println(a...: "=== 发送GET请求 ===")
 client := NewOpenAPIKeyClient(baseURL: BASE_URL1, apiKey: API_KEY1, apiSecret: API_SECRET1)
 defer func() {
  if client != nil {
   client.Close()
  }
 }()
 result, err := client.Get(urlPath: GET_URL, params: nil)