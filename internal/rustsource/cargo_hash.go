package rustsource

import (
	"encoding/binary"
	"encoding/hex"
	"math/bits"
)

type sip128State struct{ v0, v1, v2, v3 uint64 }

func (state *sip128State) round() {
	state.v0 += state.v1
	state.v2 += state.v3
	state.v1 = bits.RotateLeft64(state.v1, 13)
	state.v1 ^= state.v0
	state.v3 = bits.RotateLeft64(state.v3, 16)
	state.v3 ^= state.v2
	state.v0 = bits.RotateLeft64(state.v0, 32)
	state.v2 += state.v1
	state.v0 += state.v3
	state.v1 = bits.RotateLeft64(state.v1, 17)
	state.v1 ^= state.v2
	state.v3 = bits.RotateLeft64(state.v3, 21)
	state.v3 ^= state.v0
	state.v2 = bits.RotateLeft64(state.v2, 32)
}

// cargoShortHash ports Cargo 0.92's StableSipHasher128 short_hash for a URL
// string. Rust's Hash for str appends 0xff to make the byte stream prefix-free.
func cargoShortHash(value string) string {
	payload := append([]byte(value), 0xff)
	state := sip128State{v0: 0x736f6d6570736575, v1: 0x646f72616e646f6d ^ 0xee, v2: 0x6c7967656e657261, v3: 0x7465646279746573}
	for len(payload) >= 8 {
		word := binary.LittleEndian.Uint64(payload[:8])
		state.v3 ^= word
		state.round()
		state.v0 ^= word
		payload = payload[8:]
	}
	var tail [8]byte
	copy(tail[:], payload)
	word := binary.LittleEndian.Uint64(tail[:]) | uint64((len(value)+1)&0xff)<<56
	state.v3 ^= word
	state.round()
	state.v0 ^= word
	state.v2 ^= 0xee
	state.round()
	state.round()
	state.round()
	low := state.v0 ^ state.v1 ^ state.v2 ^ state.v3
	state.v1 ^= 0xdd
	state.round()
	state.round()
	state.round()
	high := state.v0 ^ state.v1 ^ state.v2 ^ state.v3
	combined := low*3 + high
	var encoded [8]byte
	binary.LittleEndian.PutUint64(encoded[:], combined)
	return hex.EncodeToString(encoded[:])
}
