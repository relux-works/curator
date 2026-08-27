//go:build windows

package godriver

func testExecutableBytes() []byte {
	return []byte{'M', 'Z', 0, 0, 0, 0, 0, 0}
}
