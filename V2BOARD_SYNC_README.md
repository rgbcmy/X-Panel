# V2Board 用户同步策略更新 - 快速开始

## 🎯 核心变更

**从**: 为已存在的Inbound添加用户  
**到**: 为每个用户动态创建专属Inbound

## 📋 新功能

✅ 支持用户级别的VLESS Reality配置  
✅ 自动创建和管理Inbound  
✅ 独立的端口和Reality参数  
✅ 智能跳过无配置用户  

## 🚀 快速验证

### 1. 检查V2Board API返回格式
确保API返回包含 `vless_config` 的用户数据：
```json
{
  "users": [
    {
      "id": 1,
      "uuid": "xxx",
      "email": "user@example.com",
      "speed_limit": null,
      "vless_config": {
        "inbound_port": 8024,
        "inbound_tag": "user-1-vless",
        "reality_private_key": "...",
        "reality_public_key": "...",
        ...
      }
    }
  ]
}
```

### 2. 编译部署
```bash
go build
systemctl restart x-ui
```

### 3. 查看日志
```bash
# 等待定时任务执行或手动触发
tail -f /var/log/x-ui/access.log | grep "V2board sync"
```

期望看到：
```
[INFO] V2board sync: processing 10 users
[INFO] successfully processed inbound for user 16 tag: user-16-vless-2
[INFO] V2board sync completed: processed 5 users, skipped 3 users, errors 0
```

### 4. 验证Inbound
登录管理面板，查看是否自动创建了新的Inbound。

## 📚 详细文档

- **需求和设计**: [V2BOARD_SYNC_STRATEGY.md](V2BOARD_SYNC_STRATEGY.md)
- **使用指南**: [V2BOARD_SYNC_USAGE.md](V2BOARD_SYNC_USAGE.md)  
- **变更总结**: [V2BOARD_SYNC_CHANGES.md](V2BOARD_SYNC_CHANGES.md)

## 🔍 代码变更

### 修改文件
1. `web/service/v2board.go` - 添加VlessConfig结构体
2. `web/job/v2board_sync_job.go` - 重写同步逻辑

### 核心代码片段

**新增结构体**:
```go
type VlessConfig struct {
    InboundPort        int      `json:"inbound_port"`
    InboundTag         string   `json:"inbound_tag"`
    RealityPrivateKey  string   `json:"reality_private_key"`
    // ... 更多字段
}
```

**新同步方法**:
```go
func (j *V2boardSyncJob) syncUserDrivenInbounds(allSetting *entity.AllSetting) error {
    // 1. 获取所有用户
    userList, err := j.v2boardService.GetUserList(allSetting, "", "")
    
    // 2. 遍历用户
    for _, user := range userList.Users {
        if user.VlessConfig == nil {
            continue // 跳过无配置用户
        }
        
        // 3. 创建或更新Inbound
        j.processUserInbound(&user, allSetting)
    }
}
```

## ⚠️ 注意事项

1. **端口冲突**: 确保 `inbound_port` 不与现有端口冲突
2. **Tag唯一性**: 确保 `inbound_tag` 在系统中唯一  
3. **API兼容**: V2Board必须返回新格式数据
4. **逐步迁移**: 建议先在测试环境验证

## 🐛 常见问题

**Q: 用户没有被添加？**  
A: 检查用户是否有 `vless_config` 字段，查看日志中的"skipping"信息

**Q: 端口冲突怎么办？**  
A: 系统会记录警告并跳过该用户，需要在V2Board调整端口配置

**Q: 如何手动触发同步？**  
A: 重启服务或等待定时任务执行（建议5-10分钟间隔）

**Q: 旧的Inbound会被删除吗？**  
A: 不会，新策略不会删除任何现有Inbound

## 📞 支持

如有问题，请查看详细文档或提交Issue。

---

**更新日期**: 2026-01-15  
**版本**: v1.0.0
