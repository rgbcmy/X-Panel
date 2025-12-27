# 快速修复指南 - V2Board 多 Inbound 同步

## 问题
从 v2board 同步客户端时报错：`Duplicate email`，只能同步到第一个 inbound。

## 原因
数据库 `client_traffics` 表的 `email` 字段设置了全局唯一约束，同一个 email 不能在多个 inbound 中存在。

## 解决方案
将唯一约束改为联合唯一约束 `(inbound_id, email)`，允许同一个 email 在不同 inbound 中存在。

## 🎉 自动升级（推荐）

**好消息！** 数据库迁移已经集成到启动流程中，只需更新代码并重启即可自动升级！

### 快速执行步骤

#### 1. 备份数据库（建议）
```bash
sudo cp /etc/x-ui/x-ui.db /etc/x-ui/x-ui.db.backup.$(date +%Y%m%d_%H%M%S)
```

#### 2. 停止服务
```bash
sudo x-ui stop
```

#### 3. 更新代码
```bash
cd /path/to/X-Panel
git pull  # 如果从 git 仓库
# 或者手动替换修改的文件
```

#### 4. 重新编译
```bash
go build -o x-ui main.go
sudo mv x-ui /usr/local/x-ui/
```

#### 5. 启动服务（自动迁移）
```bash
sudo x-ui start
```

**就这么简单！** 启动时会自动检测并执行数据库迁移。

#### 6. 验证
查看日志确认迁移成功：
```bash
sudo x-ui log | grep -i "client_traffics"
```

你应该看到类似这样的消息：
```
Migrating client_traffics table to support multiple inbounds per email...
client_traffics table migration completed successfully!
```

如果已经迁移过，会显示：
```
client_traffics table already migrated, skipping...
```

## 🔧 手动迁移（备选方案）

如果你需要在不重启服务的情况下手动迁移，或者遇到自动迁移问题，可以使用以下方法：

### 方式 A: 使用 Go 脚本
```bash
cd /path/to/X-Panel/migration
go run migrate_client_traffics.go
```

### 方式 B: 使用 SQL 脚本
```bash
sudo sqlite3 /etc/x-ui/x-ui.db < /path/to/X-Panel/database/migration_add_composite_unique.sql
```

## 验证同步功能

1. 登录 X-Panel 管理面板
2. 检查多个 inbound 是否都启用了 v2board 同步
3. 手动触发同步或等待自动同步
4. 检查日志，应该不再出现 "Duplicate email" 错误
5. 验证客户端列表，同一个用户应该出现在多个 inbound 中

## 如果出现问题

### 回滚数据库
```bash
sudo x-ui stop
sudo cp /etc/x-ui/x-ui.db.backup.* /etc/x-ui/x-ui.db
sudo x-ui start
```

### 检查日志
```bash
sudo x-ui log
# 或者
sudo journalctl -u x-ui -f
```

## 修改的核心文件

1. `xray/client_traffic.go` - 数据模型定义
2. `web/service/inbound.go` - 数据库操作逻辑
3. `database/migration_add_composite_unique.sql` - SQL 迁移脚本
4. `migration/migrate_client_traffics.go` - Go 迁移脚本

## 技术说明

### 数据库变更
```sql
-- 旧约束
email TEXT UNIQUE

-- 新约束
UNIQUE(inbound_id, email)
```

### 代码变更
```go
// 查询/更新/删除时，从：
WHERE email = ?

// 改为：
WHERE inbound_id = ? AND email = ?
```

## 联系方式

如有问题，请查看详细文档：
- `migration/README.md` - 完整迁移指南
- `migration/TESTING.md` - 测试指南
- `CHANGES.md` - 详细技术说明
