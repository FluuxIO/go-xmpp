# TODO

## check
### Features

- Use a config file to define the checks to perform as client on an XMPP server.

## send

### Issues

- Remove global variable (like mucToleave)
- Does not report error when trying to connect to a non open port (for example localhost with no server running).

### Features

- configuration
  - allow unencrypted
  - skip tls verification
- support muc and single user at same time
- send html -> parse console colors to xhtml (is there a easy way or lib for it ?)
