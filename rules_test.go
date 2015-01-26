package yara

import (
	"fmt"
	"testing"
)

func makeRules(t *testing.T, rule string) *Rules {
	c, err := NewCompiler()
	if c == nil || err != nil {
		t.Fatal("NewCompiler():", err)
	}
	if err = c.AddString(rule, ""); err != nil {
		t.Fatal("AddString():", err)
	}
	r, err := c.GetRules()
	if err != nil {
		t.Fatal("GetRules:", err)
	}
	return r
}

func TestSimpleMatch(t *testing.T) {
	r := makeRules(t,
		"rule test : tag1 { meta: author = \"Hilko Bengen\" strings: $a = \"abc\" fullword condition: $a }")
	m, err := r.ScanMem([]byte(" abc "), 0, 0)
	if err != nil {
		t.Errorf("ScanMem: %s", err)
	}
	t.Logf("Matches: %+v", m)
}

func assertTrueRules(t *testing.T, rules []string, data []byte) {
	for _, rule := range rules {
		r := makeRules(t, rule)
		if m, err := r.ScanMem(data, 0, 0); len(m) == 0 {
			t.Errorf("Rule < %s > did not match data < %v >", rule, data)
		} else if err != nil {
			t.Errorf("Error %s", err)
		}
	}
}

func assertFalseRules(t *testing.T, rules []string, data []byte) {
	for _, rule := range rules {
		r := makeRules(t, rule)
		if m, err := r.ScanMem(data, 0, 0); len(m) > 0 {
			t.Errorf("Rule < %s > matched data < %v >", rule, data)
		} else if err != nil {
			t.Errorf("Error %s", err)
		}
	}
}

func TestRe(t *testing.T) {
}

func TestBooleanOperators(t *testing.T) {
	assertTrueRules(t, []string{
		"rule test { condition: true }",
		"rule test { condition: true or false }",
		"rule test { condition: true and true }",
		"rule test { condition: 0x1 and 0x2}",
	}, []byte("dummy"))

	assertFalseRules(t, []string{
		"rule test { condition: false }",
		"rule test { condition: true and false }",
		"rule test { condition: false or false }",
	}, []byte("dummy"))
}

func TestComparisonOperators(t *testing.T) {
	assertTrueRules(t, []string{
		"rule test { condition: 2 > 1 }",
		"rule test { condition: 1 < 2 }",
		"rule test { condition: 2 >= 1 }",
		"rule test { condition: 1 <= 1 }",
		"rule test { condition: 1 == 1 }",
	}, []byte("dummy"))
	assertFalseRules(t, []string{
		"rule test { condition: 1 != 1}",
		"rule test { condition: 2 > 3}",
	}, []byte("dummy"))
}

func TestArithmeticOperators(t *testing.T) {
	assertTrueRules(t, []string{
		"rule test { condition: (1 + 1) * 2 == (9 - 1) \\ 2 }",
		"rule test { condition: 5 % 2 == 1 }",
	}, []byte("dummy"))
}

func TestBitwiseOperators(t *testing.T) {
	assertTrueRules(t, []string{
		"rule test { condition: 0x55 | 0xAA == 0xFF }",
		"rule test { condition: ~0xAA ^ 0x5A & 0xFF == 0x0F }",
		"rule test { condition: ~0x55 & 0xFF == 0xAA }",
		"rule test { condition: 8 >> 2 == 2 }",
		"rule test { condition: 1 << 3 == 8 }",
	}, []byte("dummy"))
}

func TestStrings(t *testing.T) {
	assertTrueRules(t, []string{
		"rule test { strings: $a = \"a\" condition: $a }",
		"rule test { strings: $a = \"ab\" condition: $a }",
		"rule test { strings: $a = \"abc\" condition: $a }",
		"rule test { strings: $a = \"xyz\" condition: $a }",
		"rule test { strings: $a = \"abc\" nocase fullword condition: $a }",
		"rule test { strings: $a = \"aBc\" nocase  condition: $a }",
		"rule test { strings: $a = \"abc\" fullword condition: $a }",
	}, []byte("---- abc ---- xyz"))
	assertFalseRules(t, []string{
		"rule test { strings: $a = \"a\" fullword condition: $a }",
		"rule test { strings: $a = \"ab\" fullword condition: $a }",
		"rule test { strings: $a = \"abc\" wide fullword condition: $a }",
	}, []byte("---- abc ---- xyz"))
	assertTrueRules(t, []string{
		"rule test { strings: $a = \"a\" wide condition: $a }",
		"rule test { strings: $a = \"a\" wide ascii condition: $a }",
		"rule test { strings: $a = \"ab\" wide condition: $a }",
		"rule test { strings: $a = \"ab\" wide ascii condition: $a }",
		"rule test { strings: $a = \"abc\" wide condition: $a }",
		"rule test { strings: $a = \"abc\" wide nocase fullword condition: $a }",
		"rule test { strings: $a = \"aBc\" wide nocase condition: $a }",
		"rule test { strings: $a = \"aBc\" wide ascii nocase condition: $a }",
		"rule test { strings: $a = \"---xyz\" wide nocase condition: $a }",
	}, []byte("---- a\x00b\x00c\x00 -\x00-\x00-\x00-\x00x\x00y\x00z\x00"))
	assertTrueRules(t, []string{
		"rule test { strings: $a = \"abc\" fullword condition: $a }",
	}, []byte("abc"))
	assertFalseRules(t, []string{
		"rule test { strings: $a = \"abc\" fullword condition: $a }",
	}, []byte("xabcx"))
	assertFalseRules(t, []string{
		"rule test { strings: $a = \"abc\" fullword condition: $a }",
	}, []byte("xabc"))
	assertFalseRules(t, []string{
		"rule test { strings: $a = \"abc\" fullword condition: $a }",
	}, []byte("abcx"))
	assertFalseRules(t, []string{
		"rule test { strings: $a = \"abc\" ascii wide fullword condition: $a }",
	}, []byte("abcx"))
	assertTrueRules(t, []string{
		"rule test { strings: $a = \"abc\" ascii wide fullword condition: $a }",
	}, []byte("a\x00abc"))
	assertTrueRules(t, []string{
		"rule test { strings: $a = \"abc\" wide fullword condition: $a }",
	}, []byte("a\x00b\x00c\x00"))
	assertFalseRules(t, []string{
		"rule test { strings: $a = \"abc\" wide fullword condition: $a }",
	}, []byte("x\x00a\x00b\x00c\x00x\x00"))
	assertFalseRules(t, []string{
		"rule test { strings: $a = \"ab\" wide fullword condition: $a }",
	}, []byte("x\x00a\x00b\x00"))
	assertFalseRules(t, []string{
		"rule test { strings: $a = \"abc\" wide fullword condition: $a }",
	}, []byte("x\x00a\x00b\x00c\x00"))
	assertTrueRules(t, []string{
		"rule test { strings: $a = \"abc\" wide fullword condition: $a }",
	}, []byte("x\x01a\x00b\x00c\x00"))
	assertTrueRules(t, []string{
		"rule test {\n" +
			"   strings:\n" +
			"     $a = \"abcdef\"\n" +
			"     $b = \"cdef\"\n" +
			"     $c = \"ef\"\n" +
			"   condition:\n" +
			"     all of them\n" +
			"}",
	}, []byte("abcdef"))

}

var pe32file = []byte{
	0x4d, 0x5a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x40, 0x00, 0x00, 0x00,
	0x50, 0x45, 0x00, 0x00, 0x4c, 0x01, 0x01, 0x00, 0x5d, 0xbe, 0x45, 0x45, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0xe0, 0x00, 0x03, 0x01, 0x0b, 0x01, 0x08, 0x00, 0x04, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x60, 0x01, 0x00, 0x00, 0x60, 0x01, 0x00, 0x00,
	0x64, 0x01, 0x00, 0x00, 0x00, 0x00, 0x40, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00,
	0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x64, 0x01, 0x00, 0x00, 0x60, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x04,
	0x00, 0x00, 0x10, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x10, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2e, 0x74, 0x65, 0x78, 0x74, 0x00, 0x00, 0x00,
	0x04, 0x00, 0x00, 0x00, 0x60, 0x01, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x60, 0x01, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x20, 0x00, 0x00, 0x60,
	0x6a, 0x2a, 0x58, 0xc3,
}

var elf32file = []byte{
	0x7f, 0x45, 0x4c, 0x46, 0x01, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x02, 0x00, 0x03, 0x00, 0x01, 0x00, 0x00, 0x00, 0x60, 0x80, 0x04, 0x08, 0x34, 0x00, 0x00, 0x00,
	0xa8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x34, 0x00, 0x20, 0x00, 0x01, 0x00, 0x28, 0x00,
	0x04, 0x00, 0x03, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80, 0x04, 0x08,
	0x00, 0x80, 0x04, 0x08, 0x6c, 0x00, 0x00, 0x00, 0x6c, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00,
	0x00, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0xb8, 0x01, 0x00, 0x00, 0x00, 0xbb, 0x2a, 0x00, 0x00, 0x00, 0xcd, 0x80, 0x00, 0x54, 0x68, 0x65,
	0x20, 0x4e, 0x65, 0x74, 0x77, 0x69, 0x64, 0x65, 0x20, 0x41, 0x73, 0x73, 0x65, 0x6d, 0x62, 0x6c,
	0x65, 0x72, 0x20, 0x32, 0x2e, 0x30, 0x35, 0x2e, 0x30, 0x31, 0x00, 0x00, 0x2e, 0x73, 0x68, 0x73,
	0x74, 0x72, 0x74, 0x61, 0x62, 0x00, 0x2e, 0x74, 0x65, 0x78, 0x74, 0x00, 0x2e, 0x63, 0x6f, 0x6d,
	0x6d, 0x65, 0x6e, 0x74, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x0b, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x00, 0x00, 0x00, 0x60, 0x80, 0x04, 0x08,
	0x60, 0x00, 0x00, 0x00, 0x0c, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x11, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x6c, 0x00, 0x00, 0x00, 0x1f, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x01, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x8b, 0x00, 0x00, 0x00, 0x1a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
}

var elf64file = []byte{
	0x7f, 0x45, 0x4c, 0x46, 0x02, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x02, 0x00, 0x3e, 0x00, 0x01, 0x00, 0x00, 0x00, 0x80, 0x00, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x40, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xc8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x40, 0x00, 0x38, 0x00, 0x01, 0x00, 0x40, 0x00, 0x04, 0x00, 0x03, 0x00,
	0x01, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x8c, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x8c, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x20, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0xb8, 0x01, 0x00, 0x00, 0x00, 0xbb, 0x2a, 0x00, 0x00, 0x00, 0xcd, 0x80, 0x00, 0x54, 0x68, 0x65,
	0x20, 0x4e, 0x65, 0x74, 0x77, 0x69, 0x64, 0x65, 0x20, 0x41, 0x73, 0x73, 0x65, 0x6d, 0x62, 0x6c,
	0x65, 0x72, 0x20, 0x32, 0x2e, 0x30, 0x35, 0x2e, 0x30, 0x31, 0x00, 0x00, 0x2e, 0x73, 0x68, 0x73,
	0x74, 0x72, 0x74, 0x61, 0x62, 0x00, 0x2e, 0x74, 0x65, 0x78, 0x74, 0x00, 0x2e, 0x63, 0x6f, 0x6d,
	0x6d, 0x65, 0x6e, 0x74, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0b, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00,
	0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80, 0x00, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0c, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x11, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x8c, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x1f, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0xab, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x1a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
}

func TestHexStrings(t *testing.T) {
	assertTrueRules(t, []string{
		"rule test { strings: $a = { 64 01 00 00 60 01 } condition: $a }",
		"rule test { strings: $a = { 64 0? 00 00 ?0 01 } condition: $a }",
		"rule test { strings: $a = { 64 01 [1-3] 60 01 } condition: $a }",
		"rule test { strings: $a = { 64 01 [1-3] (60|61) 01 } condition: $a }",
		"rule test { strings: $a = { 4D 5A [-] 6A 2A [-] 58 C3} condition: $a }",
		"rule test { strings: $a = { 4D 5A [300-] 6A 2A [-] 58 C3} condition: $a }",
	}, pe32file)
	assertFalseRules(t, []string{"rule test { strings: $a = { 4D 5A [0-300] 6A 2A } condition: $a }"}, pe32file)
	assertTrueRules(t, []string{
		"rule test { strings: $a = { 31 32 [-] 38 39 } condition: $a }",
		"rule test { strings: $a = { 31 32 [-] 33 34 [-] 38 39 } condition: $a }",
		"rule test { strings: $a = { 31 32 [1] 34 35 [2] 38 39 } condition: $a }",
		"rule test { strings: $a = { 31 32 [1-] 34 35 [1-] 38 39 } condition: $a }",
		"rule test { strings: $a = { 31 32 [0-3] 34 35 [1-] 38 39 } condition: $a }",
	}, []byte("123456789"))
	assertTrueRules(t, []string{
		"rule test { strings: $a = { 31 32 [-] 38 39 } condition: all of them }",
	}, []byte("123456789"))
	assertFalseRules(t, []string{
		"rule test { strings: $a = { 31 32 [-] 32 33 } condition: $a }",
		"rule test { strings: $a = { 35 36 [-] 31 32 } condition: $a }",
		"rule test { strings: $a = { 31 32 [2-] 34 35 } condition: $a }",
		"rule test { strings: $a = { 31 32 [0-3] 37 38 } condition: $a }",
	}, []byte("123456789"))

	rules := makeRules(t, "rule test { strings: $a = { 61 [0-3] (62|63) } condition: $a }")
	matches, _ := rules.ScanMem([]byte("abbb"), 0, 0)
	if matches[0].Strings[0].Name != "$a" ||
		matches[0].Strings[0].Offset != 0 ||
		string(matches[0].Strings[0].Data) != "ab" {
		t.Error("wrong match")
	}
}

//func TestXXX(t *testing.T) {
//	assertTrueRules(t, []string{}, []byte("dummy"))
//	assertFalseRules(t, []string{}, []byte())
//}

func TestEntrypoint(t *testing.T) {
	assertTrueRules(t, []string{
		"rule test { strings: $a = { 6a 2a 58 c3 } condition: $a at entrypoint }",
	}, pe32file)
	assertTrueRules(t, []string{
		"rule test { strings: $a = { b8 01 00 00 00 bb 2a } condition: $a at entrypoint }",
	}, elf32file)
	assertTrueRules(t, []string{
		"rule test { strings: $a = { b8 01 00 00 00 bb 2a } condition: $a at entrypoint }",
	}, elf64file)
	assertFalseRules(t, []string{
		"rule test { condition: entrypoint >= 0 }",
	}, []byte("dummy"))
}

func TestFilesize(t *testing.T) {
	assertTrueRules(t, []string{
		fmt.Sprintf("rule test { condition: filesize == %d }", len(pe32file)),
	}, pe32file)
}

func TestSomething(t *testing.T) {
	assertTrueRules(t, []string{
		"rule test { condition: 0x55 | 0xAA == 0xFF }",
		"rule test { condition: ~0xAA ^ 0x5A & 0xFF == 0x0F }",
		"rule test { condition: ~0x55 & 0xFF == 0xAA }",
		"rule test { condition: 8 >> 2 == 2 }",
		"rule test { condition: 1 << 3 == 8 }",
	}, []byte("dummy"))
}