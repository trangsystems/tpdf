/*
 * This file is subject to the terms and conditions defined in
 * file ' ', which is part of this source code package.
 */

package cmap

import (
	"sort"
	"strings"
	"testing"
)

func init() {
	// Uncomment when debugging to get debug or trace logging output.
	//common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
	//common.SetLogger(common.NewConsoleLogger(common.LogLevelTrace))
}

// cmap1Data represents a basic CMap.
const cmap1Data = `
	/CIDInit /ProcSet findresource begin
	12 dict begin
	begincmap
	/CIDSystemInfo
	<<  /Registry (Adobe)
	/Ordering (UCS)
	/Supplement 0
	>> def
	/CMapName /Adobe-Identity-UCS def
	/CMapType 2 def
	1 begincodespacerange
	<0000> <FFFF>
	endcodespacerange
	8 beginbfchar
	<0003> <0020>
	<0007> <0024>
	<0033> <0050>
	<0035> <0052>
	<0037> <0054>
	<005A> <0077>
	<005C> <0079>
	<005F> <007C>
	endbfchar
	7 beginbfrange
	<000F> <0017> <002C>
	<001B> <001D> <0038>
	<0025> <0026> <0042>
	<002F> <0031> <004C>
	<0044> <004C> <0061>
	<004F> <0053> <006C>
	<0055> <0057> <0072>
	endbfrange
	endcmap
	CMapName currentdict /CMap defineresource pop
	end
	end
`

// TestCMapParser tests basic loading of a simple CMap.
func TestCMapParser1(t *testing.T) {
	cmap, err := LoadCmapFromDataCID([]byte(cmap1Data))
	if err != nil {
		t.Error("Failed: ", err)
		return
	}

	if cmap.Name() != "Adobe-Identity-UCS" {
		t.Errorf("CMap name incorrect (%s)", cmap.Name())
		return
	}

	if cmap.Type() != 2 {
		t.Errorf("CMap type incorrect")
		return
	}

	if len(cmap.codespaces) != 1 {
		t.Errorf("len codespace != 1 (%d)", len(cmap.codespaces))
		return
	}

	if cmap.codespaces[0].Low != 0 {
		t.Errorf("code space low range != 0 (%d)", cmap.codespaces[0].Low)
		return
	}

	if cmap.codespaces[0].High != 0xFFFF {
		t.Errorf("code space high range != 0xffff (%d)", cmap.codespaces[0].High)
		return
	}

	expectedMappings := map[CharCode]rune{
		0x0003:     0x0020,
		0x005F:     0x007C,
		0x000F:     0x002C,
		0x000F + 5: 0x002C + 5,
		0x001B:     0x0038,
		0x001B + 2: 0x0038 + 2,
		0x002F:     0x004C,
		0x0044:     0x0061,
		0x004F:     0x006C,
		0x0055:     0x0072,
	}

	for k, expected := range expectedMappings {
		if v, ok := cmap.CharcodeToUnicode(k); !ok || v != string(expected) {
			t.Errorf("incorrect mapping, expecting 0x%X ??? 0x%X (%#v)", k, expected, v)
			return
		}
	}

	v, _ := cmap.CharcodeToUnicode(0x99)
	if v != MissingCodeString { //!= "notdef" {
		t.Errorf("Unmapped code, expected to map to undefined")
		return
	}

	charcodes := []byte{0x00, 0x03, 0x00, 0x0F}
	s, _ := cmap.CharcodeBytesToUnicode(charcodes)
	if s != " ," {
		t.Error("Incorrect charcode bytes ??? string mapping")
		return
	}
}

const cmap2Data = `
	/CIDInit /ProcSet findresource begin
	12 dict begin
	begincmap
	/CIDSystemInfo
	<<  /Registry (Adobe)
	/Ordering (UCS)
	/Supplement 0
	>> def
	/CMapName /Adobe-Identity-UCS def
	/CMapType 2 def
	1 begincodespacerange
	<0000> <FFFF>
	endcodespacerange
	7 beginbfrange
	<0080> <00FF> <002C>
	<802F> <902F> <0038>
	endbfrange
	endcmap
	CMapName currentdict /CMap defineresource pop
	end
	end
`

// TestCMapParser2 tests a bug that came up when 2-byte character codes had the higher byte set to 0,
// e.g. 0x0080, and the character map was not taking the number of bytes of the input codemap into account.
func TestCMapParser2(t *testing.T) {
	cmap, err := LoadCmapFromDataCID([]byte(cmap2Data))
	if err != nil {
		t.Error("Failed: ", err)
		return
	}

	if cmap.Name() != "Adobe-Identity-UCS" {
		t.Errorf("CMap name incorrect (%s)", cmap.Name())
		return
	}

	if cmap.Type() != 2 {
		t.Errorf("CMap type incorrect")
		return
	}

	if len(cmap.codespaces) != 1 {
		t.Errorf("len codespace != 1 (%d)", len(cmap.codespaces))
		return
	}

	if cmap.codespaces[0].Low != 0 {
		t.Errorf("code space low range != 0 (%d)", cmap.codespaces[0].Low)
		return
	}

	if cmap.codespaces[0].High != 0xFFFF {
		t.Errorf("code space high range != 0xffff (%d)", cmap.codespaces[0].High)
		return
	}

	expectedMappings := map[CharCode]rune{
		0x0080: 0x002C,
		0x802F: 0x0038,
	}

	for k, expected := range expectedMappings {
		if v, ok := cmap.CharcodeToUnicode(k); !ok || v != string(expected) {
			t.Errorf("incorrect mapping, expecting 0x%X ??? 0x%X (got 0x%X)", k, expected, v)
			return
		}
	}

	// Check byte sequence mappings.
	expectedSequenceMappings := []struct {
		bytes    []byte
		expected string
	}{
		{[]byte{0x80, 0x2F, 0x00, 0x80}, string([]rune{0x0038, 0x002C})},
	}

	for _, exp := range expectedSequenceMappings {
		str, _ := cmap.CharcodeBytesToUnicode(exp.bytes)
		if str != exp.expected {
			t.Errorf("Incorrect byte sequence mapping % X ??? % X (got % X)",
				exp.bytes, []rune(exp.expected), []rune(str))
			return
		}
	}
}

// cmapData3 is a CMap with a mixture of 1 and 2 byte codespaces.
const cmapData3 = `
	/CIDInit /ProcSet findresource begin
	12 dict begin begincmap
	/CIDSystemInfo
	3 dict dup begin
	/Registry (Adobe) def
	/Supplement 2 def
	end def

	/CMapName /test-1 def
	/CMapType 1 def

	4 begincodespacerange
	<00> <80>
	<8100> <9fff>
	<a0> <d0>
	<d140> <fbfc>
	endcodespacerange
	7 beginbfrange
	<00> <80> <10>
	<8100> <9f00> <1000>
	<a0> <d0> <90>
	<d140> <f000> <a000>
	endbfrange
	endcmap
`

// TestCMapParser3 test case of a CMap with mixed number of 1 and 2 bytes in the codespace range.
func TestCMapParser3(t *testing.T) {
	cmap, err := LoadCmapFromDataCID([]byte(cmapData3))
	if err != nil {
		t.Error("Failed: ", err)
		return
	}

	if cmap.Name() != "test-1" {
		t.Errorf("CMap name incorrect (%s)", cmap.Name())
		return
	}

	if cmap.Type() != 1 {
		t.Errorf("CMap type incorrect")
		return
	}

	// Check codespaces.
	expectedCodespaces := []Codespace{
		{1, 0x00, 0x80},
		{1, 0xa0, 0xd0},
		{2, 0x8100, 0x9fff},
		{2, 0xd140, 0xfbfc},
	}

	if len(cmap.codespaces) != len(expectedCodespaces) {
		t.Errorf("len codespace != %d (%d)", len(expectedCodespaces), len(cmap.codespaces))
		return
	}

	for i, cs := range cmap.codespaces {
		exp := expectedCodespaces[i]
		if cs.NumBytes != exp.NumBytes {
			t.Errorf("code space number of bytes != %d (%d) %x", exp.NumBytes, cs.NumBytes, exp)
			return
		}

		if cs.Low != exp.Low {
			t.Errorf("code space low range != %d (%d) %x", exp.Low, cs.Low, exp)
			return
		}

		if cs.High != exp.High {
			t.Errorf("code space high range != 0x%X (0x%X) %x", exp.High, cs.High, exp)
			return
		}
	}

	// Check mappings.
	expectedMappings := map[CharCode]rune{
		0x80:   0x10 + 0x80,
		0x8100: 0x1000,
		0xa0:   0x90,
		0xd140: 0xa000,
	}
	for k, expected := range expectedMappings {
		if v, ok := cmap.CharcodeToUnicode(k); !ok || v != string(expected) {
			t.Errorf("incorrect mapping: expecting 0x%02X ??? 0x%02X (got 0x%02X)", k, expected, v)
			return
		}
	}

	// Check byte sequence mappings.
	expectedSequenceMappings := []struct {
		bytes    []byte
		expected string
	}{

		{[]byte{0x80, 0x81, 0x00, 0xa1, 0xd1, 0x80, 0x00},
			string([]rune{
				0x90,
				0x1000,
				0x91,
				0xa000 + 0x40,
				0x10})},
	}

	for _, exp := range expectedSequenceMappings {
		str, _ := cmap.CharcodeBytesToUnicode(exp.bytes)
		if str != exp.expected {
			t.Errorf("Incorrect byte sequence mapping: % 02X ??? % 02X (got % 02X)",
				exp.bytes, []rune(exp.expected), []rune(str))
			return
		}
	}
}

// cmapData4 is a CMap with some utf16 encoded unicode strings that contain surrogates.
const cmap4Data = `
    /CIDInit /ProcSet findresource begin
    11 dict begin
    begincmap
    /CIDSystemInfo
    << /Registry (Adobe)
    /Ordering (UCS)
    /Supplement 0
    >> def
    /CMapName /Adobe-Identity-UCS def
    /CMapType 2 def
    1 begincodespacerange
    <0000> <FFFF>
    endcodespacerange
    15 beginbfchar
    <01E1> <002C>
    <0201> <007C>
    <059C> <21D2>
    <05CA> <2200>
    <05CC> <2203>
    <05D0> <2208>
    <0652> <2295>
    <073F> <D835DC50>
    <0749> <D835DC5A>
    <0889> <D835DC84>
    <0893> <D835DC8E>
    <08DD> <D835DC9E>
    <08E5> <D835DCA6>
    <08E7> <2133>
    <0D52> <2265>
    endbfchar
    1 beginbfrange
    <0E36> <0E37> <27F5>
    endbfrange
    endcmap
`

// TestCMapParser4 checks that ut16 encoded unicode strings are interpreted correctly.
func TestCMapParser4(t *testing.T) {
	cmap, err := LoadCmapFromDataCID([]byte(cmap4Data))
	if err != nil {
		t.Error("Failed to load CMap: ", err)
		return
	}

	if cmap.Name() != "Adobe-Identity-UCS" {
		t.Errorf("CMap name incorrect (%s)", cmap.Name())
		return
	}

	if cmap.Type() != 2 {
		t.Errorf("CMap type incorrect")
		return
	}

	if len(cmap.codespaces) != 1 {
		t.Errorf("len codespace != 1 (%d)", len(cmap.codespaces))
		return
	}

	if cmap.codespaces[0].Low != 0 {
		t.Errorf("code space low range != 0 (%d)", cmap.codespaces[0].Low)
		return
	}

	if cmap.codespaces[0].High != 0xFFFF {
		t.Errorf("code space high range != 0xffff (%d)", cmap.codespaces[0].High)
		return
	}

	expectedMappings := map[CharCode]rune{
		0x0889: '\U0001d484', // `????`
		0x0893: '\U0001d48e', // `????`
		0x08DD: '\U0001d49e', // `????`
		0x08E5: '\U0001d4a6', // `????
	}

	for k, expected := range expectedMappings {
		if v, ok := cmap.CharcodeToUnicode(k); !ok || v != string(expected) {
			t.Errorf("incorrect mapping, expecting 0x%04X ??? %+q (got %+q)", k, expected, v)
			return
		}
	}

	// Check byte sequence mappings.
	expectedSequenceMappings := []struct {
		bytes    []byte
		expected string
	}{
		{[]byte{0x07, 0x3F, 0x07, 0x49}, "\U0001d450\U0001d45a"}, // `????????`
		{[]byte{0x08, 0x89, 0x08, 0x93}, "\U0001d484\U0001d48e"}, // `????????`
		{[]byte{0x08, 0xDD, 0x08, 0xE5}, "\U0001d49e\U0001d4a6"}, // `????????`
		{[]byte{0x08, 0xE7, 0x0D, 0x52}, "\u2133\u2265"},         // `??????`
	}

	for _, exp := range expectedSequenceMappings {
		str, _ := cmap.CharcodeBytesToUnicode(exp.bytes)
		if str != exp.expected {
			t.Errorf("Incorrect byte sequence mapping % 02X ??? %+q (got %+q)",
				exp.bytes, exp.expected, str)
			return
		}
	}
}

var (
	codeToUnicode1 = map[CharCode]rune{ // 40 entries
		0x02ca: '??',
		0x02cb: '??',
		0x02cd: '??',
		0x039c: '??',
		0x039d: '??',
		0x039e: '??',
		0x039f: '??',
		0x03a0: '??',
		0x03a1: '??',
		0x03a6: '??',
		0x03b1: '??',
		0x03b2: '??',
		0x03b3: '??',
		0x03b4: '??',
		0x03b5: '??',
		0x03b6: '??',
		0x03b7: '??',
		0x03c6: '??',
		0x03c7: '??',
		0x03c9: '??',
		0x2013: '???',
		0x2014: '???',
		0x2018: '???',
		0x2019: '???',
		0x203e: '???',
		0x20ac: '???',
		0x2163: '???',
		0x2164: '???',
		0x2165: '???',
		0x2166: '???',
		0x2167: '???',
		0x2168: '???',
		0x2169: '???',
		0x2190: '???',
		0x2191: '???',
		0x2192: '???',
		0x2193: '???',
		0x2220: '???',
		0x2223: '???',
		0x222a: '???',
	}

	codeToUnicode2 = map[CharCode]rune{ // 40 entries
		0x0100: '??',
		0x0101: '??',
		0x0102: '??',
		0x0111: '??',
		0x0112: '??',
		0x0113: '??',
		0x0114: '??',
		0x0115: '??',
		0x0116: '??',
		0x011b: '??',
		0x0126: '??',
		0x0127: '??',
		0x0128: '??',
		0x0129: '??',
		0x012a: '??',
		0x012b: '??',
		0x012c: '??',
		0x013b: '??',
		0x013c: '??',
		0x013e: '??',
		0x013f: '??',
		0x0140: '??',
		0x0141: '??',
		0x0150: '??',
		0x0151: '??',
		0x0152: '??',
		0x0153: '??',
		0x0154: '??',
		0x0155: '??',
		0x015a: '??',
		0x0165: '??',
		0x0166: '??',
		0x0167: '??',
		0x0168: '??',
		0x0169: '??',
		0x016a: '??',
		0x016b: '??',
		0x017a: '??',
		0x017b: '??',
		0x017d: '??',
	}

	codeToUnicode3 = map[CharCode]rune{ // 93 entries
		0x0124: '??',
		0x0125: '??',
		0x0126: '??',
		0x0127: '??',
		0x0134: '??',
		0x0135: '??',
		0x0136: '??',
		0x0137: '??',
		0x0138: '??',
		0x0144: '??',
		0x0145: '??',
		0x0146: '??',
		0x0147: '??',
		0x0154: '??',
		0x0155: '??',
		0x0156: '??',
		0x0157: '??',
		0x0164: '??',
		0x0169: '??',
		0x0174: '??',
		0x0175: '??',
		0x0176: '??',
		0x0177: '??',
		0x0184: '??',
		0x0185: '??',
		0x0186: '??',
		0x0187: '??',
		0x0194: '??',
		0x019a: '??',
		0x01a4: '??',
		0x01a5: '??',
		0x01a6: '??',
		0x01a7: '??',
		0x01b4: '??',
		0x01b5: '??',
		0x01b6: '??',
		0x01b7: '??',
		0x01c4: '??',
		0x01cb: '??',
		0x01d4: '??',
		0x01d5: '??',
		0x01d6: '??',
		0x01d7: '??',
		0x01e4: '??',
		0x01e5: '??',
		0x01e6: '??',
		0x01e7: '??',
		0x01f4: '??',
		0x01f5: '??',
		0x0204: '??',
		0x0205: '??',
		0x0206: '??',
		0x0207: '??',
		0x0214: '??',
		0x0215: '??',
		0x0216: '??',
		0x0217: '??',
		0x0224: '??',
		0x0226: '??',
		0x0227: '??',
		0x0254: '??',
		0x0255: '??',
		0x0256: '??',
		0x0257: '??',
		0x0264: '??',
		0x0265: '??',
		0x0266: '??',
		0x0267: '??',
		0x0273: '??',
		0x0274: '??',
		0x0275: '??',
		0x0276: '??',
		0x0277: '??',
		0x0284: '??',
		0x0285: '??',
		0x0286: '??',
		0x0287: '??',
		0x0294: '??',
		0x0296: '??',
		0x0297: '??',
		0x02a4: '??',
		0x02a5: '??',
		0x02c6: '??',
		0x02c7: '??',
		0x0304: '??',
		0x0305: '??',
		0x0306: '??',
		0x0307: '??',
		0x030d: '??',
		0x0314: '??',
		0x0315: '??',
		0x0316: '??',
		0x0317: '??',
	}
)

const bfData1 = `
8 beginbfchar
<02cd> <02cd>
<03a6> <03a6>
<03c9> <03c9>
<203e> <203e>
<20ac> <20ac>
<2220> <2220>
<2223> <2223>
<222a> <222a>
endbfchar
8 beginbfrange
<02ca><02cb> <02ca>
<039c><03a1> <039c>
<03b1><03b7> <03b1>
<03c6><03c7> <03c6>
<2013><2014> <2013>
<2018><2019> <2018>
<2163><2169> <2163>
<2190><2193> <2190>
endbfrange
`

// TestBfData checks that cmap.toBfData produces the expected output.
func TestBfData(t *testing.T) {
	cmap := NewToUnicodeCMap(codeToUnicode1)

	bfDataExpected := strings.Trim(bfData1, "\n")
	bfDataTest := cmap.toBfData()

	if bfDataTest != bfDataExpected {
		t.Errorf("Incorrect bfData")
		return
	}
}

// TestBfData checks that cmap.toBfData produces the expected output.
func TestCMapCreation(t *testing.T) {
	checkCmapWriteRead(t, codeToUnicode1)
	checkCmapWriteRead(t, codeToUnicode2)
	checkCmapWriteRead(t, codeToUnicode3)
}

// checkCmapWriteRead creates CMap data from `codeToUnicode` then parses it and checks that the
// same codeToUnicode is returned.
func checkCmapWriteRead(t *testing.T, codeToUnicode map[CharCode]rune) {
	cmap0 := NewToUnicodeCMap(codeToUnicode)

	data := cmap0.Bytes()
	cmap, err := LoadCmapFromDataCID(data)
	if err != nil {
		t.Error("Failed to load CMap: ", err)
		return
	}

	codes0 := make([]CharCode, 0, len(codeToUnicode))
	for code := range codeToUnicode {
		codes0 = append(codes0, code)
	}
	sort.Slice(codes0, func(i, j int) bool { return codes0[i] < codes0[j] })
	codes := make([]CharCode, 0, len(cmap.codeToUnicode))
	for code := range cmap.codeToUnicode {
		codes = append(codes, code)
	}
	sort.Slice(codes, func(i, j int) bool { return codes[i] < codes[j] })

	if len(cmap.codeToUnicode) != len(codeToUnicode) {
		t.Errorf("Incorrect length. expected=%d test=%d", len(codeToUnicode1), len(cmap.codeToUnicode))
		return
	}

	for i, code := range codes0 {
		if code != codes[i] {
			t.Errorf("Code mismatch: i=%d expected=0x%04x test=0x%04x", i, code, codes[i])
			return
		}
		u0 := codeToUnicode[code]
		u := cmap.codeToUnicode[code]
		if u != string(u0) {
			t.Errorf("Unicode mismatch: i=%d code0=0x%04x expected=%q test=%q", i, code, u0, u)
			return
		}
	}
}
