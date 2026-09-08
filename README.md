# linode-tool

Linode 快速节点部署 CLI 工具。

目标：

- 固定 Linode $5 Nanode 1GB
- 固定 Debian 12
- 选择可用区域
- 自动创建实例
- 自动创建并绑定全开放防火墙
- 查看实例
- 删除实例

## 当前状态

MVP 开发中。

## 设计原则

- CLI 工具，不做 Web
- 无数据库
- 直接调用 Linode API
- 区域动态获取

## 计划

- [x] 项目初始化
- [ ] Linode API 客户端
- [ ] 创建实例
- [ ] Firewall 全开放
- [ ] 删除实例
- [ ] 实例列表
