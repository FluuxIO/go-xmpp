package xmpp

import (
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"gosrc.io/xmpp/stanza"
)

// TODO: Should I move this as an extension of the client?
//    I should probably make the code more modular, but keep concern separated to keep it simple.
type ServerCheck struct {
	address string
	domain  string
}

func NewChecker(address, domain string) (*ServerCheck, error) {
	client := ServerCheck{}

	var err error
	var host string
	if client.address, host, err = extractParams(address); err != nil {
		return &client, err
	}

	if domain != "" {
		client.domain = domain
	} else {
		client.domain = host
	}

	return &client, nil
}

// Check triggers actual TCP connection, based on previously defined parameters.
func (c *ServerCheck) Check() error {
	var tcpconn net.Conn
	var err error

	timeout := 15 * time.Second
	tcpconn, err = net.DialTimeout("tcp", c.address, timeout)
	if err != nil {
		return err
	}

	decoder := xml.NewDecoder(tcpconn)

	// Send stream open tag
	if _, err = fmt.Fprintf(tcpconn, clientStreamOpen, c.domain); err != nil {
		return err
	}

	// Set xml decoder and extract streamID from reply (not used for now)
	_, err = stanza.InitStream(decoder)
	if err != nil {
		return err
	}

	// extract stream features
	var f stanza.StreamFeatures
	packet, err := stanza.NextPacket(decoder)
	if err != nil {
		err = fmt.Errorf("stream open decode features: %w", err)
		return err
	}

	switch p := packet.(type) {
	case stanza.StreamFeatures:
		f = p
	case stanza.StreamError:
		return errors.New("open stream error: " + p.Error.Local)
	default:
		return errors.New("expected packet received while expecting features, got " + p.Name())
	}

	if _, ok := f.DoesStartTLS(); ok {
		_, err = fmt.Fprintf(tcpconn, "<starttls xmlns='urn:ietf:params:xml:ns:xmpp-tls'/>")
		if err != nil {
			return err
		}

		var k stanza.TLSProceed
		if err = decoder.DecodeElement(&k, nil); err != nil {
			return fmt.Errorf("expecting starttls proceed: %w", err)
		}

		var tlsConfig tls.Config
		tlsConfig.ServerName = c.domain
		tlsConn := tls.Client(tcpconn, &tlsConfig)
		// We convert existing connection to TLS
		if err = tlsConn.Handshake(); err != nil {
			return err
		}

		// We check that cert matches hostname
		if err = tlsConn.VerifyHostname(c.domain); err != nil {
			return err
		}

		if err = checkExpiration(tlsConn); err != nil {
			return err
		}
		return nil
	}
	return errors.New("TLS not supported on server")
}

// Check expiration date for the whole certificate chain and returns an error
// if the expiration date is in less than 48 hours.
func checkExpiration(tlsConn *tls.Conn) error {
	checkedCerts := make(map[string]struct{})
	for _, chain := range tlsConn.ConnectionState().VerifiedChains {
		for _, cert := range chain {
			if _, checked := checkedCerts[string(cert.Signature)]; checked {
				continue
			}
			checkedCerts[string(cert.Signature)] = struct{}{}

			// Check the expiration.
			timeNow := time.Now()
			expiresInHours := int64(cert.NotAfter.Sub(timeNow).Hours())
			// fmt.Printf("Cert '%s' expires in %d days\n", cert.Subject.CommonName, expiresInHours/24)
			if expiresInHours <= 48 {
				return fmt.Errorf("certificate '%s' will expire on %s", cert.Subject.CommonName, cert.NotAfter)
			}
		}
	}
	return nil
}

func extractParams(addr string) (string, string, error) {
	var err error
	hostport := strings.Split(addr, ":")
	if len(hostport) > 2 {
		err = errors.New("too many colons in xmpp server address")
		return addr, hostport[0], err
	}

	// Address is composed of two parts, we are good
	if len(hostport) == 2 && hostport[1] != "" {
		return addr, hostport[0], err
	}

	// Port was not passed, we append XMPP default port:
	return strings.Join([]string{hostport[0], "5222"}, ":"), hostport[0], err
}
