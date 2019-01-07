package adb

import "os"

// The status information extracted from man 2 stat

// S_IFMT 0170000      /* type of file */
// 	S_IFIFO  0010000  /* named pipe (fifo) */
// 	S_IFCHR  0020000  /* character special */
// 	S_IFDIR  0040000  /* directory */
// 	S_IFBLK  0060000  /* block special */
// 	S_IFREG  0100000  /* regular */
// 	S_IFLNK  0120000  /* symbolic link */
// 	S_IFSOCK 0140000  /* socket */
// 	S_IFWHT  0160000  /* whiteout */
// S_ISUID 0004000  /* set user id on execution */
// S_ISGID 0002000  /* set group id on execution */
// S_ISVTX 0001000  /* save swapped text even after use */
// S_IRUSR 0000400  /* read permission, owner */
// S_IWUSR 0000200  /* write permission, owner */
// S_IXUSR 0000100  /* execute/search permission, owner */

const (
	ModeDir        uint32 = 0040000
	ModeSymlink           = 0120000
	ModeSocket            = 0140000
	ModeFifo              = 0010000
	ModeCharDevice        = 0020000
	ModePerm              = 0000777
)

func fileModeFromAdb(m uint32) os.FileMode {
	mode := os.FileMode(m & ModePerm)
	if m&ModeDir != 0 {
		mode |= os.ModeDir
	}
	if m&ModeSymlink != 0 {
		mode |= os.ModeSymlink
	}
	if m&ModeSocket != 0 {
		mode |= os.ModeSocket
	}
	if m&ModeFifo != 0 {
		mode |= os.ModeNamedPipe
	}
	if m&ModeCharDevice != 0 {
		mode |= os.ModeCharDevice
	}
	return os.FileMode(mode)
}
