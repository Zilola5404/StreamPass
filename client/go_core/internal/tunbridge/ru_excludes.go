package tunbridge

import (
	"bufio"
	"bytes"
	_ "embed"
	"net/netip"
	"strings"
	"sync"
)

//go:embed data/ru_ipv4_cidrs.txt
var ruCIDRFile []byte

var (
	ruCIDRsOnce sync.Once
	ruCIDRs     []netip.Prefix
	ruCIDRErr   error
)

// LoadRUExcludes returns RU IPv4 prefixes excluded from the TUN default route
// (Android excludeRoute parity on Windows via Inet4RouteExcludeAddress).
func LoadRUExcludes() ([]netip.Prefix, error) {
	ruCIDRsOnce.Do(func() {
		ruCIDRs, ruCIDRErr = parseCIDRList(ruCIDRFile)
	})
	return ruCIDRs, ruCIDRErr
}

func parseCIDRList(raw []byte) ([]netip.Prefix, error) {
	sc := bufio.NewScanner(bytes.NewReader(raw))
	out := make([]netip.Prefix, 0, 12000)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		p, err := netip.ParsePrefix(line)
		if err != nil {
			continue
		}
		if !p.Addr().Is4() {
			continue
		}
		out = append(out, p)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
