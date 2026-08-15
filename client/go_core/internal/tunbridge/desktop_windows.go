//go:build windows

package tunbridge

import (
	"context"
	"fmt"
	"net/netip"
	"strings"

	"github.com/apernet/hysteria/core/v2/client"
	tun "github.com/sagernet/sing-tun"
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
// Caller should already have called protect.BindPhysicalUnderlay before the
// Hysteria handshake. StartDesktop re-asserts that bind and never picks StreamPass.
func StartDesktop(ctx context.Context, hyClient client.Client, mtu uint32, engine *decision.AtomicEngine, relayID string, opts Options) (*Session, error) {
	if mtu == 0 {
		mtu = 1400
	}
	if engine == nil {
		engine = decision.NewAtomicEngine(decision.NewEngine(nil, nil, decision.DefaultMode), 0)
	}

	if !protect.HasUnderlay() {
		ifIdx, ifName, err := protect.BindPhysicalUnderlay()
		if err != nil {
			return nil, fmt.Errorf("physical interface: %w", err)
		}
		if protect.IsTunnelInterface(ifName) {
			protect.Clear()
			return nil, fmt.Errorf("refusing underlay on tunnel iface %s", ifName)
		}
		logLine(fmt.Sprintf("[vpn] UNDERLAY_IF index=%d name=%q", ifIdx, ifName))
	} else {
		logLine("[vpn] UNDERLAY_REUSE session bind (skip second Get-NetRoute probe)")
	}

	// Do not attach sing-tun InterfaceMonitor on Windows desktop: Start() can
	// block forever waiting for a network-update event (VerifyWindowsTUN live hang).
	tunOptions := tun.Options{
		Name:                 windowsAdapterName,
		MTU:                  mtu,
		AutoRoute:            true,
		StrictRoute:          false,
		Inet4Address:         []netip.Prefix{TunIPv4Prefix()},
		DNSServers:           []netip.Addr{TunDNS()},
		Logger:               wintunLogger{},
		EXP_DisableDNSHijack: false,
	}

	logLine("[vpn] WINTUN_CREATE begin")
	sess, err := startStack(ctx, tunOptions, hyClient, engine, relayID, opts, stackHooks{
		AfterCreate: func() {
			logLine("[vpn] TUN_CREATED name=" + windowsAdapterName + " addr=" + TunIPv4Host() + "/30")
		},
		AfterRoute: func() {
			gw := TunIPv4Prefix().Addr().Next().String()
			protect.RegisterTunnelSessionGateway(gw)
			logLine("[vpn] ROUTES_APPLIED ipv4=0.0.0.0/0 via=" + windowsAdapterName + " gateway=" + gw)
			logLine("[vpn] DNS_READY server=" + TunDNS().String())
		},
		AfterStop: func() {
			if err := protect.ClearSessionTunnelRoutes(); err != nil {
				logLine("[vpn] session route cleanup: " + err.Error())
			}
			protect.Clear()
			logLine("[vpn] TUN_STOPPED")
		},
	})
	if err != nil {
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
