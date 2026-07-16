# Affiliate Attribution Security Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为推广入口增加可信代理解析、多维 Redis 限流和签名注册归因。

**Architecture:** 复用 Gin 可信代理和现有 Redis Lua 限流器；由落地页签发 HMAC Cookie，全局中间件验证后通过 context 将访问事件交给现有邀请绑定事务。数据库事件表记录已验证注册关联。

**Tech Stack:** Go、Gin、go-redis、PostgreSQL、Vue、Docker Compose

---

### Task 1: 扩展可组合限流键

**Files:**
- Modify: `backend/internal/middleware/rate_limiter.go`
- Test: `backend/internal/middleware/rate_limiter_test.go`

- [ ] 先写失败测试，验证自定义键经过哈希/规范化后隔离不同邀请码，并且空键 fail-close。
- [ ] 运行 `go test ./internal/middleware -run TestRateLimiterCustomKey -count=1`，确认因功能缺失失败。
- [ ] 为 `RateLimitOptions` 增加 `KeyFunc`，键值只以 SHA-256 摘要进入 Redis。
- [ ] 重跑测试确认通过。

### Task 2: 签名归因 Cookie 与可信 IP

**Files:**
- Create: `backend/internal/handler/affiliate_attribution.go`
- Modify: `backend/internal/handler/affiliate_landing_handler.go`
- Test: `backend/internal/handler/affiliate_landing_handler_test.go`

- [ ] 先写失败测试覆盖签名、篡改、过期、HttpOnly/SameSite 和可信客户端 IP。
- [ ] 运行定向测试确认失败原因正确。
- [ ] 实现版本化 HMAC Token 与验证中间件，落地页使用 `ip.GetTrustedClientIP`。
- [ ] 重跑测试确认通过。

### Task 3: 访问事件注册关联

**Files:**
- Create: `backend/migrations/180_affiliate_verified_attribution.sql`
- Modify: `backend/internal/service/affiliate_service.go`
- Modify: `backend/internal/repository/affiliate_repo.go`
- Test: `backend/migrations/affiliate_growth_migration_test.go`
- Test: `backend/internal/service/affiliate_authorization_test.go`

- [ ] 先写失败测试覆盖迁移隐私字段、context 归因和邀请码不匹配拒绝关联。
- [ ] 运行定向测试确认失败。
- [ ] 增加事件 ID 返回、context 载荷和事务内关联更新。
- [ ] 重跑测试确认通过。

### Task 4: 路由接入与生产配置文档

**Files:**
- Modify: `backend/internal/server/routes/common.go`
- Modify: `backend/internal/server/router.go`
- Modify: `deploy/config.example.yaml`
- Modify: `deploy/README.md`
- Test: `backend/internal/server/routes/routes_test.go`（若无对应文件则在 `common_test.go` 新建）

- [ ] 先写失败路由测试，覆盖 429 与归因中间件进入注册路由。
- [ ] 接入 IP、指纹、邀请码三层限流及全局 Cookie 验证中间件。
- [ ] 写明 Cloudflare+Nginx 可信代理配置，默认继续为空。
- [ ] 重跑路由测试。

### Task 5: 全量验证与运行态回归

**Files:**
- Verify only

- [ ] 运行 `go test ./...`。
- [ ] 运行 `go test -tags embed ./...`。
- [ ] 运行 `pnpm run build`。
- [ ] 重建 Docker 开发栈，确认健康、迁移存在、无效码 404、超限 429、有效 Cookie 注册归因链路。
- [ ] 检查 `git diff --check` 与工作区状态后提交并推送 `zzz/main`。
