# Fluux XMPP Changelog

## v0.5.0

### Changes

- Added support for XEP-0198 (Stream management)
- Added message queue : when using "SendX" methods on a client, messages are also stored in a queue. When requesting
acks from the server, sent messages will be discarded, and unsent ones will be sent again. (see https://xmpp.org/extensions/xep-0198.html#acking)
- Added support for stanza_errors (see https://xmpp.org/rfcs/rfc3920.html#def C.2.  Stream error namespace and https://xmpp.org/rfcs/rfc6120.html#schemas-streamerror)
- Added separate hooks for connection and reconnection on the client. One can now specify different actions to get triggered on client connect 
and reconnect, at client init time.
- Client state update is now thread safe
- Changed the Config struct to use pointer semantics
- Tests
- Refactoring, including removing some Fprintf statements in favor of Marshal + Write and using structs from the library
instead of strings

## v0.4.0

### Changes

- Added support for XEP-0060 (PubSub)  
(no support for 6.5.4 Returning Some Items yet as it needs XEP-0059, Result Sets)
- Added support for XEP-0050 (Commands)
- Added support for XEP-0004 (Forms)
- Updated the client example with a TUI
- Make keepalive interval configurable #134
- Fix updating of EventManager.CurrentState #136
- Added callbacks for error management in Component and Client. Users must now provide a callback function when using NewClient/Component.
- Moved JID from xmpp package to stanza package

## v0.3.0

### Changes

- Update requirements to go1.13
- Add a websocket transport
- Add Client.SendIQ method
- Add IQ result routes to the Router
- Fix SIGSEGV in xmpp_component (#126)
- Add tests for Component and code style fixes

## v0.2.0

### Changes

- XMPP Over Websocket support
- Add support for getting IQ responses to client IQ queries (synchronously or asynchronously, passing an handler
  function).
- Implement X-OAUTH2 authentication method. You can read more details here:
  [Understanding ejabberd OAuth Support & Roadmap: Step 4](https://blog.process-one.net/understanding-ejabberd-oauth-support-roadmap/)
- Fix issues in the stanza builder when trying to add text inside and XMPP node.
- Fix issues with unescaped % characters in XMPP payload.

### Code migration guide

TODO
