//go:build windows

package tunbridge

import (
	"context"
	"fmt"
	"net/netip"
	"strings"

	"github.com/apernet/hysteria/core/v2/client"
	tun "github.com/sagernet/sing-tun"
	"github.com/sagernet/sing/common/control"
	"github.com/sagernet/sing/common/logger"

	"streampass/go_core/internal/decision"
	"streampass/go_core/internal/protect"
)

const windowsAdapterName = "StreamPass"

type wintunLogger struct{}

func (wintunLogger) Trace(args ...any) {}
func (wintunLogger) Debug(args ...any) {}
func (wintunLogger) Info(args ...any)  { logLine("[wintun] " + fmt.Sprint(args...)) }
func (wintunLogger) Warn(args ...any)  { logLine("[wintun] warn " + fmt.Sprint(args...)) }
func (wintunLogger) Error(args ...any) { logLine("[wintun] error " + fmt.Sprint(args...)) }
func (wintunLogger) Fatal(args ...any) { logLine("[wintun] fatal " + fmt.Sprint(args...)) }
func (wintunLogger) Panic(args ...any) { logLine("[wintun] panic " + fmt.Sprint(args...)) }

var _ logger.Logger = wintunLogger{}

// StartDesktop creates a Wintun adapter (no Android fd) and installs IPv4
// default routes via AutoRoute. IPv6 is not captured (Variant B).
//
// Caller must BindInterface to the physical NIC *before* this returns routes.
func StartDesktop(ctx context.Context, hyClient client.Client, mtu uint32, engine *decision.AtomicEngine, relayID string, opts Options) (*Session, error) {
	if mtu == 0 {
		mtu = 1400
	}
	if engine == nil {
		engine = decision.NewAtomicEngine(decision.NewEngine(nil, nil, decision.DefaultMode), 0)
	}

	ifIdx, ifName, err := protect.PhysicalInterfaceIndex()
	if err != nil {
		return nil, fmt.Errorf("physical interface: %w", err)
	}
	protect.BindInterface(ifIdx)
	logLine(fmt.Sprintf("[vpn] UNDERLAY_IF index=%d name=%s", ifIdx, ifName))

	finder := control.NewDefaultInterfaceFinder()
	netMon, err := tun.NewNetworkUpdateMonitor(wintunLogger{})
	if err != nil {
		protect.Clear()
		return nil, fmt.Errorf("network monitor: %w", err)
	}
	ifMon, err := tun.NewDefaultInterfaceMonitor(netMon, wintunLogger{}, tun.DefaultInterfaceMonitorOptions{
		InterfaceFinder: finder,
	})
	if err != nil {
		protect.Clear()
		return nil, fmt.Errorf("interface monitor: %w", err)
	}
	if err := netMon.Start(); err != nil {
		protect.Clear()
		return nil, fmt.Errorf("network monitor start: %w", err)
	}
	if err := ifMon.Start(); err != nil {
		_ = netMon.Close()
		protect.Clear()
		return nil, fmt.Errorf("interface monitor start: %w", err)
	}

	tunOptions := tun.Options{
		Name:                 windowsAdapterName,
		MTU:                  mtu,
		AutoRoute:            true,
		StrictRoute:          false,
		Inet4Address:         []netip.Prefix{TunIPv4Prefix()},
		DNSServers:           []netip.Addr{TunDNS()},
		InterfaceMonitor:     ifMon,
		InterfaceFinder:      finder,
		Logger:               wintunLogger{},
		EXP_DisableDNSHijack: false,
	}

	sess, err := startStack(ctx, tunOptions, hyClient, engine, relayID, opts, stackHooks{
		AfterCreate: func() {
			logLine("[vpn] TUN_CREATED name=" + windowsAdapterName + " addr=" + TunIPv4Host() + "/30")
		},
		AfterRoute: func() {
			logLine("[vpn] ROUTES_APPLIED ipv4=0.0.0.0/0 via=" + windowsAdapterName)
			logLine("[vpn] DNS_READY server=" + TunDNS().String())
		},
		AfterStop: func() {
			_ = ifMon.Close()
			_ = netMon.Close()
			protect.Clear()
			logLine("[vpn] TUN_STOPPED")
		},
	})
	if err != nil {
		_ = ifMon.Close()
		_ = netMon.Close()
		protect.Clear()
		return nil, wrapWintunErr(err)
	}
	return sess, nil
}

func wrapWintunErr(err error) error {
	if err == nil {
		return nil
	}
	s := strings.ToLower(err.Error())
	if strings.Contains(s, "access is denied") ||
		strings.Contains(s, "elevation") ||
		strings.Contains(s, "privileg") ||
		strings.Contains(s, "required privilege") {
		return fmt.Errorf("нужны права администратора для Wintun: %w", err)
	}
	if strings.Contains(s, "wintun.dll") || strings.Contains(s, "cannot find") {
		return fmt.Errorf("положите wintun.dll рядом со streampasscore.exe: %w", err)
	}
	return err
}
