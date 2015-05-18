package xmpp

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net"
)

const (
	NsJabberClient = "jabber:client"
	NsStream       = "http://etherx.jabber.org/streams"
	NsIqAuth       = "jabber:iq:auth"
	NsIqRoster     = "jabber:iq:roster"
	NsTLS          = "urn:ietf:params:xml:ns:xmpp-tls"
	NsDisco        = "http://jabber.org/protocol/disco#items"
	NsMuc          = "http://jabber.org/protocol/muc"

	xmlStream      = "<stream:stream from='%s' to='%s' version='1.0' xml:lang='en' xmlns='%s' xmlns:stream='%s'>"
	xmlStartTLS    = "<starttls xmlns='%s'/>"
	xmlIqSet       = "<iq type='set' id='%s'><query xmlns='%s'><username>%s</username><password>%s</password><resource>%s</resource></query></iq>"
	xmlIqGet       = "<iq from='%s' to='%s' id='%s' type='get'><query xmlns='%s'/></iq>"
	xmlPresence    = "<presence from='%s'><show>%s</show></presence>"
	xmlMUCPresence = "<presence id='%s' to='%s' from='%s'><x xmlns='%s'/></presence>"
	xmlMUCMessage  = "<message from='%s' id='%s' to='%s' type='groupchat'><body>%s</body></message>"
	xmlMessage     = "<message from='%s' id='%s' to='%s' type='chat'><body>%s</body></message>"
)

type required struct{}

type features struct {
	XMLName    xml.Name  `xml:"features"`
	StartTLS   *required `xml:"starttls>required"`
	Mechanisms []string  `xml:"mechanisms>mechanism"`
}

type item struct {
	Jid         string `xml:"jid,attr"`
	Name        string `xml:"name,attr"`
	MentionName string `xml:"mention_name,attr"`
}

type query struct {
	XMLName xml.Name `xml:"query"`
	Items   []*item  `xml:"item"`
}

type body struct {
	Body string `xml:",innerxml"`
}

type Conn struct {
	incoming *xml.Decoder
	outgoing net.Conn
}

type Message struct {
	Jid         string
	MentionName string
	Body        string
}

func (c *Conn) Stream(jid, host string) {
	fmt.Fprintf(c.outgoing, xmlStream, jid, host, NsJabberClient, NsStream)
}

func (c *Conn) StartTLS() {
	fmt.Fprintf(c.outgoing, xmlStartTLS, NsTLS)
}

func (c *Conn) UseTLS() {
	c.outgoing = tls.Client(c.outgoing, &tls.Config{ServerName: "chat.hipchat.com"})
	c.incoming = xml.NewDecoder(c.outgoing)
}

func (c *Conn) Auth(user, pass, resource string) {
	fmt.Fprintf(c.outgoing, xmlIqSet, id(), NsIqAuth, user, pass, resource)
}

func (c *Conn) Features() *features {
	var f features
	c.incoming.DecodeElement(&f, nil)
	return &f
}

func (c *Conn) Next() (xml.StartElement, error) {
	var element xml.StartElement

	for {
		var err error
		var t xml.Token
		t, err = c.incoming.Token()
		if err != nil {
			return element, err
		}

		switch t := t.(type) {
		case xml.StartElement:
			element = t
			if element.Name.Local == "" {
				return element, errors.New("invalid xml response")
			}

			return element, nil
		}
	}
	panic("unreachable")
}

// Read a next XML element
func (c *Conn) NextElement() (topElem *Element, err error) {
	var elem *Element = nil
	for {
		var t xml.Token
		t, err = c.incoming.Token()
		if err != nil {
			return
		}

		switch t := t.(type) {
		case xml.StartElement:
			elem = NewElement(elem)
			if topElem == nil {
				topElem = elem
			}
			if t.Name.Local == "" {
				err = errors.New("invalid xml response")
				return
			}
			elem.StartElement = t
		case xml.EndElement:
			if elem == nil {
				continue
			}
			elem.EndElement = t
			elem = elem.Parent
			if elem == nil {
				// Top level elem ends
				return
			}
		case xml.CharData:
			if elem == nil {
				continue
			}
			elem.CharData = string(t)
		case xml.Comment:
			if elem == nil {
				continue
			}
			elem.Comment = t
		case xml.ProcInst:
			if elem == nil {
				continue
			}
			elem.ProcInst = t
		case xml.Directive:
			if elem == nil {
				continue
			}
			elem.Directive = t
		}
	}
	panic("unreachable")
}

func (c *Conn) Discover(from, to string) {
	fmt.Fprintf(c.outgoing, xmlIqGet, from, to, id(), NsDisco)
}

func (c *Conn) Body() string {
	b := new(body)
	c.incoming.DecodeElement(b, nil)
	return b.Body
}

func (c *Conn) Query() *query {
	q := new(query)
	c.incoming.DecodeElement(q, nil)
	return q
}

func (c *Conn) Presence(jid, pres string) {
	fmt.Fprintf(c.outgoing, xmlPresence, jid, pres)
}

func (c *Conn) MUCPresence(roomId, jid string) {
	fmt.Fprintf(c.outgoing, xmlMUCPresence, id(), roomId, jid, NsMuc)
}

func (c *Conn) MUCSend(to, from, body string) {
	fmt.Fprintf(c.outgoing, xmlMUCMessage, from, id(), to, html.EscapeString(body))
}

func (c *Conn) Send(to, from, body string) {
	fmt.Fprintf(c.outgoing, xmlMessage, from, id(), to, html.EscapeString(body))
}

func (c *Conn) Roster(from, to string) {
	fmt.Fprintf(c.outgoing, xmlIqGet, from, to, id(), NsIqRoster)
}

func (c *Conn) KeepAlive() {
	fmt.Fprintf(c.outgoing, " ")
}

func Dial(host string) (*Conn, error) {
	c := new(Conn)
	outgoing, err := net.Dial("tcp", host+":5222")

	if err != nil {
		return c, err
	}

	c.outgoing = outgoing
	c.incoming = xml.NewDecoder(outgoing)

	return c, nil
}

func ToMap(attr []xml.Attr) map[string]string {
	m := make(map[string]string)
	for _, a := range attr {
		m[a.Name.Local] = a.Value
	}

	return m
}

func id() string {
	b := make([]byte, 8)
	io.ReadFull(rand.Reader, b)
	return fmt.Sprintf("%x", b)
}
