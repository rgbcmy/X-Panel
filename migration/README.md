# 数据库迁移指南

## 问题描述

在从 v2board 同步客户端信息时，由于 `client_traffics` 表中 `email` 字段设置了唯一约束，导致同一个用户不能同步到多个 inbound（不同节点），出现 "Duplicate email" 错误。

## 解决方案

将 `client_traffics` 表的唯一约束从单独的 `email` 字段改为 `(inbound_id, email)` 的联合唯一约束。这样同一个 email 可以在不同的 inbound 中存在。

## 迁移步骤

### 方法一：使用 Go 脚本（推荐）

1. **停止 X-Panel 服务**
   ```bash
   x-ui stop
   ```

2. **备份数据库（重要！）**
   ```bash
   cp /etc/x-ui/x-ui.db /etc/x-ui/x-ui.db.backup
   ```

3. **运行迁移脚本**
   ```bash
   cd /path/to/X-Panel/migration
   go run migrate_client_traffics.go
   ```

4. **重启 X-Panel 服务**
   ```bash
   x-ui start
   ```

### 方法二：手动执行 SQL

1. **停止 X-Panel 服务**
   ```bash
   x-ui stop
   ```

2. **备份数据库（重要！）**
   ```bash
   cp /etc/x-ui/x-ui.db /etc/x-ui/x-ui.db.backup
   ```

3. **执行 SQL 迁移**
   ```bash
   sqlite3 /etc/x-ui/x-ui.db < /path/to/X-Panel/database/migration_add_composite_unique.sql
   ```

4. **重启 X-Panel 服务**
   ```bash
   x-ui start
   ```

## 验证迁移

迁移完成后，可以通过以下方式验证：

```bash
sqlite3 /etc/x-ui/x-ui.db "PRAGMA table_info(client_traffics);"
sqlite3 /etc/x-ui/x-ui.db "PRAGMA index_list(client_traffics);"
```

你应该看到一个名为 `idx_inbound_email` 的复合唯一索引。

## 回滚

如果迁移后出现问题，可以从备份恢复：

```bash
x-ui stop
cp /etc/x-ui/x-ui.db.backup /etc/x-ui/x-ui.db
x-ui start
```

## 注意事项

1. **务必先备份数据库**
2. 迁移前停止 X-Panel 服务
3. 迁移完成后需要重启服务
4. 如果你有现有的重复 email 数据（同一个 email 在多个 inbound 中），迁移脚本会保留第一个，删除其他的

## 修改的文件

- `xray/client_traffic.go` - 更新了 `ClientTraffic` 模型，将 `email` 的 `gorm:"unique"` 改为 `gorm:"uniqueIndex:idx_inbound_email"`，并在 `InboundId` 字段也添加了相同的索引标记
- `web/service/inbound.go` - 更新了相关的数据库查询方法，在查询和删除时同时使用 `inbound_id` 和 `email`

## 影响范围

此迁移只影响 `client_traffics` 表的结构，不会影响其他表或功能。迁移后：
- 同一个 email 可以在不同的 inbound 中存在
- v2board 同步功能可以将同一个用户同步到多个节点
- 所有查询和更新操作都会考虑 `inbound_id` 和 `email` 的组合
