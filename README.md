# linode-tool

`linode-tool` 是一个个人使用的 Linode CLI 开机工具，用于快速批量创建节点测试 VPS。

它保持单机 CLI 形式，不包含 Web、后台服务或数据库。

## 固定配置

- 套餐：`g6-nanode-1`（Linode Nanode 1 GB / $5 套餐）
- 系统：`linode/debian12`
- 实例名称：`cg-node-001`、`cg-node-002`、`cg-node-003`……
- Region：运行时通过 Linode API 动态获取，本地映射用于显示中文名称
- Firewall：为每台新实例创建并绑定一个 Firewall
  - Inbound TCP `1-65535`：允许 `0.0.0.0/0` 和 `::/0`
  - Inbound UDP `1-65535`：允许 `0.0.0.0/0` 和 `::/0`
  - Outbound：全部允许

> [!WARNING]
> 默认 Firewall 按项目用途开放全部 TCP 和 UDP 端口，不适合直接承载包含敏感数据的生产服务。

## 安装

需要 Go 1.23 或更高版本。

```bash
git clone https://github.com/pixingzoudaiyuexing/linode-tool.git
cd linode-tool
go build -o linode-tool ./cmd/linode-tool
sudo install -m 0755 linode-tool /usr/local/bin/linode-tool
```

## 配置

在 [Linode Cloud Manager](https://cloud.linode.com/profile/tokens) 创建 Personal Access Token，至少授予 Linodes 和 Firewalls 的读写权限，然后设置环境变量：

```bash
export LINODE_TOKEN=xxxx
```

可以将该命令加入当前 Shell 的配置文件。请不要把真实 Token 提交到 Git 仓库。

## 使用

### 创建实例

```bash
linode-tool create
```

命令会依次要求：

1. 选择亚洲、欧洲或美洲
2. 选择 Linode API 返回的具体 Region
3. 输入 `Root Password`（交互终端中不会回显）
4. 输入创建数量

实例按顺序逐台创建，并显示 `[当前数量/总数量]` 进度。已有同名实例时，会自动选择下一个可用的 `cg-node-NNN` 名称。实例创建成功后，工具会创建并绑定 Firewall；如果 Firewall 创建失败，工具会尝试删除刚创建的实例，避免留下不符合预期且继续计费的实例。

### 查看实例

```bash
linode-tool list
```

输出实例 ID、名称、区域、IPv4 地址和状态。

### 删除实例

```bash
linode-tool delete
```

从实例列表中选择目标，并输入 `yes` 二次确认后删除。删除 Linode 实例不可恢复，请确认 ID 和名称无误。

### 查看地区

```bash
linode-tool regions
```

显示 Linode API 当前返回且支持创建 Linode 的地区。新 Region 如果暂时没有本地中文映射，会显示 API 官方名称，但不会依赖硬编码的完整 Region 列表。
