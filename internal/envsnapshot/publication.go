package envsnapshot

import "bytes"

// Publication contains frozen output bytes from the capture coordinator. That
// coordinator owns source authorization and verified execute metadata; this
// boundary checks correspondence, never discovers or reads source paths.
type Publication struct {
	Manifest []byte
	Launch   []byte
	Files    map[string][]byte
}

func (p Publication) freeze() (Publication, map[string]bool, error) {
	p.Manifest = bytes.Clone(p.Manifest)
	p.Launch = bytes.Clone(p.Launch)
	files := make(map[string][]byte, len(p.Files))
	for name, data := range p.Files {
		files[name] = bytes.Clone(data)
	}
	p.Files = files
	m, err := DecodeManifest(p.Manifest)
	if err != nil {
		return Publication{}, nil, err
	}
	d, err := DecodeDescriptor(p.Launch)
	if err != nil {
		return Publication{}, nil, err
	}
	dirs, err := validateMetadata(m, d, p.Launch)
	if err != nil {
		return Publication{}, nil, err
	}
	if len(p.Files) != len(m.Files) {
		return Publication{}, nil, refuse("captured-entry-set")
	}
	for _, f := range m.Files {
		data, ok := p.Files[f.Path]
		if !ok || int64(len(data)) != f.Size || Digest(data) != f.SHA256 {
			return Publication{}, nil, refuse("captured-file-identity")
		}
	}
	return p, dirs, nil
}
