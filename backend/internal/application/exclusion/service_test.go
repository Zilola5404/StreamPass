package exclusion

import (
	"testing"

	domainexcl "streampass/backend/internal/domain/exclusion"
)

func TestNormalize_ok(t *testing.T) {
	got, err := Normalize([]string{" *.Bank.RU ", "mail.ru", "*.bank.ru"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "*.bank.ru" || got[1] != "mail.ru" {
		t.Fatalf("got %#v", got)
	}
}

func TestNormalize_invalid(t *testing.T) {
	if _, err := Normalize([]string{"not a domain"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestNormalize_tooMany(t *testing.T) {
	list := make([]string, domainexcl.MaxDomains+1)
	for i := range list {
		list[i] = "d" + itoa(i) + ".example.com"
	}
	if _, err := Normalize(list); err == nil {
		t.Fatal("expected too many error")
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
