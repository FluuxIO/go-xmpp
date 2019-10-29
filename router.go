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

	iqResultRoutes    map[string]*IqResultRoute
	iqResultRouteLock sync.RWMutex
}

// NewRouter returns a new router instance.
func NewRouter() *Router {
	return &Router{
		iqResultRoutes: make(map[string]*IqResultRoute),
	}
}

// route is called by the XMPP client to dispatch stanza received using the set up routes.
// It is also used by test, but is not supposed to be used directly by users of the library.
func (r *Router) route(s Sender, p stanza.Packet) {
	iq, isIq := p.(stanza.IQ)
	if isIq {
		r.iqResultRouteLock.RLock()
		route, ok := r.iqResultRoutes[iq.Id]
		r.iqResultRouteLock.RUnlock()
		if ok {
			r.iqResultRouteLock.Lock()
			delete(r.iqResultRoutes, iq.Id)
			r.iqResultRouteLock.Unlock()
			close(route.matched)
			route.handler.HandlePacket(s, p)
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

func iqNotImplemented(s Sender, iq stanza.IQ) {
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

// NewIqResultRoute register a route that will catch an IQ result stanza with
// the given Id. The route will only match ones, after which it will automatically
// be unregistered
func (r *Router) NewIqResultRoute(ctx context.Context, id string) *IqResultRoute {
	route := &IqResultRoute{
		context: ctx,
		matched: make(chan struct{}),
	}
	r.iqResultRouteLock.Lock()
	r.iqResultRoutes[id] = route
	r.iqResultRouteLock.Unlock()
	go func() {
		select {
		case <-route.context.Done():
			r.iqResultRouteLock.Lock()
			delete(r.iqResultRoutes, id)
			r.iqResultRouteLock.Unlock()
			if route.timeoutHandler != nil {
				route.timeoutHandler(route.context.Err())
			}
		case <-route.matched:
		}
	}()
	return route
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

// HandleIqResult register a temporary route
func (r *Router) HandleIqResult(id string, handler Handler) *IqResultRoute {
	return r.NewIqResultRoute(context.Background(), id).Handler(handler)
}

func (r *Router) HandleFuncIqResult(id string, f func(s Sender, p stanza.Packet)) *IqResultRoute {
	return r.NewIqResultRoute(context.Background(), id).HandlerFunc(f)
}

// ============================================================================
// IqResultRoute
type TimeoutHandlerFunc func(err error)

type IqResultRoute struct {
	context        context.Context
	matched        chan struct{}
	handler        Handler
	timeoutHandler TimeoutHandlerFunc
}

func (r *IqResultRoute) Handler(handler Handler) *IqResultRoute {
	r.handler = handler
	return r
}

func (r *IqResultRoute) HandlerFunc(f HandlerFunc) *IqResultRoute {
	return r.Handler(f)
}

func (r *IqResultRoute) TimeoutHandlerFunc(f TimeoutHandlerFunc) *IqResultRoute {
	r.timeoutHandler = f
	return r
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
	case stanza.IQ:
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
	case stanza.IQ:
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
	iq, ok := p.(stanza.IQ)
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
