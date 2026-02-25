// From https://github.com/sevlyar/go-daemon
package file

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"syscall"
)

var (
	// ErrWoldBlock indicates on locking pid-file by another process.
	ErrWouldBlock = syscall.EWOULDBLOCK
)

// LockFile wraps *os.File and provide functions for locking of files.
type LockFile struct {
	*os.File
}

// NewLockFile returns a new LockFile with the given File.
func NewLockFile(file *os.File) *LockFile {
	return &LockFile{file}
}

var (
	defaultPIDPermsNum = 0664
)

// SaveCurrentPID Writes a pid file, but first make sure it doesn't exist with a running pid.
// https://gist.github.com/davidnewhall/3627895a9fc8fa0affbd747183abca39
func SaveCurrentPID(fileName string) error {
	if piddata, err := os.ReadFile(fileName); err == nil {
		if pid, err := strconv.Atoi(string(piddata)); err == nil {
			if process, err := os.FindProcess(pid); err == nil {
				// sig is 0, then no signal is sent, but error checking is still performed;
				// this can be used to check for the existence of a process ID or process group ID.
				if err := process.Signal(syscall.Signal(0)); err != nil {
					return fmt.Errorf("running process %d error: %w", pid, err)
				}
			}
		}
	}

	// If we get here, then the pidfile didn't exist, or the pid belong to the user running this app or
	// no process with this pid exists.
	err := os.WriteFile(fileName, fmt.Appendf(nil, "%d", os.Getpid()), os.FileMode(defaultPIDPermsNum))

	return fmt.Errorf("save current pid failed: %w", err)
}

// CreatePidFile opens the named file, applies exclusive lock and writes
// current process id to file.
func CreatePidFile(name string, perm os.FileMode) (lock *LockFile, err error) {
	if lock, err = OpenLockFile(name, perm); err != nil {
		return
	}
	if err = lock.Lock(); err != nil {
		_ = lock.Remove()
		return
	}
	if err = lock.WritePid(); err != nil {
		_ = lock.Remove()
	}

	return
}

// OpenLockFile opens the named file with flags os.O_RDWR|os.O_CREATE and specified perm.
// If successful, function returns LockFile for opened file.
func OpenLockFile(name string, perm os.FileMode) (lock *LockFile, err error) {
	var file *os.File
	if file, err = os.OpenFile(name, os.O_RDWR|os.O_CREATE, perm); err == nil {
		lock = &LockFile{file}
	}

	return
}

// Lock apply exclusive lock on an open file. If file already locked, returns error.
func (file *LockFile) Lock() error {
	err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		return fmt.Errorf("locking file failed: %w", err)
	}

	return nil
}

// Unlock remove exclusive lock on an open file.
func (file *LockFile) Unlock() error {
	err := syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	if err != nil {
		return fmt.Errorf("unlocking file failed: %w", err)
	}

	return nil
}

// ReadPidFile reads process id from file with give name and returns pid.
// If unable read from a file, returns error.
func ReadPidFile(name string) (pid int, err error) {
	var file *os.File
	if file, err = os.OpenFile(name, os.O_RDONLY, os.FileMode(defaultPIDPermsNum)); err != nil {
		return
	}
	defer file.Close()

	lock := &LockFile{file}
	pid, err = lock.ReadPid()

	return
}

// WritePid writes current process id to an open file.
func (file *LockFile) WritePid() (err error) {
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return
	}

	var fileLen int
	if fileLen, err = fmt.Fprint(file, os.Getpid()); err != nil {
		return
	}
	if err = file.Truncate(int64(fileLen)); err != nil {
		return
	}
	err = file.Sync()

	return
}

// ReadPid reads process id from file and returns pid.
// If unable read from a file, returns error.
func (file *LockFile) ReadPid() (pid int, err error) {
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return
	}
	_, err = fmt.Fscan(file, &pid)

	return
}

// Remove removes lock, closes and removes an open file.
func (file *LockFile) Remove() error {
	if file != nil {
		defer file.Close()

		if err := file.Unlock(); err != nil {
			return err
		}

		name, err := GetFdName(file.Fd())
		if err != nil {
			return err
		}

		err = syscall.Unlink(name)

		if err != nil {
			return fmt.Errorf("unlinking file '%s' failed: %w", name, err)
		}

		return nil
	}

	return os.ErrInvalid
}

// GetFdName returns file name for given descriptor.
func GetFdName(fd uintptr) (name string, err error) {
	path := fmt.Sprintf("/proc/self/fd/%d", int(fd))

	var (
		fi os.FileInfo
		n  int
	)

	if fi, err = os.Lstat(path); err != nil {
		return
	}

	buf := make([]byte, fi.Size()+1)

	if n, err = syscall.Readlink(path, buf); err == nil {
		name = string(buf[:n])
	}

	return
}
