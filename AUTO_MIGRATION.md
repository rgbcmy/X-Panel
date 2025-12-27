# V2Board 多 Inbound 同步 - 自动升级版

## ✨ 重大改进

数据库迁移现在已经**完全自动化**！不需要手动执行任何迁移脚本。

## 🚀 升级步骤（超简单）

```bash
# 1. 备份数据库（建议）
sudo cp /etc/x-ui/x-ui.db /etc/x-ui/x-ui.db.backup

# 2. 停止服务
sudo x-ui stop

# 3. 更新代码并编译
cd /path/to/X-Panel
git pull
go build -o x-ui main.go
sudo mv x-ui /usr/local/x-ui/

# 4. 启动服务 - 自动完成数据库迁移！
sudo x-ui start
```

就这么简单！🎉

## 工作原理

在 `database/db.go` 的 `initModels()` 函数中添加了自动迁移逻辑：

```go
func initModels() error {
    // 特殊处理：迁移 client_traffics 表的唯一约束
    if err := migrateClientTrafficsTable(); err != nil {
        return err
    }
    
    // ... 其他表的 AutoMigrate
}
```

`migrateClientTrafficsTable()` 函数会：
- ✅ 自动检测是否需要迁移
- ✅ 创建备份表（client_traffics_backup）
- ✅ 使用事务保护，失败自动回滚
- ✅ 已迁移的数据库会自动跳过
- ✅ 完整的日志记录
- ✅ 为旧数据的 `inbound_id` 设置默认值 1（解决 NOT NULL 约束问题）

## 验证迁移

启动后检查日志：

```bash
sudo x-ui log | grep -i "client_traffics"
```

**首次迁移时会看到：**
```
Migrating client_traffics table to support multiple inbounds per email...
client_traffics table migration completed successfully!
Backup table 'client_traffics_backup' has been created and can be dropped after verification
```

**已迁移的数据库会显示：**
```
client_traffics table already migrated, skipping...
```

## 验证功能

1. 配置多个 inbound 并启用 v2board 同步
2. 确保不同 inbound 配置了不同的 v2board node ID
3. 触发同步 - 应该不再出现 "Duplicate email" 错误
4. 检查数据库：同一个用户应该出现在多个 inbound 中

```sql
-- 查看同一用户在不同 inbound 中的记录
SELECT ct.email, i.remark, ct.enable 
FROM client_traffics ct
JOIN inbounds i ON ct.inbound_id = i.id
WHERE ct.email = 'user@example.com';
```

## 清理备份表（可选）

迁移成功并验证无误后，可以删除备份表：

```bash
sqlite3 /etc/x-ui/x-ui.db "DROP TABLE IF EXISTS client_traffics_backup;"
```

## 回滚

如果出现问题：

```bash
sudo x-ui stop
sudo cp /etc/x-ui/x-ui.db.backup /etc/x-ui/x-ui.db
sudo x-ui start
```

## 技术细节

### 修改的核心文件

1. **`database/db.go`** - 添加自动迁移逻辑
2. **`xray/client_traffic.go`** - 更新模型定义，使用联合唯一索引
3. **`web/service/inbound.go`** - 更新数据库操作，支持 inbound_id

### 数据库变更

```sql
-- 旧结构
email TEXT UNIQUE

-- 新结构
UNIQUE(inbound_id, email)
CREATE UNIQUE INDEX idx_inbound_email ON client_traffics(inbound_id, email)
```

**数据迁移处理：**
- 旧数据中 `inbound_id` 为 NULL 或 0 的记录会自动设置为 1
- 使用 SQL: `COALESCE(NULLIF(inbound_id, 0), 1)` 确保所有记录都有有效的 inbound_id

## 手动迁移（备用）

如果自动迁移失败，可以使用手动方式：

### 方式 1: Go 脚本
```bash
cd migration
go run migrate_client_traffics.go
```

### 方式 2: SQL 脚本
```bash
sqlite3 /etc/x-ui/x-ui.db < database/migration_add_composite_unique.sql
```

## 文档

- `QUICKFIX.md` - 快速修复指南
- `CHANGES.md` - 详细技术说明
- `migration/README.md` - 完整迁移指南
- `migration/TESTING.md` - 测试指南

## 优势

- ✅ **零配置** - 不需要手动执行任何脚本
- ✅ **自动检测** - 智能判断是否需要迁移
- ✅ **安全可靠** - 自动备份，事务保护
- ✅ **幂等操作** - 可以重复运行，不会重复迁移
- ✅ **完整日志** - 方便问题排查
- ✅ **向后兼容** - 对旧数据库友好
