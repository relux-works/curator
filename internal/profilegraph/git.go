package profilegraph

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1" // #nosec G505 -- Git object identity, never a trust signature.
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"io"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/pkgversion"
)

// gitDB reads only object/refs families. It never loads repository config,
// alternates, hooks, attributes, credentials or a subprocess environment.
type gitObject struct {
	kind  string
	bytes []byte
}

type gitDB struct {
	b           *build
	root, label string
}

func (b *build) db(source string) gitDB {
	root := path.Join(envprofile.GraphRepository(b.req.Home, source), ".git")
	if err := b.r.bind(root, b.req.Home); err != nil {
		b.err = err
	}
	return gitDB{b, root, "git/" + source}
}
func (g gitDB) read(p string) ([]byte, error) { return g.b.r.Read(g.root, g.label, p) }
func oid(s string) bool {
	if len(s) != 40 && len(s) != 64 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil && s == strings.ToLower(s)
}
func (g gitDB) inflate(raw []byte) ([]byte, error) {
	z, err := zlib.NewReader(bytes.NewReader(raw))
	if err != nil {
		return nil, refusal("graph_invalid_input")
	}
	defer func() { _ = z.Close() }()
	result, err := io.ReadAll(io.LimitReader(z, int64(g.b.r.Limits.File)+129))
	if err != nil {
		return nil, refusal("graph_invalid_input")
	}
	if len(result) > g.b.r.Limits.File+128 {
		return nil, refusal("graph_limit_exceeded")
	}
	return result, nil
}

// Cache verified decoded objects as well as raw files. Repeated tree walks do
// not repeatedly inflate data; expanded payloads share the consumed-byte budget.
func (g gitDB) object(id string, depth int) (string, []byte, error) {
	if depth > g.b.r.Limits.Depth {
		return "", nil, refusal("graph_limit_exceeded")
	}
	key := g.label + "\x00" + id
	if cached, ok := g.b.objects[key]; ok {
		return cached.kind, cached.bytes, nil
	}
	kind, raw, err := g.readObject(id, depth)
	if err != nil {
		return "", nil, err
	}
	if len(raw) > g.b.r.Limits.File || len(raw) > g.b.r.Limits.Input-g.b.r.total {
		return "", nil, refusal("graph_limit_exceeded")
	}
	g.b.r.total += len(raw)
	if g.b.objects == nil {
		g.b.objects = map[string]gitObject{}
	}
	g.b.objects[key] = gitObject{kind, raw}
	return kind, raw, nil
}
func (g gitDB) readObject(id string, depth int) (string, []byte, error) {
	if !oid(id) {
		return "", nil, refusal("graph_invalid_input")
	}
	if depth > g.b.r.Limits.Depth {
		return "", nil, refusal("graph_limit_exceeded")
	}
	raw, err := g.read("objects/" + id[:2] + "/" + id[2:])
	if err == nil {
		raw, err = g.inflate(raw)
		if err != nil {
			return "", nil, err
		}
		header, body, ok := bytes.Cut(raw, []byte{0})
		if !ok {
			return "", nil, refusal("graph_invalid_input")
		}
		parts := strings.Split(string(header), " ")
		if len(parts) != 2 || parts[1] != strconv.Itoa(len(body)) {
			return "", nil, refusal("graph_invalid_input")
		}
		if objectHash(raw, len(id)/2) != id {
			return "", nil, refusal("graph_invalid_input")
		}
		return parts[0], body, nil
	}
	if !isMissing(err) {
		return "", nil, err
	}
	entries, err := g.b.r.directory(g.root, "objects/pack")
	if err != nil {
		return "", nil, err
	}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".idx") {
			idx, err := g.read("objects/pack/" + entry.Name())
			if err != nil {
				return "", nil, err
			}
			offset, found, err := indexOffset(idx, id)
			if err != nil {
				return "", nil, err
			}
			if !found {
				continue
			}
			pack, err := g.read("objects/pack/" + strings.TrimSuffix(entry.Name(), ".idx") + ".pack")
			if err != nil {
				return "", nil, err
			}
			kind, body, err := g.packed(pack, offset, depth, len(id)/2)
			if err != nil {
				return "", nil, err
			}
			framed := append([]byte(kind+" "+strconv.Itoa(len(body))+"\x00"), body...)
			if objectHash(framed, len(id)/2) != id {
				return "", nil, refusal("graph_invalid_input")
			}
			return kind, body, nil
		}
	}
	return "", nil, refusal("graph_missing_input")
}
func indexOffset(idx []byte, id string) (int, bool, error) {
	if len(idx) < 1032 || !bytes.Equal(idx[:4], []byte{255, 't', 'O', 'c'}) || binary.BigEndian.Uint32(idx[4:8]) != 2 {
		return 0, false, refusal("graph_invalid_input")
	}
	width := len(id) / 2
	n := uint64(binary.BigEndian.Uint32(idx[1028:1032]))
	if 1032+n*uint64(width+8)+uint64(width*2) > uint64(len(idx)) { // #nosec G115 -- width is a validated Git object ID width, 20 or 32.
		return 0, false, refusal("graph_invalid_input")
	}
	wanted, _ := hex.DecodeString(id)
	// n fits the bounded index file, so all following offsets also fit int.
	pos := sort.Search(int(n), func(i int) bool { return bytes.Compare(idx[1032+i*width:1032+(i+1)*width], wanted) >= 0 }) // #nosec G115 -- n fits the bounded index file.
	if pos == int(n) || !bytes.Equal(idx[1032+pos*width:1032+(pos+1)*width], wanted) {                                     // #nosec G115 -- n fits the bounded index file.
		return 0, false, nil
	}
	start := 1032 + int(n)*(width+4) + pos*4 // #nosec G115 -- n fits the bounded index file.
	v := binary.BigEndian.Uint32(idx[start : start+4])
	if v&0x80000000 != 0 {
		return 0, false, refusal("graph_limit_exceeded")
	}
	return int(v), true, nil
}
func (g gitDB) packed(pack []byte, offset, depth, width int) (string, []byte, error) {
	if depth > g.b.r.Limits.Depth {
		return "", nil, refusal("graph_limit_exceeded")
	}
	if len(pack) < 12 || string(pack[:4]) != "PACK" || offset < 12 || offset >= len(pack) {
		return "", nil, refusal("graph_invalid_input")
	}
	pos := offset
	c := pack[pos]
	pos++
	kind := int(c>>4) & 7
	size := int(c & 15)
	shift := 4
	for c&128 != 0 {
		if pos >= len(pack) || shift > 28 {
			return "", nil, refusal("graph_invalid_input")
		}
		c = pack[pos]
		pos++
		size |= int(c&127) << shift
		shift += 7
	}
	if size > g.b.r.Limits.File {
		return "", nil, refusal("graph_limit_exceeded")
	}
	baseOffset := 0
	baseID := ""
	switch kind {
	case 6:
		if pos >= len(pack) {
			return "", nil, refusal("graph_invalid_input")
		}
		c = pack[pos]
		pos++
		distance := int(c & 127)
		for c&128 != 0 {
			if pos >= len(pack) || distance > len(pack) {
				return "", nil, refusal("graph_invalid_input")
			}
			c = pack[pos]
			pos++
			distance = ((distance + 1) << 7) | int(c&127)
		}
		baseOffset = offset - distance
	case 7:
		if pos+width > len(pack) {
			return "", nil, refusal("graph_invalid_input")
		}
		baseID = hex.EncodeToString(pack[pos : pos+width])
		pos += width
	}
	raw, err := g.inflate(pack[pos:])
	if err != nil {
		return "", nil, err
	}
	if len(raw) != size {
		return "", nil, refusal("graph_invalid_input")
	}
	if kind == 6 || kind == 7 {
		var typ string
		var base []byte
		if kind == 6 {
			typ, base, err = g.packed(pack, baseOffset, depth+1, width)
		} else {
			typ, base, err = g.object(baseID, depth+1)
		}
		if err != nil {
			return "", nil, err
		}
		result, err := applyDelta(base, raw, g.b.r.Limits.File)
		return typ, result, err
	}
	names := map[int]string{1: "commit", 2: "tree", 3: "blob", 4: "tag"}
	typ, ok := names[kind]
	if !ok {
		return "", nil, refusal("graph_invalid_input")
	}
	return typ, raw, nil
}
func applyDelta(base, delta []byte, limit int) ([]byte, error) {
	pos := 0
	number := func() (int, bool) {
		n, shift := 0, 0
		for pos < len(delta) && shift < 35 {
			c := delta[pos]
			pos++
			n |= int(c&127) << shift
			if c&128 == 0 {
				return n, true
			}
			shift += 7
		}
		return 0, false
	}
	original, ok := number()
	if !ok || original != len(base) {
		return nil, refusal("graph_invalid_input")
	}
	size, ok := number()
	if !ok {
		return nil, refusal("graph_invalid_input")
	}
	if size > limit {
		return nil, refusal("graph_limit_exceeded")
	}
	out := make([]byte, 0, size)
	for pos < len(delta) {
		op := delta[pos]
		pos++
		if op == 0 {
			return nil, refusal("graph_invalid_input")
		}
		if op&128 == 0 {
			n := int(op)
			if pos+n > len(delta) || len(out)+n > size {
				return nil, refusal("graph_invalid_input")
			}
			out = append(out, delta[pos:pos+n]...)
			pos += n
			continue
		}
		offset, n := 0, 0
		for i := 0; i < 7; i++ {
			if op&(1<<i) != 0 {
				if pos >= len(delta) {
					return nil, refusal("graph_invalid_input")
				}
				v := int(delta[pos])
				pos++
				if i < 4 {
					offset |= v << (8 * i)
				} else {
					n |= v << (8 * (i - 4))
				}
			}
		}
		if n == 0 {
			n = 65536
		}
		if offset > len(base) || n > len(base)-offset || n > size-len(out) {
			return nil, refusal("graph_invalid_input")
		}
		out = append(out, base[offset:offset+n]...)
	}
	if len(out) != size {
		return nil, refusal("graph_invalid_input")
	}
	return out, nil
}
func (g gitDB) file(commit, rel string) ([]byte, error) {
	kind, raw, err := g.pathObject(commit, rel)
	if err != nil {
		return nil, err
	}
	if kind != "blob" {
		return nil, refusal("graph_invalid_input")
	}
	return raw, nil
}
func (g gitDB) pathObject(commit, rel string) (string, []byte, error) {
	if !cleanRelative(rel) || denied(rel) {
		return "", nil, refusal("graph_path_denied")
	}
	typ, raw, err := g.object(commit, 0)
	if err != nil {
		return "", nil, err
	}
	if typ != "commit" {
		return "", nil, refusal("graph_invalid_input")
	}
	first, _, _ := bytes.Cut(raw, []byte{'\n'})
	tree := strings.TrimPrefix(string(first), "tree ")
	width := len(commit) / 2
	if rel == "." {
		return g.object(tree, 0)
	}
	parts := strings.Split(rel, "/")
	if len(parts) > g.b.r.Limits.Depth {
		return "", nil, refusal("graph_limit_exceeded")
	}
	for i, component := range parts {
		typ, raw, err = g.object(tree, 0)
		if err != nil {
			return "", nil, err
		}
		if typ != "tree" {
			return "", nil, refusal("graph_invalid_input")
		}
		found := false
		for len(raw) > 0 {
			if err := g.b.r.inspect(); err != nil {
				return "", nil, err
			}
			entry, rest, ok := bytes.Cut(raw, []byte{0})
			if !ok || len(rest) < width {
				return "", nil, refusal("graph_invalid_input")
			}
			raw = rest[width:]
			mode, name, ok := strings.Cut(string(entry), " ")
			if !ok {
				return "", nil, refusal("graph_invalid_input")
			}
			if name != component {
				continue
			}
			if mode == "120000" {
				return "", nil, refusal("graph_symlink_escape")
			}
			tree = hex.EncodeToString(rest[:width])
			found = true
			break
		}
		if !found {
			return "", nil, refusal("graph_missing_input")
		}
		if i == len(parts)-1 {
			typ, raw, err = g.object(tree, 0)
			if err != nil {
				return "", nil, err
			}
			return typ, raw, nil
		}
	}
	return "", nil, refusal("graph_missing_input")
}
func (b *build) gitFile(source, commit, rel string) ([]byte, error) {
	return b.db(source).file(commit, rel)
}
func (b *build) gitTag(source, tag string) (string, error) {
	if !cleanRelative(tag) {
		return "", refusal("graph_invalid_input")
	}
	g := b.db(source)
	raw, err := g.read("refs/tags/" + tag)
	if isMissing(err) {
		packed, e := g.read("packed-refs")
		if e != nil {
			return "", e
		}
		for _, line := range strings.Split(string(packed), "\n") {
			fields := strings.Fields(line)
			if len(fields) == 2 && fields[1] == "refs/tags/"+tag {
				raw = []byte(fields[0])
				err = nil
				break
			}
		}
	}
	if err != nil {
		return "", err
	}
	id := strings.TrimSpace(string(raw))
	for depth := 0; depth <= b.r.Limits.Depth; depth++ {
		typ, obj, err := g.object(id, depth)
		if err != nil {
			return "", err
		}
		if typ == "commit" {
			return id, nil
		}
		if typ != "tag" {
			return "", refusal("graph_invalid_input")
		}
		line, _, _ := bytes.Cut(obj, []byte{'\n'})
		id = strings.TrimPrefix(string(line), "object ")
	}
	return "", refusal("graph_limit_exceeded")
}
func (b *build) gitCandidates(source string) ([]contextresolve.Candidate, error) {
	g := b.db(source)
	names := map[string]bool{}
	var walk func(string, int) error
	walk = func(rel string, depth int) error {
		if depth > b.r.Limits.Depth {
			return refusal("graph_limit_exceeded")
		}
		entries, err := b.r.directory(g.root, rel)
		if err != nil {
			if isMissing(err) {
				return nil
			}
			return err
		}
		for _, e := range entries {
			p := path.Join(rel, e.Name())
			if e.IsDir() {
				if err := walk(p, depth+1); err != nil {
					return err
				}
			} else {
				names[strings.TrimPrefix(p, "refs/tags/")] = true
			}
		}
		return nil
	}
	if err := walk("refs/tags", 0); err != nil {
		return nil, err
	}
	if raw, err := g.read("packed-refs"); err == nil {
		for _, line := range strings.Split(string(raw), "\n") {
			f := strings.Fields(line)
			if len(f) == 2 && strings.HasPrefix(f[1], "refs/tags/") {
				names[strings.TrimPrefix(f[1], "refs/tags/")] = true
			}
		}
	} else if !isMissing(err) {
		return nil, err
	}
	var out []contextresolve.Candidate
	for tag := range names {
		v, ok := pkgversion.ParseTag(tag)
		if !ok {
			continue
		}
		commit, err := b.gitTag(source, tag)
		if err != nil {
			return nil, err
		}
		out = append(out, contextresolve.Candidate{Tag: tag, Version: v, Commit: commit})
	}
	sort.Slice(out, func(i, j int) bool {
		c := pkgversion.Compare(out[i].Version, out[j].Version)
		if c != 0 {
			return c > 0
		}
		return out[i].Tag < out[j].Tag
	})
	return out, nil
}

func objectHash(raw []byte, width int) string {
	if width == 32 {
		sum := sha256.Sum256(raw)
		return hex.EncodeToString(sum[:])
	}
	sum := sha1.Sum(raw) // #nosec G401 -- Required Git SHA-1 object format.
	return hex.EncodeToString(sum[:])
}
func (g gitDB) treeAt(tree, rel string, width int) (string, error) {
	for _, component := range strings.Split(rel, "/") {
		kind, raw, err := g.object(tree, 0)
		if err != nil {
			return "", err
		}
		if kind != "tree" {
			return "", refusal("graph_invalid_input")
		}
		found := false
		for len(raw) > 0 {
			if err := g.b.r.inspect(); err != nil {
				return "", err
			}
			entry, rest, ok := bytes.Cut(raw, []byte{0})
			if !ok || len(rest) < width {
				return "", refusal("graph_invalid_input")
			}
			raw = rest[width:]
			mode, name, ok := strings.Cut(string(entry), " ")
			if !ok {
				return "", refusal("graph_invalid_input")
			}
			if name == component {
				if mode != "40000" {
					return "", refusal("graph_invalid_input")
				}
				tree = hex.EncodeToString(rest[:width])
				found = true
				break
			}
		}
		if !found {
			return "", refusal("graph_missing_input")
		}
	}
	return tree, nil
}
