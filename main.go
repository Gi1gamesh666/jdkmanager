//package main
//
//import (
//	"fmt"
//	"log"
//	"github.com/spf13/cobra"
//	"syscall"
//	"unsafe"
//)
//
//const (
//	HWND_BROADCAST   = 0xFFFF
//	WM_SETTINGCHANGE = 0x001A
//)
//
//func setUserEnvVar(name, value string) error {
//	// 打开注册表键
//	key, err := syscall.RegOpenKeyEx(
//		syscall.HKEY_CURRENT_USER,
//		"Environment",
//		0,
//		syscall.KEY_SET_VALUE,
//	)
//	if err != nil {
//		return fmt.Errorf("打开注册表失败: %v", err)
//	}
//	defer syscall.RegCloseKey(key)
//
//	// 转换字符串为UTF-16指针
//	namePtr, err := syscall.UTF16PtrFromString(name)
//	if err != nil {
//		return fmt.Errorf("转换变量名失败: %v", err)
//	}
//
//	valuePtr, err := syscall.UTF16PtrFromString(value)
//	if err != nil {
//		return fmt.Errorf("转换变量值失败: %v", err)
//	}
//
//	err = syscall.RegSetValueEx(
//		key,
//		namePtr,
//		0,
//		syscall.REG_SZ,
//		(*byte)(unsafe.Pointer(valuePtr)),
//		(uint32)(len(value)+1)*2,
//	)
//	if err != nil {
//		return fmt.Errorf("写入注册表失败: %v", err)
//	}
//
//	// 通知系统环境变量已更改
//	userEnv, _ := syscall.UTF16PtrFromString("Environment")
//	syscall.SendMessage(HWND_BROADCAST, WM_SETTINGCHANGE, 0, uintptr(unsafe.Pointer(userEnv)))
//
//	return nil
//}
//
//func main() {
//	// 示例：设置用户级JAVA_HOME
//	err := setUserEnvVar("JAVA_HOME", "C:\\Users\\YourName\\jdk-17")
//	if err != nil {
//		log.Fatalf("设置用户环境变量失败: %v", err)
//	}
//
//	// 示例：添加到用户PATH
//	currentPath := os.Getenv("PATH")
//	newPath := fmt.Sprintf("%s;C:\\Users\\YourName\\bin", currentPath)
//	err = setUserEnvVar("PATH", newPath)
//	if err != nil {
//		log.Fatalf("更新用户PATH失败: %v", err)
//	}
//
//	fmt.Println("用户环境变量设置成功")
//
//
//
//
//
//}

package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

func create_folder(path string) {
	if _,err := os.Stat(path); os.IsNotExist(err){
		err := os.MkdirAll(path, 0755)
		if err != nil {
			fmt.Println("创建文件夹失败：",err)
			return false
		}
		fmt.Println("文件夹创建成功")
		return true
	}
	else{
		fmt.Println("文件夹已经存在")
		return true
	}
}

func checkprotectedDirs(target string) error{
	protectedDirs := string[]{
		filepath.Join(os.Getenv("SystemRoot")),       // C:\Windows
		filepath.Join(os.Getenv("ProgramFiles")),     // C:\Program Files
		filepath.Join(os.Getenv("ProgramFiles(x86)")), // C:\Program Files (x86)
		os.Getenv("SystemDrive") + "\\",              // C:\
	}

	for _, dir := range protectedDirs {
		if dir == "" {
			continue
		}

		rel, err := filepath.Rel(dir, targetDir)
		if err == nil && !strings.HasPrefix(rel, "..") {
			return true, nil
		}else{
			return false,err
		}
	}
}

//func isAdmin() bool{
//	_,err := os.Open("\\\\\\\\.\\\\PHYSICALDRIVE0")
//	return err == nil
//}


func createSymlinkSmart(target, link string) error{

	targetinfo,err := os.Lstat(target)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("[-]jdk路径：%q 不存在",target)
			return false,nil
		}
		return false,fmt.Println("[-]检查目标路径失败：%w",err)
	}

	fl,err := os.Lstat(link)
	if err != nil {
		if os.IsNotExist(err) {
			err :=os.Symlink(target, link)
			if err == nil {
				fmt.Println("[+]成功创建链接")
			}

		}
		err := os.Remove(link)
		if err == nil {
			err :=os.Symlink(target, link)
		}
		return false,err
	}







}

func setUserEnvVar(name, value string) error {

	key, err := syscall.RegOpenKeyEx(
		syscall.HKEY_CURRENT_USER,
		"Environment",
		0,
		syscall.KEY_SET_VALUE,
	)
	if err != nil {
		return fmt.Errorf("打开注册表失败: %v", err)
	}
	defer syscall.RegCloseKey(key)

	namePtr, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return fmt.Errorf("转换变量名失败: %v", err)
	}

	valuePtr, err := syscall.UTF16PtrFromString(value)
	if err != nil {
		return fmt.Errorf("转换变量值失败: %v", err)
	}

	err = syscall.RegSetValueEx(
		key,
		namePtr,
		0,
		syscall.REG_SZ,
		(*byte)(unsafe.Pointer(valuePtr)),
		uint32(len(value)+1)*2, // UTF-16字节长度（含null终止符）
	)
	if err != nil {
		return fmt.Errorf("写入注册表失败: %v", err)
	}

	const (
		HWND_BROADCAST   = 0xFFFF
		WM_SETTINGCHANGE = 0x001A
	)
	env, _ := syscall.UTF16PtrFromString("Environment")
	syscall.SendMessage(HWND_BROADCAST, WM_SETTINGCHANGE, 0, uintptr(unsafe.Pointer(env)))

	return nil

}

func checkJava




var rootCmd = &cobra.Command{
	Use:   "jdkmanager",
	Short: "一个基于golang开发的专为解决Windows平台JDK管理困难而开发的轻量化JDK管理工具🔧",
}


var helloCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化管理器，默认情况下会使用Java目录下的jdk，如需指定jdk路径请使用",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello World")
	},
}




func init() {
	rootCmd.AddCommand(helloCmd)
}

func main() {

	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "help",
		Short:  "显示帮助信息",
		Hidden: true,
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}