/* 
 * A name server which sends back the IP address of its client, the
 * recursive resolver. When queried for type TXT, it sends back the text
 * form of the address.  When queried for type A (resp. AAAA), it sends
 * back the IPv4 (resp. v6) address.
 *
 * Similar services: whoami.ultradns.net, whoami.akamai.net. Also (but it
 * is not their normal goal): rs.dns-oarc.net, porttest.dns-oarc.net,
 * amiopen.openresolvers.org.
 *
 * Original version from:
 * Stephane Bortzmeyer <stephane+grong@bortzmeyer.org>
 *
 * Adapted to Go DNS (i.e. completely rewritten)
 * Miek Gieben <miek@miek.nl>
 */

package main

import (
	"dns"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"
)

var (
	printf   *bool
	compress *bool
	tsig     *string
)

const dom = "whoami.miek.nl."

func handleReflect(w dns.ResponseWriter, r *dns.Msg) {
	var (
		v4  bool
		rr  dns.RR
		str string
		a   net.IP
	)
	// TC must be done here
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = *compress
	if ip, ok := w.RemoteAddr().(*net.UDPAddr); ok {
		str = "Port: " + strconv.Itoa(ip.Port) + " (udp)"
		a = ip.IP
		v4 = a.To4() != nil
	}
	if ip, ok := w.RemoteAddr().(*net.TCPAddr); ok {
		str = "Port: " + strconv.Itoa(ip.Port) + " (tcp)"
		a = ip.IP
		v4 = a.To4() != nil
	}

	if v4 {
		rr = new(dns.RR_A)
		rr.(*dns.RR_A).Hdr = dns.RR_Header{Name: dom, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 0}
		rr.(*dns.RR_A).A = a.To4()
	} else {
		rr = new(dns.RR_AAAA)
		rr.(*dns.RR_AAAA).Hdr = dns.RR_Header{Name: dom, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: 0}
		rr.(*dns.RR_AAAA).AAAA = a
	}

	t := new(dns.RR_TXT)
	t.Hdr = dns.RR_Header{Name: dom, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0}
	t.Txt = []string{str}

	switch r.Question[0].Qtype {
	case dns.TypeTXT:
		m.Answer = append(m.Answer, t)
		m.Extra = append(m.Extra, rr)
	default:
		fallthrough
	case dns.TypeAAAA, dns.TypeA:
		m.Answer = append(m.Answer, rr)
		m.Extra = append(m.Extra, t)
	}

	if r.IsTsig() {
		if w.TsigStatus() == nil {
			m.SetTsig(r.Extra[len(r.Extra)-1].(*dns.RR_TSIG).Hdr.Name, dns.HmacMD5, 300, r.MsgHdr.Id, time.Now().Unix())
		} else {
			println("Status", w.TsigStatus().Error())
		}
	}
	if *printf {
		fmt.Printf("%v\n", m.String())
	}
	w.Write(m)
}

func serve(net, name, secret string) {
	switch name {
	case "":
		err := dns.ListenAndServe(":8053", net, nil)
		if err != nil {
			fmt.Printf("Failed to setup the "+net+" server: %s\n", err.Error())
		}
	default:
		err := dns.ListenAndServeTsig(":8053", net, nil, map[string]string{name: secret})
		if err != nil {
			fmt.Printf("Failed to setup the "+net+" server: %s\n", err.Error())
		}
	}
}

func main() {
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	printf = flag.Bool("print", false, "print replies")
	compress = flag.Bool("compress", false, "compress replies")
	tsig = flag.String("tsig", "", "use MD5 hmac tsig: keyname:base64")
	var name, secret string
	flag.Usage = func() {
		flag.PrintDefaults()
	}
	flag.Parse()
	if *tsig != "" {
		a := strings.SplitN(*tsig, ":", 2)
		name, secret = a[0], a[1]
	}
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	dns.HandleFunc(".", handleReflect)
	go serve("tcp", name, secret)
	go serve("udp", name, secret)
	sig := make(chan os.Signal)
	signal.Notify(sig)
forever:
	for {
		select {
		case <-sig:
			fmt.Printf("Signal received, stopping\n")
			break forever
		}
	}
}
