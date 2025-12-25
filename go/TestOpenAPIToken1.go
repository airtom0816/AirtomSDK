package main
import (
"encoding/json"
"fmt"
)
const (
// 使用指定的TOKEN
TOKEN = "eyJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJhaXJ0b20iLCJpZCI6MTAsImFjY291bnQiOiLotbXno4oiLCJzdWIiOiLotbXno4oiLCJpYXQiOjE3NjQ1NzA2NDIsImV4cCI6MzMzMDA1NzA2NDJ9.cDnyVC_FiAy96p1yg22mTiEaLA_oHeVRBVMdjh_qlsA"
BASE_URL = "http://10.0.0.132"
RID = "tushuguan"
)
func main() {
client := NewOpenAPITokenClient(baseURL: BASE_URL, token: TOKEN)
defer client.Close()
urlPath := fmt.Sprintf(format: "/openapi/asset/connection/tables?rid=%s", a...: RID)
result, err := client.Get(urlPath, params: nil)
if err != nil {
fmt.Printf(format: "请求失败: %v\n", a...: err)
return
}
formatted, err := json.MarshalIndent(v: result, prefix: "", indent: " ")
if err != nil {
fmt.Printf(format: "格式化输出失败: %v\n", a...: err)
fmt.Printf(format: "原始结果: %v\n", a...: result)
return
}
// 补充：打印格式化后的结果
fmt.Println(string(formatted))
}