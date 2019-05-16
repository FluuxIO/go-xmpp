# XMPP Check

XMPP check is a tool to check TLS certificate on a remote server.

## Installation

```
$ go get -u gosrc.io/xmpp/cmd/xmpp-check
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
