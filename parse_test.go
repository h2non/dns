package dns

import (
	"crypto/rsa"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestSignRSA(t *testing.T) {
	pub := "miek.nl. IN DNSKEY 256 3 5 AwEAAb+8lGNCxJgLS8rYVer6EnHVuIkQDghdjdtewDzU3G5R7PbMbKVRvH2Ma7pQyYceoaqWZQirSj72euPWfPxQnMy9ucCylA+FuH9cSjIcPf4PqJfdupHk9X6EBYjxrCLY4p1/yBwgyBIRJtZtAqM3ceAH2WovEJD6rTtOuHo5AluJ"

	priv := `Private-key-format: v1.3
Algorithm: 5 (RSASHA1)
Modulus: v7yUY0LEmAtLythV6voScdW4iRAOCF2N217APNTcblHs9sxspVG8fYxrulDJhx6hqpZlCKtKPvZ649Z8/FCczL25wLKUD4W4f1xKMhw9/g+ol926keT1foQFiPGsItjinX/IHCDIEhEm1m0Cozdx4AfZai8QkPqtO064ejkCW4k=
PublicExponent: AQAB
PrivateExponent: YPwEmwjk5HuiROKU4xzHQ6l1hG8Iiha4cKRG3P5W2b66/EN/GUh07ZSf0UiYB67o257jUDVEgwCuPJz776zfApcCB4oGV+YDyEu7Hp/rL8KcSN0la0k2r9scKwxTp4BTJT23zyBFXsV/1wRDK1A5NxsHPDMYi2SoK63Enm/1ptk=
Prime1: /wjOG+fD0ybNoSRn7nQ79udGeR1b0YhUA5mNjDx/x2fxtIXzygYk0Rhx9QFfDy6LOBvz92gbNQlzCLz3DJt5hw==
Prime2: wHZsJ8OGhkp5p3mrJFZXMDc2mbYusDVTA+t+iRPdS797Tj0pjvU2HN4vTnTj8KBQp6hmnY7dLp9Y1qserySGbw==
Exponent1: N0A7FsSRIg+IAN8YPQqlawoTtG1t1OkJ+nWrurPootScApX6iMvn8fyvw3p2k51rv84efnzpWAYiC8SUaQDNxQ==
Exponent2: SvuYRaGyvo0zemE3oS+WRm2scxR8eiA8WJGeOc+obwOKCcBgeZblXzfdHGcEC1KaOcetOwNW/vwMA46lpLzJNw==
Coefficient: 8+7ZN/JgByqv0NfULiFKTjtyegUcijRuyij7yNxYbCBneDvZGxJwKNi4YYXWx743pcAj4Oi4Oh86gcmxLs+hGw==
Created: 20110302104537
Publish: 20110302104537
Activate: 20110302104537`

	xk, _ := NewRR(pub)
	k := xk.(*RR_DNSKEY)
	p, err := k.NewPrivateKey(priv)
	if err != nil {
		t.Logf("%v\n", err)
		t.Fail()
	}
	switch priv := p.(type) {
	case *rsa.PrivateKey:
		if 65537 != priv.PublicKey.E {
			t.Log("Exponenent should be 65537")
			t.Fail()
		}
	default:
		t.Logf("We should have read an RSA key: %v", priv)
		t.Fail()
	}
	if k.KeyTag() != 37350 {
		t.Logf("%d %v\n", k.KeyTag(), k)
		t.Log("Keytag should be 37350")
		t.Fail()
	}

	soa := new(RR_SOA)
	soa.Hdr = RR_Header{"miek.nl.", TypeSOA, ClassINET, 14400, 0}
	soa.Ns = "open.nlnetlabs.nl."
	soa.Mbox = "miekg.atoom.net."
	soa.Serial = 1293945905
	soa.Refresh = 14400
	soa.Retry = 3600
	soa.Expire = 604800
	soa.Minttl = 86400

	sig := new(RR_RRSIG)
	sig.Hdr = RR_Header{"miek.nl.", TypeRRSIG, ClassINET, 14400, 0}
	sig.Expiration = 1296534305 // date -u '+%s' -d"2011-02-01 04:25:05"
	sig.Inception = 1293942305  // date -u '+%s' -d"2011-01-02 04:25:05"
	sig.KeyTag = k.KeyTag()
	sig.SignerName = k.Hdr.Name
	sig.Algorithm = k.Algorithm

	sig.Sign(p, []RR{soa})
	if sig.Signature != "D5zsobpQcmMmYsUMLxCVEtgAdCvTu8V/IEeP4EyLBjqPJmjt96bwM9kqihsccofA5LIJ7DN91qkCORjWSTwNhzCv7bMyr2o5vBZElrlpnRzlvsFIoAZCD9xg6ZY7ZyzUJmU6IcTwG4v3xEYajcpbJJiyaw/RqR90MuRdKPiBzSo=" {
		t.Log("Signature is not correct")
		t.Logf("%v\n", sig)
		t.Fail()
	}
}

func TestSignECDSA(t *testing.T) {
	pub := `example.net. 3600 IN DNSKEY 257 3 14 (
	xKYaNhWdGOfJ+nPrL8/arkwf2EY3MDJ+SErKivBVSum1
	w/egsXvSADtNJhyem5RCOpgQ6K8X1DRSEkrbYQ+OB+v8
	/uX45NBwY8rp65F6Glur8I/mlVNgF6W/qTI37m40 )`
	priv := `Private-key-format: v1.2
Algorithm: 14 (ECDSAP384SHA384)
PrivateKey: WURgWHCcYIYUPWgeLmiPY2DJJk02vgrmTfitxgqcL4vwW7BOrbawVmVe0d9V94SR`

	eckey, err := NewRR(pub)
	if err != nil {
		t.Fatal(err.Error())
	}
	privkey, err := eckey.(*RR_DNSKEY).NewPrivateKey(priv)
	if err != nil {
		t.Fatal(err.Error())
	}
	ds := eckey.(*RR_DNSKEY).ToDS(SHA384)
	if ds.KeyTag != 10771 {
		t.Fatal("Wrong keytag on DS")
	}
	if ds.Digest != "72d7b62976ce06438e9c0bf319013cf801f09ecc84b8d7e9495f27e305c6a9b0563a9b5f4d288405c3008a946df983d6" {
		t.Fatal("Wrong DS Digest")
	}
	a, _ := NewRR("www.example.net. 3600 IN A 192.0.2.1")
	sig := new(RR_RRSIG)
	sig.Hdr = RR_Header{"example.net.", TypeRRSIG, ClassINET, 14400, 0}
	sig.Expiration, _ = DateToTime("20100909102025")
	sig.Inception, _ = DateToTime("20100812102025")
	sig.KeyTag = eckey.(*RR_DNSKEY).KeyTag()
	sig.SignerName = eckey.(*RR_DNSKEY).Hdr.Name
	sig.Algorithm = eckey.(*RR_DNSKEY).Algorithm

	sig.Sign(privkey, []RR{a})

	t.Logf("%s", sig.String())
	if e := sig.Verify(eckey.(*RR_DNSKEY), []RR{a}); e != nil {
		t.Logf("Failure to validate: %s", e.Error())
		t.Fail()
	}
}

func TestDotInName(t *testing.T) {
	buf := make([]byte, 20)
	PackDomainName("aa\\.bb.nl.", buf, 0, nil, false)
	// index 3 must be a real dot
	if buf[3] != '.' {
		t.Log("Dot should be a real dot")
		t.Fail()
	}

	if buf[6] != 2 {
		t.Log("This must have the value 2")
		t.Fail()
	}
	dom, _, _ := UnpackDomainName(buf, 0)
	// printing it should yield the backspace again
	if dom != "aa\\.bb.nl." {
		t.Log("Dot should have been escaped: " + dom)
		t.Fail()
	}
}

func TestParseZone(t *testing.T) {
	zone := `z1.miek.nl. 86400 IN RRSIG NSEC 8 3 86400 20110823011301 20110724011301 12051 miek.nl. lyRljEQFOmajcdo6bBI67DsTlQTGU3ag9vlE07u7ynqt9aYBXyE9mkasAK4V0oI32YGb2pOSB6RbbdHwUmSt+cYhOA49tl2t0Qoi3pH21dicJiupdZuyjfqUEqJlQoEhNXGtP/pRvWjNA4pQeOsOAoWq/BDcWCSQB9mh2LvUOH4= ; {keyid = sksak}
z2.miek.nl.  86400   IN      NSEC    miek.nl. TXT RRSIG NSEC
$TTL 100
z3.miek.nl.  IN      NSEC    miek.nl. TXT RRSIG NSEC`
	to := ParseZone(strings.NewReader(zone), "", "")
	i := 0
	for x := range to {
		if x.Error == nil {
			switch i {
			case 0:
				if x.RR.Header().Name != "z1.miek.nl." {
					t.Log("Failed to parse z1")
					t.Fail()
				}
			case 1:
				if x.RR.Header().Name != "z2.miek.nl." {
					t.Log("Failed to parse z2")
					t.Fail()
				}
			case 2:
				if x.RR.String() != "z3.miek.nl.\t100\tIN\tNSEC\tmiek.nl. TXT RRSIG NSEC" {
					t.Logf("Failed to parse z3 %s", x.RR.String())
					t.Fail()
				}
			}
		} else {
			t.Logf("Failed to parse: %v\n", x.Error)
			t.Fail()
		}
		i++
	}
}

func TestDomainName(t *testing.T) {
	tests := []string{"r\\.gieben.miek.nl.", "www\\.www.miek.nl.",
		"www.*.miek.nl.", "www.*.miek.nl.",
	}
	dbuff := make([]byte, 40)

	for _, ts := range tests {
		if _, ok := PackDomainName(ts, dbuff, 0, nil, false); !ok {
			t.Log("Not a valid domain name")
			t.Fail()
			continue
		}
		n, _, ok := UnpackDomainName(dbuff, 0)
		if !ok {
			t.Log("Failed to unpack packed domain name")
			t.Fail()
			continue
		}
		if ts != n {
			t.Logf("Must be equal: in: %s, out: %s\n", ts, n)
			t.Fail()
		}
	}
}

func TestParseDirectiveMisc(t *testing.T) {
	tests := map[string]string{
		"$ORIGIN miek.nl.\na IN NS b": "a.miek.nl.\t3600\tIN\tNS\tb.miek.nl.",
		"$TTL 2H\nmiek.nl. IN NS b.":  "miek.nl.\t7200\tIN\tNS\tb.",
		"miek.nl. 1D IN NS b.":        "miek.nl.\t86400\tIN\tNS\tb.",
		`name. IN SOA  a6.nstld.com. hostmaster.nic.name. (
        203362132 ; serial
        5m        ; refresh (5 minutes)
        5m        ; retry (5 minutes)
        2w        ; expire (2 weeks)
        300       ; minimum (5 minutes)
)`: "name.\t3600\tIN\tSOA\ta6.nstld.com. hostmaster.nic.name. 203362132 300 300 1209600 300",
		". 3600000  IN  NS ONE.MY-ROOTS.NET.":        ".\t3600000\tIN\tNS\tONE.MY-ROOTS.NET.",
		"ONE.MY-ROOTS.NET. 3600000 IN A 192.168.1.1": "ONE.MY-ROOTS.NET.\t3600000\tIN\tA\t192.168.1.1",
	}
	for i, o := range tests {
		rr, e := NewRR(i)
		if e != nil {
			t.Log("Failed to parse RR: " + e.Error())
			t.Fail()
			continue
		}
		if rr.String() != o {
			t.Logf("`%s' should be equal to\n`%s', but is     `%s'\n", i, o, rr.String())
			t.Fail()
		} else {
			t.Logf("RR is OK: `%s'", rr.String())
		}
	}
}

func TestParseNSEC(t *testing.T) {
	nsectests := map[string]string{
		"nl. IN NSEC3PARAM 1 0 5 30923C44C6CBBB8F":                                                                                                 "nl.\t3600\tIN\tNSEC3PARAM\t1 0 5 30923C44C6CBBB8F",
		"p2209hipbpnm681knjnu0m1febshlv4e.nl. IN NSEC3 1 1 5 30923C44C6CBBB8F P90DG1KE8QEAN0B01613LHQDG0SOJ0TA NS SOA TXT RRSIG DNSKEY NSEC3PARAM": "p2209hipbpnm681knjnu0m1febshlv4e.nl.\t3600\tIN\tNSEC3\t1 1 5 30923C44C6CBBB8F P90DG1KE8QEAN0B01613LHQDG0SOJ0TA NS SOA TXT RRSIG DNSKEY NSEC3PARAM",
		"localhost.dnssex.nl. IN NSEC www.dnssex.nl. A RRSIG NSEC":                                                                                 "localhost.dnssex.nl.\t3600\tIN\tNSEC\twww.dnssex.nl. A RRSIG NSEC",
		"localhost.dnssex.nl. IN NSEC www.dnssex.nl. A RRSIG NSEC TYPE65534":                                                                       "localhost.dnssex.nl.\t3600\tIN\tNSEC\twww.dnssex.nl. A RRSIG NSEC TYPE65534",
	}
	for i, o := range nsectests {
		rr, e := NewRR(i)
		if e != nil {
			t.Log("Failed to parse RR: " + e.Error())
			t.Fail()
			continue
		}
		if rr.String() != o {
			t.Logf("`%s' should be equal to\n`%s', but is     `%s'\n", i, o, rr.String())
			t.Fail()
		} else {
			t.Logf("RR is OK: `%s'", rr.String())
		}
	}
}

func TestParseLOC(t *testing.T) {
	lt := map[string]string{
		"SW1A2AA.find.me.uk.	LOC	51 30 12.748 N 00 07 39.611 W 0.00m 0.00m 0.00m 0.00m":
		  "SW1A2AA.find.me.uk.\t3600\tIN\tLOC\t51 30 12.748 N 00 07 39.611 W 0.00m 0.00m 0.00m 0.00m",
		"SW1A2AA.find.me.uk.	LOC	51 0 0.0 N 00 07 39.611 W 0.00m 0.00m 0.00m 0.00m":
		  "SW1A2AA.find.me.uk.\t3600\tIN\tLOC\t51 00 0.000 N 00 07 39.611 W 0.00m 0.00m 0.00m 0.00m",
	}
	for i, o := range lt {
		rr, e := NewRR(i)
		if e != nil {
			t.Log("Failed to parse RR: " + e.Error())
			t.Fail()
			continue
		}
		if rr.String() != o {
			t.Logf("`%s' should be equal to\n`%s', but is     `%s'\n", i, o, rr.String())
			t.Fail()
		} else {
			t.Logf("RR is OK: `%s'", rr.String())
		}
	}
}

func TestQuotes(t *testing.T) {
	tests := map[string]string{
		`t.example.com. IN TXT "a bc"`: "t.example.com.\t3600\tIN\tTXT\t\"a bc\"",
		`t.example.com. IN TXT "a
 bc"`: "t.example.com.\t3600\tIN\tTXT\t\"a\\n bc\"",
		`t.example.com. IN TXT "a"`:                                                          "t.example.com.\t3600\tIN\tTXT\t\"a\"",
		`t.example.com. IN TXT "aa"`:                                                         "t.example.com.\t3600\tIN\tTXT\t\"aa\"",
		`t.example.com. IN TXT "aaa" ;`:                                                      "t.example.com.\t3600\tIN\tTXT\t\"aaa\"",
		`t.example.com. IN TXT "abc" "DEF"`:                                                  "t.example.com.\t3600\tIN\tTXT\t\"abc\" \"DEF\"",
		`t.example.com. IN TXT "abc" ( "DEF" )`:                                              "t.example.com.\t3600\tIN\tTXT\t\"abc\" \"DEF\"",
		`t.example.com. IN TXT aaa ;`:                                                        "t.example.com.\t3600\tIN\tTXT\t\"aaa \"",
		`t.example.com. IN TXT aaa aaa;`:                                                     "t.example.com.\t3600\tIN\tTXT\t\"aaa aaa\"",
		`t.example.com. IN TXT aaa aaa`:                                                      "t.example.com.\t3600\tIN\tTXT\t\"aaa aaa\"",
		`t.example.com. IN TXT aaa`:                                                          "t.example.com.\t3600\tIN\tTXT\t\"aaa\"",
		"cid.urn.arpa. NAPTR 100 50 \"s\" \"z3950+I2L+I2C\"    \"\" _z3950._tcp.gatech.edu.": "cid.urn.arpa.\t3600\tIN\tNAPTR\t100 50 \"s\" \"z3950+I2L+I2C\" \"\" _z3950._tcp.gatech.edu.",
		"cid.urn.arpa. NAPTR 100 50 \"s\" \"rcds+I2C\"         \"\" _rcds._udp.gatech.edu.":  "cid.urn.arpa.\t3600\tIN\tNAPTR\t100 50 \"s\" \"rcds+I2C\" \"\" _rcds._udp.gatech.edu.",
		"cid.urn.arpa. NAPTR 100 50 \"s\" \"http+I2L+I2C+I2R\" \"\" _http._tcp.gatech.edu.":  "cid.urn.arpa.\t3600\tIN\tNAPTR\t100 50 \"s\" \"http+I2L+I2C+I2R\" \"\" _http._tcp.gatech.edu.",
		"cid.urn.arpa. NAPTR 100 10 \"\" \"\" \"/urn:cid:.+@([^\\.]+\\.)(.*)$/\\2/i\" .":     "cid.urn.arpa.\t3600\tIN\tNAPTR\t100 10 \"\" \"\" \"/urn:cid:.+@([^\\.]+\\.)(.*)$/\\2/i\" .",
	}
	for i, o := range tests {
		rr, e := NewRR(i)
		if e != nil {
			t.Log("Failed to parse RR: " + e.Error())
			t.Fail()
			continue
		}
		if rr.String() != o {
			t.Logf("`%s' should be equal to\n`%s', but is\n`%s'\n", i, o, rr.String())
			t.Fail()
		} else {
			t.Logf("RR is OK: `%s'", rr.String())
		}
	}
}

func TestParseBrace(t *testing.T) {
	tests := map[string]string{
		"(miek.nl.) 3600 IN A 127.0.0.1":                 "miek.nl.\t3600\tIN\tA\t127.0.0.1",
		"miek.nl. (3600) IN MX (10) elektron.atoom.net.": "miek.nl.\t3600\tIN\tMX\t10 elektron.atoom.net.",
		`miek.nl. IN (
                        3600 A 127.0.0.1)`: "miek.nl.\t3600\tIN\tA\t127.0.0.1",
		"(miek.nl.) (A) (127.0.0.1)":                          "miek.nl.\t3600\tIN\tA\t127.0.0.1",
		"miek.nl A 127.0.0.1":                                 "miek.nl.\t3600\tIN\tA\t127.0.0.1",
		"_ssh._tcp.local. 60 IN (PTR) stora._ssh._tcp.local.": "_ssh._tcp.local.\t60\tIN\tPTR\tstora._ssh._tcp.local.",
		"miek.nl. NS ns.miek.nl":                              "miek.nl.\t3600\tIN\tNS\tns.miek.nl.",
		`(miek.nl.) (
                        (IN) 
                        (AAAA)
                        (::1) )`: "miek.nl.\t3600\tIN\tAAAA\t::1",
		`(miek.nl.) (
                        (IN) 
                        (AAAA)
                        (::1))`: "miek.nl.\t3600\tIN\tAAAA\t::1",
		`((m)(i)ek.(n)l.) (SOA) (soa.) (soa.) (
                                2009032802 ; serial
                                21600      ; refresh (6 hours)
                                7(2)00       ; retry (2 hours)
                                604()800     ; expire (1 week)
                                3600       ; minimum (1 hour)
                        )`: "miek.nl.\t3600\tIN\tSOA\tsoa. soa. 2009032802 21600 7200 604800 3600",
		"miek\\.nl. IN A 127.0.0.1": "miek\\.nl.\t3600\tIN\tA\t127.0.0.1",
		"miek.nl. IN A 127.0.0.1":   "miek.nl.\t3600\tIN\tA\t127.0.0.1",
		"miek.nl. A 127.0.0.1":      "miek.nl.\t3600\tIN\tA\t127.0.0.1",
		`miek.nl.       86400 IN SOA elektron.atoom.net. miekg.atoom.net. (
                                2009032802 ; serial
                                21600      ; refresh (6 hours)
                                7200       ; retry (2 hours)
                                604800     ; expire (1 week)
                                3600       ; minimum (1 hour)
                        )`: "miek.nl.\t86400\tIN\tSOA\telektron.atoom.net. miekg.atoom.net. 2009032802 21600 7200 604800 3600",
	}
	for i, o := range tests {
		rr, e := NewRR(i)
		if e != nil {
			t.Log("Failed to parse RR: " + e.Error())
			t.Fail()
			continue
		}
		if rr.String() != o {
			t.Logf("`%s' should be equal to\n`%s', but is     `%s'\n", i, o, rr.String())
			t.Fail()
		} else {
			t.Logf("RR is OK: `%s'", rr.String())
		}
	}
}

func TestParseFailure(t *testing.T) {
	tests := []string{"miek.nl. IN A 327.0.0.1",
		"miek.nl. IN AAAA ::x",
		"miek.nl. IN MX a0 miek.nl.",
		"miek.nl aap IN MX mx.miek.nl.",
		"miek.nl. IN CNAME ",
		"miek.nl. PA MX 10 miek.nl.",
		"miek.nl. ) IN MX 10 miek.nl.",
	}

	for _, s := range tests {
		_, err := NewRR(s)
		if err == nil {
			t.Log("Should have triggered an error")
			t.Fail()
		}
	}
}

// A bit useless, how to use b.N?
func BenchmarkZoneParsing(b *testing.B) {
	f, err := os.Open("t/miek.nl.signed_test")
	if err != nil {
		return
	}
	defer f.Close()
	to := ParseZone(f, "", "t/miek.nl.signed_test")
	for x := range to {
		x = x
	}
}

func TestZoneParsing(t *testing.T) {
	f, err := os.Open("t/miek.nl.signed_test")
	if err != nil {
		return
	}
	defer f.Close()
	start := time.Now().UnixNano()
	to := ParseZone(f, "", "t/miek.nl.signed_test")
	var i int
	for x := range to {
		t.Logf("%s\n", x.RR)
		i++
	}
	delta := time.Now().UnixNano() - start
	t.Logf("%d RRs parsed in %.2f s (%.2f RR/s)", i, float32(delta)/1e9, float32(i)/(float32(delta)/1e9))
}

// name.	3600	IN	SOA	a6.nstld.com. hostmaster.nic.name. 203362132 300 300 1209600 300
// name.	10800	IN	NS	name.
// name.	10800	IN	NS	g6.nstld.com.
// name.	7200	IN	NS	h6.nstld.com.
// name.	3600	IN	NS	j6.nstld.com.
// name.	3600	IN	NS	k6.nstld.com.
// name.	10800	IN	NS	l6.nstld.com.
// name.	10800	IN	NS	a6.nstld.com.
// name.	10800	IN	NS	c6.nstld.com.
// name.	10800	IN	NS	d6.nstld.com.
// name.	10800	IN	NS	f6.nstld.com.
// name.	10800	IN	NS	m6.nstld.com.
// 0-0onlus.name.	10800	IN	NS	ns7.ehiweb.it.
// 0-0onlus.name.	10800	IN	NS	ns8.ehiweb.it.
// 0-g.name.	10800	IN	MX	10 mx01.nic.name.
// 0-g.name.	10800	IN	MX	10 mx02.nic.name.
// 0-g.name.	10800	IN	MX	10 mx03.nic.name.
// 0-g.name.	10800	IN	MX	10 mx04.nic.name.
// 0-g.name.    10800   IN      TXT     "10 mx\"04.nic"
// moutamassey.0-g.name.name.	10800	IN	NS	ns01.yahoodomains.jp.
// moutamassey.0-g.name.name.	10800	IN	NS	ns02.yahoodomains.jp.
func ExampleZone() {
	zone := `$ORIGIN .
$TTL 3600       ; 1 hour
name                    IN SOA  a6.nstld.com. hostmaster.nic.name. (
                                203362132  ; serial
                                300        ; refresh (5 minutes)
                                300        ; retry (5 minutes)
                                1209600    ; expire (2 weeks)
                                300        ; minimum (5 minutes)
                                )
$TTL 10800      ; 3 hours
@	10800	IN	NS	@
               IN       NS      g6.nstld.com.
               7200     NS      h6.nstld.com.
             3600 IN    NS      j6.nstld.com.
             IN 3600    NS      k6.nstld.com.
                        NS      l6.nstld.com.
                        NS      a6.nstld.com.
                        NS      c6.nstld.com.
                        NS      d6.nstld.com.
                        NS      f6.nstld.com.
                        NS      m6.nstld.com.
$ORIGIN name.
0-0onlus                NS      ns7.ehiweb.it.
                        NS      ns8.ehiweb.it.
0-g                     MX      10 mx01.nic
                        MX      10 mx02.nic
                        MX      10 mx03.nic
                        MX      10 mx04.nic
                        TXT     "10 mx\"04.nic"
$ORIGIN 0-g.name
moutamassey             NS      ns01.yahoodomains.jp.
                        NS      ns02.yahoodomains.jp.
`
	to := ParseZone(strings.NewReader(zone), "", "testzone")
	for x := range to {
		fmt.Printf("%s\n", x.RR)
	}
}

// www.example.com.	3600	IN	HIP	 2 200100107B1A74DF365639CC39F1D578 AwEAAbdxyhNuSutc5EMzxTs9LBPCIkOFH8cIvM4p9+LrV4e19WzK00+CI6zBCQTdtWsuxKbWIy87UOoJTwkUs7lBu+Upr1gsNrut79ryra+bSRGQb1slImA8YVJyuIDsj7kwzG7jnERNqnWxZ48AWkskmdHaVDP4BcelrTI3rMXdXF5D rvs.example.com.
func ExampleHIP() {
	h := `www.example.com      IN  HIP ( 2 200100107B1A74DF365639CC39F1D578
                AwEAAbdxyhNuSutc5EMzxTs9LBPCIkOFH8cIvM4p
9+LrV4e19WzK00+CI6zBCQTdtWsuxKbWIy87UOoJTwkUs7lBu+Upr1gsNrut79ryra+bSRGQ
b1slImA8YVJyuIDsj7kwzG7jnERNqnWxZ48AWkskmdHaVDP4BcelrTI3rMXdXF5D
        rvs.example.com. )`
	if hip, err := NewRR(h); err == nil {
		fmt.Printf("%s\n", hip.String())
	}
}

// example.com.	1000	IN	SOA	master.example.com. admin.example.com. 1 4294967294 4294967293 4294967295 100
func ExampleSOA() {
	s := "example.com. 1000 SOA master.example.com. admin.example.com. 1 4294967294 4294967293 4294967295 100"
	if soa, err := NewRR(s); err == nil {
		fmt.Printf("%s\n", soa.String())
	}
}

func TestLineNumberError(t *testing.T) {
	s := "example.com. 1000 SOA master.example.com. admin.example.com. monkey 4294967294 4294967293 4294967295 100"
	if _, err := NewRR(s); err != nil {
		if err.Error() != "dns: bad SOA zone parameter: \"monkey\" at line: 1:68" {
			t.Logf("Not expecting this error: " + err.Error())
			t.Fail()
		}
	}
}

// Test with no known RR on the line
func TestLineNumberError2(t *testing.T) {
	s := "example.com. 1000 SO master.example.com. admin.example.com. 1 4294967294 4294967293 4294967295 100"
	_, err := NewRR(s)
	if err == nil {
		t.Fail()
	} else {
		//		fmt.Printf("%s\n", err.Error())
	}
}
