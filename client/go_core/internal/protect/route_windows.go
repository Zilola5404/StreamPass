//go:build windows

package protect

import (
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var (
	sessionMu       sync.Mutex
	sessionGateways []string
)

// RegisterTunnelSessionGateway records the IPv4 gateway installed for this VPN
// session (TUN Addr().Next()). Used to remove only our routes on stop/cleanup.
func RegisterTunnelSessionGateway(gateway string) {
	gw := strings.TrimSpace(gateway)
	if gw == "" {
		return
	}
	sessionMu.Lock()
	defer sessionMu.Unlock()
	for _, existing := range sessionGateways {
		if existing == gw {
			return
		}
	}
	sessionGateways = append(sessionGateways, gw)
}

// ClearSessionTunnelRoutes removes default routes recorded for the active or
// last StreamPass session. Best-effort; may require Administrator.
func ClearSessionTunnelRoutes() error {
	sessionMu.Lock()
	gws := append([]string(nil), sessionGateways...)
	sessionGateways = nil
	sessionMu.Unlock()

	var lastErr error
	for _, gw := range gws {
		if err := deleteDefaultRouteViaGateway(gw); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// ClearStaleTunnelDefaultRoute removes leftover 0.0.0.0/0 routes from previous
// StreamPass crashes. Prefers session-owned gateways, then scans route table for
// StreamPass TUN subnet gateways (10.10.0.x).
func clearStaleTunnelDefaultRouteImpl() error {
	_ = ClearSessionTunnelRoutes()

	var lastErr error
	removed := 0
	for _, gw := range findStaleStreamPassGateways() {
		if err := deleteDefaultRouteViaGateway(gw); err != nil {
			lastErr = err
			continue
		}
		removed++
	}
	if removed == 0 && lastErr != nil {
		return lastErr
	}
	return nil
}

func deleteDefaultRouteViaGateway(gateway string) error {
	gw := strings.TrimSpace(gateway)
	if gw == "" {
		return nil
	}
	cmd := exec.Command("route", "delete", "0.0.0.0", "mask", "0.0.0.0", gw)
	out, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(out))
	if err != nil {
		lower := strings.ToLower(text)
		if strings.Contains(lower, "not found") ||
			strings.Contains(lower, "element not found") ||
			strings.Contains(lower, "не найден") {
			return nil
		}
		return fmt.Errorf("route delete via %s: %w (%s)", gw, err, text)
	}
	return nil
}

var staleGatewayRE = regexp.MustCompile(`^\s*0\.0\.0\.0\s+0\.0\.0\.0\s+(10\.10\.0\.\d+)\s+`)

func findStaleStreamPassGateways() []string {
	out, err := exec.Command("route", "print", "-4").CombinedOutput()
	if err != nil {
		return []string{"10.10.0.2"}
	}
	seen := map[string]struct{}{}
	var gws []string
	for _, line := range strings.Split(string(out), "\n") {
		m := staleGatewayRE.FindStringSubmatch(line)
		if len(m) < 2 {
			continue
		}
		gw := m[1]
		if _, ok := seen[gw]; ok {
			continue
		}
		seen[gw] = struct{}{}
		gws = append(gws, gw)
	}
	if len(gws) == 0 {
		return []string{"10.10.0.2"}
	}
	return gws
}

// defaultRouteInterfaceIndex returns the Windows default-route NIC before TUN.
func defaultRouteInterfaceIndex() (index int, name string, ok bool) {
	out, err := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command",
		`$r = Get-NetRoute -DestinationPrefix '0.0.0.0/0' -ErrorAction SilentlyContinue | Sort-Object RouteMetric | Select-Object -First 1; if (-not $r) { exit 1 }; $alias = (Get-NetIPInterface -InterfaceIndex $r.InterfaceIndex -AddressFamily IPv4 -ErrorAction SilentlyContinue).InterfaceAlias; Write-Output ("{0}|{1}" -f $r.InterfaceIndex, $alias)`,
	).CombinedOutput()
	if err != nil {
		return 0, "", false
	}
	parts := strings.Split(strings.TrimSpace(string(out)), "|")
	if len(parts) < 2 {
		return 0, "", false
	}
	idx, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || idx <= 0 {
		return 0, "", false
	}
	name = strings.TrimSpace(parts[1])
	if name == "" {
		iface, err := net.InterfaceByIndex(idx)
		if err != nil {
			return 0, "", false
		}
		name = iface.Name
	}
	if IsTunnelInterface(name) {
		return 0, "", false
	}
	return idx, name, true
}
