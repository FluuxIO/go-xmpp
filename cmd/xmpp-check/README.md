# XMPP Check

XMPP check is a tool to check TLS certificate on a remote server.

## Installation

To install `xmpp-check` in your Go path:

```
$ go get -u gosrc.io/xmpp/...
```

## Usage

If you server is on standard port and XMPP domains matches the hostname you can simply use:

```
$ xmpp-check myhost.net
2019/05/16 16:04:36 All checks passed
```

You can also pass the port and the XMPP domain if different from the server hostname:

```
$ xmpp-check myhost.net:5222 xmppdomain.net
2019/05/16 16:05:21 All checks passed
```

Error code will be non-zero in case of error. You can thus use it directly with your usual 
monitoring scripts.
