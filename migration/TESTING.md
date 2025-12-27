# V2Board 同步多 Inbound 测试说明

## 测试场景

验证同一个 v2board 用户可以同步到多个不同的 inbound（不同节点）。

## 前置条件

1. 已完成数据库迁移（参见 migration/README.md）
2. 已配置 v2board 连接信息
3. 有至少 2 个 inbound 启用了 v2board 同步，并配置了不同的 v2board node ID

## 测试步骤

### 1. 准备测试环境

创建两个测试 inbound：

- **Inbound 1**
  - 名称: "测试节点1"
  - 协议: vless
  - V2Board 启用: ✓
  - V2Board Node ID: "1"
  - V2Board Node Type: "v2ray"

- **Inbound 2**
  - 名称: "测试节点2"  
  - 协议: vless
  - V2Board 启用: ✓
  - V2Board Node ID: "2"
  - V2Board Node Type: "v2ray"

### 2. 在 V2Board 中配置

确保 v2board 中：
- 节点 1 和节点 2 都已创建
- 至少有一个用户订阅了这两个节点

### 3. 触发同步

手动触发 v2board 同步任务，或等待自动同步。

### 4. 验证结果

检查数据库中 `client_traffics` 表：

```bash
sqlite3 /etc/x-ui/x-ui.db "SELECT inbound_id, email, enable FROM client_traffics WHERE email LIKE '%test%' ORDER BY inbound_id;"
```

**预期结果：**
- 同一个用户的 email 应该出现在两个不同的 inbound_id 中
- 两条记录的 enable 字段都应该是 true（1）

### 5. 功能验证

1. 客户端使用订阅链接应该能看到两个节点
2. 两个节点都应该可以正常连接
3. 流量统计应该分别记录在各自的 inbound 中

## 预期行为

- ✅ 同步成功，没有 "Duplicate email" 错误
- ✅ 同一个用户可以在多个 inbound 中存在
- ✅ 每个 inbound 的客户端配置独立管理
- ✅ 流量统计分别记录

## 常见问题

### Q: 同步后只能看到一个 inbound 的客户端？

A: 检查：
1. 两个 inbound 是否都启用了 v2board 同步
2. v2board node ID 是否配置正确
3. 查看日志确认同步是否成功

### Q: 仍然出现 "Duplicate email" 错误？

A: 检查：
1. 是否已正确执行数据库迁移
2. 是否重启了 X-Panel 服务
3. 检查数据库索引：`sqlite3 /etc/x-ui/x-ui.db "PRAGMA index_list(client_traffics);"`

### Q: 旧数据如何处理？

A: 
- 迁移脚本会自动处理现有数据
- 如果有冲突，会保留第一条记录
- 建议在测试环境先验证

## 数据库查询示例

### 查看所有 inbound 的客户端分布
```sql
SELECT 
    i.remark as inbound_name,
    COUNT(ct.id) as client_count
FROM inbounds i
LEFT JOIN client_traffics ct ON i.id = ct.inbound_id
WHERE i.v2board_enabled = 1
GROUP BY i.id, i.remark;
```

### 查看同一个 email 在不同 inbound 中的记录
```sql
SELECT 
    ct.email,
    i.remark as inbound_name,
    ct.enable,
    ct.up + ct.down as total_traffic
FROM client_traffics ct
JOIN inbounds i ON ct.inbound_id = i.id
WHERE ct.email = 'user@example.com'
ORDER BY i.id;
```

### 查看有多个 inbound 的用户
```sql
SELECT 
    email,
    COUNT(DISTINCT inbound_id) as inbound_count
FROM client_traffics
GROUP BY email
HAVING inbound_count > 1
ORDER BY inbound_count DESC;
```
