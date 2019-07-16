# XMPP Check

XMPP check is a tool to check TLS certificate on a remote server.

## Installation

To install `xmpp-check` in your Go path:

```
$ go get -u gosrc.io/xmpp/cmd/xmpp-check
```

## Usage

```
$ xmpp-check --help
Usage:
  xmpp-check <host[:port]> [flags]

Examples:
xmpp-check chat.sum7.eu:5222 --domain meckerspace.de

Flags:
  -d, --domain string   domain if host handle multiple domains
  -h, --help            help for xmpp-check
```

If you server is on standard port and XMPP domains matches the hostname you can simply use:

```
$ xmpp-check chat.sum7.eu
 info All checks passed
   ⇢  address="chat.sum7.eu" domain=""
   ⇢  main.go:43 main.runCheck
   ⇢  2019-07-16T22:01:39.765+02:00
```

You can also pass the port and the XMPP domain if different from the server hostname:

```
$ xmpp-check chat.sum7.eu:5222 --domain meckerspace.de
 info All checks passed
   ⇢  address="chat.sum7.eu:5222" domain="meckerspace.de"
   ⇢  main.go:43 main.runCheck
   ⇢  2019-07-16T22:01:33.270+02:00
```

Error code will be non-zero in case of error. You can thus use it directly with your usual 
monitoring scripts.
