package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	phpBaseURL = "https://windows.php.net/downloads/releases/archives/"
)

func main() {
	// 获取命令行参数
	args := os.Args[1:]

	// 如果没有参数，执行默认的更新操作
	if len(args) == 0 {
		updateProgram()
		return
	}

	// 处理各种命令
	switch args[0] {
	case "list":
		listVersions()
	case "install":
		if len(args) < 2 {
			fmt.Println("请指定要安装的版本，例如：pvm install 7.4")
			return
		}
		installVersion(args[1])
	default:
		fmt.Println("未知命令。可用命令：")
		fmt.Println("  pvm list - 列出所有版本")
		fmt.Println("  pvm install <版本> - 安装指定版本")
		fmt.Println("  pvm - 更新到最新版本")
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

func listVersions() {
	// 获取当前目录
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("获取当前目录失败: %v\n", err)
		return
	}

	// 列出当前目录下的所有 pvm_*.exe 文件
	files, err := filepath.Glob(filepath.Join(currentDir, "pvm_*.exe"))
	if err != nil {
		fmt.Printf("列出文件失败: %v\n", err)
		return
	}

	if len(files) == 0 {
		fmt.Println("没有找到任何版本")
		return
	}

	fmt.Println("可用版本：")
	for _, file := range files {
		// 获取文件信息
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		// 提取版本号（去掉 pvm_ 前缀和 .exe 后缀）
		version := filepath.Base(file)
		version = version[4 : len(version)-4]
		fmt.Printf("  %s (大小: %d 字节, 修改时间: %s)\n", 
			version, 
			info.Size(), 
			info.ModTime().Format("2006-01-02 15:04:05"))
	}
}

func downloadPHP(version string) error {
	// 构建下载 URL
	url := fmt.Sprintf("%sphp-%s-Win32-VC15-x64.zip", phpBaseURL, version)
	
	// 创建下载目录
	downloadDir := filepath.Join(os.TempDir(), "pvm")
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return fmt.Errorf("创建下载目录失败: %v", err)
	}

	// 下载文件
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("下载失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	// 保存文件
	out, err := os.Create(filepath.Join(downloadDir, fmt.Sprintf("php-%s.zip", version)))
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("保存文件失败: %v", err)
	}

	return nil
}

func installVersion(version string) {
	// 获取当前目录
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("获取当前目录失败: %v\n", err)
		return
	}

	// 检查版本文件是否存在
	versionFile := filepath.Join(currentDir, fmt.Sprintf("pvm_%s.exe", version))
	if _, err := os.Stat(versionFile); os.IsNotExist(err) {
		fmt.Printf("正在下载 PHP %s...\n", version)
		if err := downloadPHP(version); err != nil {
			fmt.Printf("下载失败: %v\n", err)
			return
		}
		fmt.Printf("下载完成，正在安装...\n")
	}

	// 检查程序是否在运行
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq pvm.exe")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		// 如果程序在运行，则关闭它
		killCmd := exec.Command("taskkill", "/F", "/IM", "pvm.exe")
		killCmd.Run()
		time.Sleep(time.Second) // 等待程序完全关闭
	}

	// 删除当前版本
	os.Remove(filepath.Join(currentDir, "pvm.exe"))

	// 复制指定版本
	copyCmd := exec.Command("copy", "/Y", versionFile, "pvm.exe")
	if err := copyCmd.Run(); err != nil {
		fmt.Printf("安装版本 %s 失败: %v\n", version, err)
		return
	}

	fmt.Printf("成功安装版本 %s\n", version)

	// 启动新程序
	startCmd := exec.Command("start", "pvm.exe")
	startCmd.Run()

	// 等待5秒
	time.Sleep(5 * time.Second)
} 