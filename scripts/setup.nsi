; NSIS 安装脚本 - GoalfyMediaConverter (简化版)
; 仅包含核心功能,不依赖额外插件

!include "MUI2.nsh"

; 应用信息
!define APP_NAME "GoalfyMediaConverter"
!define APP_VERSION "1.0.0"
!define APP_PUBLISHER "Goalfy"
!define APP_EXE "ffmpeg-binary.exe"
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

    ; 下载 FFmpeg (简化版 - 提示用户手动下载)
    DetailPrint "注意: 需要安装 FFmpeg"
    DetailPrint "请访问: https://www.gyan.dev/ffmpeg/builds/"
    DetailPrint "下载后将 ffmpeg.exe 复制到: $INSTDIR\bin\"

    ; 创建开始菜单快捷方式
    CreateDirectory "$SMPROGRAMS\${APP_NAME}"
    CreateShortcut "$SMPROGRAMS\${APP_NAME}\${APP_NAME}.lnk" "$INSTDIR\${APP_EXE}"
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

    ; 开机自启动(可选)
    MessageBox MB_YESNO "是否设置开机自动启动?" IDYES autostart IDNO skipAutostart
    autostart:
        WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Run" \
                         "${APP_NAME}" "$INSTDIR\${APP_EXE}"
    skipAutostart:

    ; 启动服务
    DetailPrint "正在启动服务..."
    Exec '"$INSTDIR\${APP_EXE}"'

    DetailPrint "安装完成!"
    DetailPrint "服务地址: http://127.0.0.1:28888"
SectionEnd

; 卸载部分
Section "Uninstall"
    ; 停止服务
    ExecWait 'taskkill /F /IM ${APP_EXE}'

    ; 删除文件
    Delete "$INSTDIR\${APP_EXE}"
    Delete "$INSTDIR\Uninstall.exe"
    RMDir /r "$INSTDIR\bin"
    RMDir /r "$INSTDIR\data"
    RMDir /r "$INSTDIR\temp"
    RMDir /r "$INSTDIR\output"
    RMDir "$INSTDIR"

    ; 删除开始菜单快捷方式
    Delete "$SMPROGRAMS\${APP_NAME}\${APP_NAME}.lnk"
    Delete "$SMPROGRAMS\${APP_NAME}\卸载.lnk"
    RMDir "$SMPROGRAMS\${APP_NAME}"

    ; 删除注册表项
    DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}"
    DeleteRegValue HKCU "Software\Microsoft\Windows\CurrentVersion\Run" "${APP_NAME}"
SectionEnd