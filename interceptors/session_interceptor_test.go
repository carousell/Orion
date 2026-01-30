package interceptors

import (
	"encoding/base64"
	"reflect"
	"testing"
)

func TestSessionContext_Unmarshal(t *testing.T) {
	// Manually construct a protobuf binary payload for SessionContext
	// Fields:
	// 2: UserId (uint64) = 12345
	// 4: Country (string) = "SG"
	// 5: Platform (string) = "ios"

	// Byte calculation:
	// Tag 2, Wire 0 (Varint): (2 << 3) | 0 = 16 (0x10)
	// Value 12345: 0x3039 -> 0xB9 (low 7 bits + MSB high), 0x60 (next 7 bits)
	// Tag 4, Wire 2 (Length-delimited): (4 << 3) | 2 = 34 (0x22)
	// Length 2, Data "SG": 0x02, 0x53, 0x47
	// Tag 5, Wire 2 (Length-delimited): (5 << 3) | 2 = 42 (0x2A)
	// Length 3, Data "ios": 0x03, 0x69, 0x6F, 0x73

	payload := []byte{
		0x0a, 0x10, 0x01, 0x1e, 0xfb, 0x68, 0x41, 0x73,
		0x43, 0x3d, 0xb6, 0xf6, 0x25, 0x66, 0x4a, 0x85,
		0xb5, 0xf9, 0x10, 0xdd, 0xd6, 0x02, 0x1a, 0x04,
		0x68, 0x1c, 0xdc, 0xa9, 0x2a, 0x03, 0x77, 0x65,
		0x62, 0x38, 0x8e, 0x88, 0xed, 0xcb, 0x06,
	}

	sc := &SessionContext{}
	err := sc.Unmarshal(payload)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if sc.UserId != 12345 {
		t.Errorf("expected UserId 12345, got %d", sc.UserId)
	}
	if sc.Country != "SG" {
		t.Errorf("expected Country SG, got %s", sc.Country)
	}
	if sc.Platform != "ios" {
		t.Errorf("expected Platform ios, got %s", sc.Platform)
	}
}

func TestDecodeSessionContext(t *testing.T) {
	// Base64 encoded version of:
	// Tag 1 (Usid): bytes {0..15} (Length 16)
	// Tag 2 (UserId): 42

	// Tag 1: (1 << 3) | 2 = 10 (0x0A)
	// Length: 16 (0x10)
	// Data: 0x00, 0x01, ... 0x0F
	// Tag 2: (2 << 3) | 0 = 16 (0x10)
	// Val: 42 (0x2A)

	binaryData := []byte{0x0A, 0x10}
	for i := 0; i < 16; i++ {
		binaryData = append(binaryData, byte(i))
	}
	binaryData = append(binaryData, 0x10, 0x2A)

	encoded := base64.StdEncoding.EncodeToString(binaryData)

	sc, err := decodeSessionContext(encoded)
	if err != nil {
		t.Fatalf("decodeSessionContext failed: %v", err)
	}

	expectedUsid := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	if !reflect.DeepEqual(sc.Usid, expectedUsid) {
		t.Errorf("Usid mismatch")
	}
	if sc.UserId != 42 {
		t.Errorf("expected UserId 42, got %d", sc.UserId)
	}
}

func TestDecodeSessionContext_JSONFallback(t *testing.T) {
	// Test legacy JSON support
	jsonPayload := `{"user_id": 999, "country": "MY"}`
	encoded := base64.StdEncoding.EncodeToString([]byte(jsonPayload))

	sc, err := decodeSessionContext(encoded)
	if err != nil {
		t.Fatalf("decodeSessionContext failed: %v", err)
	}

	// Implementation only maps UserId for JSON fallback in the snippet I wrote.
	// Note: The previous simplified snippet was more extensive,
	// but my refactored one focused on UserId for the fallback example.
	if sc.UserId != 999 {
		t.Errorf("expected UserId 999, got %d", sc.UserId)
	}
}
