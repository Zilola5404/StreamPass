package dnscache

import (
	"testing"

	"golang.org/x/net/dns/dnsmessage"
)

func TestRememberIP_HostForIP(t *testing.T) {
	RememberIP("Ya.Ru.", "77.88.8.8")
	if got := HostForIP("77.88.8.8"); got != "ya.ru" {
		t.Fatalf("HostForIP= %q want ya.ru", got)
	}
	RememberIP("youtube.com", "142.251.208.118")
	if got := HostForIP("142.251.208.118"); got != "youtube.com" {
		t.Fatalf("HostForIP= %q", got)
	}
}

func TestAnycastKeepsRelayHost(t *testing.T) {
	ip := "162.159.129.67"
	RememberIP("www.indeed.com", ip)
	RememberIP("cdn.example.com", ip) // anycast overwrite of "last"
	if got := HostForIP(ip); got != "cdn.example.com" {
		t.Fatalf("HostForIP last=%q", got)
	}
	hosts := HostsForIP(ip)
	found := false
	for _, h := range hosts {
		if h == "www.indeed.com" {
			found = true
		}
	}
	if !found {
		t.Fatalf("HostsForIP=%v missing www.indeed.com", hosts)
	}
}

func TestRelayHostNearIP_sameSlash24(t *testing.T) {
	PinRelayIP("www.upwork.com", "104.18.188.234")
	if got := RelayHostNearIP("104.18.188.200"); got != "www.upwork.com" {
		t.Fatalf("RelayHostNearIP=%q want www.upwork.com", got)
	}
	if got := RelayHostForIP("104.18.188.200"); got != "" {
		t.Fatalf("exact pin should miss sibling, got %q", got)
	}
}

func TestExtractAIPs_simpleA(t *testing.T) {
	msg := dnsmessage.Message{
		Header:    dnsmessage.Header{ID: 1, Response: true, RCode: dnsmessage.RCodeSuccess},
		Questions: []dnsmessage.Question{{Name: dnsmessage.MustNewName("2ip.ru."), Type: dnsmessage.TypeA, Class: dnsmessage.ClassINET}},
		Answers: []dnsmessage.Resource{{
			Header: dnsmessage.ResourceHeader{Name: dnsmessage.MustNewName("2ip.ru."), Type: dnsmessage.TypeA, Class: dnsmessage.ClassINET, TTL: 60},
			Body:   &dnsmessage.AResource{A: [4]byte{188, 40, 167, 82}},
		}},
	}
	raw, err := msg.Pack()
	if err != nil {
		t.Fatal(err)
	}
	ips := ExtractAIPs(raw)
	if len(ips) != 1 || ips[0] != "188.40.167.82" {
		t.Fatalf("ExtractAIPs=%v want [188.40.167.82]", ips)
	}
	IndexAnswers("2ip.ru", raw)
	PinDirectIP("2ip.ru", ips[0])
	if DirectHostNearIP("188.40.167.82") != "2ip.ru" {
		t.Fatalf("DirectHostNearIP=%q", DirectHostNearIP("188.40.167.82"))
	}
}

func TestExtractAIPs_additionalSection(t *testing.T) {
	msg := dnsmessage.Message{
		Header:    dnsmessage.Header{ID: 1, Response: true, RCode: dnsmessage.RCodeSuccess},
		Questions: []dnsmessage.Question{{Name: dnsmessage.MustNewName("2ip.ru."), Type: dnsmessage.TypeA, Class: dnsmessage.ClassINET}},
		Answers: []dnsmessage.Resource{{
			Header: dnsmessage.ResourceHeader{Name: dnsmessage.MustNewName("2ip.ru."), Type: dnsmessage.TypeCNAME, Class: dnsmessage.ClassINET, TTL: 60},
			Body:   &dnsmessage.CNAMEResource{CNAME: dnsmessage.MustNewName("2ip.ru.cdn.example.")},
		}},
		Additionals: []dnsmessage.Resource{{
			Header: dnsmessage.ResourceHeader{Name: dnsmessage.MustNewName("2ip.ru.cdn.example."), Type: dnsmessage.TypeA, Class: dnsmessage.ClassINET, TTL: 60},
			Body:   &dnsmessage.AResource{A: [4]byte{188, 40, 167, 82}},
		}},
	}
	raw, err := msg.Pack()
	if err != nil {
		t.Fatal(err)
	}
	ips := ExtractAIPs(raw)
	if len(ips) != 1 || ips[0] != "188.40.167.82" {
		t.Fatalf("ExtractAIPs=%v want [188.40.167.82]", ips)
	}
}

