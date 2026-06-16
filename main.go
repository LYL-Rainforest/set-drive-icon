package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

const DRIVE_FIXED = 3

var (
	modkernel32    = syscall.NewLazyDLL("kernel32.dll")
	procGetDrives  = modkernel32.NewProc("GetLogicalDrives")
	procGetDriveType = modkernel32.NewProc("GetDriveTypeW")
)

func getDriveLetters() []string {
	ret, _, _ := procGetDrives.Call()
	bitmask := uint32(ret)
	var drives []string
	for i := 0; i < 26; i++ {
		if bitmask&(1<<i) != 0 {
			drives = append(drives, string(rune('A'+i)))
		}
	}
	return drives
}

func isFixedDrive(letter string) bool {
	root, _ := syscall.UTF16PtrFromString(letter + ":\\")
	ret, _, _ := procGetDriveType.Call(uintptr(unsafe.Pointer(root)))
	return ret == DRIVE_FIXED
}

func main() {
	ok, _ := isAdmin()
	if !ok {
		fmt.Println("需要管理员权限运行")
		os.Exit(1)
	}

	for _, letter := range getDriveLetters() {
		if !isFixedDrive(letter) {
			continue
		}

		icoPath := letter + ":\\icon.ico"
		if _, err := os.Stat(icoPath); os.IsNotExist(err) {
			continue
		}

		regPath := fmt.Sprintf(`SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\DriveIcons\%s\DefaultIcon`, letter)
		k, _, err := registry.CreateKey(registry.LOCAL_MACHINE, regPath, registry.SET_VALUE)
		if err != nil {
			fmt.Printf("创建注册表键失败 %s: %v\n", letter, err)
			continue
		}
		val := icoPath + ",0"
		err = k.SetStringValue("", val)
		k.Close()
		if err != nil {
			fmt.Printf("设置注册表值失败 %s: %v\n", letter, err)
			continue
		}
		fmt.Printf("设置图标 %s -> %s\n", letter, val)
	}
}

func isAdmin() (bool, error) {
	modadvapi32 := syscall.NewLazyDLL("advapi32.dll")
	procOpenProcessToken := modadvapi32.NewProc("OpenProcessToken")
	procGetTokenInformation := modadvapi32.NewProc("GetTokenInformation")

	var token syscall.Handle
	curProc, _ := syscall.GetCurrentProcess()
	ret, _, _ := procOpenProcessToken.Call(
		uintptr(curProc),
		uintptr(0x0008), // TOKEN_QUERY
		uintptr(unsafe.Pointer(&token)),
	)
	if ret == 0 {
		return false, fmt.Errorf("OpenProcessToken failed")
	}
	defer syscall.CloseHandle(token)

	const TokenElevation = 20
	var elevation uint32
	var retLen uint32
	ret, _, _ = procGetTokenInformation.Call(
		uintptr(token),
		uintptr(TokenElevation),
		uintptr(unsafe.Pointer(&elevation)),
		uintptr(unsafe.Sizeof(elevation)),
		uintptr(unsafe.Pointer(&retLen)),
	)
	if ret == 0 {
		return false, fmt.Errorf("GetTokenInformation failed")
	}
	return elevation != 0, nil
}
