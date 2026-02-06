# V2Board 用户同步功能 - 使用示例

## 配置说明

在使用新的用户驱动同步功能之前，需要确保以下配置已正确设置：

### 1. V2Board集成配置
在系统设置中启用V2Board集成：
- V2Board URL: `https://your-v2board-domain.com`
- V2Board Token: `your-api-token`
- 启用V2Board: `true`

### 2. 定时任务配置
确保V2Board同步任务已添加到定时任务中，建议每5-10分钟执行一次。

## API示例

### V2Board API响应示例

```json
{
  "users": [
    {
      "id": 16,
      "uuid": "31f9a03b-ced0-4d9d-9325-66ca6eb77ca3",
      "email": "user@example.com",
      "speed_limit": 0,
      "vless_config": {
        "inbound_port": 8024,
        "inbound_tag": "user-16-vless-2",
        "reality_private_key": "X87Ee4L278zkkqwGN+TfDUin5HBbHNSmz87AMupvfLY",
        "reality_public_key": "4vQfQrcDaSb2J1Ai4MB9nGcaZegb+3QF6PclOuwWO2w",
        "reality_short_id": "23180e5d6724",
        "reality_dest": "tesla.com",
        "reality_server_names": ["www.tesla.com"],
        "spider_x": "/",
        "fingerprint": "firefox",
        "flow": "xtls-rprx-vision"
      }
    },
    {
      "id": 17,
      "uuid": "40320bf6-175a-482e-800b-d357df8d2d9f",
      "email": "another@example.com",
      "speed_limit": null
    }
  ]
}
```

## 功能行为

### 用户有 vless_config
当用户对象包含 `vless_config` 字段时：

1. **首次同步**:
   - 自动创建新的Inbound
   - 端口使用 `vless_config.inbound_port`
   - Tag使用 `vless_config.inbound_tag`
   - 配置Reality相关参数
   - 将用户添加为第一个客户端

2. **后续同步**:
   - 根据 `inbound_tag` 查找现有Inbound
   - 如果用户已存在，更新其配置（速度限制、flow等）
   - 如果用户不存在，添加为新客户端
   - 保留其他已有客户端

### 用户无 vless_config
当用户对象不包含 `vless_config` 字段时：
- 跳过该用户，不做任何处理
- 在日志中记录调试信息

## 生成的Inbound配置示例

### Inbound基本信息
```json
{
  "id": 1,
  "remark": "V2Board User 16 - user@example.com",
  "enable": true,
  "port": 8024,
  "protocol": "vless",
  "tag": "user-16-vless-2",
  "v2boardEnabled": true,
  "v2boardNodeId": "user-16",
  "v2boardNodeType": "vless"
}
```

### Settings配置
```json
{
  "clients": [
    {
      "id": "31f9a03b-ced0-4d9d-9325-66ca6eb77ca3",
      "email": "user@example.com",
      "enable": true,
      "flow": "xtls-rprx-vision",
      "speedLimit": 0,
      "totalGB": 0,
      "expiryTime": 0,
      "limitIp": 0
    }
  ],
  "decryption": "none"
}
```

### StreamSettings配置（Reality）
```json
{
  "network": "tcp",
  "security": "reality",
  "realitySettings": {
    "show": false,
    "dest": "tesla.com",
    "xver": 0,
    "serverNames": ["www.tesla.com"],
    "privateKey": "X87Ee4L278zkkqwGN+TfDUin5HBbHNSmz87AMupvfLY",
    "shortIds": ["23180e5d6724"],
    "spiderX": "/",
    "fingerprint": "firefox"
  }
}
```

### Sniffing配置
```json
{
  "enabled": true,
  "destOverride": ["http", "tls", "quic"]
}
```

## 日志示例

### 成功同步日志
```
[INFO] V2board sync: processing 10 users
[DEBUG] user 1 has no vless_config, skipping
[DEBUG] user 2 has no vless_config, skipping
[INFO] created new inbound for user 16 port: 8024 tag: user-16-vless-2
[INFO] successfully processed inbound for user 16 tag: user-16-vless-2
[DEBUG] updated existing client in inbound user-17-vless-1 for user 17
[INFO] successfully processed inbound for user 17 tag: user-17-vless-1
```

### 错误处理日志
```
[WARNING] failed to process inbound for user 20 : failed to add inbound: port 8080 already in use
[WARNING] failed to process inbound for user 21 : failed to marshal stream settings: invalid json
```

## 常见问题

### Q1: 端口冲突怎么办？
A: 系统会记录警告日志并跳过该用户。需要在V2Board中为用户分配不同的端口。

### Q2: Tag重复怎么办？
A: 如果Tag已存在，系统会更新现有Inbound而不是创建新的。确保V2Board返回的tag是唯一的。

### Q3: 用户删除后Inbound会被删除吗？
A: 当前版本不会自动删除Inbound。如果需要清理，需要手动删除或实现清理逻辑。

### Q4: 支持多节点吗？
A: 当前实现需要配置默认的nodeId和nodeType。如需支持多节点，需要修改 `syncUserDrivenInbounds` 方法的参数传递。

### Q5: 如何验证同步是否成功？
A: 查看系统日志，检查是否有成功创建或更新Inbound的日志。也可以在管理面板的Inbound列表中查看新创建的Inbound。

## 迁移指南

### 从旧策略迁移
如果之前使用的是Inbound驱动模式：

1. **备份数据**: 导出当前所有Inbound配置
2. **更新代码**: 部署新版本代码
3. **测试同步**: 手动触发一次同步任务
4. **验证结果**: 检查日志和Inbound列表
5. **清理旧配置**: 删除不再使用的Inbound（可选）

### 注意事项
- 新旧策略互不兼容，需要完全切换
- 确保V2Board API返回包含 `vless_config` 的用户数据
- 首次同步可能会创建大量Inbound，注意端口分配

## 性能考虑

- 每次同步遍历所有用户，建议用户数较少时使用（< 1000用户）
- 对于大规模用户，建议实现增量同步机制
- 定时任务间隔不宜过短，建议5-10分钟
- 注意监控系统资源占用，特别是数据库连接
