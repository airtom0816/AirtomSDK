# OpenAPITokenClient.py
import json
from typing import Dict, Any, Optional, Union
from urllib.parse import urljoin

# 导入之前翻译的 HttpClient
from HttpClient import HttpClient, HttpClientOption


class OpenAPITokenClient:
    """OpenAPI Token 认证客户端"""

    def __init__(self, base_url: str, token: str):
        """
        初始化 OpenAPI Token 客户端

        Args:
            base_url: API 基础URL
            token: 认证令牌
        """
        self.token = token
        self.base_url = base_url.rstrip('/') + '/'  # 确保以斜杠结尾

        # 创建 HTTP 客户端配置
        headers = {"token": token}
        option = HttpClientOption(header=headers)
        self.client = HttpClient(option)

    def get(self, url: str, params: Optional[Dict[str, Any]] = None) -> Any:
        """
        发送 GET 请求

        Args:
            url: 相对URL路径
            params: 查询参数（可选）

        Returns:
            响应数据

        Raises:
            Exception: 请求失败时抛出
        """
        # 构建完整URL
        full_url = urljoin(self.base_url, url.lstrip('/'))

        # 发送 GET 请求
        response = self.client.get(full_url)

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
        发送 POST 请求

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

        # 发送 POST JSON 请求
        response = self.client.post_json(full_url, data)

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
        发送 POST 表单请求

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

        # 发送 POST 表单请求
        response = self.client.post_form(full_url, data)

        # 检查响应状态
        if response.status_code >= 400:
            raise Exception(f"HTTP {response.status_code}: {response.text}")

        # 尝试解析 JSON 响应
        try:
            return response.json()
        except json.JSONDecodeError:
            return response.text

    def put(self, url: str, data: Dict[str, Any]) -> Any:
        """
        发送 PUT 请求

        Args:
            url: 相对URL路径
            data: 请求数据

        Returns:
            响应数据

        Raises:
            Exception: 请求失败时抛出

        Note:
            由于 HttpClient 没有直接实现 PUT 方法，
            这里使用 POST 并添加 X-HTTP-Method-Override 头
        """
        # 构建完整URL
        full_url = urljoin(self.base_url, url.lstrip('/'))

        # 添加方法重写头
        headers = {"X-HTTP-Method-Override": "PUT"}

        # 发送请求
        response = self.client.post_json(full_url, data, headers=headers)

        # 检查响应状态
        if response.status_code >= 400:
            raise Exception(f"HTTP {response.status_code}: {response.text}")

        # 尝试解析 JSON 响应
        try:
            return response.json()
        except json.JSONDecodeError:
            return response.text

    def delete(self, url: str) -> Any:
        """
        发送 DELETE 请求

        Args:
            url: 相对URL路径

        Returns:
            响应数据

        Raises:
            Exception: 请求失败时抛出
        """
        # 构建完整URL
        full_url = urljoin(self.base_url, url.lstrip('/'))

        # 发送 GET 请求（添加 DELETE 指示）
        headers = {"X-HTTP-Method-Override": "DELETE"}
        response = self.client.get(full_url, headers=headers)

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


class OpenAPITokenClientV2(OpenAPITokenClient):
    """
    OpenAPI Token 认证客户端（增强版）
    支持更多认证方式和配置选项
    """

    def __init__(self,
                 base_url: str,
                 token: str,
                 timeout: Optional[int] = 30,
                 verify_ssl: bool = True,
                 proxy: Optional[str] = None,
                 auth_header_name: str = "Authorization",
                 auth_header_format: str = "Bearer {}"):
        """
        初始化 OpenAPI Token 客户端（增强版）

        Args:
            base_url: API 基础URL
            token: 认证令牌
            timeout: 超时时间（秒）
            verify_ssl: 是否验证 SSL 证书
            proxy: 代理地址（格式: "host:port"）
            auth_header_name: 认证头名称
            auth_header_format: 认证头格式
        """
        self.token = token
        self.base_url = base_url.rstrip('/') + '/'
        self.auth_header_name = auth_header_name
        self.auth_header_format = auth_header_format

        # 格式化认证头
        auth_header_value = auth_header_format.format(token)

        # 创建 HTTP 客户端配置
        headers = {
            auth_header_name: auth_header_value,
            "User-Agent": "OpenAPI-Token-Client/1.0",
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

    def refresh_token(self, new_token: str) -> None:
        """
        刷新认证令牌

        Args:
            new_token: 新的认证令牌
        """
        self.token = new_token

        # 更新认证头
        auth_header_value = self.auth_header_format.format(new_token)
        self.client.session.headers[self.auth_header_name] = auth_header_value

    def get_with_auth_type(self,
                           url: str,
                           auth_type: str = "token") -> Any:
        """
        发送 GET 请求（支持多种认证类型）

        Args:
            url: 相对URL路径
            auth_type: 认证类型 ("token", "bearer", "basic")

        Returns:
            响应数据

        Raises:
            Exception: 请求失败时抛出
        """
        # 根据认证类型设置不同的头
        if auth_type == "bearer":
            headers = {"Authorization": f"Bearer {self.token}"}
        elif auth_type == "basic":
            # Base64 编码 token
            import base64
            encoded_token = base64.b64encode(self.token.encode()).decode()
            headers = {"Authorization": f"Basic {encoded_token}"}
        else:  # 默认使用 token
            headers = {"token": self.token}

        # 构建完整URL
        full_url = urljoin(self.base_url, url.lstrip('/'))

        # 发送 GET 请求
        response = self.client.get(full_url, headers=headers)

        # 检查响应状态
        if response.status_code >= 400:
            raise Exception(f"HTTP {response.status_code}: {response.text}")

        # 尝试解析 JSON 响应
        try:
            return response.json()
        except json.JSONDecodeError:
            return response.text

    def request(self,
                method: str,
                url: str,
                data: Optional[Dict[str, Any]] = None,
                params: Optional[Dict[str, Any]] = None) -> Any:
        """
        发送通用请求

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

        # 根据方法发送请求
        if method.upper() == "GET":
            response = self.client.get(full_url)
        elif method.upper() == "POST":
            response = self.client.post_json(full_url, data or {})
        elif method.upper() == "PUT":
            # 添加方法重写头
            headers = {"X-HTTP-Method-Override": "PUT"}
            response = self.client.post_json(full_url, data or {}, headers=headers)
        elif method.upper() == "DELETE":
            # 添加方法重写头
            headers = {"X-HTTP-Method-Override": "DELETE"}
            response = self.client.get(full_url, headers=headers)
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


class TokenManager:
    """Token 管理器，支持自动刷新"""

    def __init__(self,
                 base_url: str,
                 token: str,
                 refresh_url: Optional[str] = None,
                 refresh_interval: int = 3600):  # 默认1小时刷新一次
        """
        初始化 Token 管理器

        Args:
            base_url: API 基础URL
            token: 初始令牌
            refresh_url: 令牌刷新接口URL
            refresh_interval: 刷新间隔（秒）
        """
        self.base_url = base_url
        self.current_token = token
        self.refresh_url = refresh_url
        self.refresh_interval = refresh_interval
        self.last_refresh_time = 0

        # 创建客户端
        self.client = OpenAPITokenClientV2(
            base_url=base_url,
            token=token,
            auth_header_name="Authorization",
            auth_header_format="Bearer {}"
        )

    def should_refresh(self) -> bool:
        """检查是否应该刷新令牌"""
        current_time = time.time()
        return (self.refresh_url and
                current_time - self.last_refresh_time > self.refresh_interval)

    def refresh(self) -> bool:
        """
        刷新令牌

        Returns:
            是否刷新成功
        """
        if not self.refresh_url:
            return False

        try:
            # 调用刷新接口
            response = self.client.post(self.refresh_url, {
                "token": self.current_token
            })

            # 假设响应中包含新令牌
            if isinstance(response, dict) and "access_token" in response:
                new_token = response["access_token"]
                self.current_token = new_token
                self.client.refresh_token(new_token)
                self.last_refresh_time = time.time()
                return True

        except Exception as e:
            print(f"刷新令牌失败: {e}")

        return False

    def get(self, url: str, auto_refresh: bool = True) -> Any:
        """
        发送 GET 请求（支持自动刷新令牌）

        Args:
            url: 相对URL路径
            auto_refresh: 是否自动刷新令牌

        Returns:
            响应数据
        """
        if auto_refresh and self.should_refresh():
            self.refresh()

        return self.client.get(url)

    def post(self, url: str, data: Dict[str, Any], auto_refresh: bool = True) -> Any:
        """
        发送 POST 请求（支持自动刷新令牌）

        Args:
            url: 相对URL路径
            data: 请求数据
            auto_refresh: 是否自动刷新令牌

        Returns:
            响应数据
        """
        if auto_refresh and self.should_refresh():
            self.refresh()

        return self.client.post(url, data)

    def close(self) -> None:
        """关闭连接"""
        self.client.close()

