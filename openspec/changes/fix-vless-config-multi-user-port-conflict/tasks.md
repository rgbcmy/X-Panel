## 1. 新增基于端口的入站查找辅助方法

- [x] 1.1 在 `web/job/v2board_sync_job.go` 的 `V2boardSyncJob` 中新增 `getInboundByPort(port int) (*model.Inbound, error)`——遍历 `GetAllInbounds()` 并返回第一个 `Port` 等于请求端口的入站；若无匹配则返回 `nil, fmt.Errorf("record not found")`。

## 2. 在 processUserInbound 中增加端口回退逻辑

- [x] 2.1 在 `processUserInbound` 中，当基于标签的查找未找到已有入站（`existingInbound == nil`）后，调用 `getInboundByPort(config.InboundPort)`。
- [x] 2.2 若找到端口匹配的入站，则调用 `updateUserInbound(portInbound, user)` 并返回——不再尝试创建新入站。
- [x] 2.3 仅当标签查找和端口查找均未找到已有入站时，才调用 `createUserInbound(user)`。

## 3. 验证与测试

- [x] 3.1 执行 `go build` 构建项目，确认零编译错误。
- [ ] 3.2 通过日志手动验证：使用 ≥2 个共享相同 `inbound_port` 的用户运行一次同步周期，确认 `processedCount` 等于带有 `vless_config` 的用户总数，且 `errorCount` 为 0。
- [ ] 3.3 通过面板 UI 确认所有用户均以客户端身份出现在共享入站中。
