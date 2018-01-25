/*
Fluux XMPP is a Go XMPP library, focusing on simplicity, simple automation, and IoT.

The goal is to make simple to write simple adhoc XMPP clients:

 - For automation (like for example monitoring of an XMPP service),
 - For building connected "things" by plugging them on an XMPP server,
 - For writing simple chatbots to control a service or a thing.

Fluux XMPP can be used to build XMPP clients or XMPP components.

Clients

Fluux XMPP can be use to create fully interactive XMPP clients (for
example console-based), but it is more commonly used to build automated
clients (connected devices, automation scripts, chatbots, etc.).

Components

XMPP components can typically be used to extends the features of an XMPP
server, in a portable way, using component protocol over persistent TCP
connections.

Compliance

Fluux XMPP has been primarily tested with ejabberd (https://www.ejabberd.im)
but it should work with any XMPP compliant server.

*/
package xmpp
