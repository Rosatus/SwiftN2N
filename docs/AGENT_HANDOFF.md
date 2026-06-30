# SwiftN2N Agent Handoff

本文档面向后续接手 SwiftN2N 的开发 agent。它记录当前项目状态、架构边界、关键实现、已验证结论、常见故障和继续开发路线。

最后更新：2026-06-30

## 当前结论

SwiftN2N 是一个 Wails v2 桌面 GUI，用来启动、停止和观察官方 n2n v3 `edge` 二进制。项目不编译、不 patch n2n 源码，只负责：

- 解析 GUI 配置。
- 组装 `edge` 命令参数。
- 通过普通进程或 Linux `pkexec` privileged helper 启动 `edge`。
- 实时读取 stdout/stderr。
- 把关键日志映射为 UI 状态。
- 停止 helper/edge，避免残留进程。
- 打包官方 `edge` 二进制到 Wails 构建产物。

截至本文档更新时，Linux amd64 的核心链路已跑通：

- GUI 正常渲染。
- `Check Edge` 能检测官方 edge。
- Linux 普通用户 GUI 可通过 Polkit/pkexec 拉起 root helper。
- helper 可启动 edge，并将日志回传 GUI。
- `Disconnect` 已改为 privileged stop helper，不再由普通 GUI 直接 kill root/helper 进程。
- 日志噪音已做后端聚合降噪。
- 示例 supernode 配置已用 edge 本体验证可注册。真实 operator 信息不要写入公开仓库。

## 技术栈

- Wails v2
- Go
- React + TypeScript
- Vite
- Plain CSS
- 官方 n2n v3 `edge` 二进制

Linux/Pop!_OS/Ubuntu 24.04 构建使用 WebKitGTK 4.1：

```bash
wails build -clean -tags "webkit2_41"
```

Makefile 默认已经传入该 tag。

## 重要路径

```text
.
├── app.go                         # Wails 绑定 facade
├── main.go                        # Wails 入口 + hidden helper mode
├── internal/edge                  # n2n edge 配置、进程管理、helper、日志解析
├── internal/platform              # 平台进程组、kill、pkexec/manifest 差异
├── frontend/src                   # React UI
├── frontend/wailsjs               # Wails 自动生成绑定
├── scripts/download-edge-linux.sh # 官方 edge 下载与 SHA256 校验
├── Makefile                       # test/build/release/dev 入口
└── docs/AGENT_HANDOFF.md          # 本文档
```

核心文件职责：

```text
internal/edge/types.go       数据结构：Config/Status/LogEvent/EnvironmentStatus
internal/edge/config.go      GUI Config -> edge args/env
internal/edge/binary.go      edge 二进制定位
internal/edge/manager.go     Start/Stop/日志泵/状态机/Wails events
internal/edge/helper.go      --helper edge 与 --helper stop
internal/edge/parser.go      n2n 日志 -> SwiftN2N 状态
internal/platform/*.go       跨平台进程组、pkexec、taskkill
frontend/src/lib/edgeConfig.ts 默认配置/本地存储/日志工具
frontend/src/components/*    配置面板、连接动画、日志终端
```

## 常用命令

开发依赖：

```bash
cd <repo>
cd frontend && npm install
```

测试：

```bash
make test
cd frontend && npm run build
cd frontend && npm audit --audit-level=high
```

下载官方 Linux amd64 edge：

```bash
make edge-linux-amd64
```

开发运行：

```bash
make dev-linux
```

生产构建：

```bash
make build-linux
```

release archive：

```bash
make release-linux
```

输出：

```text
build/bin/SwiftN2N
build/bin/bin/linux/amd64/edge
dist/SwiftN2N-linux-amd64.tar.gz
```

## 官方 edge 二进制策略

项目通过 `scripts/download-edge-linux.sh` 下载官方 ntop/n2n GitHub release：

```text
https://github.com/ntop/n2n/releases/download/3.0/n2n_3.0.0-1038_amd64.deb
```

固定 SHA256：

```text
dd29694cf8c22487618218687c0b81961736ae1afe1c43bcfde6d48f017f16f6
```

脚本解包 `/usr/sbin/edge` 到：

```text
bin/linux/amd64/edge
```

`make build-linux` 后会复制到：

```text
build/bin/bin/linux/amd64/edge
```

不要从 Go 里编译 n2n，不要修改 n2n 源码。后续要支持更多平台时，应添加对应官方/可信二进制下载与 checksum。

补充：ntop/n2n `3.0` GitHub release 当前只发布 Linux rpm/deb 资产，没有 Windows/macOS `edge` 资产。因此本项目的 CI 策略是：

- Linux amd64：下载官方 deb，并校验 SHA256。
- Windows amd64：在 `windows-latest` runner 上从 `ntop/n2n` 的 `3.0-stable` 分支源码构建 `edge.exe`。
- macOS amd64：在 `macos-13` runner 上从 `ntop/n2n` 的 `3.0-stable` 分支源码构建 `edge`。

相关脚本：

```text
scripts/download-edge-linux.sh       Linux 官方 deb 下载
scripts/build-edge-from-source.sh    Windows/macOS 源码构建 edge
scripts/prepare-edge.sh              按目标平台准备 edge
scripts/package-release.sh           把 Wails 产物和 edge 打包
```

GitHub Actions workflow：

```text
.github/workflows/release.yml
```

workflow 在 push/PR 时上传 artifacts；推送 `v*` tag 时还会把构建产物上传到 GitHub Release。

## 配置模型

前端传给 Go 的结构在 `internal/edge/types.go`：

```go
type Config struct {
    EdgePath            string   `json:"edgePath"`
    Supernodes          []string `json:"supernodes"`
    Community           string   `json:"community"`
    Address             string   `json:"address"`
    Key                 string   `json:"key"`
    Cipher              string   `json:"cipher"`
    HeaderEncryption    bool     `json:"headerEncryption"`
    MAC                 string   `json:"mac"`
    DeviceName          string   `json:"deviceName"`
    MTU                 int      `json:"mtu"`
    LocalPort           string   `json:"localPort"`
    ManagementPort      int      `json:"managementPort"`
    Verbose             int      `json:"verbose"`
    AuthUsername        string   `json:"authUsername"`
    AuthPassword        string   `json:"authPassword"`
    FederationPublicKey string   `json:"federationPublicKey"`
    SupernodeOnly       string   `json:"supernodeOnly"`
    Compression         string   `json:"compression"`
    AcceptMulticast     bool     `json:"acceptMulticast"`
    EnableRouting       bool     `json:"enableRouting"`
    Routes              []string `json:"routes"`
    TrafficRules        []string `json:"trafficRules"`
    WindowsMetric       int      `json:"windowsMetric"`
}
```

参数组装在 `internal/edge/config.go`。重要规则：

- `community` 作为 `-c <community>` 传入。
- `key` 通过 `N2N_KEY` 环境变量传入，不用 `-k`，避免进程列表泄露。
- `authPassword` 通过 `N2N_PASSWORD` 环境变量传入，不用 `-J`。
- `cipher` 映射：
  - `none` -> `-A1`
  - `twofish` -> `-A2`
  - `aes` -> `-A3`
  - `chacha20` -> `-A4`
  - `speck` -> `-A5`
- `headerEncryption` 为 true 时加 `-H`。
- Linux/macOS 加 `-f`，Windows 不强加。
- `verbose` 控制 `-v` 重复次数，当前默认是 `0`。
- Windows metric 只在 Windows 加 `-x`。

敏感字段不要持久化。前端 `saveStoredConfig` 已排除：

- `key`
- `authPassword`

日志输出会对 secrets 做 redact。

## Linux 权限模型

Linux 不能像 Windows manifest 那样让整个 Wails app 自动 UAC。当前成熟方案是：

```text
普通用户 GUI
  -> pkexec SwiftN2N --helper edge
      -> root helper
          -> edge
              -> n2n drop privileges to nobody
```

### 启动 helper

入口：`main.go`

```text
SwiftN2N --helper edge
```

实现：`internal/edge/helper.go`

流程：

1. helper 从 stdin 读取 JSON config。
2. `ResolvePath` 找 edge。
3. `BuildCommand` 组装 args/env。
4. 启动 edge。
5. stdout/stderr 被扫描成 JSON line event 输出给 GUI。
6. helper 收到 SIGTERM 后终止 edge，3 秒后仍未退出则 SIGKILL。

### 停止 helper

之前出现过 `edge process did not exit after forced kill`。根因是 GUI 普通用户不能可靠 kill root helper / nobody edge。

现在 Stop 使用隐藏命令：

```text
SwiftN2N --helper stop <helper-pid>
```

流程：

1. GUI 的 `StopEdge()` 检测该 edge 是否通过 helper 启动。
2. 如果是普通用户 + helper 模式，则通过 `pkexec` 启动 stop helper。
3. stop helper 校验 `<pid>` 的 `/proc/<pid>/cmdline` 确实是 `SwiftN2N --helper edge`。
4. stop helper 向 root helper 发 SIGTERM。
5. root helper 负责停止 edge。
6. 6 秒未退出则 stop helper 发 SIGKILL。

真实烟测曾通过：

```text
helper <pid> stopped
```

并确认 helper/edge 无残留。

## Wails 事件协议

Go -> Frontend：

```text
edge:status  Status
edge:log     LogEvent
edge:exit    ExitEvent
```

Frontend -> Go：

```text
StartEdge(config)
StopEdge()
GetStatus()
ValidateConfig(config)
GetEdgeVersion(config)
CheckEnvironment(config)
```

绑定文件由 Wails 生成在：

```text
frontend/wailsjs/go/main/App.js
frontend/wailsjs/runtime/runtime.js
```

如果改了 Go 暴露方法，运行 `wails dev` 或 `wails build` 重新生成绑定。

## 状态解析

解析逻辑在 `internal/edge/parser.go`。

成功连接 supernode 的关键日志包括：

```text
Rx REGISTER_SUPER_ACK from ...
[OK] edge <<< ================ >>> supernode
Rx PONG from supernode ...
edge connected to supernode
registered with supernode
```

这些会映射为：

```text
state = connected
message = edge connected
```

失败/重试日志：

```text
supernode not responding
no response from supernode
authentication error
permission denied
address already in use
unable to open
failed to add supernode
```

注意：

- `attempts left N` 不代表已经失败。只要后续收到 `Rx REGISTER_SUPER_ACK` 或 `[OK]` 就是正常。
- `dropping Tx multicast` 通常只是 TAP 上的多播噪音，不是连接失败根因。
- 单个 edge 连接 supernode 成功只能证明控制通道正常；真正端到端 overlay 需要第二台 edge 加入后互 ping。

## 日志降噪

实现位置：`internal/edge/manager.go`

后端会保留第一条噪音日志，然后按时间窗口输出汇总：

```text
suppressed 33 repeated noisy log lines: dropping Tx multicast
```

当前降噪模式：

- `dropping Tx multicast`
- `DROP packet before first registration with supernode`
- `Rx TAP packet`
- broadcast `Tx PACKET ... FF:FF:FF:FF:FF:FF`

不会过滤关键事件：

- `Rx REGISTER_SUPER_ACK`
- `[OK] edge <<< >>> supernode`
- `Rx PONG from supernode`
- `authentication error`
- `supernode not responding`
- process exit/error

测试：`internal/edge/log_filter_test.go`

## 前端状态与本地存储

默认配置在：

```text
frontend/src/lib/edgeConfig.ts
```

当前重要默认值：

```ts
cipher: 'chacha20'
headerEncryption: false
mtu: 1290
verbose: 0
```

如果用户之前在 localStorage 保存过旧配置，旧值会继续覆盖默认值。例如旧配置里可能仍然有：

```text
headerEncryption = true
verbose = 1
```

排查 GUI 行为时不要只看默认代码，也要考虑本地存储。必要时可通过 UI 重新导入/编辑配置，或清空应用 localStorage。

## 已验证 supernode 结论

参考服务端部署手册：

```text
/path/to/private/N2N_SUPERNODE.md
```

该手册记录：

```text
supernode: supernode.example.net:9077
公网 IP: 203.0.113.10
主协议: 9077/udp
配置: /etc/n2n/supernode-9077.conf
内容:
  -p=9077
  -F=example-federation
  -t=5646
```

服务端没有要求客户端启用 `-H` header encryption。

### 手动 edge 测试结论

1. GUI 曾失败的参数：

```bash
edge -d swiftn2n-test1 \
  -a 10.77.0.201 \
  -c example-community \
  -l supernode.example.net:9077 \
  -A4 -H -M 1290 -f -v
```

结果：

```text
supernode not responding
```

2. 按手册最小参数，不加 `-H`：

```bash
sudo env N2N_KEY='<key>' edge \
  -d swiftn2n-test3 \
  -a 10.77.0.203 \
  -c example-community \
  -l supernode.example.net:9077 \
  -M 1290 -f -v
```

结果：

```text
Rx REGISTER_SUPER_ACK from C6:17:A2:42:27:B6 [203.0.113.10:9077]
[OK] edge <<< ================ >>> supernode
Rx PONG from supernode C6:17:A2:42:27:B6
```

3. `-A4` 不带 `-H`：

```bash
edge ... -A4 -M 1290 -f -v
```

结果：成功注册。

4. 默认 AES 带 `-H`：

```bash
edge ... -H -M 1290 -f -v
```

结果：失败，`supernode not responding`。

明确结论：

```text
失败主因是客户端启用了 -H header encryption，而当前 supernode/预期配置不匹配。
cipher 不是注册失败主因；ChaCha20 不带 -H 可以注册。
```

推荐 GUI 配置：

```text
supernode: supernode.example.net:9077
community: example-community
address: 10.77.0.13 或目标虚拟网段地址
community key: 强密钥，且不要等于 community
cipher: ChaCha20
Header encryption: off
MTU: 1290
verbose: 0
supernode only: off
```

如果日志显示：

```text
using null cipher.
WARNING: encryption is disabled in edge
```

说明 GUI 里 cipher 被选成 `none` 或配置导致 `-A1`。这能连上，但没有加密，不建议正式使用。

如果日志显示：

```text
WARNING: community and encryption key must differ
```

说明 community 和 key 一样。要更换 key。

## 手动诊断命令

查看 SwiftN2N/helper/edge 进程：

```bash
ps -eo pid,ppid,pgid,user,stat,args \
  | rg '(<repo>|\\./build/bin)/(build/bin/SwiftN2N|build/bin/bin/linux/amd64/edge|bin/linux/amd64/edge)|SwiftN2N --helper edge|./build/bin/SwiftN2N'
```

强制清理本项目残留进程：

```bash
sudo pkill -TERM -f '<repo>/build/bin/(SwiftN2N --helper edge|bin/linux/amd64/edge)' || true
sleep 2
sudo pkill -KILL -f '<repo>/build/bin/(SwiftN2N --helper edge|bin/linux/amd64/edge)' || true
```

查看 TAP：

```bash
ip -br addr show | rg 'edge|n2n|swift|10\.'
ip route | rg '10\.'
```

DNS：

```bash
getent ahostsv4 supernode.example.net
```

抓 UDP 9077：

```bash
sudo timeout 15s tcpdump -n -i any 'host 203.0.113.10 and udp port 9077'
```

成功时应看到 Out 和 In：

```text
Out IP <client>.<port> > 203.0.113.10.9077: UDP
In  IP 203.0.113.10.9077 > <client>.<port>: UDP
```

手动 edge 最小测试：

```bash
sudo env N2N_KEY='<strong-secret>' \
  <repo>/bin/linux/amd64/edge \
  -d swiftn2n-test \
  -a 10.77.0.203 \
  -c example-community \
  -l supernode.example.net:9077 \
  -M 1290 \
  -f \
  -v
```

成功标志：

```text
Rx REGISTER_SUPER_ACK
[OK] edge <<< ================ >>> supernode
Rx PONG from supernode
```

测试结束后确认无残留：

```bash
ps -eo pid,ppid,pgid,user,stat,args | rg 'swiftn2n-test|edge' | rg -v rg
ip -br addr show | rg 'swiftn2n-test|10\.77'
```

## 已修复过的问题

### GUI 一闪后空背景

原因之一是前端读取 `environment.missing.length`，但 Go nil slice 被 JSON 序列化成 `null`。

修复：

- Go `EnvironmentStatus.Missing` 初始化为 `[]string{}`。
- 前端 `ConfigPanel` 对 `environment.missing` 做 `Array.isArray` 容错。
- 添加 `internal/edge/types_test.go` 确保 `missing` 序列化为 `[]`。

### Vite/Wails 静态资源路径

生产包需要相对资源路径。

修复：

```ts
// frontend/vite.config.ts
export default defineConfig({
  base: './',
  plugins: [react()]
})
```

### Disconnect 后 edge 残留

原因：普通 GUI 不能可靠 kill root helper / nobody edge。

修复：

- 新增 `--helper stop <pid>`。
- `Manager.Stop()` 在 helper 模式下通过 pkexec stop helper 停止 root helper。
- helper 收到 SIGTERM 后负责停止 edge。

### `supernode not responding`

排查后确认：

- supernode UDP 9077 是通的。
- edge 本体可成功注册。
- GUI 失败主要来自 `Header encryption` 开启。

修复：

- 默认 `headerEncryption=false`。
- parser 识别 `Rx REGISTER_SUPER_ACK`、`[OK] edge`、`Rx PONG from supernode` 为 connected。

### 日志过吵

修复：

- 默认 `verbose=0`。
- 后端聚合高频噪音日志。

## 测试覆盖

当前 Go 测试：

```text
internal/edge/config_test.go      参数组装、secret 不进入 args、非法值拒绝
internal/edge/parser_test.go      日志状态分类
internal/edge/types_test.go       missing [] 序列化
internal/edge/log_filter_test.go  噪音日志识别
```

运行：

```bash
go test ./...
```

前端目前没有单测。至少要持续保证：

```bash
cd frontend && npm run build
cd frontend && npm audit --audit-level=high
```

## 后续路线建议

优先级从高到低：

1. 增加配置迁移版本号
   - 当前 localStorage 旧值可能覆盖新默认值。
   - 建议 storage key 或 profile schema 加 version。
   - 可对旧配置自动把 `headerEncryption=true` 改为 false，但要谨慎，避免破坏其他用户环境。

2. 增加 Connection Health 面板
   - 显示 last ACK 时间。
   - 显示 last PONG 时间。
   - 显示 edge PID/helper PID。
   - 显示 TAP name/IP/MAC。

3. 增加 profile presets
   - 示例 supernode preset：`cipher=chacha20`、`headerEncryption=false`、`mtu=1290`。
   - 自定义 profile 导入导出已存在，但可更成熟。

4. 增加环境检查
   - `pkexec` 是否存在。
   - Polkit agent/session bus 是否可用。
   - edge 可执行权限。
   - TAP 权限。
   - supernode DNS。
   - UDP 抓包需要 root，不建议自动做，但可以给诊断命令。

5. 前端改善
   - 用明确的状态文案区分：
     - process started
     - TAP ready
     - supernode registered
     - peer reachable
   - 当前 connected 代表 supernode 注册成功，不代表已有 peer 可通信。

6. Windows/macOS
   - Windows manifest 已请求管理员，但进程树终止和 TAP/WinTUN 驱动检查还需要实机验证。
   - macOS privileged helper 尚未实现，当前只能提示权限需求。

7. Release automation
   - 补 Linux desktop file/icon 安装脚本。
   - 补 Windows/macOS binary bundling。
   - 补 checksum manifest。

## 代码编辑注意事项

- 不要把 key/password 放进命令行参数。
- 不要把 key/password 写入 localStorage、日志或文件。
- 不要让普通 GUI 直接处理 sudo 密码；使用 Polkit/pkexec。
- 不要对 n2n 源码做本地 patch。
- 不要默认启用 `-H`，除非服务端/客户端全链路明确支持 header encryption。
- 如果添加新的日志状态分类，先补 `parser_test.go`。
- 如果添加新的噪音过滤模式，先确认不会吞关键错误，再补 `log_filter_test.go`。
- 如果修改 Go 暴露方法，需要重新生成 Wails bindings。

## 快速接手 Checklist

接手后先执行：

```bash
cd <repo>
go test ./...
cd frontend && npm run build && npm audit --audit-level=high
cd ..
make release-linux
```

确认当前进程：

```bash
ps -eo pid,ppid,pgid,user,stat,args \
  | rg '(<repo>|\\./build/bin)/(build/bin/SwiftN2N|build/bin/bin/linux/amd64/edge)|SwiftN2N --helper edge|./build/bin/SwiftN2N' \
  | rg -v rg
```

推荐连接参数：

```text
supernode: supernode.example.net:9077
community: example-community
address: 10.77.0.13
cipher: ChaCha20
Header encryption: off
MTU: 1290
verbose: 0
```

成功连接 supernode 时应看到：

```text
Rx REGISTER_SUPER_ACK
[OK] edge <<< ================ >>> supernode
Rx PONG from supernode
```
