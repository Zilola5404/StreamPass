package dnscache

import (
	"testing"
	"time"
)

func TestCache_putGetRaw(t *testing.T) {
	c := New(8)
	c.PutRaw(1, "example.com.", []byte{0xAB, 0xCD, 0x01}, time.Minute)
	got, ok := c.GetRaw(1, "example.com.")
	if !ok || len(got) != 3 || got[0] != 0xAB {
		t.Fatalf("got=%v ok=%v", got, ok)
	}
}

func TestCache_miss(t *testing.T) {
	c := New(8)
	if _, ok := c.GetRaw(1, "missing.com."); ok {
		t.Fatal("expected miss")
	}
}

func TestCache_typeIsolation(t *testing.T) {
	c := New(8)
	c.PutRaw(1, "a.com.", []byte{1}, time.Minute)
	c.PutRaw(28, "a.com.", []byte{28}, time.Minute)
	a, _ := c.GetRaw(1, "a.com.")
	aaaa, _ := c.GetRaw(28, "a.com.")
	if a[0] != 1 || aaaa[0] != 28 {
		t.Fatalf("a=%v aaaa=%v", a, aaaa)
	}
}
