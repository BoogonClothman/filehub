# FileHub

局域网文件共享中心 — 以本机为存储，其他设备通过浏览器访问、上传、下载、管理文件。

单文件 Go 可执行文件，零运行时依赖。

## 快速开始

从 [Releases](../../releases) 下载 `filehub.exe`，双击运行。

浏览器打开 `http://localhost:5000`，局域网设备访问 `http://<你的IP>:5000`。

## 命令行参数

```
filehub.exe                  # 默认端口 5000，数据目录 ./data
filehub.exe -port 8080       # 自定义端口
filehub.exe -data D:\myfiles # 自定义数据目录
```

## 功能

- 📁 目录浏览（面包屑导航）
- 📤 文件上传（按钮 + 拖拽）
- 📥 文件下载
- 🖼️ 图片预览
- ✏️ 文件/目录重命名
- 🗑️ 删除（支持递归删除）
- 📂 新建目录
- 🖱️ 右键菜单

## 从源码编译

需要 Go 1.24+。

```bash
git clone https://github.com/BoogonClothman/filehub.git
cd filehub
go build -o filehub.exe .
```

## 注册为 Windows 服务

以**管理员身份**运行：

```bash
scripts\install-service.bat    # 注册并启动服务，开机自启
scripts\uninstall-service.bat  # 停止并卸载服务
```

注册后服务随系统启动，无需手动打开控制台窗口。

## 技术栈

- Go 标准库 (`net/http`, `embed`)
- 前端：原生 HTML/CSS/JS，SPA，零前端依赖

## 配置

首次运行自动生成 `config.json`：

```json
{
  "port": 5000,
  "dataRoot": "./data",
  "maxUploadMB": 100
}
```

## 许可证

MIT
