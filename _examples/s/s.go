// A nameserver in Go
// Zones must be defined by hand currently (this is a work-in-progress in the
// Go DNS library)
// It tries to do the correct things, formerr, nxdomain and answers
// it does ONLINE signing, with a predefined key, there is no ZSK/KSK destinction
// It DOES NOT DO NSEC/NSEC3 yet

package main
// a (simple) dig-like query tool in 99 lines of Go
import (
	"net"
	"dns"
	"dns/resolver"
	"os"
	"flag"
	"fmt"
	"strings"
)

func main() {
	var dnssec *bool = flag.Bool("dnssec", false, "Request DNSSEC records")
	var port *string = flag.String("port", "53", "Set the query port")
	var aa *bool = flag.Bool("aa", false, "Set AA flag in query")
	var ad *bool = flag.Bool("ad", false, "Set AD flag in query")
	var cd *bool = flag.Bool("cd", false, "Set CD flag in query")
	var rd *bool = flag.Bool("rd", true, "Unset RD flag in query")
	var tcp *bool = flag.Bool("tcp", false, "TCP mode")
        var nsid *bool = flag.Bool("nsid", false, "Ask for the NSID")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [@server] [qtype] [qclass] [name ...]\n", os.Args[0])
		flag.PrintDefaults()
	}

	nameserver := "@127.0.0.1"      // Default nameserver
	qtype := uint16(dns.TypeA)      // Default qtype
	qclass := uint16(dns.ClassINET) // Default qclass
	var qname []string

	flag.Parse()

FLAGS:
	for i := 0; i < flag.NArg(); i++ {
		// If it starts with @ it is a nameserver
		if flag.Arg(i)[0] == '@' {
			nameserver = flag.Arg(i)
			continue FLAGS
		}
                // First class, then type, to make ANY queries possible
		// And if it looks like type, it is a type
		for k, v := range dns.Rr_str {
			if v == strings.ToUpper(flag.Arg(i)) {
				qtype = k
				continue FLAGS
			}
		}
		// If it looks like a class, it is a class
		for k, v := range dns.Class_str {
			if v == strings.ToUpper(flag.Arg(i)) {
				qclass = k
				continue FLAGS
			}
		}
		// Anything else is a qname
		qname = append(qname, flag.Arg(i))
	}
	r := new(resolver.Resolver)
        r.FromFile("/etc/resolv.conf")
	r.Timeout = 2
	r.Port = *port
	r.Tcp = *tcp
	r.Attempts = 1
	qr := r.NewQuerier()
	// @server may be a name, resolv that 
	var err os.Error
	nameserver = string([]byte(nameserver)[1:]) // chop off @
	_, addr, err := net.LookupHost(nameserver)
	if err == nil {
		r.Servers = addr
	} else {
		r.Servers = []string{nameserver}
	}

	m := new(dns.Msg)
	m.MsgHdr.Authoritative = *aa
	m.MsgHdr.AuthenticatedData = *ad
	m.MsgHdr.CheckingDisabled = *cd
	m.MsgHdr.RecursionDesired = *rd
	m.Question = make([]dns.Question, 1)
	if *dnssec || *nsid {
		opt := new(dns.RR_OPT)
		opt.Hdr = dns.RR_Header{Name: "", Rrtype: dns.TypeOPT}
		opt.SetVersion(0)
		opt.SetDo()
		opt.SetUDPSize(4096)
                if *nsid {
                        opt.Option = make([]dns.Option, 1)
                        opt.Option[0].Code = dns.OptionCodeNSID
                        opt.Option[0].Data = ""
                }
		m.Extra = make([]dns.RR, 1)
		m.Extra[0] = opt
	}

	for _, v := range qname {
		m.Question[0] = dns.Question{v, qtype, qclass}
                m.SetId()
		qr <- resolver.Msg{m, nil, nil}
		in := <-qr
		if in.Dns != nil {
                        if m.Id != in.Dns.Id {
                                fmt.Printf("Id mismatch\n")
                        }
			fmt.Printf("%v\n", in.Dns)
                        fmt.Printf("%s\n", in.Meta)
		} else {
                        fmt.Printf("%v\n", in.Error.String())
                }
	}
	qr <- resolver.Msg{}
	<-qr
}