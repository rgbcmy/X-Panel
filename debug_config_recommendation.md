# X-Panel 本地调试 REALITY 配置建议

## 问题原因
REALITY 协议需要与真实的 TLS 服务器进行握手，本地调试时如果：
1. 网络无法访问 target 服务器
2. target 服务器 TLS 配置变化
3. 防火墙/代理阻止访问

就会出现 "target sent incorrect server hello or handshake incomplete" 错误。

## 解决方案

### 选项 1：禁用 REALITY（推荐用于纯本地调试）
如果只是调试业务逻辑，不需要测试 REALITY 协议本身：
1. 登录 X-Panel Web 界面
2. 进入【入站列表】
3. 编辑对应的入站
4. 将 Security 改为 `none`
5. 保存并重启

### 选项 2：更换可靠的 target
选择本地网络稳定可达的网站：

```json
{
  "realitySettings": {
    "target": "www.apple.com:443",
    "serverNames": ["www.apple.com"]
  }
}
```

测试 target 是否可达：
```bash
curl -I https://www.apple.com --connect-timeout 5
openssl s_client -connect www.apple.com:443 -servername www.apple.com < /dev/null
```

### 选项 3：使用本地 TLS 服务器
搭建本地 TLS 服务器作为 target：
```bash
# 使用 Nginx 或 Caddy 在本地监听 443
target: "127.0.0.1:8443"
```

## 推荐的本地调试流程

1. **开发阶段**：使用 `security: none` 快速迭代
2. **功能测试**：使用稳定的 target（如 apple.com）测试 REALITY
3. **生产部署**：根据实际需求配置 target

## 常用 target 选择

| 网站 | Target | ServerNames | 说明 |
|-----|--------|-------------|------|
| Apple | www.apple.com:443 | www.apple.com | 全球稳定 |
| Cloudflare | www.cloudflare.com:443 | www.cloudflare.com | CDN稳定 |
| Microsoft | www.microsoft.com:443 | www.microsoft.com | 企业级 |
| Yahoo | www.yahoo.com:443 | www.yahoo.com | 全球可达 |

## 验证配置

```bash
# 1. 检查 target 可达性
curl -I https://TARGET_DOMAIN --connect-timeout 5

# 2. 检查 TLS 握手
openssl s_client -connect TARGET_DOMAIN:443 -servername TARGET_DOMAIN

# 3. 重启服务
./x-ui-bin restart
# 或
systemctl restart x-ui

# 4. 查看日志
tail -f error.log
```
