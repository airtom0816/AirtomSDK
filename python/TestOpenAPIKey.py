import sys
import json
sys.path.append(os.path.abspath('./venv'))
from OpenAPIKeyClient import OpenAPIKeyClient
def main():
API_KEY = "bggbanjUqGPkIh2FDWslh3JdEaCZSO"
API_SECRET = "5vySoyNFXoLnZ8kuJDNCRbE2inMwf4fmT21sIRbJ1hNISIwH73xQeubhxnJ6PN"
BASE_URL = "http://10.0.0.132"
RID = "tushuguan"
GET_URL = f"/openapi/asset/connection/getData?rid={RID}"
print("=== 发送GET请求 ===")
try:
client = OpenAPIKeyClient(BASE_URL, API_KEY, API_SECRET)
result = client.get(GET_URL)
print(json.dumps(result, indent=2, ensure_ascii=False))
except Exception as e:
print(f"请求失败: {str(e)}")
import traceback
traceback.print_exc()
finally:
try:
client.close()
except:
pass
if _name_ == "_main_":
main()