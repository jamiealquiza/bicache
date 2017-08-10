// Package fnv implements allocation-free
// 32 and 64 bit FNV-1 hash variants.
package fnv

const (
	offset32 = 0x811c9dc5
	prime32  = 0x1000193
	offset64 = 0xcbf29ce484222325
	prime64  = 0x100000001b3
)

// Hash32a takes a string and
// returns a 32 bit FNV-1a.
func Hash32a(s string) uint32 {
	var h uint32 = offset32
	for _, c := range []byte(s) {
		h ^= uint32(c)
		h *= prime32
	}

	return h
}

// Hash32 takes a string and
// returns a 32 bit FNV-1.
func Hash32(s string) uint32 {
	var h uint32 = offset32
	for _, c := range []byte(s) {
		h *= prime32
		h ^= uint32(c)
	}

	return h
}

// Hash64a takes a string and
// returns a 64 bit FNV-1a.
func Hash64a(s string) uint64 {
	var h uint64 = offset64
	for _, c := range []byte(s) {
		h ^= uint64(c)
		h *= prime64
	}

	return h
}

// Hash64 takes a string and
// returns a 64 bit FNV-1.
func Hash64(s string) uint64 {
	var h uint64 = offset64
	for _, c := range []byte(s) {
		h *= prime64
		h ^= uint64(c)
	}

	return h
}
