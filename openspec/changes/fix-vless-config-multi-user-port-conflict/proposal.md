## 问题背景

V2Board 同步任务目前只能成功为第一个用户创建 VLESS 入站。当多个用户共享相同的 `inbound_port` 时，第二个及之后的用户都会遭遇"端口已存在"错误——原因是 `processUserInbound` 仅通过 `inbound_tag` 查找已有入站，从不按端口查找，因此每次都会尝试为用户新建入站，触发端口唯一性约束。

## 变更内容

- 在 `processUserInbound` 中增加基于端口的入站查找：创建新入站前，先检查请求端口上是否已存在入站。
- 若找到端口匹配的入站，则将用户作为客户端加入该入站，而不是创建重复入站。
- 所有 `vless_config.inbound_port` 对应已有入站的用户，均会被正确地追加或更新为该入站的客户端。

## 能力范围

### 新增能力

- `multi-user-port-shared-inbound`：允许多个 V2Board 用户共享同一 `inbound_port`，每个用户作为独立客户端加入同一入站。

### 变更现有能力

<!-- 没有现有规范层级的需求发生变化 -->

## 影响范围

- **`web/job/v2board_sync_job.go`**：`processUserInbound` 新增基于端口的二次查找；新增辅助方法 `getInboundByPort`。
- 无 API 或数据结构变更；无破坏性变更。
- 所有带有 `vless_config` 的用户均可被成功配置，不受端口是否已被占用影响。
