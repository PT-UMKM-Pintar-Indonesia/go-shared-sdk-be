package sdk_helper

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"net"
	"net/url"
	"reflect"
	"sort"
	"testing"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
)

type named struct {
	ID   int
	Name string
}

func TestBuildSliceByWithVariants(t *testing.T) {
	byID := func(n named) int { return n.ID }
	eqName := func(a, b named) bool { return a.Name == b.Name }

	slices := [][]named{
		{{1, "a"}, {2, "b"}},
		{{2, "b"}, {3, "c"}},
	}

	union := UnionBy(slices, byID)
	if len(union) != 3 {
		t.Fatalf("UnionBy got %v", union)
	}

	unionW := UnionWith(slices, eqName)
	if len(unionW) != 3 {
		t.Fatalf("UnionWith got %v", unionW)
	}

	inter := IntersectionBy(slices, byID)
	if len(inter) != 1 || inter[0].ID != 2 {
		t.Fatalf("IntersectionBy got %v", inter)
	}

	interW := IntersectionWith(slices, eqName)
	if len(interW) != 1 || interW[0].Name != "b" {
		t.Fatalf("IntersectionWith got %v", interW)
	}

	xor := XorBy(slices, byID)
	if len(xor) != 2 {
		t.Fatalf("XorBy got %v", xor)
	}

	xorW := XorWith(slices, eqName)
	if len(xorW) != 2 {
		t.Fatalf("XorWith got %v", xorW)
	}

	// Single-slice and empty shortcuts.
	if got := IntersectionBy([][]named{{{1, "a"}}}, byID); len(got) != 1 {
		t.Fatalf("single IntersectionBy got %v", got)
	}

	if got := IntersectionBy(nil, byID); len(got) != 0 {
		t.Fatalf("nil IntersectionBy got %v", got)
	}

	if got := IntersectionWith([][]named{{{1, "a"}}}, eqName); len(got) != 1 {
		t.Fatalf("single IntersectionWith got %v", got)
	}

	if got := IntersectionWith(nil, eqName); len(got) != 0 {
		t.Fatalf("nil IntersectionWith got %v", got)
	}

	if got := XorWith(nil, eqName); len(got) != 0 {
		t.Fatalf("nil XorWith got %v", got)
	}
}

func TestMapObjectKeys(t *testing.T) {
	safe := NewMapObjectSafe[string, int](map[string]int{"a": 1, "b": 2})

	ks := safe.Keys()
	sort.Strings(ks)

	if !reflect.DeepEqual(ks, []string{"a", "b"}) {
		t.Fatalf("safe keys %v", ks)
	}

	unsafe := NewMapObjectUnSafe[string, int](map[string]int{"x": 3})

	ku := unsafe.Keys()
	if len(ku) != 1 || ku[0] != "x" {
		t.Fatalf("unsafe keys %v", ku)
	}

	if got := Keys(map[int]string{5: "e", 6: "f"}); len(got) != 2 {
		t.Fatalf("free Keys %v", got)
	}
}

func TestSetIsSupersetOf(t *testing.T) {
	big := NewSetSafe(1, 2, 3)
	small := NewSetSafe(1, 2)

	if !big.IsSupersetOf(small) {
		t.Fatal("big must be superset of small")
	}

	if small.IsSupersetOf(big) {
		t.Fatal("small must not be superset of big")
	}

	bigU := NewSetUnSafe("a", "b")
	subU := NewSetUnSafe("a")
	emptyU := NewSetUnSafe[string]()

	if !bigU.IsSupersetOf(subU) || !bigU.IsSupersetOf(emptyU) {
		t.Fatal("unsafe superset checks failed")
	}

	if subU.IsSupersetOf(bigU) {
		t.Fatal("unsafe subset must not be superset")
	}
}

func TestIsPreWriteNetworkError(t *testing.T) {
	if !isPreWriteNetworkError(&net.DNSError{Err: "no such host", Name: "x"}) {
		t.Fatal("DNS errors are pre-write")
	}

	if !isPreWriteNetworkError(&net.OpError{Op: "dial", Err: errors.New("refused")}) {
		t.Fatal("dial errors are pre-write")
	}

	if isPreWriteNetworkError(&net.OpError{Op: "read", Err: errors.New("reset")}) {
		t.Fatal("read errors are not pre-write")
	}

	if isPreWriteNetworkError(errors.New("plain")) {
		t.Fatal("non-network errors are not pre-write")
	}

	if isPreWriteNetworkError(nil) {
		t.Fatal("nil must not be treated as pre-write")
	}
}

func TestToFormValues(t *testing.T) {
	v := url.Values{"a": {"1"}}

	if got, ok := toFormValues(v); !ok || got.Get("a") != "1" {
		t.Fatal("url.Values passthrough failed")
	}

	if got, ok := toFormValues(map[string]string{"b": "2"}); !ok || got.Get("b") != "2" {
		t.Fatal("map[string]string conversion failed")
	}

	if got, ok := toFormValues(map[string][]string{"c": {"3", "4"}}); !ok || len(got["c"]) != 2 {
		t.Fatal("map[string][]string conversion failed")
	}

	if got, ok := toFormValues(struct{ X int }{1}); ok || got != nil {
		t.Fatal("unsupported type must reject")
	}
}

func TestConstructorSmoke(t *testing.T) {
	if NewCrypto() == nil {
		t.Fatal("NewCrypto nil")
	}

	if NewRandom() == nil {
		t.Fatal("NewRandom nil")
	}

	if NewSignature() == nil {
		t.Fatal("NewSignature nil")
	}

	if NewSql() == nil {
		t.Fatal("NewSql nil")
	}
}

func TestCertPublicKeyRawToKeyAndBase64(t *testing.T) {
	cert := NewCert()

	gen := cert.GenerateKey(&sdk_dto.GeneratePrivateKey{
		KeySize:        sdk_cons.KEY_SIZE_2048,
		PrivateKeyType: sdk_cons.PRIVPKCS8,
		PublicKeyType:  sdk_cons.PUBPKCS8,
	})
	if gen.Error != nil {
		t.Fatal(gen.Error)
	}

	pub := cert.PublicKeyRawToKey(&sdk_dto.PublicKeyRawToKey{
		KeyType:      sdk_cons.PUBPKCS8,
		KeyRawPublic: gen.KeyRawPublic,
	})
	if pub.Error != nil || pub.KeyPublic == nil {
		t.Fatalf("pkcs8 public got err %v", pub.Error)
	}

	gen1 := cert.GenerateKey(&sdk_dto.GeneratePrivateKey{
		KeySize:        sdk_cons.KEY_SIZE_2048,
		PrivateKeyType: sdk_cons.PRIVPKCS1,
		PublicKeyType:  sdk_cons.PUBPKCS1,
	})
	if gen1.Error != nil {
		t.Fatal(gen1.Error)
	}

	pub1 := cert.PublicKeyRawToKey(&sdk_dto.PublicKeyRawToKey{
		KeyType:      sdk_cons.PUBPKCS1,
		KeyRawPublic: gen1.KeyRawPublic,
	})
	if pub1.Error != nil || pub1.KeyPublic == nil {
		t.Fatalf("pkcs1 public got err %v", pub1.Error)
	}

	if res := cert.PublicKeyRawToKey(&sdk_dto.PublicKeyRawToKey{KeyRawPublic: []byte("no pem")}); res.Error == nil {
		t.Fatal("garbage PEM must error")
	}

	// Non-RSA public key must be rejected.
	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	der, err := x509MarshalEC(&ecKey.PublicKey)
	if err != nil {
		t.Fatal(err)
	}

	ecPEM := pem.EncodeToMemory(&pem.Block{Type: sdk_cons.PUBPKCS8, Bytes: der})

	if res := cert.PublicKeyRawToKey(&sdk_dto.PublicKeyRawToKey{
		KeyType:      sdk_cons.PUBPKCS8,
		KeyRawPublic: ecPEM,
	}); res.Error == nil {
		t.Fatal("EC key must be rejected")
	}

	// Base64 converters.
	b64Priv := base64.StdEncoding.EncodeToString(gen.KeyRawPrivate)

	privRaw := cert.PrivateKeyBase64ToRaw(&sdk_dto.PrivateKeyBase64ToRaw{KeyRawPrivate: b64Priv})
	if privRaw.Error != nil || privRaw.KeyType != sdk_cons.PRIVPKCS8 || string(privRaw.KeyRawPrivate) != string(gen.KeyRawPrivate) {
		t.Fatalf("priv base64 err %v type %q", privRaw.Error, privRaw.KeyType)
	}

	b64Pub := base64.StdEncoding.EncodeToString(gen.KeyRawPublic)

	pubRaw := cert.PublicKeyBase64ToRaw(&sdk_dto.PublicKeyBase64ToRaw{KeyRawPublic: b64Pub})
	if pubRaw.Error != nil || pubRaw.KeyType != sdk_cons.PUBPKCS8 {
		t.Fatalf("pub base64 err %v type %q", pubRaw.Error, pubRaw.KeyType)
	}

	if res := cert.PrivateKeyBase64ToRaw(&sdk_dto.PrivateKeyBase64ToRaw{KeyRawPrivate: "!!!"}); res.Error == nil {
		t.Fatal("invalid base64 must error")
	}

	notPEM := base64.StdEncoding.EncodeToString([]byte("random bytes"))
	if res := cert.PublicKeyBase64ToRaw(&sdk_dto.PublicKeyBase64ToRaw{KeyRawPublic: notPEM}); res.Error == nil {
		t.Fatal("non-PEM payload must error")
	}

	// A private PEM fed to the public converter fails the type check.
	if res := cert.PublicKeyBase64ToRaw(&sdk_dto.PublicKeyBase64ToRaw{KeyRawPublic: b64Priv}); res.Error == nil {
		t.Fatal("wrong PEM type must error")
	}
}

func x509MarshalEC(pub *ecdsa.PublicKey) ([]byte, error) {
	return x509.MarshalPKIXPublicKey(pub)
}
