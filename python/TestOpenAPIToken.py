import json
from OpenAPITokenClient import OpenAPITokenClient
TOKEN = "eyJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJhaXJ0b20iLCJpZCI6MTAsImFjY291bnQiOiLotbXno4oiLCJzdWIiOiLotbXno4oiLCJpYXQiOjE3NjQ1NzA2NDIsImV4cCI6MzMzMDA1NzA2NDJ9.cDnyVC_FiAy96p1yg22mTiEaLA_oHeVRBVMdjh_qlsA"
BASE_URL = "http://10.0.0.132"
RID = "tushuguan"
client = OpenAPITokenClient(BASE_URL, TOKEN)
result = client.get(f"/openapi/asset/connection/getData?rid={RID}")
client.close()
print(json.dumps(result, indent=2, ensure_ascii=False))