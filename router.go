package xmpp

import (
	"context"
	"encoding/xml"
	"strings"
	"sync"

	"gosrc.io/xmpp/stanza"
)

/*
The XMPP router helps client and component developers select which XMPP they would like to process,
and associate processing code depending on the router configuration.

Here are important rules to keep in mind while setting your routes and matchers:
- Routes are evaluated in the order they are set.
- When a route matches, it is executed and all others routes are ignored. For each packet, only a single
  route is executed.
- An empty route will match everything. Adding an empty route as the last route in your router will
  allow you to get all stanzas that did not match any previous route. You can for example use this to
  log all unexpected stanza received by your client or component.

TODO: Automatically reply to IQ that do not match any route, to comply to XMPP standard.
*/

type Router struct {
	// Routes to be matched, in order.
	routes []*Route

	IQResultRoutes    map[string]*IQResultRoute
	IQResultRouteLock sync.RWMutex
}

// NewRouter returns a new router instance.
func NewRouter() *Router {
	return &Router{
		IQResultRoutes: make(map[string]*IQResultRoute),
	}
}

// route is called by the XMPP client to dispatch stanza received using the set up routes.
// It is also used by test, but is not supposed to be used directly by users of the library.
func (r *Router) route(s Sender, p stanza.Packet) {
	a, isA := p.(stanza.SMAnswer)
	if isA {
		switch tt := s.(type) {
		case *Client:
			lastAcked := a.H
			SendMissingStz(int(lastAcked), s, tt.Session.SMState.UnAckQueue)
		case *Component:
		// TODO
		default:
		}
	}
	iq, isIq := p.(*stanza.IQ)
	if isIq {
		r.IQResultRouteLock.RLock()
		route, ok := r.IQResultRoutes[iq.Id]
		r.IQResultRouteLock.RUnlock()
		if ok {
			r.IQResultRouteLock.Lock()
			delete(r.IQResultRoutes, iq.Id)
			r.IQResultRouteLock.Unlock()
			route.result <- *iq
			close(route.result)
			return
		}
		if iq.Any != nil && iq.Any.XMLName.Local == "ping" {
			_ = s.Send(&stanza.IQ{Attrs: stanza.Attrs{
				Id:   iq.Id,
				Type: stanza.IQTypeResult,
				From: iq.To,
				To:   iq.From,
			}})
			return
		}
	}

	var match RouteMatch
	if r.Match(p, &match) {
		// If we match, route the packet
		match.Handler.HandlePacket(s, p)
		return
	}

	// If there is no match and we receive an iq set or get, we need to send a reply
	if isIq && (iq.Type == stanza.IQTypeGet || iq.Type == stanza.IQTypeSet) {
		iqNotImplemented(s, iq)
	}
}

// SendMissingStz sends all stanzas that did not reach the server, according to the response to an ack request (see XEP-0198, acks)
func SendMissingStz(lastSent int, s Sender, uaq *stanza.UnAckQueue) error {
	uaq.RWMutex.Lock()
	if len(uaq.Uslice) <= 0 {
		uaq.RWMutex.Unlock()
		return nil
	}
	last := uaq.Uslice[len(uaq.Uslice)-1]
	if last.Id > lastSent {
		// Remove sent stanzas from the queue
		uaq.PopN(lastSent - last.Id)
		// Re-send non acknowledged stanzas
		for _, elt := range uaq.PopN(len(uaq.Uslice)) {
			eltStz := elt.(*stanza.UnAckedStz)
			err := s.SendRaw(eltStz.Stz)
			if err != nil {
				return err
			}

		}
		// Ask for updates on stanzas we just sent to the entity. Not sure I should leave this. Maybe let users call ack again by themselves ?
		s.Send(stanza.SMRequest{})
	}
	uaq.RWMutex.Unlock()
	return nil
}

func iqNotImplemented(s Sender, iq *stanza.IQ) {
	err := stanza.Err{
		XMLName: xml.Name{Local: "error"},
		Code:    501,
		Type:    "cancel",
		Reason:  "feature-not-implemented",
	}
	reply := iq.MakeError(err)
	_ = s.Send(reply)
}

// NewRoute registers an empty routes
func (r *Router) NewRoute() *Route {
	route := &Route{}
	r.routes = append(r.routes, route)
	return route
}

// NewIQResultRoute register a route that will catch an IQ result stanza with
// the given Id. The route will only match ones, after which it will automatically
// be unregistered
func (r *Router) NewIQResultRoute(ctx context.Context, id string) chan stanza.IQ {
	route := NewIQResultRoute(ctx)
	r.IQResultRouteLock.Lock()
	r.IQResultRoutes[id] = route
	r.IQResultRouteLock.Unlock()

	// Start a go function to make sure the route is unregistered when the context
	// is done.
	go func() {
		<-route.context.Done()
		r.IQResultRouteLock.Lock()
		delete(r.IQResultRoutes, id)
		r.IQResultRouteLock.Unlock()
	}()

	return route.result
}

func (r *Router) Match(p stanza.Packet, match *RouteMatch) bool {
	for _, route := range r.routes {
		if route.Match(p, match) {
			return true
		}
	}
	return false
}

// Handle registers a new route with a matcher for a given packet name (iq, message, presence)
// See Route.Packet() and Route.Handler().
func (r *Router) Handle(name string, handler Handler) *Route {
	return r.NewRoute().Packet(name).Handler(handler)
}

// HandleFunc registers a new route with a matcher for for a given packet name (iq, message, presence)
// See Route.Path() and Route.HandlerFunc().
func (r *Router) HandleFunc(name string, f func(s Sender, p stanza.Packet)) *Route {
	return r.NewRoute().Packet(name).HandlerFunc(f)
}

// ============================================================================

// TimeoutHandlerFunc is a function type for handling IQ result timeouts.
type TimeoutHandlerFunc func(err error)

// IQResultRoute is a temporary route to match IQ result stanzas
type IQResultRoute struct {
	context context.Context
	result  chan stanza.IQ
}

// NewIQResultRoute creates a new IQResultRoute instance
func NewIQResultRoute(ctx context.Context) *IQResultRoute {
	return &IQResultRoute{
		context: ctx,
		result:  make(chan stanza.IQ),
	}
}

// ============================================================================
// IQ result handler

// IQResultHandler is a utility interface for IQ result handlers
type IQResultHandler interface {
	HandleIQ(ctx context.Context, s Sender, iq stanza.IQ)
}

// IQResultHandlerFunc is an adapter to allow using functions as IQ result handlers.
type IQResultHandlerFunc func(ctx context.Context, s Sender, iq stanza.IQ)

// HandleIQ is a proxy function to implement IQResultHandler using a function.
func (f IQResultHandlerFunc) HandleIQ(ctx context.Context, s Sender, iq stanza.IQ) {
	f(ctx, s, iq)
}

// ============================================================================
// Route

type Handler interface {
	HandlePacket(s Sender, p stanza.Packet)
}

type Route struct {
	handler Handler
	// Matchers are used to "specialize" routes and focus on specific packet features
	matchers []Matcher
}

func (r *Route) Handler(handler Handler) *Route {
	r.handler = handler
	return r
}

// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as XMPP handlers. If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// Handler that calls f.
type HandlerFunc func(s Sender, p stanza.Packet)

// HandlePacket calls f(s, p)
func (f HandlerFunc) HandlePacket(s Sender, p stanza.Packet) {
	f(s, p)
}

// HandlerFunc sets a handler function for the route
func (r *Route) HandlerFunc(f HandlerFunc) *Route {
	return r.Handler(f)
}

// AddMatcher adds a matcher to the route
func (r *Route) AddMatcher(m Matcher) *Route {
	r.matchers = append(r.matchers, m)
	return r
}

func (r *Route) Match(p stanza.Packet, match *RouteMatch) bool {
	for _, m := range r.matchers {
		if matched := m.Match(p, match); !matched {
			return false
		}
	}

	// We have a match, let's pass info route match info
	match.Route = r
	match.Handler = r.handler
	return true
}

// --------------------
// Match on packet name

type nameMatcher string

func (n nameMatcher) Match(p stanza.Packet, match *RouteMatch) bool {
	var name string
	// TODO: To avoid type switch everywhere in matching, I think we will need to have
	//    to move to a concrete type for packets, to make matching and comparison more natural.
	//    Current code structure is probably too rigid.
	// Maybe packet types should even be from an enum.
	switch p.(type) {
	case stanza.Message:
		name = "message"
	case *stanza.IQ:
		name = "iq"
	case stanza.Presence:
		name = "presence"
	}
	if name == string(n) {
		return true
	}
	return false
}

// Packet matches on a packet name (iq, message, presence, ...)
// It matches on the Local part of the xml.Name
func (r *Route) Packet(name string) *Route {
	name = strings.ToLower(name)
	return r.AddMatcher(nameMatcher(name))
}

// -------------------------
// Match on stanza type

// nsTypeMather matches on a list of IQ  payload namespaces
type nsTypeMatcher []string

func (m nsTypeMatcher) Match(p stanza.Packet, match *RouteMatch) bool {
	var stanzaType stanza.StanzaType
	switch packet := p.(type) {
	case *stanza.IQ:
		stanzaType = packet.Type
	case stanza.Presence:
		stanzaType = packet.Type
	case stanza.Message:
		if packet.Type == "" {
			// optional on message, normal is the default type
			stanzaType = "normal"
		} else {
			stanzaType = packet.Type
		}
	default:
		return false
	}
	return matchInArray(m, string(stanzaType))
}

// IQNamespaces adds an IQ matcher, expecting both an IQ and a
func (r *Route) StanzaType(types ...string) *Route {
	for k, v := range types {
		types[k] = strings.ToLower(v)
	}
	return r.AddMatcher(nsTypeMatcher(types))
}

// -------------------------
// Match on IQ and namespace

// nsIqMather matches on a list of IQ  payload namespaces
type nsIQMatcher []string

func (m nsIQMatcher) Match(p stanza.Packet, match *RouteMatch) bool {
	iq, ok := p.(*stanza.IQ)
	if !ok {
		return false
	}
	if iq.Payload == nil {
		return false
	}
	return matchInArray(m, iq.Payload.Namespace())
}

// IQNamespaces adds an IQ matcher, expecting both an IQ and a
func (r *Route) IQNamespaces(namespaces ...string) *Route {
	for k, v := range namespaces {
		namespaces[k] = strings.ToLower(v)
	}
	return r.AddMatcher(nsIQMatcher(namespaces))
}

// ============================================================================
// Matchers

// Matchers are used to "specialize" routes and focus on specific packet features.
// You can register attach them to a route via the AddMatcher method.
type Matcher interface {
	Match(stanza.Packet, *RouteMatch) bool
}

// RouteMatch extracts and gather match information
type RouteMatch struct {
	Route   *Route
	Handler Handler
}

// matchInArray is a generic matching function to check if a string is a list
// of specific function
func matchInArray(arr []string, value string) bool {
	for _, str := range arr {
		if str == value {
			return true
		}
	}
	return false
}
