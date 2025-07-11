package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unsafe"
)

//func create_folder(path string) bool {
//	if _, err := os.Stat(path); os.IsNotExist(err) {
//		err := os.MkdirAll(path, 0755)
//		if err != nil {
//			fmt.Println("创建文件夹失败：", err)
//			return false
//		}
//		fmt.Println("文件夹创建成功")
//		return true
//	} else {
//		fmt.Println("文件夹已经存在")
//		return true
//	}
//}

func checkprotectedDirs(target string) (error, bool) {
	protectedDirs := []string{
		filepath.Join(os.Getenv("SystemRoot")),        // C:\Windows
		filepath.Join(os.Getenv("ProgramFiles")),      // C:\Program Files
		filepath.Join(os.Getenv("ProgramFiles(x86)")), // C:\Program Files (x86)
		os.Getenv("SystemDrive") + "\\",               // C:\
	}

	for _, dir := range protectedDirs {
		if dir == "" {
			continue
		}

		rel, err := filepath.Rel(dir, target)
		if err == nil && !strings.HasPrefix(rel, "..") {
			return nil, true
		}
	}
	return nil, false
}

//func isAdmin() bool{
//	_,err := os.Open("\\\\\\\\.\\\\PHYSICALDRIVE0")
//	return err == nil
//}

func pathExists(path string) (bool, error) {
	_, err := os.Lstat(path)
	if err == nil {
		return true, nil // 路径存在
	}
	if os.IsNotExist(err) {
		return false, nil // 路径不存在
	}
	return false, err // 其他错误（如权限不足）
}

//func createSymlinkSmart(target, link string) error {
//
//	if exists, err := pathExists(target); err != nil {
//		return fmt.Errorf("检查目标失败: %w", err)
//	} else if !exists {
//		return fmt.Errorf("目标路径不存在: %q", target)
//	}
//
//	if exists, err := pathExists(link); err != nil {
//		return fmt.Errorf("检查链接失败: %w", err)
//	} else if exists {
//		// 存在则删除
//		if err := os.Remove(link); err != nil {
//			return fmt.Errorf("删除旧路径失败: %w", err)
//		}
//	}
//
//	if err := os.Symlink(target, link); err != nil {
//		return fmt.Errorf("创建链接失败: %w", err)
//	}
//
//	fmt.Printf("[+]成功创建链接 %q -> %q\n", link, target)
//	return nil
//}

func formatPath(path string) string {
	path = strings.TrimSpace(path)
	return strings.TrimRight(path, `/\`)
}

func setUserEnvVar(name, value string, mode int) error {
	var key registry.Key
	key, err := registry.OpenKey(
		registry.CURRENT_USER,
		"Environment",
		registry.WRITE)
	if err != nil {
		return fmt.Errorf("[-]打开注册表失败: %v", err)
	}

	defer key.Close()

	if mode == 0 {

		if err := key.SetStringValue(name, value); err != nil {
			return fmt.Errorf("[-]写入注册表失败: %v\n", err)
		}

		if err != nil {
			return fmt.Errorf("[-]写入注册表失败: %v", err)
		}
	}

	if mode == 1 {

		path, _, err := key.GetStringValue("Path")
		if err != nil && err != registry.ErrNotExist {
			return fmt.Errorf("读取PATH失败: %v", err)
		}

		newPath := formatPath(value)
		existingPaths := strings.Split(path, ";")

		// 检查是否已存在（不区分大小写）
		for _, p := range existingPaths {
			if p == "" {
				continue
			}
			if strings.EqualFold(formatPath(p), newPath) {
				return nil // 已存在则跳过
			}
		}

		// 追加新路径（自动处理分隔符）
		var newPATH string
		if path == "" {
			newPATH = newPath
		} else {
			newPATH = path + ";" + newPath
		}

		// 写入注册表
		if err := key.SetStringValue("Path", newPATH); err != nil {
			return fmt.Errorf("写入PATH失败: %v", err)
		}

	}

	const (
		HWND_BROADCAST   = 0xFFFF
		WM_SETTINGCHANGE = 0x001A
	)
	env, _ := windows.UTF16PtrFromString("Environment")

	user32 := windows.NewLazyDLL("user32.dll")
	SendMessage := user32.NewProc("SendMessageW")

	Ret, _, err := SendMessage.Call(HWND_BROADCAST, WM_SETTINGCHANGE, 0, uintptr(unsafe.Pointer(env)))

	if Ret == 0 {
		return fmt.Errorf("[-]SendMessage 失败: %v", err)
	}

	return nil
}

func checkJavaHome() (string, error, bool) {
	exedir, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("[-]获取当前路径失败: %v", err), false
	}

	dir := filepath.Dir(exedir)

	jdkpath := "jdk"
	javahome := filepath.Join(dir, jdkpath)

	if _, err := os.Stat(javahome); os.IsNotExist(err) {
		return "", fmt.Errorf("[-]JDK路径不存在: %s", javahome), false
	}

	err = setUserEnvVar("JAVA_HOME", javahome, 0)
	if err != nil {
		return "", fmt.Errorf("[-]设置失败: %v\n", err), false
	} else {
		fmt.Println("[+]设置成功")
	}

	fmt.Println("[+] JAVA_HOME设置成功:", javahome)
	return javahome, err, true

}

func searchJDK() ([]string, error) {

	exedir, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("[-]获取当前路径失败: %v", err)
	}

	dir := filepath.Dir(exedir)

	jdkpath := "jdk"
	javahome := filepath.Join(dir, jdkpath)

	if _, err := os.Stat(javahome); os.IsNotExist(err) {
		return nil, fmt.Errorf("[-]JDK路径不存在: %s", javahome)
	}

	entries, err := os.ReadDir(javahome)
	if err != nil {
		return nil, fmt.Errorf("[-]读取目标目录失败: %v", err)
	}

	var dirs []string
	javaPattern := regexp.MustCompile(`^(jdk|jre)-?`)

	for _, entry := range entries {
		if entry.IsDir() && javaPattern.MatchString(entry.Name()) {
			dirs = append(dirs, entry.Name())
		}
	}

	return dirs, nil

}

func selectVersion(versions []string) (string, error) {

	fmt.Println("可用的Java版本:")
	for i, version := range versions {
		fmt.Printf("[%d] %s\n", i+1, version)
	}

	for {
		fmt.Println("\n请选择Java版本(输入序号):")
		var input string
		_, err := fmt.Scanln(&input)
		if err != nil {
			return "", fmt.Errorf("[-]输入错误")
		}

		choice, err := strconv.Atoi(strings.TrimSpace(input))
		if err != nil {
			return "", fmt.Errorf("[-]类型转换失败: %v", err)
			continue
		}

		if choice < 1 || choice > len(versions) {
			fmt.Printf("[-]错误: 请输入 1-%d 之间的数字\n", len(versions))
			continue
		}

		return versions[choice-1], nil

	}

}

var rootCmd = &cobra.Command{
	Use:   "jdkmanager",
	Short: "一个基于golang开发的专为解决Windows平台JDK管理困难而开发的轻量化JDK管理工具🔧",
}

var helpCmd = &cobra.Command{
	Use:    "help",
	Short:  "显示帮助信息",
	Hidden: true,
}

func init() {
	rootCmd.AddCommand(helpCmd)
}

func main() {

	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	jdks, _ := searchJDK()
	_java, _ := selectVersion(jdks)

	exedir, err := os.Executable()
	if err != nil {
	}

	dir := filepath.Dir(exedir)

	jdkpath := "jdk"
	javahome := filepath.Join(dir, jdkpath)

	if _, err := os.Stat(javahome); os.IsNotExist(err) {
	}

	java_target := filepath.Join(javahome, _java, "bin")

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "初始化管理器，默认情况下会使用Java目录下的jdk，如需指定jdk路径请使用",
		Run: func(cmd *cobra.Command, args []string) {
			_, err, check := checkJavaHome()
			if check == true {
				fmt.Println("[+]设置完成,请使用version选择java版本")
			}
			if err != nil {
				fmt.Println(err)
			}
		},
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "初始化完成后，选择Java版本",
		Run: func(cmd *cobra.Command, args []string) {

			err := setUserEnvVar("PATH", java_target, 1)

			if err != nil {
				fmt.Println(err)
			}
		},
	}

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
