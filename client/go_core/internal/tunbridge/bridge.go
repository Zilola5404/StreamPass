package tunbridge

import (
	"context"
	"net"
	"net/netip"
	"sync"
	"time"

	"github.com/apernet/hysteria/core/v2/client"
	"github.com/sagernet/sing/common/buf"
	"github.com/sagernet/sing/common/bufio"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	tun "github.com/sagernet/sing-tun"
)

type Session struct {
	cancel context.CancelFunc
	stop   func()
}

func (s *Session) Close() {
	if s == nil {
		return
	}
	if s.cancel != nil {
		s.cancel()
	}
	if s.stop != nil {
		s.stop()
	}
}

func Start(ctx context.Context, fd int, hyClient client.Client, mtu uint32) (*Session, error) {
	if mtu == 0 {
		mtu = 1400
	}

	tunOptions := tun.Options{
		FileDescriptor: fd,
		MTU:            mtu,
		AutoRoute:      false,
		Inet4Address: []netip.Prefix{
			netip.PrefixFrom(netip.MustParseAddr("10.10.0.2"), 32),
		},
	}

	tunDev, err := tun.New(tunOptions)
	if err != nil {
		return nil, err
	}

	runCtx, cancel := context.WithCancel(ctx)
	stack, err := tun.NewStack("system", tun.StackOptions{
		Context:    runCtx,
		Tun:        tunDev,
		TunOptions: tunOptions,
		Handler:    &hyHandler{client: hyClient},
		UDPTimeout: 5 * time.Minute,
	})
	if err != nil {
		cancel()
		_ = tunDev.Close()
		return nil, err
	}

	if err := tunDev.Start(); err != nil {
		cancel()
		_ = stack.Close()
		_ = tunDev.Close()
		return nil, err
	}
	if err := stack.Start(); err != nil {
		cancel()
		_ = stack.Close()
		_ = tunDev.Close()
		return nil, err
	}

	stop := func() {
		_ = stack.Close()
		_ = tunDev.Close()
	}
	return &Session{cancel: cancel, stop: stop}, nil
}

type hyHandler struct {
	client client.Client
}

func (h *hyHandler) PrepareConnection(
	network string,
	source M.Socksaddr,
	destination M.Socksaddr,
	routeContext tun.DirectRouteContext,
	timeout time.Duration,
) (tun.DirectRouteDestination, error) {
	return nil, nil
}

func (h *hyHandler) NewConnectionEx(
	ctx context.Context,
	conn net.Conn,
	source M.Socksaddr,
	destination M.Socksaddr,
	onClose N.CloseHandlerFunc,
) {
	defer conn.Close()
	if onClose != nil {
		defer onClose(nil)
	}

	remote, err := h.client.TCP(destination.String())
	if err != nil {
		return
	}
	defer remote.Close()
	_ = bufio.CopyConn(ctx, conn, remote)
}

func (h *hyHandler) NewPacketConnectionEx(
	ctx context.Context,
	conn N.PacketConn,
	source M.Socksaddr,
	destination M.Socksaddr,
	onClose N.CloseHandlerFunc,
) {
	defer conn.Close()
	if onClose != nil {
		defer onClose(nil)
	}

	hyUDP, err := h.client.UDP()
	if err != nil {
		return
	}

	destAddr := destination.String()
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		readBuffer := buf.NewPacket()
		defer readBuffer.Release()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			_, err := conn.ReadPacket(readBuffer)
			if err != nil {
				return
			}
			payload := append([]byte(nil), readBuffer.Bytes()...)
			readBuffer.Reset()
			if err := hyUDP.Send(payload, destAddr); err != nil {
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			data, addr, err := hyUDP.Receive()
			if err != nil {
				return
			}
			if addr != destAddr {
				continue
			}
			writeBuffer := buf.As(data)
			if err := conn.WritePacket(writeBuffer, destination); err != nil {
				writeBuffer.Release()
				return
			}
			writeBuffer.Release()
		}
	}()

	wg.Wait()
	_ = hyUDP.Close()
}
