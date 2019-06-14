package xmpp

import (
	"strings"
)

/*
The XMPP router helps client and component developers select which XMPP they would like to process,
and associate processing code depending on the router configuration.

TODO: Automatically reply to IQ that do not match any route, to comply to XMPP standard.
*/

type Router struct {
	// Routes to be matched, in order.
	routes []*Route
}

// NewRouter returns a new router instance.
func NewRouter() *Router {
	return &Router{}
}

func (r *Router) Route(s Sender, p Packet) {
	var match RouteMatch
	if r.Match(p, &match) {
		match.Handler.HandlePacket(s, p)
	}
}

// NewRoute registers an empty routes
func (r *Router) NewRoute() *Route {
	route := &Route{}
	r.routes = append(r.routes, route)
	return route
}

func (r *Router) Match(p Packet, match *RouteMatch) bool {
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
func (r *Router) HandleFunc(name string, f func(s Sender, p Packet)) *Route {
	return r.NewRoute().Packet(name).HandlerFunc(f)
}

// ============================================================================
// Route
type Handler interface {
	HandlePacket(s Sender, p Packet)
}

type Route struct {
	handler Handler
	// Matchers are used to "specialize" routes and focus on specific packet features
	matchers []matcher
}

func (r *Route) Handler(handler Handler) *Route {
	r.handler = handler
	return r
}

// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as XMPP handlers. If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// Handler that calls f.
type HandlerFunc func(s Sender, p Packet)

// HandlePacket calls f(s, p)
func (f HandlerFunc) HandlePacket(s Sender, p Packet) {
	f(s, p)
}

// HandlerFunc sets a handler function for the route
func (r *Route) HandlerFunc(f HandlerFunc) *Route {
	return r.Handler(f)
}

// addMatcher adds a matcher to the route
func (r *Route) addMatcher(m matcher) *Route {
	r.matchers = append(r.matchers, m)
	return r
}

func (r *Route) Match(p Packet, match *RouteMatch) bool {
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

func (n nameMatcher) Match(p Packet, match *RouteMatch) bool {
	var name string
	// TODO: To avoid type switch everywhere in matching, I think we will need to have
	//    to move to a concrete type for packets, to make matching and comparison more natural.
	//    Current code structure is probably too rigid.
	// Maybe packet types should even be from an enum.
	switch p.(type) {
	case Message:
		name = "message"
	case IQ:
		name = "iq"
	case Presence:
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
	return r.addMatcher(nameMatcher(name))
}

// -------------------------
// Match on IQ and namespace

// nsIqMather matches on a list of IQ  payload namespaces
type nsIQMatcher []string

func (m nsIQMatcher) Match(p Packet, match *RouteMatch) bool {
	// TODO
	iq, ok := p.(IQ)
	if !ok {
		return false
	}
	if len(iq.Payload) < 1 {
		return false
	}
	return matchInArray(m, iq.Payload[0].Namespace())
}

// IQNamespaces adds an IQ matcher, expecting both an IQ and a
func (r *Route) IQNamespaces(namespaces ...string) *Route {
	for k, v := range namespaces {
		namespaces[k] = strings.ToLower(v)
	}
	return r.addMatcher(nsIQMatcher(namespaces))
}

// ============================================================================
// Matchers

// Matchers are used to "specialize" routes and focus on specific packet features
type matcher interface {
	Match(Packet, *RouteMatch) bool
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
