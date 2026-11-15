# set-drive-icon

Windows 命令行工具：**用磁盘根目录的 `icon.ico` 自定义该盘在「此电脑」和资源管理器中的图标**。

把任意 `.ico` 改名为 `icon.ico` 放到盘根目录（如 `D:\icon.ico`），以管理员身份运行一次即可生效。

## 工作原理

1. 枚举本机所有盘符（`GetLogicalDrives`），筛选固定磁盘（`GetDriveTypeW` == `DRIVE_FIXED`），排除 U 盘、网络盘、光驱；
2. 检测 `<盘符>:\icon.ico` 是否存在；
3. 写入注册表，接管 Explorer 的磁盘图标：

```
HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\DriveIcons\<盘符>\DefaultIcon
    (默认值) = <盘符>:\icon.ico,0
```

4. 运行前自检管理员权限（`TokenElevation`），权限不足时直接退出，不会写一半。

## 使用方法

1. 准备 `icon.ico`，放到要改图标的磁盘根目录，例如 `D:\icon.ico`；
2. 右键程序 →「以管理员身份运行」（或在管理员终端中运行）；
3. 看到输出 `设置图标 D -> D:\icon.ico,0` 即成功，刷新资源管理器（`F5`）后生效。

## 构建

仅支持 Windows（注册表 + Win32 API），交叉编译：

```bash
GOOS=windows GOARCH=amd64 go build -o set-drive-icon.exe
```

依赖：Go 1.26+、`golang.org/x/sys`。

## 恢复默认图标

删除对应盘符的注册表项即可，例如恢复 D 盘：

```reg
[-HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\DriveIcons\D]
```

或在注册表编辑器中删除 `DriveIcons` 下对应盘符的子项，再刷新资源管理器。

## 注意事项

- 仅支持 Windows，必须管理员权限运行；
- `.ico` 建议包含多尺寸（16 / 32 / 48 / 256 px），生效的是第 0 帧；
- 写入的是 `HKLM`，对本机所有用户生效；
- 个别情况下刷新不生效时，重启资源管理器（任务管理器中结束 `explorer.exe` 后重新运行）。
