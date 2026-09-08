# linode-tool

`linode-tool` 是一个面向个人节点测试的 Linode CLI 工具。

它直接调用 Linode API，快速批量创建和管理 VPS。项目保持简单：没有 Web 界面、后台服务、数据库、Docker 或多云适配层。

## 功能

- 交互选择 Asia、Europe、America 三级地区菜单
- Region 列表运行时从 Linode API 获取，不硬编码完整区域列表
- 批量创建 Linode 实例并显示进度
- 创建实例后自动创建并绑定全开放 Firewall
- 查看实例 ID、名称、区域、IPv4 和状态
- 交互选择并确认删除实例

## 固定配置

每次创建都使用以下配置，命令行不会提供套餐或系统选择：

| 配置项 | 固定值 |
| --- | --- |
| 套餐 | `g6-nanode-1` |
| 系统 | `linode/debian12` |
| 名称 | `cg-node-001`、`cg-node-002`、`cg-node-003`…… |

已有同名实例时，工具会自动寻找下一个可用编号。

Firewall 使用全放行策略：

- Inbound：全部协议、全部端口、所有 IPv4/IPv6 来源（`ACCEPT`）
- Outbound：全部协议、全部端口、所有目标（`ACCEPT`）

Firewall 名称包含 Linode 实例 ID，避免实例名称被重复使用时与历史 Firewall 重名。

## 安装

### 一键安装

适用于 Debian/Ubuntu。系统已有 Go 1.23 或更高版本时直接复用；没有 Go 或版本过低时，脚本会自动安装官方 Go 1.23.12：

```bash
curl -fsSL https://raw.githubusercontent.com/pixingzoudaiyuexing/linode-tool/main/install.sh | sh
```

安装脚本会检查：

- 当前系统是否为 Debian 或 Ubuntu
- Go 是否存在且版本不低于 1.23；必要时自动安装 Go 1.23.12
- 是否具备 root 或 sudo 权限
- `/usr/local/bin/linode-tool` 是否成功安装并可执行

自动安装的 Go 放在 `/usr/local/lib/linode-tool/go`，不会覆盖系统已有的 Go。脚本会直接下载 GitHub `main` 分支源码后本地构建，避免 Go 模块代理缓存旧分支版本。脚本只执行一次构建和安装，不会创建服务或自动更新。检查失败时会输出明确错误并退出。

### 手动构建

适用于其他已安装 Go 1.23 或更高版本的开发环境：

```bash
git clone https://github.com/pixingzoudaiyuexing/linode-tool.git
cd linode-tool
go build -o linode-tool ./cmd/linode-tool
sudo install -m 0755 linode-tool /usr/local/bin/linode-tool
```

## 配置 Token

在 [Linode Cloud Manager](https://cloud.linode.com/profile/tokens) 创建 Personal Access Token。Token 需要能够读取和管理 Linodes、Regions、Firewalls。

可以提前在当前 Shell 中设置：

```bash
export LINODE_TOKEN=xxxx
```

如果没有设置环境变量，运行 `linode-tool` 时会进入交互界面并隐藏输入 Linode API Token。Token 只在当前进程内使用，不会写入文件。

不要把真实 Token 写入 Git 仓库、脚本或公开日志。

## 使用

### 交互菜单

直接运行：

```bash
linode-tool
```

会进入主菜单，可选择创建实例、查看实例、删除实例、查看地区或退出。原有的 `linode-tool create`、`linode-tool list`、`linode-tool delete` 和 `linode-tool regions` 子命令仍然可用。

### 创建实例

```bash
linode-tool create
```

交互流程：

1. 选择一级地区：亚洲、欧洲或美洲
2. 选择 Linode API 返回的具体 Region
3. 输入 `Root Password`
4. 输入创建数量

示例进度：

```text
[1/5] 创建中...
[1/5] 创建成功: cg-node-001 (ID: 123456, Firewall ID: 7890)
```

在真实终端中输入 Root Password 时不会回显。实例创建成功但 Firewall 创建失败时，工具会尝试删除刚创建的实例，并报告清理结果。

### 查看实例

```bash
linode-tool list
```

输出字段：`ID`、`名称`、`区域`、`IP`、`状态`。

### 删除实例

```bash
linode-tool delete
```

工具会列出实例供选择，最后一项为“全部删除”。删除单台实例时，确认提示直接回车即同意，也可以输入 `y` 或 `yes`；选择“全部删除”时必须手动输入完整的 `yes`。删除 Linode 实例不可恢复，请确认实例名称和 ID。

### 查看地区

```bash
linode-tool regions
```

该命令用于查看当前 API 返回的可用地区。已维护中文名称的 Region 显示中文；新 Region 暂无映射时显示 Linode 官方名称。

## 安全提示

默认 Firewall 会放行所有入站和出站流量，仅适合节点测试。不要直接将该默认策略用于承载敏感数据的生产服务。

创建和删除操作会真实修改 Linode 账户资源并产生费用。首次使用建议先创建数量为 `1` 的实例，确认区域、IP 和 Firewall 状态后再进行批量操作。

## 开发验证

本地执行：

```bash
go test ./...
go vet ./...
go build ./...
```
