/*
Fluux XMPP is an modern and full-featured XMPP library that can be used to build clients or
server components.

The goal is to make simple to write modern compliant XMPP software:

 - For automation (like for example monitoring of an XMPP service),
 - For building connected "things" by plugging them on an XMPP server,
 - For writing simple chatbots to control a service or a thing.
 - For writing XMPP servers components. Fluux XMPP supports:
    - XEP-0114: Jabber Component Protocol
    - XEP-0355: Namespace Delegation
    - XEP-0356: Privileged Entity

The library is designed to have minimal dependencies. For now, the library does not depend on any other library.

The library includes a StreamManager that provides features like autoreconnect exponential back-off.

The library is implementing latest versions of the XMPP specifications (RFC 6120 and RFC 6121), and includes
support for many extensions.

Clients

Fluux XMPP can be use to create fully interactive XMPP clients (for
example console-based), but it is more commonly used to build automated
clients (connected devices, automation scripts, chatbots, etc.).

Components

XMPP components can typically be used to extends the features of an XMPP
server, in a portable way, using component protocol over persistent TCP
connections.

Component protocol is defined in XEP-114 (https://xmpp.org/extensions/xep-0114.html).

Compliance

Fluux XMPP has been primarily tested with ejabberd (https://www.ejabberd.im)
but it should work with any XMPP compliant server.

*/
package xmpp
