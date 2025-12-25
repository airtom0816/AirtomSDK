# HttpClient.py
import json
import ssl
import time
from typing import Dict, List, Optional, Any, Union
from urllib.parse import urlparse
import requests
from requests.adapters import HTTPAdapter
from requests.packages.urllib3.util.ssl_ import create_urllib3_context
from requests.packages.urllib3.poolmanager import PoolManager


class HttpClientOption:
    """HTTP 客户端配置选项"""

    def __init__(self,
                 header: Optional[Dict[str, Any]] = None,
                 cookie: Optional[Dict[str, Any]] = None,
                 proxy_address: Optional[str] = None,
                 socket_timeout: Optional[int] = None,
                 connect_timeout: Optional[int] = None,
                 ignore_ssl: bool = True):
        """
        初始化 HttpClient 配置

        Args:
            header: 请求头字典
            cookie: Cookie 字典
            proxy_address: 代理地址 (格式: "host:port")
            socket_timeout: socket 超时时间（毫秒）
            connect_timeout: 连接超时时间（毫秒）
            ignore_ssl: 是否忽略 SSL 证书验证
        """
        self.header = header or {}
        self.cookie = cookie or {}
        self.proxy_address = proxy_address
        self.socket_timeout = socket_timeout
        self.connect_timeout = connect_timeout
        self.ignore_ssl = ignore_ssl

        # 默认超时设置（单位：秒）
        self.default_socket_timeout = 60  # 60秒
        self.default_connect_timeout = 60  # 60秒
        self.connection_request_timeout = 5  # 5秒


class IgnoreSSLAdapter(HTTPAdapter):
    """忽略 SSL 证书验证的适配器"""

    def init_poolmanager(self, *args, **kwargs):
        ctx = create_urllib3_context()
        ctx.check_hostname = False
        ctx.verify_mode = ssl.CERT_NONE
        kwargs['ssl_context'] = ctx
        return super().init_poolmanager(*args, **kwargs)


class HTTPResponse:
    """HTTP 响应封装类"""

    def __init__(self, request_method: str, request_url: str,
                 status_code: int, headers: Dict[str, str],
                 content: bytes, elapsed_time: float):
        """
        初始化 HTTP 响应

        Args:
            request_method: 请求方法
            request_url: 请求URL
            status_code: 状态码
            headers: 响应头
            content: 响应内容
            elapsed_time: 请求耗时（毫秒）
        """
        self.request_method = request_method
        self.request_url = request_url
        self.status_code = status_code
        self.headers = headers
        self.content = content
        self.elapsed_time = elapsed_time

    @property
    def text(self) -> str:
        """获取响应文本"""
        return self.content.decode('utf-8', errors='ignore')

    def json(self) -> Any:
        """解析 JSON 响应"""
        return json.loads(self.text)

    def __str__(self) -> str:
        return (f"HTTPResponse(status_code={self.status_code}, "
                f"elapsed_time={self.elapsed_time}ms, "
                f"content_length={len(self.content)})")


class HttpClient:
    """HTTP 客户端类"""

    def __init__(self, option: Optional[Union[Dict[str, Any], HttpClientOption]] = None):
        """
        初始化 HTTP 客户端

        Args:
            option: 配置选项，可以是字典或 HttpClientOption 对象
        """
        if option is None:
            option = {}

        # 转换配置选项
        if isinstance(option, dict):
            self.option = HttpClientOption(
                header=option.get('header'),
                cookie=option.get('cookie'),
                proxy_address=option.get('proxy_address'),
                socket_timeout=option.get('socket_timeout'),
                connect_timeout=option.get('connect_timeout'),
                ignore_ssl=option.get('ignore_ssl', True)
            )
        else:
            self.option = option

        # 创建会话
        self.session = requests.Session()

        # 设置默认请求头
        if self.option.header:
            self.session.headers.update(
                {k: str(v) for k, v in self.option.header.items()}
            )

        # 设置 Cookie
        if self.option.cookie:
            for key, value in self.option.cookie.items():
                self.session.cookies.set(key, str(value))

        # 设置代理
        if self.option.proxy_address:
            proxy_parts = self.option.proxy_address.split(':')
            if len(proxy_parts) != 2:
                raise ValueError("请输入正确的proxyAddress地址 (格式: host:port)")

            proxy_host = proxy_parts[0]
            proxy_port = int(proxy_parts[1])
            self.session.proxies = {
                'http': f'http://{proxy_host}:{proxy_port}',
                'https': f'http://{proxy_host}:{proxy_port}'
            }

        # 设置 SSL 验证
        self.session.verify = not self.option.ignore_ssl

        # 添加忽略 SSL 的适配器
        if self.option.ignore_ssl:
            adapter = IgnoreSSLAdapter()
            self.session.mount('https://', adapter)
            self.session.mount('http://', adapter)

        # 设置连接池
        adapter = HTTPAdapter(
            pool_connections=200,
            pool_maxsize=20,
            max_retries=3
        )
        self.session.mount('https://', adapter)
        self.session.mount('http://', adapter)

    def _build_timeout(self) -> tuple:
        """构建超时配置"""
        connect_timeout = (
            self.option.connect_timeout / 1000
            if self.option.connect_timeout
            else self.option.default_connect_timeout
        )

        read_timeout = (
            self.option.socket_timeout / 1000
            if self.option.socket_timeout
            else self.option.default_socket_timeout
        )

        return (connect_timeout, read_timeout)

    def _build_response(self, response: requests.Response,
                        start_time: float) -> HTTPResponse:
        """构建 HTTPResponse 对象"""
        elapsed_time = (time.time() - start_time) * 1000  # 转换为毫秒

        return HTTPResponse(
            request_method=response.request.method,
            request_url=response.request.url,
            status_code=response.status_code,
            headers=dict(response.headers),
            content=response.content,
            elapsed_time=elapsed_time
        )

    def get(self, url: str, headers: Optional[Dict[str, Any]] = None) -> HTTPResponse:
        """
        GET 请求

        Args:
            url: 请求URL
            headers: 请求头

        Returns:
            HTTPResponse 对象
        """
        start_time = time.time()
        timeout = self._build_timeout()

        response = self.session.get(
            url,
            headers=headers,
            timeout=timeout
        )

        return self._build_response(response, start_time)

    def post_json(self, url: str, data: Any,
                  headers: Optional[Dict[str, Any]] = None) -> HTTPResponse:
        """
        POST JSON 请求

        Args:
            url: 请求URL
            data: 请求数据（可以是字典或 JSON 字符串）
            headers: 请求头

        Returns:
            HTTPResponse 对象
        """
        start_time = time.time()
        timeout = self._build_timeout()

        # 准备请求头
        request_headers = {'Content-Type': 'application/json; charset=utf-8'}
        if headers:
            request_headers.update(headers)

        # 准备请求数据
        if isinstance(data, str):
            json_data = data
        else:
            json_data = json.dumps(data, ensure_ascii=False)

        response = self.session.post(
            url,
            data=json_data,
            headers=request_headers,
            timeout=timeout
        )

        return self._build_response(response, start_time)

    def post_form(self, url: str, data: Dict[str, Any],
                  headers: Optional[Dict[str, Any]] = None) -> HTTPResponse:
        """
        POST 表单请求

        Args:
            url: 请求URL
            data: 表单数据
            headers: 请求头

        Returns:
            HTTPResponse 对象
        """
        start_time = time.time()
        timeout = self._build_timeout()

        # 准备请求头
        request_headers = {
            'Content-Type': 'application/x-www-form-urlencoded; charset=UTF-8'
        }
        if headers:
            request_headers.update(headers)

        response = self.session.post(
            url,
            data=data,
            headers=request_headers,
            timeout=timeout
        )

        return self._build_response(response, start_time)

    def get_json(self, url: str, data: Any) -> HTTPResponse:
        """
        GET 请求（带 JSON 请求体）

        Args:
            url: 请求URL
            data: 请求数据（可以是字典或 JSON 字符串）

        Returns:
            HTTPResponse 对象
        """
        start_time = time.time()
        timeout = self._build_timeout()

        # 准备请求数据
        if isinstance(data, str):
            json_data = data
        else:
            json_data = json.dumps(data, ensure_ascii=False)

        # 使用 requests 的 request 方法发送带请求体的 GET 请求
        response = self.session.request(
            'GET',
            url,
            data=json_data,
            headers={'Content-Type': 'application/json; charset=utf-8'},
            timeout=timeout
        )

        return self._build_response(response, start_time)

    def upload_files(self, url: str, files: List[Any]) -> HTTPResponse:
        """
        上传文件

        Args:
            url: 上传URL
            files: 文件列表

        Returns:
            HTTPResponse 对象
        """
        start_time = time.time()
        timeout = self._build_timeout()

        # 准备文件字典
        files_dict = {}
        for i, file in enumerate(files):
            # 假设 file 有 filename 和 content 属性
            # 根据实际情况调整
            if hasattr(file, 'filename') and hasattr(file, 'content'):
                files_dict[f'file_{i}'] = (file.filename, file.content)
            else:
                # 如果文件是字典形式
                files_dict[f'file_{i}'] = (file.get('filename'), file.get('content'))

        response = self.session.post(
            url,
            files=files_dict,
            timeout=timeout
        )

        return self._build_response(response, start_time)

    def close(self) -> None:
        """关闭 HTTP 客户端"""
        if self.session:
            self.session.close()

    def __enter__(self):
        """上下文管理器入口"""
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """上下文管理器退出"""
        self.close()


