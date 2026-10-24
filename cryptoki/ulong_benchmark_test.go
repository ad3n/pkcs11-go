package cryptoki

import "testing"

var (
	ulongBenchmarkBytes     []byte
	ulongBenchmarkAttribute *Attribute
)

func BenchmarkULongToBytes(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		ulongBenchmarkBytes = ULongToBytes(0x11223344)
	}
}

func BenchmarkNewIntegerAttribute(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		ulongBenchmarkAttribute = NewAttribute(CKA_KEY_TYPE, uint(CKK_AES))
	}
}

func TestULongOwnedAllocation(t *testing.T) {
	first := ULongToBytes(0x11223344)
	second := ULongToBytes(0x11223344)
	first[0] ^= 0xff
	if BytesToULong(second) != 0x11223344 {
		t.Fatal("returned buffers alias")
	}

	if n := testing.AllocsPerRun(1000, func() { ulongBenchmarkBytes = ULongToBytes(0x11223344) }); n != 1 {
		t.Fatalf("encoding allocations = %v; want 1", n)
	}
}

func TestULongDecodeBoundaries(t *testing.T) {
	for _, size := range []int{0, 1, ULongSize - 1, ULongSize, ULongSize + 1, 4096} {
		input := make([]byte, size)
		for i := range input {
			input[i] = 0xff
		}

		width := min(size, ULongSize)
		var want uint
		for range width {
			want = want<<8 | 0xff
		}

		if got := BytesToULong(input); got != want {
			t.Fatalf("length %d: got %x, want %x", size, got, want)
		}

		for _, value := range input {
			if value != 0xff {
				t.Fatal("decoder mutated input")
			}
		}
	}
}
