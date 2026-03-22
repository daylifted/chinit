package loader

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// ExecuteInMemory executes a binary in memory without writing to disk.
// Uses memfd_create to create an anonymous file in memory.
func ExecuteInMemory(binaryData []byte, args []string) error {
	fd, err := memfdCreate("packed-binary", 0)
	if err != nil {
		return fmt.Errorf("memfd_create failed: %w", err)
	}

	if _, err := syscall.Write(fd, binaryData); err != nil {
		syscall.Close(fd)
		return fmt.Errorf("failed to write to memfd: %w", err)
	}

	if _, err := syscall.Seek(fd, 0, 0); err != nil {
		syscall.Close(fd)
		return fmt.Errorf("failed to seek: %w", err)
	}

	execPath := fmt.Sprintf("/proc/self/fd/%d", fd)
	env := os.Environ()

	err = syscall.Exec(execPath, args, env)

	syscall.Close(fd)
	return fmt.Errorf("exec failed: %w", err)
}

// memfdCreate creates an anonymous file in memory (Linux 3.17+).
func memfdCreate(name string, flags uint) (int, error) {
	const SYS_MEMFD_CREATE = 319

	nameBytes := append([]byte(name), 0)

	r1, _, errno := syscall.Syscall(
		SYS_MEMFD_CREATE,
		uintptr(unsafe.Pointer(&nameBytes[0])),
		uintptr(flags),
		0,
	)

	if errno != 0 {
		return -1, errno
	}

	return int(r1), nil
}
