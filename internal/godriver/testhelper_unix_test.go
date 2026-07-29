//go:build !windows

package godriver

func testExecutableBytes() []byte {
	return []byte{'\x7f', 'E', 'L', 'F', 2, 1, 1, 0}
}
