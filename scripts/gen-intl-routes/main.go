// Generate intl_ipv4_routes.txt = 0.0.0.0/0 minus Russian CIDRs (for Android API < 33).
// Usage: go run ./scripts/gen-intl-routes
package main

import (
	"bufio"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"sort"
)

func main() {
	root := findRoot()
	ruFile := filepath.Join(root, "client", "android", "app", "src", "main", "assets", "ru_ipv4_cidrs.txt")
	outFile := filepath.Join(root, "client", "android", "app", "src", "main", "assets", "intl_ipv4_routes.txt")

	excl, err := loadPrefixes(ruFile)
	if err != nil {
		panic(err)
	}
	sortPrefixes(excl)

	universe := netip.MustParsePrefix("0.0.0.0/0")
	routes := subtractAll(universe, excl)

	f, err := os.Create(outFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, p := range routes {
		fmt.Fprintf(w, "%s\n", p.String())
	}
	w.Flush()

	fmt.Printf("ru excludes: %d\n", len(excl))
	fmt.Printf("intl routes: %d -> %s\n", len(routes), outFile)
}

func findRoot() string {
	if _, err := os.Stat("client/android"); err == nil {
		return "."
	}
	if _, err := os.Stat("../client/android"); err == nil {
		return ".."
	}
	return "../.."
}

func loadPrefixes(path string) ([]netip.Prefix, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []netip.Prefix
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		if line == "" {
			continue
		}
		p, err := netip.ParsePrefix(line)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", line, err)
		}
		out = append(out, p.Masked())
	}
	return out, s.Err()
}

func sortPrefixes(in []netip.Prefix) {
	sort.Slice(in, func(i, j int) bool {
		if in[i].Addr().Compare(in[j].Addr()) != 0 {
			return in[i].Addr().Less(in[j].Addr())
		}
		return in[i].Bits() < in[j].Bits()
	})
}

func subtractAll(base netip.Prefix, excludes []netip.Prefix) []netip.Prefix {
	cur := []netip.Prefix{base.Masked()}
	for _, ex := range excludes {
		var next []netip.Prefix
		for _, p := range cur {
			next = append(next, subtractOne(p, ex)...)
		}
		cur = next
	}
	return cur
}

func subtractOne(base, ex netip.Prefix) []netip.Prefix {
	if !base.Overlaps(ex) {
		return []netip.Prefix{base}
	}
	if ex.Bits() <= base.Bits() {
		if ex.Contains(base.Addr()) {
			return nil
		}
	}
	queue := []netip.Prefix{base}
	var kept []netip.Prefix
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		if !p.Overlaps(ex) {
			kept = append(kept, p)
			continue
		}
		if ex.Contains(p.Addr()) && ex.Bits() <= p.Bits() {
			if p.Bits() >= 32 {
				continue
			}
		}
		if p.Bits() >= 32 {
			if !ex.Contains(p.Addr()) {
				kept = append(kept, p)
			}
			continue
		}
		left, right, ok := splitPrefix(p)
		if !ok {
			kept = append(kept, p)
			continue
		}
		queue = append(queue, left, right)
	}
	return kept
}

func splitPrefix(p netip.Prefix) (left, right netip.Prefix, ok bool) {
	if p.Bits() >= 32 {
		return netip.Prefix{}, netip.Prefix{}, false
	}
	bits := p.Bits() + 1
	base := p.Addr()
	// next network = base + 2^(32-bits)
	var add [4]byte
	a := base.As4()
	copy(add[:], a[:])
	inc := uint32(1) << (32 - bits)
	v := uint32(a[0])<<24 | uint32(a[1])<<16 | uint32(a[2])<<8 | uint32(a[3])
	v += inc
	next := netip.AddrFrom4([4]byte{byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)})
	left = netip.PrefixFrom(base, bits)
	right = netip.PrefixFrom(next, bits)
	return left.Masked(), right.Masked(), true
}
