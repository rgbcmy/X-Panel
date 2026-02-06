# V2Board 用户同步策略变更文档

## 变更概述

### 当前策略（旧）
- **模式**: Inbound驱动模式
- **流程**: 
  1. 遍历系统中所有已存在的Inbound
  2. 为每个启用了V2Board的Inbound获取用户列表
  3. 将用户添加到对应的Inbound中

### 新策略
- **模式**: 用户驱动模式
- **流程**:
  1. 从V2Board API `/api/v1/server/UniProxy/user` 获取所有用户列表
  2. 遍历用户列表
  3. 检查用户是否包含 `vless_config` 配置
  4. 为有 `vless_config` 的用户动态创建或更新对应的Inbound
  5. 将用户添加到其专属的Inbound中
  6. 没有 `vless_config` 的用户跳过处理

## 新API结构

### 接口信息
- **URL**: `/api/v1/server/UniProxy/user`
- **Method**: GET
- **参数**: `node_id`, `node_type`, `token`

### 响应数据结构

```json
{
  "users": [
    {
      "id": 1,
      "uuid": "40320bf6-175a-482e-800b-d357df8d2d9f",
      "email": "user@example.com",
      "speed_limit": null,
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
    }
  ]
}
```

### 字段说明

#### User字段
- `id`: 用户ID（整型）
- `uuid`: 用户UUID，用作VLESS客户端ID
- `email`: 用户邮箱，用于标识客户端
- `speed_limit`: 速度限制（KB/s），null表示不限速

#### VlessConfig字段（可选）
- `inbound_port`: Inbound监听端口
- `inbound_tag`: Inbound标签（唯一标识）
- `reality_private_key`: Reality协议私钥
- `reality_public_key`: Reality协议公钥
- `reality_short_id`: Reality短ID
- `reality_dest`: Reality目标域名
- `reality_server_names`: Reality服务器名称列表
- `spider_x`: SpiderX路径
- `fingerprint`: 浏览器指纹（如firefox、chrome等）
- `flow`: 流控类型（如xtls-rprx-vision）

## 实现策略

### 1. 数据结构更新
需要在 `web/service/v2board.go` 中更新 `User` 结构体，添加 `VlessConfig` 字段支持。

```go
type VlessConfig struct {
    InboundPort        int      `json:"inbound_port"`
    InboundTag         string   `json:"inbound_tag"`
    RealityPrivateKey  string   `json:"reality_private_key"`
    RealityPublicKey   string   `json:"reality_public_key"`
    RealityShortId     string   `json:"reality_short_id"`
    RealityDest        string   `json:"reality_dest"`
    RealityServerNames []string `json:"reality_server_names"`
    SpiderX            string   `json:"spider_x"`
    Fingerprint        string   `json:"fingerprint"`
    Flow               string   `json:"flow"`
}

type User struct {
    Id          int          `json:"id"`
    Uuid        string       `json:"uuid"`
    SpeedLimit  int          `json:"speed_limit"`
    Email       string       `json:"email"`
    VlessConfig *VlessConfig `json:"vless_config,omitempty"`
}
```

### 2. 同步逻辑重构

#### 核心流程
1. **获取用户列表**: 调用V2Board API获取所有用户
2. **遍历用户**: 处理每个用户
3. **检查配置**: 判断用户是否有 `vless_config`
4. **创建/更新Inbound**:
   - 根据 `inbound_tag` 查找是否已存在对应的Inbound
   - 不存在则创建新Inbound
   - 存在则更新Inbound配置
5. **添加客户端**: 将用户作为客户端添加到Inbound中
6. **更新数据库**: 保存Inbound配置

#### 关键处理点

##### Inbound配置生成
- **Protocol**: `vless`
- **Port**: 使用 `vless_config.inbound_port`
- **Tag**: 使用 `vless_config.inbound_tag`（必须唯一）
- **Settings**: 包含客户端列表和decryption设置
- **StreamSettings**: 配置Reality相关参数

##### Reality StreamSettings结构
```go
{
  "network": "tcp",
  "security": "reality",
  "realitySettings": {
    "show": false,
    "dest": "tesla.com",
    "xver": 0,
    "serverNames": ["www.tesla.com"],
    "privateKey": "xxx",
    "shortIds": ["23180e5d6724"],
    "spiderX": "/",
    "fingerprint": "firefox"
  }
}
```

##### 客户端配置
```go
{
  "id": "user-uuid",
  "email": "user@example.com",
  "flow": "xtls-rprx-vision",
  "enable": true,
  "speedLimit": 0,
  "totalGB": 0,
  "expiryTime": 0,
  "limitIp": 0
}
```

### 3. 错误处理

- **API调用失败**: 记录日志，跳过本次同步
- **Inbound创建失败**: 记录日志，继续处理下一个用户
- **Tag重复**: 使用现有Inbound，更新配置和客户端
- **端口冲突**: 记录警告，跳过该用户

### 4. 清理策略

对于不再存在于V2Board用户列表中的用户：
- 方案1: 禁用对应的客户端（保留Inbound）
- 方案2: 删除对应的Inbound（如果只有一个客户端）
- **当前采用**: 方案1，只禁用客户端，保留Inbound结构

## 优势对比

### 新策略优势
1. **灵活性**: 每个用户可以有独立的Inbound配置
2. **隔离性**: 用户间配置互不影响
3. **个性化**: 支持用户级别的Reality、端口等配置
4. **可扩展**: 便于后续添加更多用户级配置

### 旧策略劣势
1. **固定性**: 所有用户共享相同的Inbound配置
2. **限制性**: 无法为不同用户设置不同的端口、Reality配置
3. **依赖性**: 需要提前在系统中创建Inbound

## 迁移注意事项

1. **兼容性**: 保持对旧API的兼容（如果V2Board同时返回旧格式）
2. **数据迁移**: 无需迁移，新策略会自动创建所需Inbound
3. **配置验证**: 确保V2Board返回的配置参数有效
4. **资源管理**: 注意端口占用和标签唯一性
5. **日志记录**: 详细记录创建、更新、删除操作便于调试

## 测试建议

1. 测试有 `vless_config` 的用户能否正确创建Inbound
2. 测试无 `vless_config` 的用户是否被正确跳过
3. 测试重复同步时的更新逻辑
4. 测试用户删除后的清理逻辑
5. 测试端口冲突和标签冲突的处理
6. 测试Reality配置是否正确应用
