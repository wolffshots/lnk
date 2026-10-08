// Package fs provides file system operations for lnk.
package fs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/yarlson/lnk/internal/lnkerror"
)

// Sentinel errors for file system operations.
var (
	ErrFileNotExists   = errors.New("File or directory not found")
	ErrFileCheck       = errors.New("Unable to access file. Please check file permissions and try again.")
	ErrUnsupportedType = errors.New("Cannot manage this type of file")
	ErrNotManaged      = errors.New("File is not managed by lnk")
	ErrSymlinkRead     = errors.New("Unable to read symlink. The file may be corrupted or have invalid permissions.")
	ErrDirCreate       = errors.New("Failed to create directory. Please check permissions and available disk space.")
	ErrRelativePath    = errors.New("Unable to create symlink due to path configuration issues. Please check file locations.")
	ErrOutsideHome     = errors.New("Cannot manage a file outside the home directory on Windows")
	ErrCrossDevice     = errors.New("The lnk repository and the file are on different drives or file systems")
	ErrFileInUse       = errors.New("Cannot move the file because Windows denied access")
	ErrSymlinkDenied   = errors.New("Windows did not allow lnk to create a symlink")
)

// Windows API error codes. The syscall package names them on Windows only.
const (
	winErrAccessDenied     = syscall.Errno(5)
	winErrNotSameDevice    = syscall.Errno(17)
	winErrSharingViolation = syscall.Errno(32)
	winErrPrivilegeNotHeld = syscall.Errno(1314)
)

const sameDriveSuggestion = "keep the lnk repository on the same drive as your files, set LNK_HOME to change its location"

// FileSystem handles file system operations
type FileSystem struct{}

// New creates a new FileSystem instance
func New() *FileSystem {
	return &FileSystem{}
}

// ValidateFileForAdd validates that a file or directory can be added to lnk
func (fs *FileSystem) ValidateFileForAdd(filePath string) error {
	// Check if file exists and get its info
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return lnkerror.WithPath(ErrFileNotExists, filePath)
		}

		return lnkerror.WithPath(ErrFileCheck, filePath)
	}

	// Allow both regular files and directories
	if !info.Mode().IsRegular() && !info.IsDir() {
		return lnkerror.WithPathAndSuggestion(ErrUnsupportedType, filePath, "lnk can only manage regular files and directories")
	}

	return nil
}

// ValidateSymlinkForRemove validates that a symlink can be removed from lnk
func (fs *FileSystem) ValidateSymlinkForRemove(filePath, repoPath string) error {
	// Check if file exists and is a symlink
	info, err := os.Lstat(filePath) // Use Lstat to not follow symlinks
	if err != nil {
		if os.IsNotExist(err) {
			return lnkerror.WithPath(ErrFileNotExists, filePath)
		}

		return lnkerror.WithPath(ErrFileCheck, filePath)
	}

	if info.Mode()&os.ModeSymlink == 0 {
		return lnkerror.WithPathAndSuggestion(ErrNotManaged, filePath, "use 'lnk add' to manage this file first")
	}

	// Get symlink target and resolve to absolute path
	target, err := os.Readlink(filePath)
	if err != nil {
		return lnkerror.WithPath(ErrSymlinkRead, filePath)
	}

	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(filePath), target)
	}

	// Clean paths and check if target is inside the repository
	target = filepath.Clean(target)
	repoPath = filepath.Clean(repoPath)

	if !strings.HasPrefix(target, repoPath+string(filepath.Separator)) && target != repoPath {
		return lnkerror.WithPathAndSuggestion(ErrNotManaged, filePath, "use 'lnk add' to manage this file first")
	}

	return nil
}

// Move moves a file or directory from source to destination based on the file info
func (fs *FileSystem) Move(src, dst string, info os.FileInfo) error {
	if info.IsDir() {
		return fs.MoveDirectory(src, dst)
	}
	return fs.MoveFile(src, dst)
}

// MoveFile moves a file from source to destination
func (fs *FileSystem) MoveFile(src, dst string) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return lnkerror.WithPath(ErrDirCreate, filepath.Dir(dst))
	}

	// Move the file
	return rename(src, dst)
}

// CreateSymlink creates a relative symlink from target to linkPath
func (fs *FileSystem) CreateSymlink(target, linkPath string) error {
	// Calculate relative path from linkPath to target
	relTarget, err := filepath.Rel(filepath.Dir(linkPath), target)
	if err != nil {
		// On Windows, a relative symlink cannot cross volumes.
		if runtime.GOOS == "windows" && !strings.EqualFold(filepath.VolumeName(linkPath), filepath.VolumeName(target)) {
			return lnkerror.WithPathAndSuggestion(ErrCrossDevice, linkPath, sameDriveSuggestion)
		}
		return lnkerror.Wrap(ErrRelativePath)
	}

	// Create the symlink
	return mapSymlinkError(os.Symlink(relTarget, linkPath))
}

// MoveDirectory moves a directory from source to destination recursively
func (fs *FileSystem) MoveDirectory(src, dst string) error {
	// Ensure destination parent directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return lnkerror.WithPath(ErrDirCreate, filepath.Dir(dst))
	}

	// Move the directory
	return rename(src, dst)
}

// rename moves src to dst and maps the failures a user can fix to sentinel errors.
func rename(src, dst string) error {
	return mapRenameError(os.Rename(src, dst), src)
}

func mapRenameError(err error, src string) error {
	if err == nil {
		return nil
	}

	if runtime.GOOS == "windows" {
		if errors.Is(err, winErrNotSameDevice) {
			return lnkerror.WithPathAndSuggestion(ErrCrossDevice, src, sameDriveSuggestion)
		}
		// Windows refuses to move a directory while a program has a file inside it open.
		if errors.Is(err, winErrAccessDenied) || errors.Is(err, winErrSharingViolation) {
			return lnkerror.WithPathAndSuggestion(ErrFileInUse, src, "close any program that has the file, or a file inside the directory, open and try again")
		}
		return err
	}

	if errors.Is(err, syscall.EXDEV) {
		return lnkerror.WithPathAndSuggestion(ErrCrossDevice, src, sameDriveSuggestion)
	}

	return err
}

func mapSymlinkError(err error) error {
	if runtime.GOOS == "windows" && errors.Is(err, winErrPrivilegeNotHeld) {
		return lnkerror.WithSuggestion(ErrSymlinkDenied, "turn on Developer Mode in Windows Settings, or run lnk as administrator")
	}
	return err
}

// GetRelativePath converts an absolute path to a relative path from the home directory.
// The result uses forward slashes on every platform, so a .lnk file is portable.
func GetRelativePath(absPath string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	relPath, err := filepath.Rel(homeDir, absPath)
	if err != nil {
		// On Windows, Rel fails for an absolute path on a different volume.
		if runtime.GOOS == "windows" && filepath.IsAbs(absPath) {
			return "", lnkerror.WithPath(ErrOutsideHome, absPath)
		}
		return "", fmt.Errorf("failed to get relative path: %w", err)
	}

	if strings.HasPrefix(relPath, "..") {
		if runtime.GOOS == "windows" {
			return "", lnkerror.WithPath(ErrOutsideHome, absPath)
		}
		return strings.TrimPrefix(absPath, "/"), nil
	}

	return filepath.ToSlash(relPath), nil
}
