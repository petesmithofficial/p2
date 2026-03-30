package powers

import (
	"math/big"
	"testing"
)

func TestByExponent(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		exponent uint
		want     Entry
	}{
		{
			exponent: 0,
			want: Entry{
				Exponent: 0,
				Value:    1,
			},
		},
		{
			exponent: 5,
			want: Entry{
				Exponent: 5,
				Value:    32,
			},
		},
		{
			exponent: 32,
			want: Entry{
				Exponent: 32,
				Value:    4294967296,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(RawUint(uint64(tc.exponent)), func(t *testing.T) {
			got, ok := ByExponent(tc.exponent)
			if !ok {
				t.Fatalf("ByExponent(%d) returned ok = false", tc.exponent)
			}

			if got != tc.want {
				t.Fatalf("ByExponent(%d) = %#v, want %#v", tc.exponent, got, tc.want)
			}
		})
	}
}

func TestByExponentOutOfRange(t *testing.T) {
	t.Parallel()

	if _, ok := ByExponent(33); ok {
		t.Fatal("ByExponent(33) returned ok = true, want false")
	}
}

func TestFormatUint(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		value     uint64
		useCommas bool
		want      string
	}{
		{value: 1, useCommas: true, want: "1"},
		{value: 1024, useCommas: true, want: "1,024"},
		{value: 32768, useCommas: true, want: "32,768"},
		{value: 4294967296, useCommas: true, want: "4,294,967,296"},
		{value: 32768, useCommas: false, want: "32768"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.want, func(t *testing.T) {
			if got := FormatUint(tc.value, tc.useCommas); got != tc.want {
				t.Fatalf("FormatUint(%d) = %q, want %q", tc.value, got, tc.want)
			}
		})
	}
}

func TestFormatEntries(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		items     []Entry
		useCommas bool
		want      string
	}{
		{
			name: "single",
			items: []Entry{
				{Exponent: 5, Value: 32},
			},
			useCommas: true,
			want:      "2^5 = 32",
		},
		{
			name: "aligned list",
			items: []Entry{
				{Exponent: 0, Value: 1},
				{Exponent: 10, Value: 1024},
				{Exponent: 16, Value: 65536},
			},
			useCommas: true,
			want:      "2^0  = 1\n2^10 = 1,024\n2^16 = 65,536",
		},
		{
			name: "no commas",
			items: []Entry{
				{Exponent: 15, Value: 32768},
			},
			useCommas: false,
			want:      "2^15 = 32768",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := FormatEntries(tc.items, tc.useCommas); got != tc.want {
				t.Fatalf("FormatEntries() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestBetween(t *testing.T) {
	t.Parallel()

	got := Between(5, 8)
	want := []Entry{
		{Exponent: 5, Value: 32},
		{Exponent: 6, Value: 64},
		{Exponent: 7, Value: 128},
		{Exponent: 8, Value: 256},
	}

	if len(got) != len(want) {
		t.Fatalf("Between() returned %d entries, want %d", len(got), len(want))
	}

	for idx := range got {
		if got[idx] != want[idx] {
			t.Fatalf("Between()[%d] = %#v, want %#v", idx, got[idx], want[idx])
		}
	}
}

func TestClosestTo(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		target string
		want   []Entry
	}{
		{
			name:   "33",
			target: "33",
			want: []Entry{
				{Exponent: 5, Value: 32},
			},
		},
		{
			name:   "30000",
			target: "30000",
			want: []Entry{
				{Exponent: 15, Value: 32768},
			},
		},
		{
			name:   "48",
			target: "48",
			want: []Entry{
				{Exponent: 5, Value: 32},
				{Exponent: 6, Value: 64},
			},
		},
		{
			name:   "above max",
			target: "5000000000",
			want: []Entry{
				{Exponent: 32, Value: 4294967296},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			target, ok := new(big.Int).SetString(tc.target, 10)
			if !ok {
				t.Fatalf("failed to parse target %q", tc.target)
			}

			got := ClosestTo(target)
			if len(got) != len(tc.want) {
				t.Fatalf("ClosestTo(%s) returned %d entries, want %d", tc.target, len(got), len(tc.want))
			}

			for idx := range got {
				if got[idx] != tc.want[idx] {
					t.Fatalf("ClosestTo(%s)[%d] = %#v, want %#v", tc.target, idx, got[idx], tc.want[idx])
				}
			}
		})
	}
}
