package main
import (
 "fmt"
)
const (
 // 使用指定的TOKEN
 TOKEN1= "eyJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJhaXJ0b20iLCJpZCI6MTAsImFjY291bnQiOiLotbXno4oiLCJzdWIiOiLotbXno4oiLCJpYXQiOjE3NjQ1NzA2NDIsImV4cCI6MzMzMDA1NzA2NDJ9.cDnyVC_FiAy96p1yg22mTiEaLA_oHeVRBVMdjh_qlsA"
 BASE_URL1 = "http://10.0.0.132"
)
func main() {
 client := NewOpenAPITokenClient(baseURL: BASE_URL1, token: TOKEN1)
 defer client.Close()
 params := map[string]string{
  "rid": "tushuguan",
 }
 result, err := client.Get(urlPath: "/openapi/asset/connection/getData", params)
 if err != nil {
  fmt.Printf(format: "请求失败：%v\n", a...: err)
  return
 }
 fmt.Printf(format: "%v\n", a...: result)
}