package artifactpolicy

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

type blobStore struct {
	directory string
	files     []*os.File
}

type blob struct {
	file   *os.File
	offset int64
	size   int64
	sha256 string
}

func newBlobStore() (*blobStore, error) {
	directory, err := os.MkdirTemp("", "curator-artifact-policy-")
	if err != nil {
		return nil, err
	}
	if err := os.Chmod(directory, 0o700); err != nil { // #nosec G302 -- this is a private directory, not a regular file.
		_ = os.RemoveAll(directory)
		return nil, err
	}
	return &blobStore{directory: directory}, nil
}

func (store *blobStore) close() {
	for _, file := range store.files {
		_ = file.Close()
	}
	_ = os.RemoveAll(store.directory)
}

func (store *blobStore) newFile() (*os.File, error) {
	file, err := os.CreateTemp(store.directory, "blob-")
	if err != nil {
		return nil, err
	}
	store.files = append(store.files, file)
	return file, nil
}

func (store *blobStore) captureRoot(ctx context.Context, payload Payload, limits LimitVector) (blob, error) {
	if payload.Reader == nil {
		return blob{}, fmt.Errorf("payload reader is nil")
	}
	if payload.Size < 0 {
		return blob{}, fmt.Errorf("payload size is negative")
	}
	if payload.Size > limits.MaxRawPayloadBytes {
		return blob{}, &limitFailure{name: "max_raw_payload_bytes", limit: limits.MaxRawPayloadBytes, observed: payload.Size}
	}
	file, err := store.newFile()
	if err != nil {
		return blob{}, err
	}
	digest := sha256.New()
	reader := &contextReader{ctx: ctx, reader: io.LimitReader(payload.Reader, payload.Size+1)}
	written, err := io.Copy(io.MultiWriter(file, digest), reader)
	if err != nil {
		return blob{}, err
	}
	if written != payload.Size {
		return blob{}, fmt.Errorf("payload size mismatch: observed %d, declared %d", written, payload.Size)
	}
	return blob{
		file: file, size: written,
		sha256: "sha256:" + hex.EncodeToString(digest.Sum(nil)),
	}, nil
}

func (store *blobStore) appendBlob(ctx context.Context, destination *os.File, reader io.Reader, declared int64) (blob, error) {
	if declared < 0 {
		return blob{}, fmt.Errorf("declared member size is negative")
	}
	offset, err := destination.Seek(0, io.SeekEnd)
	if err != nil {
		return blob{}, err
	}
	digest := sha256.New()
	bounded := io.LimitReader(&contextReader{ctx: ctx, reader: reader}, declared+1)
	written, err := io.Copy(io.MultiWriter(destination, digest), bounded)
	if err != nil {
		return blob{}, err
	}
	if written != declared {
		return blob{}, fmt.Errorf("member size mismatch: observed %d, declared %d", written, declared)
	}
	return blob{
		file: destination, offset: offset, size: written,
		sha256: "sha256:" + hex.EncodeToString(digest.Sum(nil)),
	}, nil
}

func (store *blobStore) appendUnknownBounded(
	ctx context.Context,
	destination *os.File,
	reader io.Reader,
	maximum int64,
) (blob, int64, bool, error) {
	if maximum < 0 {
		return blob{}, 0, false, fmt.Errorf("maximum member size is negative")
	}
	offset, err := destination.Seek(0, io.SeekEnd)
	if err != nil {
		return blob{}, 0, false, err
	}
	digest := sha256.New()
	contextual := &contextReader{ctx: ctx, reader: reader}
	written := int64(0)
	if maximum > 0 {
		written, err = io.CopyN(io.MultiWriter(destination, digest), contextual, maximum)
		if err != nil && err != io.EOF {
			return blob{}, written, false, err
		}
	}
	captured := blob{
		file: destination, offset: offset, size: written,
		sha256: "sha256:" + hex.EncodeToString(digest.Sum(nil)),
	}
	if err == io.EOF {
		return captured, written, false, nil
	}
	var sentinel [1]byte
	read, readErr := io.ReadFull(contextual, sentinel[:])
	observed := written + int64(read)
	if read > 0 {
		// The sentinel proves the closed limit was crossed but is deliberately
		// not written to the spool. No output beyond the allowed maximum is
		// retained or consumed after this byte.
		return captured, observed, true, nil
	}
	if readErr == io.EOF {
		return captured, observed, false, nil
	}
	return blob{}, observed, false, readErr
}

func hashBlob(ctx context.Context, item blob) (string, error) {
	digest := sha256.New()
	written, err := io.Copy(digest, &contextReader{ctx: ctx, reader: item.reader()})
	if err != nil {
		return "", err
	}
	if written != item.size {
		return "", fmt.Errorf("blob changed while hashing: observed %d, expected %d", written, item.size)
	}
	return "sha256:" + hex.EncodeToString(digest.Sum(nil)), nil
}

func (item blob) reader() *io.SectionReader {
	return io.NewSectionReader(item.file, item.offset, item.size)
}

func (item blob) readAt(buffer []byte, offset int64) (int, error) {
	if offset < 0 || offset >= item.size {
		return 0, io.EOF
	}
	return item.reader().ReadAt(buffer, offset)
}

func (item blob) prefix(maximum int64) ([]byte, error) {
	if maximum < 0 {
		return nil, fmt.Errorf("negative prefix size")
	}
	size := item.size
	if size > maximum {
		size = maximum
	}
	payload := make([]byte, size)
	if size == 0 {
		return payload, nil
	}
	_, err := item.reader().ReadAt(payload, 0)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return payload, nil
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (reader *contextReader) Read(payload []byte) (int, error) {
	if err := contextError(reader.ctx); err != nil {
		return 0, err
	}
	return reader.reader.Read(payload)
}
