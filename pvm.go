package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	phpBaseURL = "https://windows.php.net/downloads/releases/archives/"
	phpNewURL  = "https://windows.php.net/downloads/releases/"
)

func init() {
	// 检查临时目录
	tmpDir := filepath.Join(os.TempDir(), "pvm")
	os.MkdirAll(tmpDir, 0755)

	// 显示欢迎信息
	fmt.Println("PVM - PHP 版本管理器")
	fmt.Println("===================")

	// 获取 PHP 安装目录
	phpHome, err := getPHPHome()
	if err == nil {
		fmt.Printf("PHP 目录: %s\n", phpHome)
	}
}

func main() {
	// 获取命令行参数
	args := os.Args[1:]

	// 如果没有参数，执行默认的更新操作
	if len(args) == 0 {
		fmt.Println("使用说明：")
		fmt.Println("  pvm list - 列出所有已安装的版本")
		fmt.Println("  pvm install <版本> - 安装指定版本")
		fmt.Println("  pvm use <版本> - 切换到指定版本")
		fmt.Println("  pvm check - 查看PHP官网上可用的版本")
		fmt.Println("  pvm - 显示帮助信息")
		return
	}

	// 处理各种命令
	switch args[0] {
	case "list":
		listVersions()
	case "install":
		if len(args) < 2 {
			fmt.Println("请指定要安装的版本，例如：pvm install 7.4")
			fmt.Println("您可以使用 pvm check 命令查看可用的版本")
			return
		}
		installVersion(args[1])
	case "use":
		if len(args) < 2 {
			fmt.Println("请指定要使用的版本，例如：pvm use 7.4")
			fmt.Println("您可以使用 pvm list 命令查看已安装的版本")
			return
		}
		useVersion(args[1])
	case "check":
		checkAvailableVersions()
	default:
		fmt.Println("未知命令。可用命令：")
		fmt.Println("  pvm list - 列出所有已安装的版本")
		fmt.Println("  pvm install <版本> - 安装指定版本")
		fmt.Println("  pvm use <版本> - 切换到指定版本")
		fmt.Println("  pvm check - 查看PHP官网上可用的版本")
	}
}

func updateProgram() {
	// 获取当前目录
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("获取当前目录失败: %v\n", err)
		return
	}

	// 程序路径
	pvmPath := filepath.Join(currentDir, "pvm.exe")

	// 检查程序是否在运行
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq pvm.exe")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		// 如果程序在运行，则关闭它
		killCmd := exec.Command("taskkill", "/F", "/IM", "pvm.exe")
		killCmd.Run()
		time.Sleep(time.Second) // 等待程序完全关闭
	}

	// 删除旧程序
	os.Remove(pvmPath)

	// 复制新程序
	copyCmd := exec.Command("copy", "/Y", "pvm_new.exe", "pvm.exe")
	copyCmd.Run()

	// 启动新程序
	startCmd := exec.Command("start", "pvm.exe")
	startCmd.Run()

	// 等待5秒
	time.Sleep(5 * time.Second)
}

func getPHPHome() (string, error) {
	// 使用固定的 PHP 安装目录
	phpsDir := "D:\\app\\pvm\\phps"
	phpHomeDir := "D:\\app\\pvm\\php_home"

	// 创建 phps 目录
	fmt.Printf("使用 PHP 目录: %s\n", phpsDir)
	if err := os.MkdirAll(phpsDir, 0755); err != nil {
		return "", fmt.Errorf("创建 phps 目录失败: %v", err)
	}

	// 创建 php_home 目录
	if err := os.MkdirAll(phpHomeDir, 0755); err != nil {
		return "", fmt.Errorf("创建 php_home 目录失败: %v", err)
	}

	return phpsDir, nil
}

func listVersions() {
	// 获取 PHP 安装目录
	phpHome, err := getPHPHome()
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	// 版本映射文件
	versionFile := filepath.Join(phpHome, "versions.json")
	var versionMap map[string]string

	// 读取版本映射
	data, err := os.ReadFile(versionFile)
	if err == nil {
		// 文件存在，解析 JSON
		if err := json.Unmarshal(data, &versionMap); err != nil {
			versionMap = make(map[string]string)
		}
	} else {
		// 文件不存在，创建空映射
		versionMap = make(map[string]string)
	}

	// 列出所有 PHP 安装目录
	dirs, err := filepath.Glob(filepath.Join(phpHome, "php-*"))
	if err != nil {
		fmt.Printf("列出目录失败: %v\n", err)
		return
	}

	if len(dirs) == 0 {
		fmt.Println("没有找到任何版本")
		return
	}

	// 获取当前使用的版本
	currentVersion := getCurrentVersion()

	fmt.Println("已安装的 PHP 版本:")
	for shortVersion, dirName := range versionMap {
		fullPath := filepath.Join(phpHome, dirName)
		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		isCurrent := ""
		if fullPath == currentVersion {
			isCurrent = " (当前使用)"
		}

		fmt.Printf("  %s => %s%s\n", shortVersion, dirName, isCurrent)
		fmt.Printf("      大小: %d 字节, 修改时间: %s\n", info.Size(), info.ModTime().Format("2006-01-02 15:04:05"))
	}

	// 列出未映射的目录
	unmapped := make([]string, 0)
	for _, dir := range dirs {
		dirName := filepath.Base(dir)
		found := false
		for _, mappedDir := range versionMap {
			if mappedDir == dirName {
				found = true
				break
			}
		}
		if !found {
			unmapped = append(unmapped, dirName)
		}
	}

	if len(unmapped) > 0 {
		fmt.Println("\n未映射的 PHP 安装目录:")
		for _, dirName := range unmapped {
			fullPath := filepath.Join(phpHome, dirName)
			isCurrent := ""
			if fullPath == currentVersion {
				isCurrent = " (当前使用)"
			}
			fmt.Printf("  %s%s\n", dirName, isCurrent)
		}
	}
}

func getLatestVersion(majorMinor string) (string, error) {
	// 获取目录列表
	resp, err := http.Get(phpBaseURL)
	if err != nil {
		return "", fmt.Errorf("获取版本列表失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取版本列表失败: %v", err)
	}

	// 使用多个正则表达式匹配不同格式的版本号
	patterns := []string{
		fmt.Sprintf(`php-%s\.\d+-Win32-vc15-x64\.zip`, majorMinor),
		fmt.Sprintf(`php-%s\.\d+-Win32-vc16-x64\.zip`, majorMinor),
		fmt.Sprintf(`php-%s\.\d+-Win32-vs16-x64\.zip`, majorMinor),
		fmt.Sprintf(`php-%s\.\d+-nts-Win32-vc15-x64\.zip`, majorMinor),
		fmt.Sprintf(`php-%s\.\d+-nts-Win32-vc16-x64\.zip`, majorMinor),
		fmt.Sprintf(`php-%s\.\d+-nts-Win32-vs16-x64\.zip`, majorMinor),
		fmt.Sprintf(`php-%s\.\d+-nts-Win32-vs17-x64\.zip`, majorMinor),
		fmt.Sprintf(`php-%s\.\d+-Win32-vs17-x64\.zip`, majorMinor),
		fmt.Sprintf(`php-%s\.\d+-Win32-vs16-x64\.zip`, majorMinor),
		fmt.Sprintf(`php-%s\.\d+-Win32-vs17-x64\.zip`, majorMinor),
		fmt.Sprintf(`php-%s-Win32-vc15-x64\.zip`, majorMinor), // 添加没有微版本的模式
		fmt.Sprintf(`php-%s-Win32-vc16-x64\.zip`, majorMinor),
		fmt.Sprintf(`php-%s-Win32-vs16-x64\.zip`, majorMinor),
		fmt.Sprintf(`php-%s-Win32-vs17-x64\.zip`, majorMinor),
		fmt.Sprintf(`php-%s-nts-Win32-vc15-x64\.zip`, majorMinor),
		fmt.Sprintf(`php-%s-nts-Win32-vc16-x64\.zip`, majorMinor),
		fmt.Sprintf(`php-%s-nts-Win32-vs16-x64\.zip`, majorMinor),
		fmt.Sprintf(`php-%s-nts-Win32-vs17-x64\.zip`, majorMinor),
	}

	var matches []string
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		found := re.FindAllString(string(body), -1)
		matches = append(matches, found...)
	}

	if len(matches) == 0 {
		// 尝试搜索当前PHP版本信息
		tryOtherVersions := fmt.Sprintf(`查找版本 %s 失败。
尝试以下操作:
1. 检查版本号是否正确，例如使用 "8.2" 而不是 "8.4"
2. 如果是最新版本，检查是否需要完整版本号，例如 "8.2.0" 而不是 "8.2"
3. 使用 "pvm list" 命令查看已安装的版本
4. 访问 https://windows.php.net/download 查看可用的版本`, majorMinor)
		return "", fmt.Errorf(tryOtherVersions)
	}

	// 提取版本号（使用第一个匹配项）
	version := matches[0]
	version = strings.TrimPrefix(version, "php-")
	version = strings.TrimSuffix(version, "-Win32-vc15-x64.zip")
	version = strings.TrimSuffix(version, "-Win32-vc16-x64.zip")
	version = strings.TrimSuffix(version, "-Win32-vs16-x64.zip")
	version = strings.TrimSuffix(version, "-Win32-vs17-x64.zip")
	version = strings.TrimSuffix(version, "-nts-Win32-vc15-x64.zip")
	version = strings.TrimSuffix(version, "-nts-Win32-vc16-x64.zip")
	version = strings.TrimSuffix(version, "-nts-Win32-vs16-x64.zip")
	version = strings.TrimSuffix(version, "-nts-Win32-vs17-x64.zip")

	return version, nil
}

func downloadPHP(version string) (string, string, error) {
	// 获取完整版本号
	fullVersion, err := getLatestVersion(version)
	if err != nil {
		return "", "", err
	}

	fmt.Printf("找到最新版本: %s\n", fullVersion)

	// 尝试不同的下载 URL 格式
	urls := []string{
		// 从发布目录下载
		fmt.Sprintf("%sphp-%s-Win32-vc15-x64.zip", phpNewURL, fullVersion),
		fmt.Sprintf("%sphp-%s-Win32-vc16-x64.zip", phpNewURL, fullVersion),
		fmt.Sprintf("%sphp-%s-Win32-vs16-x64.zip", phpNewURL, fullVersion),
		fmt.Sprintf("%sphp-%s-Win32-vs17-x64.zip", phpNewURL, fullVersion),
		fmt.Sprintf("%sphp-%s-nts-Win32-vc15-x64.zip", phpNewURL, fullVersion),
		fmt.Sprintf("%sphp-%s-nts-Win32-vc16-x64.zip", phpNewURL, fullVersion),
		fmt.Sprintf("%sphp-%s-nts-Win32-vs16-x64.zip", phpNewURL, fullVersion),
		fmt.Sprintf("%sphp-%s-nts-Win32-vs17-x64.zip", phpNewURL, fullVersion),

		// 从归档目录下载
		fmt.Sprintf("%sphp-%s-Win32-vc15-x64.zip", phpBaseURL, fullVersion),
		fmt.Sprintf("%sphp-%s-Win32-vc16-x64.zip", phpBaseURL, fullVersion),
		fmt.Sprintf("%sphp-%s-Win32-vs16-x64.zip", phpBaseURL, fullVersion),
		fmt.Sprintf("%sphp-%s-nts-Win32-vc15-x64.zip", phpBaseURL, fullVersion),
		fmt.Sprintf("%sphp-%s-nts-Win32-vc16-x64.zip", phpBaseURL, fullVersion),
		fmt.Sprintf("%sphp-%s-nts-Win32-vs16-x64.zip", phpBaseURL, fullVersion),
		fmt.Sprintf("%sphp-%s-nts-Win32-vs17-x64.zip", phpBaseURL, fullVersion),
		fmt.Sprintf("%sphp-%s-Win32-vs17-x64.zip", phpBaseURL, fullVersion),
	}

	var resp *http.Response
	var downloadErr error
	var successfulURL string

	fmt.Println("尝试下载以下 URL:")
	for _, url := range urls {
		fmt.Printf("  %s\n", url)
		resp, downloadErr = http.Get(url)
		if downloadErr == nil && resp.StatusCode == http.StatusOK {
			fmt.Printf("成功下载: %s\n", url)
			successfulURL = url
			break
		}
		if resp != nil {
			fmt.Printf("  状态码: %d\n", resp.StatusCode)
			resp.Body.Close()
		} else {
			fmt.Printf("  错误: %v\n", downloadErr)
		}
	}

	if successfulURL == "" {
		return "", "", fmt.Errorf("下载失败，未找到可用的下载链接。\n请访问 https://windows.php.net/download 查看可用的PHP版本。")
	}
	defer resp.Body.Close()

	// 获取完整的目录名称
	parts := strings.Split(successfulURL, "/")
	filename := parts[len(parts)-1]
	dirName := strings.TrimSuffix(filename, ".zip")

	fmt.Printf("将使用目录名: %s\n", dirName)

	// 创建下载目录
	downloadDir := filepath.Join(os.TempDir(), "pvm")
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return "", "", fmt.Errorf("创建下载目录失败: %v", err)
	}

	outputFile := filepath.Join(downloadDir, filename)
	fmt.Printf("保存到: %s\n", outputFile)

	// 保存文件
	out, err := os.Create(outputFile)
	if err != nil {
		return "", "", fmt.Errorf("创建文件失败: %v", err)
	}
	defer out.Close()

	n, err := io.Copy(out, resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("保存文件失败: %v", err)
	}
	fmt.Printf("下载完成，文件大小: %d 字节\n", n)

	// 验证文件是否存在
	_, err = os.Stat(outputFile)
	if err != nil {
		return "", "", fmt.Errorf("下载后文件不存在: %v", err)
	}

	return outputFile, dirName, nil
}

func updatePATH(phpHome string) error {
	// 获取当前 PATH 环境变量（获取系统级别 PATH）
	cmd := exec.Command("reg", "query", "HKLM\\SYSTEM\\CurrentControlSet\\Control\\Session Manager\\Environment", "/v", "PATH")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("获取系统 PATH 环境变量失败: %v", err)
	}

	// 解析注册表输出获取系统 PATH
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")
	var path string
	for _, line := range lines {
		if strings.Contains(line, "PATH") && strings.Contains(line, "REG_") {
			parts := strings.SplitN(line, "REG_", 2)
			if len(parts) > 1 {
				regValueParts := strings.SplitN(parts[1], "    ", 2)
				if len(regValueParts) > 1 {
					path = strings.TrimSpace(regValueParts[1])
					break
				}
			}
		}
	}

	if path == "" {
		path = os.Getenv("PATH") // 如果无法获取系统 PATH，则使用当前 PATH
		fmt.Println("无法从注册表获取系统 PATH，使用当前会话的 PATH 作为备用")
	}

	// 符号链接目标目录
	phpHomeDir := "D:\\app\\pvm\\php_home"

	// 检查 PHP 目录是否已经在 PATH 中
	phpInPath := false
	paths := strings.Split(path, ";")
	for _, p := range paths {
		if strings.EqualFold(p, phpHomeDir) {
			phpInPath = true
			fmt.Printf("PHP 目录已在系统 PATH 中: %s\n", phpHomeDir)
			break
		}
	}

	if !phpInPath {
		fmt.Printf("更新系统 PATH 环境变量...\n")
		fmt.Printf("将 %s 添加到系统 PATH\n", phpHomeDir)

		// 构建新的 PATH（添加到开头）
		newPath := phpHomeDir
		if path != "" {
			newPath = newPath + ";" + path
		}

		fmt.Println("以管理员权限设置系统 PATH 环境变量...")

		// 创建批处理文件来设置系统环境变量
		batFile := filepath.Join(os.TempDir(), "pvm_setenv.bat")
		batContent := fmt.Sprintf(`@echo off
echo Setting system PATH environment variable...
setx PATH "%s" /M
IF %%ERRORLEVEL%% NEQ 0 (
    echo Error: Failed to set system environment variable!
    pause
    exit /b 1
)
echo System environment variable updated!
`, newPath)

		if err := os.WriteFile(batFile, []byte(batContent), 0644); err != nil {
			return fmt.Errorf("创建批处理文件失败: %v", err)
		}

		// 使用 PowerShell 以管理员权限运行批处理文件
		fmt.Printf("请在弹出的 UAC 提示中选择\"是\"\n")
		psCmd := fmt.Sprintf(`Start-Process -FilePath "%s" -Verb RunAs -Wait`, batFile)
		cmd = exec.Command("powershell", "-Command", psCmd)
		output, err = cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("更新系统 PATH 环境变量失败: %v, 输出: %s", err, string(output))
		}

		fmt.Printf("系统 PATH 环境变量已永久更新!\n")
	}

	// 创建用于在当前窗口直接运行的批处理文件
	fmt.Println("为当前会话创建临时环境变量更新脚本...")
	currSessionBat := filepath.Join(os.TempDir(), "pvm_current_session.bat")
	currSessionContent := fmt.Sprintf(`@echo off
echo Setting PATH environment variable for current session...
set "PATH=%s;%%PATH%%"
echo Current PHP version:
php -v
echo.
echo Press any key to continue...
pause > nul
`, phpHomeDir)

	if err := os.WriteFile(currSessionBat, []byte(currSessionContent), 0644); err != nil {
		fmt.Printf("警告: 创建当前会话环境变量更新文件失败: %v\n", err)
	} else {
		// 立即运行此文件
		fmt.Println("正在使用新 PHP 启动命令提示符...")

		// 修复命令语法，避免特殊字符问题
		batLauncher := filepath.Join(os.TempDir(), "pvm_launch.bat")
		launchContent := fmt.Sprintf(`@echo off
cd /d "%s"
set "PATH=%s;%%PATH%%"
echo PHP environment set to %s
echo.
php -v
`, phpHome, phpHomeDir, filepath.Base(phpHome))

		if err := os.WriteFile(batLauncher, []byte(launchContent), 0644); err != nil {
			fmt.Printf("警告: 创建启动脚本失败: %v\n", err)
		} else {
			// 使用简单命令启动批处理文件
			startCmd := exec.Command("cmd", "/C", "start", "cmd", "/K", batLauncher)
			startCmd.Start()
		}

		fmt.Printf("\n当前会话的 PHP 路径尚未更新。\n")
		fmt.Printf("您有两个选择:\n")
		fmt.Printf("1. 使用刚刚打开的新命令提示符窗口\n")
		fmt.Printf("2. 运行以下命令更新当前窗口: %s\n\n", currSessionBat)
	}

	return nil
}

func extractZip(zipFile, destDir string) error {
	// 确保目标目录存在
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %v", err)
	}

	fmt.Printf("解压 %s 到 %s\n", zipFile, destDir)

	// 使用 7zip 解压（如果有的话）
	sevenZipCmd := exec.Command("7z", "x", "-o"+destDir, "-y", zipFile)
	output, err := sevenZipCmd.CombinedOutput()
	if err == nil {
		fmt.Printf("7zip 解压成功\n")
		return nil
	}

	fmt.Printf("7zip 解压失败: %v, 尝试 PowerShell...\n", err)

	// 如果 7zip 失败，使用 PowerShell
	psCmd := exec.Command("powershell", "-Command", fmt.Sprintf(
		`Expand-Archive -Path "%s" -DestinationPath "%s" -Force`,
		zipFile, destDir))

	// 捕获命令输出
	output, err = psCmd.CombinedOutput()
	if err != nil {
		// 如果 PowerShell 也失败，尝试使用 unzip 命令
		fmt.Printf("PowerShell 解压失败: %v, 尝试 unzip...\n", err)

		unzipCmd := exec.Command("unzip", "-o", zipFile, "-d", destDir)
		output, err = unzipCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("所有解压方法都失败: %v, 输出: %s", err, string(output))
		}
	}

	fmt.Printf("解压完成\n")
	return nil
}

func installVersion(version string) {
	// 获取 PHP 安装目录
	phpHome, err := getPHPHome()
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	// 下载 PHP
	fmt.Printf("正在下载 PHP %s...\n", version)
	downloadedFile, dirName, err := downloadPHP(version)
	if err != nil {
		fmt.Printf("下载失败: %v\n", err)
		return
	}
	fmt.Printf("下载完成，正在安装...\n")

	// PHP 版本安装目录 (使用从下载 URL 提取的目录名)
	versionDir := filepath.Join(phpHome, dirName)
	fmt.Printf("PHP 版本安装目录: %s\n", versionDir)

	// 如果目录已存在，先询问是否覆盖
	if _, err := os.Stat(versionDir); err == nil {
		fmt.Printf("版本 %s 已存在，是否覆盖？(y/n): ", version)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("操作已取消")
			return
		}
		// 删除已存在的目录
		os.RemoveAll(versionDir)
	}

	fmt.Printf("下载的文件: %s\n", downloadedFile)

	// 检查文件是否存在
	if _, err := os.Stat(downloadedFile); os.IsNotExist(err) {
		fmt.Printf("错误: 下载的文件不存在: %s\n", downloadedFile)
		return
	}

	// 创建临时解压目录
	tempExtractDir := filepath.Join(os.TempDir(), "pvm", "extract-"+dirName)
	os.RemoveAll(tempExtractDir) // 清除之前的目录
	os.MkdirAll(tempExtractDir, 0755)

	// 解压 PHP 文件到临时目录
	if err := extractZip(downloadedFile, tempExtractDir); err != nil {
		fmt.Printf("安装失败: %v\n", err)
		return
	}

	// 检查解压后的目录结构
	extractedItems, _ := filepath.Glob(filepath.Join(tempExtractDir, "*"))
	fmt.Printf("解压目录内容: %v\n", extractedItems)

	// 确保目标目录存在
	os.MkdirAll(versionDir, 0755)

	// 移动文件到最终目录
	// 如果解压出的是单一目录，则将其内容移动到版本目录
	if len(extractedItems) == 1 && isDir(extractedItems[0]) {
		// 复制单一目录中的所有文件
		fmt.Printf("移动目录 %s 中的内容到 %s\n", extractedItems[0], versionDir)
		moveCmd := exec.Command("xcopy", filepath.Join(extractedItems[0], "*"), versionDir, "/E", "/I", "/Y")
		if output, err := moveCmd.CombinedOutput(); err != nil {
			fmt.Printf("移动文件失败: %v, 输出: %s\n", err, string(output))
			return
		}
	} else {
		// 否则直接移动所有文件
		fmt.Printf("移动 %s 中的内容到 %s\n", tempExtractDir, versionDir)
		moveCmd := exec.Command("xcopy", filepath.Join(tempExtractDir, "*"), versionDir, "/E", "/I", "/Y")
		if output, err := moveCmd.CombinedOutput(); err != nil {
			fmt.Printf("移动文件失败: %v, 输出: %s\n", err, string(output))
			return
		}
	}

	// 列出版本目录内容
	files, _ := filepath.Glob(filepath.Join(versionDir, "*"))
	fmt.Printf("版本目录内容: %v\n", files)

	// 创建 php.ini 文件（从 php.ini-development 复制）
	iniDev := filepath.Join(versionDir, "php.ini-development")
	iniFile := filepath.Join(versionDir, "php.ini")
	if _, err := os.Stat(iniDev); err == nil {
		fmt.Printf("复制 %s 到 %s\n", iniDev, iniFile)
		// 使用文件操作而不是命令
		iniData, err := os.ReadFile(iniDev)
		if err == nil {
			os.WriteFile(iniFile, iniData, 0644)
		} else {
			fmt.Printf("复制 php.ini 失败: %v\n", err)
		}
	} else {
		fmt.Printf("找不到 php.ini-development: %v\n", err)
	}

	// 保存版本信息
	saveVersionInfo(version, dirName)

	fmt.Printf("PHP %s 安装完成\n", version)

	// 自动切换到这个版本
	useVersion(version)
}

// 检查路径是否是目录
func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func saveVersionInfo(version, dirName string) error {
	// 获取 PHP 安装目录
	phpHome, err := getPHPHome()
	if err != nil {
		return err
	}

	// 版本映射文件
	versionFile := filepath.Join(phpHome, "versions.json")

	// 读取现有版本映射
	var versionMap map[string]string
	data, err := os.ReadFile(versionFile)
	if err == nil {
		// 文件存在，解析 JSON
		if err := json.Unmarshal(data, &versionMap); err != nil {
			versionMap = make(map[string]string)
		}
	} else {
		// 文件不存在，创建新映射
		versionMap = make(map[string]string)
	}

	// 添加或更新版本映射
	versionMap[version] = dirName

	// 保存回文件
	data, err = json.MarshalIndent(versionMap, "", "  ")
	if err != nil {
		return fmt.Errorf("保存版本信息失败: %v", err)
	}

	if err := os.WriteFile(versionFile, data, 0644); err != nil {
		return fmt.Errorf("写入版本文件失败: %v", err)
	}

	fmt.Printf("版本信息已保存\n")
	return nil
}

func getVersionDir(version string) (string, error) {
	// 获取 PHP 安装目录
	phpHome, err := getPHPHome()
	if err != nil {
		return "", err
	}

	// 版本映射文件
	versionFile := filepath.Join(phpHome, "versions.json")

	// 读取版本映射
	var versionMap map[string]string
	data, err := os.ReadFile(versionFile)
	if err == nil {
		// 文件存在，解析 JSON
		if err := json.Unmarshal(data, &versionMap); err != nil {
			return "", fmt.Errorf("解析版本映射失败: %v", err)
		}

		// 查找版本
		if dirName, ok := versionMap[version]; ok {
			return filepath.Join(phpHome, dirName), nil
		}
	}

	// 版本不在映射中，尝试直接匹配目录
	pattern := filepath.Join(phpHome, fmt.Sprintf("php-%s*", version))
	matches, err := filepath.Glob(pattern)
	if err == nil && len(matches) > 0 {
		// 找到匹配的目录
		return matches[0], nil
	}

	return "", fmt.Errorf("找不到版本 %s 的安装目录", version)
}

func useVersion(version string) {
	// 获取版本目录
	versionDir, err := getVersionDir(version)
	if err != nil {
		fmt.Printf("版本 %s 不存在，是否要安装？(y/n): ", version)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) == "y" {
			installVersion(version)
			return
		}
		fmt.Println("操作已取消")
		return
	}

	fmt.Printf("找到 PHP 目录: %s\n", versionDir)

	// 验证 php.exe 是否存在
	phpExe := filepath.Join(versionDir, "php.exe")
	if _, err := os.Stat(phpExe); os.IsNotExist(err) {
		fmt.Printf("警告: php.exe 不存在于 %s，安装可能不完整\n", versionDir)
		fmt.Printf("是否重新安装 PHP %s? (y/n): ", version)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) == "y" {
			installVersion(version)
			return
		}
	}

	// PHP_HOME 目录路径
	phpHomeDir := "D:\\app\\pvm\\php_home"

	// 清理现有的PHP_HOME目录
	fmt.Printf("清理目录: %s\n", phpHomeDir)
	if err := os.RemoveAll(phpHomeDir); err != nil {
		fmt.Printf("警告: 无法删除旧目录: %v，尝试清空目录内容\n", err)

		// 尝试清空目录内容
		if err := removeContents(phpHomeDir); err != nil {
			fmt.Printf("警告: 无法清空目录内容: %v\n", err)
		}
	}

	// 等待1秒，确保文件系统操作完成
	time.Sleep(1 * time.Second)

	// 创建新的空目录
	fmt.Printf("创建新目录: %s\n", phpHomeDir)
	if err := os.MkdirAll(phpHomeDir, 0755); err != nil {
		fmt.Printf("创建目录失败: %v，尝试使用其他方法\n", err)
		return
	}

	// 使用手动文件复制方法而不是xcopy
	fmt.Printf("正在将PHP文件从 %s 复制到 %s\n", versionDir, phpHomeDir)

	// 枚举源目录中的所有文件
	err = copyDirectory(versionDir, phpHomeDir)
	if err != nil {
		fmt.Printf("复制文件失败: %v\n", err)
		fmt.Println("尝试使用robocopy命令...")

		// 如果Go的文件复制方法失败，尝试使用robocopy命令
		robocopyCmd := exec.Command("robocopy", versionDir, phpHomeDir, "/E", "/NFL", "/NDL")
		output, err := robocopyCmd.CombinedOutput()
		if err != nil {
			// robocopy的返回码不是标准的0=成功，需要特殊处理
			exitCode := robocopyCmd.ProcessState.ExitCode()
			if exitCode >= 8 {
				fmt.Printf("robocopy失败，返回码: %d, 输出: %s\n", exitCode, string(output))

				// 最后尝试使用批处理文件进行复制
				fmt.Println("尝试使用批处理文件进行复制...")
				copyBat := filepath.Join(os.TempDir(), "pvm_copy.bat")
				copyContent := fmt.Sprintf(`@echo off
echo 正在复制PHP文件...
md "%s" 2>nul
xcopy "%s\*.*" "%s\" /E /I /Y
if errorlevel 1 (
  echo 复制失败
  exit /b 1
)
echo 复制成功
`, phpHomeDir, versionDir, phpHomeDir)

				if err := os.WriteFile(copyBat, []byte(copyContent), 0644); err != nil {
					fmt.Printf("创建复制批处理文件失败: %v\n", err)
					return
				}

				copyCmd := exec.Command("cmd", "/C", copyBat)
				output, err = copyCmd.CombinedOutput()
				if err != nil {
					fmt.Printf("批处理复制失败: %v, 输出: %s\n", err, string(output))
					return
				}
			}
		}
	}

	fmt.Println("文件复制成功")

	// 创建一个批处理文件，用于在需要时刷新环境变量
	refreshBat := filepath.Join(phpHomeDir, "refresh_env.bat")
	refreshContent := fmt.Sprintf(`@echo off
echo 当前PHP版本: %s
php -v
`, version)
	if err := os.WriteFile(refreshBat, []byte(refreshContent), 0644); err != nil {
		fmt.Printf("警告: 创建刷新脚本失败: %v\n", err)
	}

	// 更新 PATH 环境变量（只添加php_home目录）
	if err := updatePATH(versionDir); err != nil {
		fmt.Printf("警告: %v\n", err)
	} else {
		fmt.Printf("已成功切换到版本 %s\n", version)
		fmt.Printf("环境变量已设置，当前会话和未来会话都将使用 PHP %s\n", version)
	}
}

// 使用Go原生函数复制目录
func copyDirectory(src, dst string) error {
	// 获取源目录信息
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// 创建目标目录
	if err = os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	// 读取源目录内容
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// 复制每个文件或子目录
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// 递归复制子目录
			if err = copyDirectory(srcPath, dstPath); err != nil {
				fmt.Printf("复制目录 %s 失败: %v\n", srcPath, err)
				return err
			}
		} else {
			// 复制文件
			if err = copyFile(srcPath, dstPath); err != nil {
				fmt.Printf("复制文件 %s 失败: %v\n", srcPath, err)
				return err
			}
		}
	}

	return nil
}

// 复制单个文件
func copyFile(src, dst string) error {
	// 打开源文件
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	// 创建目标文件
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	// 复制内容
	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	// 刷新缓冲区
	err = out.Sync()
	if err != nil {
		return err
	}

	// 复制文件权限
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}

// 移除目录中的所有内容但保留目录本身
func removeContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()

	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}

	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}

	return nil
}

func getCurrentVersion() string {
	path := os.Getenv("PATH")
	paths := strings.Split(path, ";")

	for _, p := range paths {
		// 检查路径是否包含 php 目录
		if strings.Contains(strings.ToLower(p), "php") {
			// 检查 php.exe 是否存在
			phpExe := filepath.Join(p, "php.exe")
			if _, err := os.Stat(phpExe); err == nil {
				return p
			}
		}
	}

	return ""
}

// 检查PHP官网上可用的版本
func checkAvailableVersions() {
	fmt.Println("正在查询PHP可用版本信息...")

	// 获取最新版本目录
	resp, err := http.Get(phpNewURL)
	if err != nil {
		fmt.Printf("获取版本列表失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取版本列表失败: %v\n", err)
		return
	}

	// 查找所有版本号
	reVersion := regexp.MustCompile(`php-(\d+\.\d+\.\d+)-`)
	matches := reVersion.FindAllStringSubmatch(string(body), -1)

	if len(matches) == 0 {
		fmt.Println("未找到可用的PHP版本信息")
		fmt.Println("请访问 https://windows.php.net/download 查看可用的PHP版本")
		return
	}

	// 用于去重的map
	versions := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			versions[match[1]] = true
		}
	}

	// 输出所有可用版本
	fmt.Println("在PHP官网上找到以下可用版本:")

	// 按主版本号分类
	majorVersions := make(map[string][]string)
	for version := range versions {
		parts := strings.Split(version, ".")
		if len(parts) >= 2 {
			majorVersion := parts[0] + "." + parts[1]
			majorVersions[majorVersion] = append(majorVersions[majorVersion], version)
		}
	}

	// 按主版本号排序输出
	for majorVersion, subVersions := range majorVersions {
		fmt.Printf("PHP %s 系列:\n", majorVersion)
		for _, version := range subVersions {
			fmt.Printf("  - %s\n", version)
		}
		fmt.Println()
	}

	fmt.Println("提示: 安装时可以使用简化版本号，例如:")
	fmt.Println("  pvm install 8.2 - 会安装8.2系列的最新版本")
	fmt.Println("  pvm install 8.2.0 - 会精确安装8.2.0版本")
}
