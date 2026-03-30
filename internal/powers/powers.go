package powers

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

const MaxExponent int64 = 32

type Entry struct {
	Exponent uint
	Value    uint64
}

var (
	entries    = buildEntries()
	entryBigs  = buildEntryBigs(entries)
	maxEntry   = entries[len(entries)-1]
	maxEntryBI = entryBigs[len(entryBigs)-1]
)

func All() []Entry {
	out := make([]Entry, len(entries))
	copy(out, entries)
	return out
}

func Between(lower, upper int) []Entry {
	if lower < 0 {
		lower = 0
	}
	if upper > int(MaxExponent) {
		upper = int(MaxExponent)
	}
	if lower > upper {
		return nil
	}

	out := make([]Entry, upper-lower+1)
	copy(out, entries[lower:upper+1])
	return out
}

func ByExponent(exp uint) (Entry, bool) {
	if exp > uint(MaxExponent) {
		return Entry{}, false
	}

	return entries[exp], true
}

func ClosestTo(target *big.Int) []Entry {
	if target == nil || target.Sign() <= 0 {
		return []Entry{entries[0]}
	}

	if target.Cmp(maxEntryBI) >= 0 {
		return []Entry{maxEntry}
	}

	for idx := 1; idx < len(entryBigs); idx++ {
		current := entryBigs[idx]
		if target.Cmp(current) > 0 {
			continue
		}

		lower := entries[idx-1]
		upper := entries[idx]

		lowerDistance := new(big.Int).Sub(target, entryBigs[idx-1])
		upperDistance := new(big.Int).Sub(current, target)

		switch lowerDistance.Cmp(upperDistance) {
		case 0:
			return []Entry{lower, upper}
		case -1:
			return []Entry{lower}
		default:
			return []Entry{upper}
		}
	}

	return []Entry{maxEntry}
}

func FormatEntries(items []Entry, useCommas bool) string {
	if len(items) == 0 {
		return ""
	}

	width := exponentWidth(items)
	parts := make([]string, 0, len(items))
	for _, item := range items {
		parts = append(parts, FormatEntry(item, useCommas, width))
	}

	return strings.Join(parts, "\n")
}

func FormatEntry(item Entry, useCommas bool, exponentWidth int) string {
	return fmt.Sprintf("2^%-*d = %s", exponentWidth, item.Exponent, FormatUint(item.Value, useCommas))
}

func FormatUint(value uint64, useCommas bool) string {
	digits := strconv.FormatUint(value, 10)
	if !useCommas || len(digits) <= 3 {
		return digits
	}

	return formatUintWithCommas(digits)
}

func RawUint(value uint64) string {
	return strconv.FormatUint(value, 10)
}

func exponentWidth(items []Entry) int {
	width := 1
	for _, item := range items {
		length := len(strconv.FormatUint(uint64(item.Exponent), 10))
		if length > width {
			width = length
		}
	}

	return width
}

func formatUintWithCommas(digits string) string {
	if len(digits) <= 3 {
		return digits
	}

	firstGroup := len(digits) % 3
	if firstGroup == 0 {
		firstGroup = 3
	}

	var builder strings.Builder
	builder.Grow(len(digits) + (len(digits)-1)/3)
	builder.WriteString(digits[:firstGroup])

	for idx := firstGroup; idx < len(digits); idx += 3 {
		builder.WriteByte(',')
		builder.WriteString(digits[idx : idx+3])
	}

	return builder.String()
}

func buildEntries() []Entry {
	items := make([]Entry, 0, MaxExponent+1)
	value := uint64(1)

	for exponent := int64(0); exponent <= MaxExponent; exponent++ {
		items = append(items, Entry{
			Exponent: uint(exponent),
			Value:    value,
		})
		value <<= 1
	}

	return items
}

func buildEntryBigs(items []Entry) []*big.Int {
	result := make([]*big.Int, 0, len(items))
	for _, item := range items {
		result = append(result, new(big.Int).SetUint64(item.Value))
	}
	return result
}
