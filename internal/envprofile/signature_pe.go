package envprofile

import (
	"bytes"
	"debug/pe"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

// Import DLL identity is independent of symbol names, including ordinal imports.
// Every descriptor, lookup/name and corresponding IAT slot must be backed by one
// unambiguous section. Never use debug/pe.ImportedSymbols as a validity check.
func signaturePEImports(file *pe.File) ([]string, error) {
	var directory pe.DataDirectory
	var width uint64
	switch header := file.OptionalHeader.(type) {
	case *pe.OptionalHeader32:
		if header.NumberOfRvaAndSizes <= pe.IMAGE_DIRECTORY_ENTRY_IMPORT {
			return nil, errors.New("missing import directory")
		}
		directory = header.DataDirectory[pe.IMAGE_DIRECTORY_ENTRY_IMPORT]
		width = 4
	case *pe.OptionalHeader64:
		if header.NumberOfRvaAndSizes <= pe.IMAGE_DIRECTORY_ENTRY_IMPORT {
			return nil, errors.New("missing import directory")
		}
		directory = header.DataDirectory[pe.IMAGE_DIRECTORY_ENTRY_IMPORT]
		width = 8
	default:
		return nil, errors.New("unsupported PE optional header")
	}
	if directory.VirtualAddress == 0 || directory.Size < 40 {
		return nil, errors.New("missing import descriptors")
	}
	if _, _, err := signaturePERange(file, uint64(directory.VirtualAddress), uint64(directory.Size)); err != nil {
		return nil, err
	}
	var libraries []string
	sawNamed := false
	for offset := uint64(0); offset+20 <= uint64(directory.Size); offset += 20 {
		descriptor, err := signaturePERead(file, uint64(directory.VirtualAddress)+offset, 20)
		if err != nil {
			return nil, err
		}
		if bytes.Equal(descriptor, make([]byte, 20)) {
			if !sawNamed {
				return nil, errors.New("ordinal-only import inventory is not qualified")
			}
			return libraries, nil
		}
		lookup := uint64(binary.LittleEndian.Uint32(descriptor))
		name := uint64(binary.LittleEndian.Uint32(descriptor[12:]))
		first := uint64(binary.LittleEndian.Uint32(descriptor[16:]))
		library, err := signaturePEName(file, name)
		if err != nil || first == 0 {
			return nil, errors.New("invalid import DLL or IAT")
		}
		if lookup == 0 {
			lookup = first
		}
		named, err := signaturePEThunks(file, lookup, first, width)
		if err != nil {
			return nil, err
		}
		sawNamed = sawNamed || named
		libraries = append(libraries, library)
	}
	return nil, errors.New("unterminated import directory")
}

func signaturePERange(file *pe.File, rva, size uint64) (*pe.Section, int64, error) {
	if rva == 0 || rva > math.MaxUint32 || size > math.MaxUint32 || rva+size > uint64(math.MaxUint32)+1 {
		return nil, 0, errors.New("import RVA overflow")
	}
	var match *pe.Section
	var offset uint64
	for _, section := range file.Sections {
		start := uint64(section.VirtualAddress)
		backed := uint64(section.Size)
		if section.VirtualSize != 0 && uint64(section.VirtualSize) < backed {
			backed = uint64(section.VirtualSize)
		}
		if rva >= start && rva-start < backed && size <= backed-(rva-start) {
			if match != nil {
				return nil, 0, errors.New("ambiguous import RVA")
			}
			match, offset = section, rva-start
		}
	}
	if match == nil {
		return nil, 0, errors.New("unmapped import RVA")
	}
	return match, int64(offset), nil
}

func signaturePERead(file *pe.File, rva, size uint64) ([]byte, error) {
	section, offset, err := signaturePERange(file, rva, size)
	if err != nil {
		return nil, err
	}
	data := make([]byte, size)
	if _, err := section.ReadAt(data, offset); err != nil {
		return nil, fmt.Errorf("unbacked import RVA: %w", err)
	}
	return data, nil
}

func signaturePEName(file *pe.File, rva uint64) (string, error) {
	var name []byte
	for {
		data, err := signaturePERead(file, rva, 1)
		if err != nil {
			return "", err
		}
		if data[0] == 0 {
			if len(name) == 0 {
				return "", errors.New("empty import name")
			}
			return string(name), nil
		}
		if data[0] > 127 || data[0] < 32 {
			return "", errors.New("non-ASCII import name")
		}
		name = append(name, data[0])
		rva++
	}
}

func signaturePEThunks(file *pe.File, lookup, first, width uint64) (bool, error) {
	named := false
	flag := uint64(1) << (width*8 - 1)
	for index := uint64(0); ; index++ {
		entry, err := signaturePERead(file, lookup+index*width, width)
		if err != nil {
			return false, err
		}
		iat, err := signaturePERead(file, first+index*width, width)
		if err != nil {
			return false, err
		}
		value, address := uint64(binary.LittleEndian.Uint32(entry)), uint64(binary.LittleEndian.Uint32(iat))
		if width == 8 {
			value, address = binary.LittleEndian.Uint64(entry), binary.LittleEndian.Uint64(iat)
		}
		if value == 0 {
			if address != 0 {
				return false, errors.New("unterminated IAT")
			}
			return named, nil
		}
		if address == 0 {
			return false, errors.New("truncated IAT")
		}
		if value&flag != 0 {
			if value & ^(flag|0xffff) != 0 {
				return false, errors.New("invalid ordinal import")
			}
			continue
		}
		named = true
		if _, err := signaturePERead(file, value, 2); err != nil {
			return false, err
		}
		if _, err := signaturePEName(file, value+2); err != nil {
			return false, err
		}
	}
}
