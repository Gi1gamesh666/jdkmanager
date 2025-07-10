package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"

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


func createSymlinkSmart(target, link string) error{

	if exists, err := pathExists(target); err != nil {
		return fmt.Errorf("检查目标失败: %w", err)
	} else if !exists {
		return fmt.Errorf("目标路径不存在: %q", target)
	}

	if exists, err := pathExists(link); err != nil {
		return fmt.Errorf("检查链接失败: %w", err)
	} else if exists {
		// 存在则删除
		if err := os.Remove(link); err != nil {
			return fmt.Errorf("删除旧路径失败: %w", err)
		}
	}

	if err := os.Symlink(target, link); err != nil {
		return fmt.Errorf("创建链接失败: %w", err)
	}

	fmt.Printf("[+]成功创建链接 %q -> %q\n", link, target)
	return nil
}


func setUserEnvVar(name, value string) error {

	key, err := syscall.RegOpenKeyEx(
		syscall.HKEY_CURRENT_USER,
		"Environment",
		0,
		syscall.KEY_SET_VALUE,
	)
	if err != nil {
		return fmt.Errorf("[-]打开注册表失败: %v", err)
	}
	defer syscall.RegCloseKey(key)

	namePtr, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return fmt.Errorf("[-]转换变量名失败: %v", err)
	}

	valuePtr, err := syscall.UTF16PtrFromString(value)
	if err != nil {
		return fmt.Errorf("[-]转换变量值失败: %v", err)
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
		return fmt.Errorf("[-]写入注册表失败: %v", err)
	}

	const (
		HWND_BROADCAST   = 0xFFFF
		WM_SETTINGCHANGE = 0x001A
	)
	env, _ := syscall.UTF16PtrFromString("Environment")
	syscall.SendMessage(HWND_BROADCAST, WM_SETTINGCHANGE, 0, uintptr(unsafe.Pointer(env)))

	return nil

}

func checkJavaHome (){
	dir ,err := filepath.Dir(os.Executable())
	if err != nil {
		return fmt.Errorf("[-]获取当前路径失败: %v", err)
	}

	jdkpath := "jdk"
	javahome := filepath.Join(dir,jdkpath)

	if _, err := os.Stat(javahome); os.IsNotExist(err) {
		return fmt.Errorf("[-]JDK路径不存在: %s", javahome)
	}

	err := setUserEnvVar("JAVA_HOME", javahome)
	if err != nil {
		return fmt.Errorf("[-]设置失败: %v\n", err)
	}else{
		fmt.Println("[+]设置成功")
	}

	fmt.Println("[+] JAVA_HOME设置成功:", javahome)
	return nil

}

func listJDK (){

	dir ,err := filepath.Dir(os.Executable())
	if err != nil {
		return fmt.Errorf("[-]获取当前路径失败: %v", err)
	}

	jdkpath := "jdk"
	javahome := filepath.Join(dir,jdkpath)

	if _, err := os.Stat(javahome); os.IsNotExist(err) {
		return fmt.Errorf("[-]JDK路径不存在: %s", javahome)
	}

	entries ,err := os.ReadDir(javahome)
	if err != nil {
		return fmt.Errorf("[-]读取目标目录失败: %v", err)
	}

	var dirs []string
	javaPattern := regexp.MustCompile(`^(jdk|jre)-?`)

	for _, entry := range entries {
		if entry.IsDir()&&javaPattern.MatchString(entry.name()){
			dirs = append(dirs, entry.name())
		}
	}

	return dirs,nil

}

func chooseJava(javaVersion int)  {



}

















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