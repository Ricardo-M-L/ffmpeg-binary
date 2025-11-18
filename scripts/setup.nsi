; NSIS 安装脚本 - GoalfyMediaConverter (Windows 服务版本)
; 将程序安装为 Windows 服务,后台运行,无黑窗口

!include "MUI2.nsh"

; 应用信息
!define APP_NAME "GoalfyMediaConverter"
!define APP_VERSION "1.0.0"
!define APP_PUBLISHER "Goalfy"
!define APP_EXE "ffmpeg-binary.exe"
!define SERVICE_NAME "GoalfyMediaConverter"
!define INSTALL_DIR "$PROGRAMFILES64\${APP_NAME}"

; 安装包输出配置
Name "${APP_NAME}"
OutFile "../dist/windows/${APP_NAME}-Setup.exe"
InstallDir "${INSTALL_DIR}"
RequestExecutionLevel admin

; 界面配置
!define MUI_ABORTWARNING

; 安装页面
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

; 卸载页面
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

; 语言
!insertmacro MUI_LANGUAGE "SimpChinese"
!insertmacro MUI_LANGUAGE "English"

; 安装部分
Section "Install"
    SetOutPath "$INSTDIR"

    ; 复制主程序
    File "../build/windows/${APP_EXE}"

    ; 创建必要的目录
    CreateDirectory "$INSTDIR\bin"
    CreateDirectory "$INSTDIR\data"
    CreateDirectory "$INSTDIR\temp"
    CreateDirectory "$INSTDIR\output"
    CreateDirectory "$INSTDIR\logs"

    ; 提示 FFmpeg 安装
    DetailPrint "提示: 需要安装 FFmpeg"
    DetailPrint "请访问: https://www.gyan.dev/ffmpeg/builds/"
    DetailPrint "下载后将 ffmpeg.exe 复制到: $INSTDIR\bin\"

    ; 停止可能正在运行的旧服务
    DetailPrint "检查并停止旧服务..."
    ExecWait '"$INSTDIR\${APP_EXE}" stop-service' $0
    Sleep 1000

    ; 卸载旧服务(如果存在)
    DetailPrint "卸载旧服务(如果存在)..."
    ExecWait '"$INSTDIR\${APP_EXE}" uninstall-service' $0
    Sleep 1000

    ; 安装 Windows 服务
    DetailPrint "正在安装 Windows 服务..."
    ExecWait '"$INSTDIR\${APP_EXE}" install-service' $0
    ${If} $0 == 0
        DetailPrint "✓ Windows 服务安装成功"
    ${Else}
        DetailPrint "✗ Windows 服务安装失败,错误代码: $0"
        MessageBox MB_OK|MB_ICONEXCLAMATION "服务安装失败,错误代码: $0$\n$\n可能需要手动安装服务:$\n$INSTDIR\${APP_EXE} install-service"
    ${EndIf}

    Sleep 1000

    ; 启动 Windows 服务
    DetailPrint "正在启动服务..."
    ExecWait 'sc start ${SERVICE_NAME}' $0
    ${If} $0 == 0
        DetailPrint "✓ 服务启动成功"
    ${Else}
        DetailPrint "尝试备用方法启动服务..."
        ExecWait '"$INSTDIR\${APP_EXE}" start-service' $0
        ${If} $0 == 0
            DetailPrint "✓ 服务启动成功"
        ${Else}
            DetailPrint "✗ 服务启动失败,错误代码: $0"
            MessageBox MB_OK|MB_ICONEXCLAMATION "服务启动失败$\n$\n请在安装完成后手动启动:$\n1. 打开 服务 (services.msc)$\n2. 找到 'Goalfy Media Converter Service'$\n3. 右键点击,选择 '启动'"
        ${EndIf}
    ${EndIf}

    ; 创建开始菜单快捷方式(用于管理服务)
    CreateDirectory "$SMPROGRAMS\${APP_NAME}"

    ; 创建管理快捷方式
    FileOpen $0 "$SMPROGRAMS\${APP_NAME}\启动服务.bat" w
    FileWrite $0 '@echo off$\r$\n'
    FileWrite $0 'sc start ${SERVICE_NAME}$\r$\n'
    FileWrite $0 'echo 服务已启动$\r$\n'
    FileWrite $0 'pause$\r$\n'
    FileClose $0

    FileOpen $0 "$SMPROGRAMS\${APP_NAME}\停止服务.bat" w
    FileWrite $0 '@echo off$\r$\n'
    FileWrite $0 'sc stop ${SERVICE_NAME}$\r$\n'
    FileWrite $0 'echo 服务已停止$\r$\n'
    FileWrite $0 'pause$\r$\n'
    FileClose $0

    FileOpen $0 "$SMPROGRAMS\${APP_NAME}\查看服务状态.bat" w
    FileWrite $0 '@echo off$\r$\n'
    FileWrite $0 'sc query ${SERVICE_NAME}$\r$\n'
    FileWrite $0 'pause$\r$\n'
    FileClose $0

    FileOpen $0 "$SMPROGRAMS\${APP_NAME}\打开服务管理器.bat" w
    FileWrite $0 '@echo off$\r$\n'
    FileWrite $0 'services.msc$\r$\n'
    FileClose $0

    FileOpen $0 "$SMPROGRAMS\${APP_NAME}\打开Web界面.url" w
    FileWrite $0 '[InternetShortcut]$\r$\n'
    FileWrite $0 'URL=http://127.0.0.1:28888$\r$\n'
    FileClose $0

    CreateShortcut "$SMPROGRAMS\${APP_NAME}\卸载.lnk" "$INSTDIR\Uninstall.exe"

    ; 写入卸载信息
    WriteUninstaller "$INSTDIR\Uninstall.exe"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" \
                     "DisplayName" "${APP_NAME}"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" \
                     "UninstallString" "$INSTDIR\Uninstall.exe"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" \
                     "Publisher" "${APP_PUBLISHER}"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" \
                     "DisplayVersion" "${APP_VERSION}"

    DetailPrint "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    DetailPrint "✓ 安装完成!"
    DetailPrint ""
    DetailPrint "服务已作为 Windows 服务运行 (后台,无窗口)"
    DetailPrint "服务将在开机时自动启动"
    DetailPrint ""
    DetailPrint "Web 界面: http://127.0.0.1:28888"
    DetailPrint ""
    DetailPrint "服务管理:"
    DetailPrint "  - 使用开始菜单中的快捷方式"
    DetailPrint "  - 或打开 '服务' (services.msc)"
    DetailPrint "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    ; 显示完成消息
    MessageBox MB_OK|MB_ICONINFORMATION "安装完成!$\n$\nGoalfyMediaConverter 已作为 Windows 服务安装$\n$\n✓ 服务已启动,后台运行 (无窗口)$\n✓ 开机自动启动已启用$\n$\nWeb 界面: http://127.0.0.1:28888$\n$\n可在开始菜单中找到服务管理工具"
SectionEnd

; 卸载部分
Section "Uninstall"
    ; 停止并卸载服务
    DetailPrint "正在停止服务..."
    ExecWait 'sc stop ${SERVICE_NAME}' $0
    Sleep 2000

    DetailPrint "正在卸载服务..."
    ExecWait '"$INSTDIR\${APP_EXE}" uninstall-service' $0
    Sleep 1000

    ; 删除文件
    Delete "$INSTDIR\${APP_EXE}"
    Delete "$INSTDIR\Uninstall.exe"
    RMDir /r "$INSTDIR\bin"
    RMDir /r "$INSTDIR\data"
    RMDir /r "$INSTDIR\temp"
    RMDir /r "$INSTDIR\output"
    RMDir /r "$INSTDIR\logs"
    RMDir "$INSTDIR"

    ; 删除开始菜单快捷方式
    Delete "$SMPROGRAMS\${APP_NAME}\启动服务.bat"
    Delete "$SMPROGRAMS\${APP_NAME}\停止服务.bat"
    Delete "$SMPROGRAMS\${APP_NAME}\查看服务状态.bat"
    Delete "$SMPROGRAMS\${APP_NAME}\打开服务管理器.bat"
    Delete "$SMPROGRAMS\${APP_NAME}\打开Web界面.url"
    Delete "$SMPROGRAMS\${APP_NAME}\卸载.lnk"
    RMDir "$SMPROGRAMS\${APP_NAME}"

    ; 删除注册表项
    DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}"

    MessageBox MB_OK "GoalfyMediaConverter 已成功卸载"
SectionEnd