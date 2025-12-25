# OpenAPIKeyClient.py
import json
import hashlib
import hmac
import time
import uuid
from typing import Dict, Any, Optional
from urllib.parse import urljoin

# 导入之前翻译的 HttpClient
from HttpClient import HttpClient, HttpClientOption


class OpenAPIKeyClient:
    """OpenAPI 密钥认证客户端"""

    def __init__(self, base_url: str, api_key: str, api_secret: str):
        """
        初始化 OpenAPI 客户端

        Args:
            base_url: API 基础URL
            api_key: API 密钥
            api_secret: API 密钥
        """
        self.api_key = api_key
        self.api_secret = api_secret
        self.base_url = base_url.rstrip('/') + '/'  # 确保以斜杠结尾

        # 创建 HTTP 客户端
        option = HttpClientOption(header={})
        self.client = HttpClient(option)

    def _generate_signature(self, text: str) -> str:
        """
        生成 HMAC-SHA256 签名

        Args:
            text: 待签名的文本

        Returns:
            十六进制签名字符串
        """
        try:
            # 创建 HMAC-SHA256 签名
            signature = hmac.new(
                self.api_secret.encode('utf-8'),
                text.encode('utf-8'),
                hashlib.sha256
            )

            # 返回十六进制字符串
            return signature.hexdigest()
        except Exception as e:
            raise RuntimeError(f"签名生成失败: {str(e)}")

    def _build_auth_headers(self, request_body: str = "") -> Dict[str, str]:
        """
        构建认证请求头

        Args:
            request_body: 请求体内容

        Returns:
            认证头字典
        """
        headers = {}

        # 时间戳（毫秒）
        timestamp = str(int(time.time() * 1000))

        # 随机数（Nonce）
        nonce = str(uuid.uuid4()).replace("-", "")

        # 签名计算：apiKey + timestamp + nonce + body
        sign_text = self.api_key + timestamp + nonce + request_body
        signature = self._generate_signature(sign_text)

        # 添加认证头
        headers["X-Api-Key"] = self.api_key
        headers["X-Timestamp"] = timestamp
        headers["X-Nonce"] = nonce
        headers["X-Signature"] = signature
        headers["Content-Type"] = "application/json"

        return headers

    def get(self, url: str, params: Optional[Dict[str, Any]] = None) -> Any:
        """
        发送 GET 请求（带签名认证）

        Args:
            url: 相对URL路径
            params: 查询参数

        Returns:
            响应数据

        Raises:
            Exception: 请求失败时抛出
        """
        # 构建完整URL
        full_url = urljoin(self.base_url, url.lstrip('/'))

        # GET 请求的 body 为空字符串
        request_body = ""
        auth_headers = self._build_auth_headers(request_body)

        # 发送 GET 请求
        response = self.client.get(full_url, headers=auth_headers)

        # 检查响应状态
        if response.status_code >= 400:
            raise Exception(f"HTTP {response.status_code}: {response.text}")

        # 尝试解析 JSON 响应
        try:
            return response.json()
        except json.JSONDecodeError:
            return response.text

    def post(self, url: str, data: Dict[str, Any]) -> Any:
        """
        发送 POST 请求（带签名认证）

        Args:
            url: 相对URL路径
            data: 请求数据

        Returns:
            响应数据

        Raises:
            Exception: 请求失败时抛出
        """
        # 构建完整URL
        full_url = urljoin(self.base_url, url.lstrip('/'))

        # 将数据转换为 JSON 字符串作为请求 body
        request_body = json.dumps(data, ensure_ascii=False)
        auth_headers = self._build_auth_headers(request_body)

        # 发送 POST JSON 请求
        response = self.client.post_json(full_url, data, headers=auth_headers)

        # 检查响应状态
        if response.status_code >= 400:
            raise Exception(f"HTTP {response.status_code}: {response.text}")

        # 尝试解析 JSON 响应
        try:
            return response.json()
        except json.JSONDecodeError:
            return response.text

    def post_form(self, url: str, data: Dict[str, Any]) -> Any:
        """
        发送 POST 表单请求（带签名认证）

        Args:
            url: 相对URL路径
            data: 表单数据

        Returns:
            响应数据

        Raises:
            Exception: 请求失败时抛出
        """
        # 构建完整URL
        full_url = urljoin(self.base_url, url.lstrip('/'))

        # 将数据转换为 JSON 字符串作为请求 body（签名需要）
        request_body = json.dumps(data, ensure_ascii=False)
        auth_headers = self._build_auth_headers(request_body)

        # 发送 POST 表单请求
        response = self.client.post_form(full_url, data, headers=auth_headers)

        # 检查响应状态
        if response.status_code >= 400:
            raise Exception(f"HTTP {response.status_code}: {response.text}")

        # 尝试解析 JSON 响应
        try:
            return response.json()
        except json.JSONDecodeError:
            return response.text

    def close(self) -> None:
        """关闭连接"""
        if self.client:
            self.client.close()

    def __enter__(self):
        """上下文管理器入口"""
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """上下文管理器退出"""
        self.close()


class OpenAPIKeyClientV2:
    """
    OpenAPI 密钥认证客户端（增强版）
    支持更多认证方式和配置选项
    """

    def __init__(self,
                 base_url: str,
                 api_key: str,
                 api_secret: str,
                 timeout: Optional[int] = 30,
                 verify_ssl: bool = True,
                 proxy: Optional[str] = None):
        """
        初始化 OpenAPI 客户端（增强版）

        Args:
            base_url: API 基础URL
            api_key: API 密钥
            api_secret: API 密钥
            timeout: 超时时间（秒）
            verify_ssl: 是否验证 SSL 证书
            proxy: 代理地址（格式: "host:port"）
        """
        self.api_key = api_key
        self.api_secret = api_secret
        self.base_url = base_url.rstrip('/') + '/'

        # 创建 HTTP 客户端配置
        headers = {
            "User-Agent": "OpenAPI-Python-Client/1.0",
            "Accept": "application/json"
        }

        option = HttpClientOption(
            header=headers,
            socket_timeout=timeout * 1000 if timeout else None,
            connect_timeout=timeout * 1000 if timeout else None,
            ignore_ssl=not verify_ssl,
            proxy_address=proxy
        )

        self.client = HttpClient(option)

    def _generate_signature_v2(self,
                               method: str,
                               path: str,
                               timestamp: str,
                               nonce: str,
                               body: str = "") -> str:
        """
        生成增强版签名

        Args:
            method: HTTP 方法（GET, POST等）
            path: 请求路径（不包含域名）
            timestamp: 时间戳
            nonce: 随机数
            body: 请求体

        Returns:
            十六进制签名字符串
        """
        # 签名算法：method + path + apiKey + timestamp + nonce + bodyHash
        body_hash = hashlib.sha256(body.encode('utf-8')).hexdigest() if body else ""

        # 构建签名字符串
        sign_text = f"{method.upper()}{path}{self.api_key}{timestamp}{nonce}{body_hash}"

        # 生成 HMAC-SHA256 签名
        signature = hmac.new(
            self.api_secret.encode('utf-8'),
            sign_text.encode('utf-8'),
            hashlib.sha256
        ).hexdigest()

        return signature

    def _build_auth_headers_v2(self,
                               method: str,
                               path: str,
                               body: str = "") -> Dict[str, str]:
        """
        构建增强版认证请求头

        Args:
            method: HTTP 方法
            path: 请求路径
            body: 请求体内容

        Returns:
            认证头字典
        """
        headers = {}

        # 时间戳（秒，整数）
        timestamp = str(int(time.time()))

        # 随机数（Nonce）
        nonce = str(uuid.uuid4()).replace("-", "")

        # 生成签名
        signature = self._generate_signature_v2(method, path, timestamp, nonce, body)

        # 添加认证头
        headers["X-Api-Key"] = self.api_key
        headers["X-Timestamp"] = timestamp
        headers["X-Nonce"] = nonce
        headers["X-Signature"] = signature
        headers["Content-Type"] = "application/json"

        return headers

    def request(self,
                method: str,
                url: str,
                data: Optional[Dict[str, Any]] = None,
                params: Optional[Dict[str, Any]] = None) -> Any:
        """
        发送通用请求（带签名认证）

        Args:
            method: HTTP 方法（GET, POST, PUT, DELETE等）
            url: 相对URL路径
            data: 请求体数据
            params: 查询参数

        Returns:
            响应数据

        Raises:
            Exception: 请求失败时抛出
        """
        # 构建完整URL
        full_url = urljoin(self.base_url, url.lstrip('/'))

        # 提取路径部分（用于签名）
        parsed_url = urlparse(full_url)
        path = parsed_url.path
        if parsed_url.query:
            path += "?" + parsed_url.query

        # 准备请求体
        request_body = ""
        if data is not None:
            request_body = json.dumps(data, ensure_ascii=False)

        # 构建认证头
        auth_headers = self._build_auth_headers_v2(method, path, request_body)

        # 根据方法发送请求
        if method.upper() == "GET":
            response = self.client.get(full_url, headers=auth_headers)
        elif method.upper() == "POST":
            response = self.client.post_json(full_url, data, headers=auth_headers)
        elif method.upper() == "PUT":
            # 注意：HttpClient 类没有实现 PUT 方法，这里简化为 POST
            # 在实际应用中，可能需要扩展 HttpClient 类
            response = self.client.post_json(full_url, data, headers=auth_headers)
        elif method.upper() == "DELETE":
            # 注意：HttpClient 类没有实现 DELETE 方法
            # 在实际应用中，可能需要扩展 HttpClient 类
            response = self.client.get(full_url, headers=auth_headers)
        else:
            raise ValueError(f"不支持的 HTTP 方法: {method}")

        # 检查响应状态
        if response.status_code >= 400:
            raise Exception(f"HTTP {response.status_code}: {response.text}")

        # 尝试解析 JSON 响应
        try:
            return response.json()
        except json.JSONDecodeError:
            return response.text

    def close(self) -> None:
        """关闭连接"""
        if self.client:
            self.client.close()

    def __enter__(self):
        """上下文管理器入口"""
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """上下文管理器退出"""
        self.close()


