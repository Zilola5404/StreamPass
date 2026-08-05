package hyconfig

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/apernet/hysteria/core/v2/client"
	"github.com/apernet/hysteria/extras/v2/obfs"

	"streampass/go_core/internal/protect"
)

const defaultMTU = 1400

// DefaultMTU is the fallback TUN MTU for mobile tunnels.
func DefaultMTU() uint32 { return defaultMTU }

// Parsed holds hysteria client settings derived from a relay connection string.
type Parsed struct {
	ServerHost string
	Auth       string
	SNI        string
	ObfsType   string
	ObfsPass   string
	Insecure   bool
	PinSHA256  string
	MTU        uint32
}

// BuildClientConfig resolves relay parameters and builds a hysteria client config.
func BuildClientConfig(connectionConfig, relayHost string, relayPort int) (*client.Config, *Parsed, error) {
	parsed, err := Parse(connectionConfig, relayHost, relayPort)
	if err != nil {
		return nil, nil, err
	}

	serverAddr, err := net.ResolveUDPAddr("udp", parsed.ServerHost)
	if err != nil {
		return nil, nil, fmt.Errorf("resolve server: %w", err)
	}

	tlsConfig := client.TLSConfig{
		InsecureSkipVerify: parsed.Insecure,
	}
	if parsed.SNI != "" {
		tlsConfig.ServerName = parsed.SNI
	} else if host, _, err := net.SplitHostPort(parsed.ServerHost); err == nil && host != "" {
		tlsConfig.ServerName = host
	}

	if parsed.PinSHA256 != "" {
		pin := normalizeCertHash(parsed.PinSHA256)
		tlsConfig.VerifyPeerCertificate = func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
			if len(rawCerts) == 0 {
				return fmt.Errorf("no certificate presented")
			}
			sum := sha256.Sum256(rawCerts[0])
			if hex.EncodeToString(sum[:]) != pin {
				return fmt.Errorf("certificate pin mismatch")
			}
			return nil
		}
	}

	cfg := &client.Config{
		ServerAddr: serverAddr,
		Auth:       parsed.Auth,
		TLSConfig:  tlsConfig,
		// Always use a ConnFactory so the QUIC UDP underlay can be protected
		// via VpnService.protect after TUN default route is installed.
		ConnFactory: &protectedConnFactory{
			obfsType: parsed.ObfsType,
			obfsPass: parsed.ObfsPass,
		},
	}
	return cfg, parsed, nil
}

// Parse reads a hysteria2 URI or falls back to relay host/port fields.
func Parse(connectionConfig, relayHost string, relayPort int) (*Parsed, error) {
	connectionConfig = strings.TrimSpace(connectionConfig)
	if connectionConfig == "" {
		if relayHost == "" {
			return nil, fmt.Errorf("connection config and relay host are empty")
		}
		if relayPort <= 0 {
			relayPort = 443
		}
		return &Parsed{
			ServerHost: net.JoinHostPort(relayHost, strconv.Itoa(relayPort)),
			Insecure:   true,
			MTU:        defaultMTU,
		}, nil
	}

	u, err := url.Parse(connectionConfig)
	if err != nil {
		return nil, fmt.Errorf("parse connection config: %w", err)
	}
	if u.Scheme != "hysteria2" && u.Scheme != "hy2" {
		return nil, fmt.Errorf("unsupported scheme %q", u.Scheme)
	}

	parsed := &Parsed{
		ServerHost: u.Host,
		MTU:        defaultMTU,
	}
	if parsed.ServerHost == "" {
		return nil, fmt.Errorf("connection config missing server host")
	}

	if u.User != nil {
		username := u.User.Username()
		password, hasPassword := u.User.Password()
		if hasPassword {
			parsed.Auth = username + ":" + password
		} else {
			parsed.Auth = username
		}
	}

	q := u.Query()
	if obfsType := q.Get("obfs"); obfsType != "" {
		parsed.ObfsType = strings.ToLower(obfsType)
		parsed.ObfsPass = q.Get("obfs-password")
	}
	if sni := q.Get("sni"); sni != "" {
		parsed.SNI = sni
	}
	if insecure, err := strconv.ParseBool(q.Get("insecure")); err == nil {
		parsed.Insecure = insecure
	} else {
		// Relay certificates are self-signed in MVP deployments.
		parsed.Insecure = q.Get("pinSHA256") == ""
	}
	if pin := q.Get("pinSHA256"); pin != "" {
		parsed.PinSHA256 = pin
		parsed.Insecure = false
	}
	return parsed, nil
}

type protectedConnFactory struct {
	obfsType string
	obfsPass string
}

func (f *protectedConnFactory) New(addr net.Addr) (net.PacketConn, error) {
	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		return nil, err
	}
	if err := protect.Conn(conn); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("protect underlay udp: %w", err)
	}
	if f.obfsType == "" {
		return conn, nil
	}
	switch f.obfsType {
	case "salamander":
		if f.obfsPass == "" {
			_ = conn.Close()
			return nil, fmt.Errorf("salamander obfs requires obfs-password")
		}
		ob, err := obfs.NewSalamanderObfuscator([]byte(f.obfsPass))
		if err != nil {
			_ = conn.Close()
			return nil, err
		}
		return obfs.WrapPacketConn(conn, ob), nil
	default:
		_ = conn.Close()
		return nil, fmt.Errorf("unsupported obfs type %q", f.obfsType)
	}
}

func wrapObfs(conn net.PacketConn, obfsType, obfsPass string) (net.PacketConn, error) {
	switch obfsType {
	case "", "none":
		return conn, nil
	case "salamander":
		if obfsPass == "" {
			_ = conn.Close()
			return nil, fmt.Errorf("salamander obfs requires obfs-password")
		}
		ob, err := obfs.NewSalamanderObfuscator([]byte(obfsPass))
		if err != nil {
			_ = conn.Close()
			return nil, err
		}
		return obfs.WrapPacketConn(conn, ob), nil
	default:
		_ = conn.Close()
		return nil, fmt.Errorf("unsupported obfs type %q", obfsType)
	}
}

func normalizeCertHash(value string) string {
	value = strings.TrimSpace(value)
	if decoded, err := base64.StdEncoding.DecodeString(value); err == nil {
		return hex.EncodeToString(decoded)
	}
	return strings.ToLower(strings.ReplaceAll(value, ":", ""))
}
