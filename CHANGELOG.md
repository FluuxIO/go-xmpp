# Fluux XMPP Changelog

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