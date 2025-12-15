# X-Panel 发布指南

本文档详细说明如何发布 X-Panel 的新版本。

## 📋 目录

- [准备工作](#准备工作)
- [发布流程](#发布流程)
- [数据库迁移说明](#数据库迁移说明)
- [回滚操作](#回滚操作)
- [常见问题](#常见问题)

---

## 🔧 准备工作

### 1. 确保所有修改已完成并测试通过

```bash
# 在本地测试编译
cd /path/to/X-Panel
CGO_ENABLED=1 go build -o x-ui main.go

# 运行测试（如果有）
go test ./...

# 检查代码格式
go fmt ./...
```

### 2. 更新版本号

编辑 `config/version` 文件，更新版本号：

```bash
# 查看当前版本
cat config/version

# 编辑版本号（例如：v26.0.0）
echo "v26.0.0" > config/version
```

### 3. 更新 CHANGELOG（推荐）

在项目根目录创建或更新 `CHANGELOG.md`：

```markdown
## [v26.0.0] - 2025-12-15

### 新增功能
- ✨ 新增 V2Board 对接支持
- ✨ 支持配置 V2Board 节点 ID 和节点类型

### 数据库变更
- 🗄️ Inbound 表新增字段：
  - `v2board_enabled` - 是否启用 V2Board 对接
  - `v2board_node_id` - V2Board 节点 ID
  - `v2board_node_type` - V2Board 节点类型

### 改进
- 💡 优化数据库自动迁移逻辑
- 📝 完善文档说明

### 修复
- 🐛 修复某个 bug 描述
```

---

## 🚀 发布流程

### 步骤 1：提交所有修改

```bash
# 查看当前修改状态
git status

# 添加所有修改的文件
git add .

# 提交修改（使用语义化提交信息）
git commit -m "feat: add v2board integration support

- Add v2board integration fields to Inbound model
- Add V2boardEnabled, V2boardNodeId, V2boardNodeType fields
- Database auto-migrate on startup
- No data loss for existing installations"
```

**提交信息规范：**
- `feat:` - 新功能
- `fix:` - 修复 bug
- `docs:` - 文档更新
- `refactor:` - 代码重构
- `perf:` - 性能优化
- `test:` - 测试相关
- `chore:` - 构建/工具链相关

### 步骤 2：推送到 GitHub

```bash
# 推送到 main 分支
git push origin main

# 等待推送完成，确认没有错误
```

### 步骤 3：创建版本标签

```bash
# 创建带注释的标签（推荐）
git tag -a v26.0.0 -m "Release v26.0.0: V2Board Integration

新增功能：
- V2Board 对接支持
- 节点 ID 和类型配置

数据库自动迁移，无需手动操作"

# 或创建轻量级标签
git tag v26.0.0

# 推送标签到 GitHub
git push origin v26.0.0
```

### 步骤 4：在 GitHub 上创建 Release

1. **访问 Release 页面**
   
   打开浏览器访问：`https://github.com/rgbcmy/x-panel/releases/new`

2. **填写 Release 信息**

   - **Choose a tag**: 选择刚才创建的标签 `v26.0.0`
   - **Release title**: `X-Panel v26.0.0 - V2Board Integration`
   - **Describe this release**: 填写详细的更新说明

   ```markdown
   ## 🎉 新功能
   
   ### V2Board 对接支持
   - ✅ 新增 V2Board 集成功能
   - ✅ 支持配置节点 ID 和节点类型
   - ✅ 可在面板中启用/禁用 V2Board 对接
   
   ## 🔄 数据库更新
   
   本版本新增以下数据库字段，**启动时自动迁移，无需手动操作**：
   
   | 字段名 | 类型 | 默认值 | 说明 |
   |--------|------|--------|------|
   | `v2board_enabled` | boolean | false | 是否启用 V2Board 对接 |
   | `v2board_node_id` | string | - | V2Board 节点 ID |
   | `v2board_node_type` | string | - | V2Board 节点类型 |
   
   **⚠️ 重要提示**：
   - 现有安装的用户，更新后会自动添加新字段
   - 不会丢失任何现有数据
   - 不会影响现有配置
   - 数据库备份位置：`/etc/x-ui/x-ui.db`（建议更新前手动备份）
   
   ## 📦 安装与更新
   
   ### 新安装
   ```bash
   bash <(curl -Ls https://raw.githubusercontent.com/rgbcmy/x-panel/main/install.sh)
   ```
   
   ### 更新现有安装
   
   **方法 1：使用安装脚本（推荐）**
   ```bash
   bash <(curl -Ls https://raw.githubusercontent.com/rgbcmy/x-panel/main/install.sh)
   ```
   
   **方法 2：使用 x-ui 命令**
   ```bash
   x-ui update
   ```
   
   **方法 3：手动更新**
   ```bash
   systemctl stop x-ui
   rm -rf /usr/local/x-ui/
   cd /usr/local/
   wget -N --no-check-certificate https://github.com/rgbcmy/x-panel/releases/download/v26.0.0/x-ui-linux-amd64.tar.gz
   tar -xzf x-ui-linux-amd64.tar.gz
   rm x-ui-linux-amd64.tar.gz
   systemctl start x-ui
   ```
   
   ## 🖥️ 支持平台
   
   ### Linux
   - ✅ amd64 (x86_64)
   - ✅ arm64 (aarch64)
   - ✅ armv7
   - ✅ armv6
   - ✅ armv5
   - ✅ 386 (x86)
   - ✅ s390x
   
   ### Windows
   - ✅ amd64 (x86_64)
   - ✅ 386 (x86)
   
   ### 系统要求
   - CentOS 7+
   - Ubuntu 20.04+
   - Debian 11+
   - Fedora 36+
   - Arch Linux / Manjaro
   - Alpine Linux
   - AlmaLinux 9+
   - Rocky Linux 9+
   - Oracle Linux 8+
   
   ## 📝 配置说明
   
   ### 启用 V2Board 对接
   
   1. 登录 X-Panel 管理面板
   2. 进入入站配置页面
   3. 编辑需要对接的入站
   4. 在配置中找到 "V2Board 设置" 区域
   5. 启用 V2Board 对接
   6. 填写节点 ID 和节点类型
   7. 保存配置
   
   ## 🐛 已知问题
   
   - 暂无
   
   ## 🔗 相关链接
   
   - [项目主页](https://github.com/rgbcmy/x-panel)
   - [使用文档](https://github.com/rgbcmy/x-panel/wiki)
   - [问题反馈](https://github.com/rgbcmy/x-panel/issues)
   - [TG 交流群](https://t.me/XUI_CN)
   
   ## 📊 完整更新日志
   
   查看 [CHANGELOG.md](https://github.com/rgbcmy/x-panel/blob/main/CHANGELOG.md) 获取完整的版本历史。
   
   ---
   
   **感谢所有贡献者和用户的支持！** ❤️
   ```

3. **设置 Release 类型**
   
   - 取消勾选 **"Set as a pre-release"**（如果是正式版本）
   - 勾选 **"Set as the latest release"**

4. **点击 "Publish release"**

### 步骤 5：等待自动构建

发布 Release 后，GitHub Actions 会自动开始构建：

1. **查看构建进度**
   
   访问：`https://github.com/rgbcmy/x-panel/actions`

2. **构建内容**
   
   - 编译所有平台的二进制文件（Linux: 7个架构，Windows: 2个架构）
   - 下载 Xray-core 依赖
   - 下载 geo 数据文件（geoip.dat, geosite.dat, IR, RU 版本）
   - 打包成 tar.gz（Linux）和 zip（Windows）
   - 自动上传到 Release 页面

3. **构建时间**
   
   大约 **15-20 分钟**，构建完成后所有文件会自动出现在 Release 页面

4. **构建产物**
   
   ```
   x-ui-linux-amd64.tar.gz
   x-ui-linux-arm64.tar.gz
   x-ui-linux-armv7.tar.gz
   x-ui-linux-armv6.tar.gz
   x-ui-linux-armv5.tar.gz
   x-ui-linux-386.tar.gz
   x-ui-linux-s390x.tar.gz
   x-ui-windows-amd64.zip
   x-ui-windows-386.zip
   ```

### 步骤 6：验证发布

1. **检查 Release 页面**
   
   确认所有构建产物都已上传

2. **测试安装脚本**
   
   在新服务器上测试：
   ```bash
   bash <(curl -Ls https://raw.githubusercontent.com/rgbcmy/x-panel/main/install.sh)
   ```

3. **测试更新功能**
   
   在已安装的服务器上测试：
   ```bash
   x-ui update
   ```

4. **验证数据库迁移**
   
   ```bash
   # 检查数据库字段
   sqlite3 /etc/x-ui/x-ui.db "PRAGMA table_info(inbounds);" | grep v2board
   
   # 应该看到三个新字段：
   # v2board_enabled
   # v2board_node_id
   # v2board_node_type
   ```

---

## 🗄️ 数据库迁移说明

### 自动迁移机制

X-Panel 使用 GORM 的 `AutoMigrate` 功能，在每次启动时自动检查并更新数据库结构。

**工作原理**（`database/db.go` 第 29-44 行）：

```go
func initModels() error {
    models := []any{
        &model.User{},
        &model.Inbound{},  // 包含 V2Board 字段
        &model.LotteryWin{},
        // ...
    }
    for _, model := range models {
        if err := db.AutoMigrate(model); err != nil {
            return err
        }
    }
    return nil
}
```

### 迁移特性

- ✅ **自动添加新字段**：新增的字段会自动添加到表中
- ✅ **保留现有数据**：不会删除或修改现有数据
- ✅ **设置默认值**：新字段会使用模型中定义的默认值
- ✅ **幂等性**：重复执行不会出错
- ❌ **不会删除字段**：已删除的字段仍保留在数据库中（安全考虑）

### 手动备份数据库（可选但推荐）

```bash
# 备份数据库
cp /etc/x-ui/x-ui.db /etc/x-ui/x-ui.db.backup.$(date +%Y%m%d_%H%M%S)

# 验证备份
ls -lh /etc/x-ui/x-ui.db*
```

### 回滚数据库（如果需要）

```bash
# 停止服务
systemctl stop x-ui

# 恢复备份
cp /etc/x-ui/x-ui.db.backup.20251215_120000 /etc/x-ui/x-ui.db

# 启动服务
systemctl start x-ui
```

---

## ⏮️ 回滚操作

### 如果发布有问题，可以快速回滚：

### 1. 在 GitHub 上标记为 Pre-release

1. 访问出问题的 Release 页面
2. 点击 "Edit release"
3. 勾选 "Set as a pre-release"
4. 取消勾选 "Set as the latest release"
5. 保存

### 2. 删除标签并重新发布

```bash
# 删除本地标签
git tag -d v26.0.0

# 删除远程标签
git push --delete origin v26.0.0

# 在 GitHub 上删除 Release

# 修复问题后重新发布
git tag v26.0.0
git push origin v26.0.0
```

### 3. 用户端回滚到旧版本

```bash
# 安装指定版本
bash <(curl -Ls https://raw.githubusercontent.com/rgbcmy/x-panel/main/install.sh)
# 在脚本中选择自定义版本，输入旧版本号

# 或手动下载旧版本
cd /usr/local
systemctl stop x-ui
rm -rf x-ui/
wget -N https://github.com/rgbcmy/x-panel/releases/download/v2.3.0/x-ui-linux-amd64.tar.gz
tar -xzf x-ui-linux-amd64.tar.gz
rm x-ui-linux-amd64.tar.gz
systemctl start x-ui
```

---

## ❓ 常见问题

### Q1: 构建失败怎么办？

**A**: 检查 GitHub Actions 日志：
1. 访问 `https://github.com/rgbcmy/x-panel/actions`
2. 点击失败的工作流
3. 查看错误日志
4. 常见问题：
   - Go 依赖问题：检查 `go.mod` 和 `go.sum`
   - 交叉编译问题：检查 CGO 配置
   - 下载依赖失败：网络问题，重新运行工作流

### Q2: 如何测试 Release 不实际发布？

**A**: 使用 `workflow_dispatch` 手动触发：
1. 访问 Actions 页面
2. 选择 "Release X-Panel" 工作流
3. 点击 "Run workflow"
4. 构建产物会上传到 Artifacts，不会发布到 Release

### Q3: 数据库迁移失败怎么办？

**A**: 
```bash
# 1. 检查日志
journalctl -u x-ui -n 100

# 2. 手动检查数据库
sqlite3 /etc/x-ui/x-ui.db

# 3. 如果有备份，回滚
systemctl stop x-ui
cp /etc/x-ui/x-ui.db.backup /etc/x-ui/x-ui.db
systemctl start x-ui

# 4. 提交 Issue 到 GitHub
```

### Q4: 如何发布热修复版本？

**A**: 
```bash
# 使用三位版本号的修订版本
# v26.0.0 -> v2.4.1

echo "v2.4.1" > config/version
git add config/version
git commit -m "fix: hotfix for xxx issue"
git push origin main
git tag v2.4.1
git push origin v2.4.1

# 在 Release 中说明这是热修复版本
```

### Q5: 如何查看当前安装的版本？

**A**: 
```bash
# 方法 1
/usr/local/x-ui/x-ui -v

# 方法 2
x-ui

# 方法 3
cat /usr/local/x-ui/config/version
```

### Q6: Docker 镜像什么时候构建？

**A**: Docker 镜像在推送 tag 时自动构建：
- 推送 `v*.*.*` 格式的 tag 会触发 Docker 构建
- 镜像会推送到：
  - Docker Hub: `rgbcmy/x-panel:v26.0.0`
  - GHCR: `ghcr.io/rgbcmy/x-panel:v26.0.0`

---

## 📚 参考资源

- [语义化版本规范](https://semver.org/lang/zh-CN/)
- [约定式提交规范](https://www.conventionalcommits.org/zh-hans/)
- [GitHub Actions 文档](https://docs.github.com/cn/actions)
- [GORM 迁移文档](https://gorm.io/zh_CN/docs/migration.html)

---

## 📝 发布清单

发布前请确认以下事项：

- [ ] 所有功能已完成并测试通过
- [ ] 代码已格式化（`go fmt ./...`）
- [ ] 更新了 `config/version` 文件
- [ ] 更新了 `CHANGELOG.md`（如果有）
- [ ] 更新了相关文档
- [ ] 提交了所有修改到 Git
- [ ] 推送到 GitHub
- [ ] 创建了版本标签
- [ ] 在 GitHub 上创建了 Release
- [ ] 填写了详细的 Release Notes
- [ ] 等待 GitHub Actions 构建完成
- [ ] 验证所有构建产物已上传
- [ ] 测试安装脚本
- [ ] 测试更新功能
- [ ] 验证数据库迁移
- [ ] 在 TG 群通知更新

---

**祝发布顺利！** 🎉
