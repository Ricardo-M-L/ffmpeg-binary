; NSIS 安装脚本 - GoalfyMediaConverter (Windows 服务版本)
; 将程序安装为 Windows 服务,后台运行,无黑窗口

!include "MUI2.nsh"
!include "LogicLib.nsh"

; 应用信息
!define APP_NAME "GoalfyMediaConverter"
!define APP_VERSION "1.0.0"
!define APP_PUBLISHER "Goalfy"
!define APP_EXE "ffmpeg-binary.exe"
!define SERVICE_NAME "GoalfyMediaConverter"
!define INSTALL_DIR "$PROGRAMFILES64\${APP_NAME}"
!define FFMPEG_URL "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip"

; 安装包输出配置
Name "${APP_NAME}"
OutFile "../dist/windows/${APP_NAME}-Setup.exe"
InstallDir "${INSTALL_DIR}"
RequestExecutionLevel admin

; 界面配置
!define MUI_ABORTWARNING
!define MUI_ICON "${NSISDIR}\Contrib\Graphics\Icons\modern-install.ico"
!define MUI_UNICON "${NSISDIR}\Contrib\Graphics\Icons\modern-uninstall.ico"

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

; ========== 辅助函数 ==========
; StrStr 函数 - 在字符串中查找子字符串
; 用法: Push "haystack" ; Push "needle" ; Call StrStr ; Pop $result
Function StrStr
    Exch $R1 ; 要查找的字符串 (needle)
    Exch
    Exch $R2 ; 被搜索的字符串 (haystack)
    Push $R3
    Push $R4
    Push $R5

    StrLen $R3 $R1
    StrCpy $R4 0

    ; 如果 needle 为空,返回 haystack
    StrCmp $R1 "" +1 +4
    StrCpy $R1 $R2
    Goto done

    loop:
        StrCpy $R5 $R2 $R3 $R4
        StrCmp $R5 $R1 done
        StrCmp $R5 "" failed
        IntOp $R4 $R4 + 1
        Goto loop

    failed:
        StrCpy $R1 ""
        Goto done

    done:
        Pop $R5
        Pop $R4
        Pop $R3
        Pop $R2
        Exch $R1
FunctionEnd

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

    ; ========== FFmpeg 检测 ==========
    DetailPrint "检测 FFmpeg..."

    ; 检查 bin 目录是否已有 ffmpeg.exe
    IfFileExists "$INSTDIR\bin\ffmpeg.exe" ffmpeg_exists check_system_ffmpeg

    check_system_ffmpeg:
        ; 检查系统 PATH 中是否有 ffmpeg
        nsExec::ExecToStack 'where ffmpeg.exe'
        Pop $0
        ${If} $0 == 0
            DetailPrint "系统已安装 FFmpeg"
            Goto ffmpeg_ready
        ${EndIf}

        ; FFmpeg 不存在,自动下载安装
        DetailPrint "未检测到 FFmpeg,准备自动下载..."

    download_ffmpeg:
        DetailPrint "开始下载 FFmpeg (约 70MB)..."

        ; 使用 inetc 插件下载(带进度条)
        inetc::get /caption "下载 FFmpeg" /canceltext "取消" "${FFMPEG_URL}" "$TEMP\ffmpeg.zip" /end
        Pop $0

        ${If} $0 == "OK"
            DetailPrint "下载完成,正在解压..."

            ; 创建解压脚本
            FileOpen $1 "$TEMP\extract_ffmpeg.ps1" w
            FileWrite $1 'chcp 65001 > $$null$\r$\n'
            FileWrite $1 'try {$\r$\n'
            FileWrite $1 '    Expand-Archive -Path "$$env:TEMP\ffmpeg.zip" -DestinationPath "$$env:TEMP\ffmpeg_tmp" -Force$\r$\n'
            FileWrite $1 '    $$exe = Get-ChildItem -Path "$$env:TEMP\ffmpeg_tmp" -Filter "ffmpeg.exe" -Recurse | Select -First 1$\r$\n'
            FileWrite $1 '    if ($$exe) {$\r$\n'
            FileWrite $1 '        Copy-Item $$exe.FullName "$INSTDIR\bin\ffmpeg.exe" -Force$\r$\n'
            FileWrite $1 '        exit 0$\r$\n'
            FileWrite $1 '    }$\r$\n'
            FileWrite $1 '    exit 1$\r$\n'
            FileWrite $1 '} catch { exit 1 }$\r$\n'
            FileClose $1

            nsExec::ExecToStack 'powershell -NoProfile -ExecutionPolicy Bypass -File "$TEMP\extract_ffmpeg.ps1"'
            Pop $0

            Delete "$TEMP\extract_ffmpeg.ps1"
            Delete "$TEMP\ffmpeg.zip"
            RMDir /r "$TEMP\ffmpeg_tmp"

            ${If} $0 == 0
                DetailPrint "FFmpeg 安装成功"
            ${Else}
                MessageBox MB_OK|MB_ICONEXCLAMATION "FFmpeg 解压失败!$\n$\n请稍后手动下载安装。"
            ${EndIf}
        ${Else}
            DetailPrint "下载失败或已取消"
            MessageBox MB_OK|MB_ICONINFORMATION "已取消 FFmpeg 下载。$\n$\n您可以稍后手动下载安装。"
        ${EndIf}
        Goto ffmpeg_ready

    ffmpeg_exists:
        DetailPrint "检测到已安装的 FFmpeg"

    ffmpeg_ready:
    ; ========== FFmpeg 检测完成 ==========

    ; ========== Windows 服务检测与安装 ==========
    DetailPrint "检测 Windows 服务..."

    ; 检查服务是否存在并且正在运行
    nsExec::ExecToStack 'sc query ${SERVICE_NAME}'
    Pop $0 ; 返回值
    Pop $1 ; 输出内容

    ${If} $0 == 0
        ; 服务存在,检查是否运行中
        Push $1
        Push "RUNNING"
        Call StrStr
        Pop $2

        ${If} $2 != ""
            ; 服务已安装且正在运行
            DetailPrint "检测到服务已安装且正在运行,跳过安装"
            Goto service_ready
        ${Else}
            ; 服务存在但未运行,先停止并卸载旧服务
            DetailPrint "检测到已停止的服务,准备重新安装..."
            nsExec::ExecToLog 'sc stop ${SERVICE_NAME}'
            Sleep 1000
            nsExec::ExecToLog '"$INSTDIR\${APP_EXE}" uninstall-service'
            Sleep 1000
        ${EndIf}
    ${Else}
        ; 服务不存在
        DetailPrint "未检测到服务,准备安装..."
    ${EndIf}

    ; 安装 Windows 服务
    DetailPrint "安装 Windows 服务..."
    nsExec::ExecToStack '"$INSTDIR\${APP_EXE}" install-service'
    Pop $0 ; 返回值
    Pop $1 ; 输出内容

    ${If} $0 == 0
        DetailPrint "服务安装成功"
    ${Else}
        DetailPrint "服务安装失败 (错误代码: $0)"
        MessageBox MB_OK|MB_ICONEXCLAMATION "服务安装失败!$\n$\n请确保以管理员权限运行安装程序。"
        Abort
    ${EndIf}

    Sleep 1000

    ; 启动 Windows 服务
    DetailPrint "启动服务..."
    nsExec::ExecToStack 'sc start ${SERVICE_NAME}'
    Pop $0 ; 返回值

    ${If} $0 == 0
        DetailPrint "服务启动成功"
    ${Else}
        DetailPrint "尝试备用方法启动..."
        nsExec::ExecToStack '"$INSTDIR\${APP_EXE}" start-service'
        Pop $0
        ${If} $0 != 0
            MessageBox MB_OK|MB_ICONEXCLAMATION "服务启动失败!$\n$\n请手动启动服务:$\n1. 按 Win+R 输入 services.msc$\n2. 找到 Goalfy Media Converter Service$\n3. 右键点击选择启动"
        ${EndIf}
    ${EndIf}

service_ready:
    ; ========== 服务检测完成 ==========

    ; 创建开始菜单快捷方式
    CreateDirectory "$SMPROGRAMS\${APP_NAME}"

    ; 创建管理脚本
    FileOpen $0 "$SMPROGRAMS\${APP_NAME}\启动服务.bat" w
    FileWrite $0 '@echo off$\r$\n'
    FileWrite $0 'chcp 65001 >nul$\r$\n'
    FileWrite $0 'echo 正在启动服务...$\r$\n'
    FileWrite $0 'sc start ${SERVICE_NAME}$\r$\n'
    FileWrite $0 'if %errorlevel% == 0 ($\r$\n'
    FileWrite $0 '    echo 服务启动成功!$\r$\n'
    FileWrite $0 ') else ($\r$\n'
    FileWrite $0 '    echo 服务启动失败,错误代码: %errorlevel%$\r$\n'
    FileWrite $0 ')$\r$\n'
    FileWrite $0 'pause$\r$\n'
    FileClose $0

    FileOpen $0 "$SMPROGRAMS\${APP_NAME}\停止服务.bat" w
    FileWrite $0 '@echo off$\r$\n'
    FileWrite $0 'chcp 65001 >nul$\r$\n'
    FileWrite $0 'echo 正在停止服务...$\r$\n'
    FileWrite $0 'sc stop ${SERVICE_NAME}$\r$\n'
    FileWrite $0 'if %errorlevel% == 0 ($\r$\n'
    FileWrite $0 '    echo 服务已停止!$\r$\n'
    FileWrite $0 ') else ($\r$\n'
    FileWrite $0 '    echo 服务停止失败,错误代码: %errorlevel%$\r$\n'
    FileWrite $0 ')$\r$\n'
    FileWrite $0 'pause$\r$\n'
    FileClose $0

    FileOpen $0 "$SMPROGRAMS\${APP_NAME}\查看服务状态.bat" w
    FileWrite $0 '@echo off$\r$\n'
    FileWrite $0 'chcp 65001 >nul$\r$\n'
    FileWrite $0 'echo 查询服务状态:$\r$\n'
    FileWrite $0 'sc query ${SERVICE_NAME}$\r$\n'
    FileWrite $0 'pause$\r$\n'
    FileClose $0

    FileOpen $0 "$SMPROGRAMS\${APP_NAME}\打开服务管理器.bat" w
    FileWrite $0 '@echo off$\r$\n'
    FileWrite $0 'start services.msc$\r$\n'
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

    DetailPrint "安装完成!"

    ; 显示完成消息
    MessageBox MB_OK|MB_ICONINFORMATION "安装完成!$\n$\nGoalfyMediaConverter 已安装为 Windows 服务$\n$\n服务已启动,后台运行$\n开机自动启动已启用$\n$\nWeb 界面: http://127.0.0.1:28888"
SectionEnd

; 卸载部分
Section "Uninstall"
    ; 停止服务 (静默执行)
    DetailPrint "停止服务..."
    nsExec::ExecToLog 'sc stop ${SERVICE_NAME}'
    Sleep 2000

    ; 卸载服务 (静默执行)
    DetailPrint "卸载服务..."
    nsExec::ExecToLog '"$INSTDIR\${APP_EXE}" uninstall-service'
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
