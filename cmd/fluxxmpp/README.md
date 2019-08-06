# fluxxmpp

fluxxIO's xmpp comandline tool

## Installation

To install `fluxxmpp` in your Go path:

```
$ go get -u gosrc.io/xmpp/cmd/fluxxmpp
```

## Usage

```
$ fluxxmpp --help
fluxxIO's xmpp comandline tool

Usage:
  fluxxmpp [command]

Available Commands:
  check       is a command-line to check if you XMPP TLS certificate is valid and warn you before it expires
  help        Help about any command
  send        is a command-line tool to send to send XMPP messages to users

Flags:
  -h, --help   help for fluxxmpp

Use "fluxxmpp [command] --help" for more information about a command.
```

### check tls

```
$ fluxxmpp check --help
is a command-line to check if you XMPP TLS certificate is valid and warn you before it expires

Usage:
  fluxxmpp check <host[:port]> [flags]

Examples:
fluxxmpp check chat.sum7.eu:5222 --domain meckerspace.de

Flags:
  -d, --domain string   domain if host handle multiple domains
  -h, --help            help for check
```

### sending messages

```
$ fluxxmpp send --help
is a command-line tool to send to send XMPP messages to users

Usage:
  fluxxmpp send <recipient,> [message] [flags]

Examples:
fluxxmpp send to@chat.sum7.eu "Hello World!"

Flags:
      --addr string       host[:port]
      --config string     config file (default is ~/.config/fluxxmpp.yml)
  -h, --help              help for send
      --jid string        using jid (required)
  -m, --muc               recipient is a muc (join it before sending messages)
      --password string   using password for your jid (required)
```


## Examples

### check tls

If you server is on standard port and XMPP domains matches the hostname you can simply use:

```
$ fluxxmpp check chat.sum7.eu
 info All checks passed
   ⇢  address="chat.sum7.eu" domain=""
   ⇢  main.go:43 main.runCheck
   ⇢  2019-07-16T22:01:39.765+02:00
```

You can also pass the port and the XMPP domain if different from the server hostname:

```
$ fluxxmpp check chat.sum7.eu:5222 --domain meckerspace.de
 info All checks passed
   ⇢  address="chat.sum7.eu:5222" domain="meckerspace.de"
   ⇢  main.go:43 main.runCheck
   ⇢  2019-07-16T22:01:33.270+02:00
```

Error code will be non-zero in case of error. You can thus use it directly with your usual 
monitoring scripts.


### sending messages

Message from arguments:
```bash
$ fluxxmpp send to@example.org "Hello World!"
 info client connected
   ⇢  cmd.go:56 main.glob..func1.1
   ⇢  2019-07-17T23:42:43.310+02:00
 info send message
   ⇢  muc=false text="Hello World!" to="to@example.org"
   ⇢  send.go:31 main.send
   ⇢  2019-07-17T23:42:43.310+02:00
```

Message from STDIN:
```bash
$  journalctl -f | fluxxmpp send to@example.org -
 info client connected
   ⇢  cmd.go:56 main.glob..func1.1
   ⇢  2019-07-17T23:40:03.177+02:00
 info send message
   ⇢  muc=false text="-- Logs begin at Mon 2019-07-08 22:16:54 CEST. --" to="to@example.org"
   ⇢  send.go:31 main.send
   ⇢  2019-07-17T23:40:03.178+02:00
 info send message
   ⇢  muc=false text="Jul 17 23:36:46 RECHNERNAME systemd[755]: Started Fetch mails." to="to@example.org"
   ⇢  send.go:31 main.send
   ⇢  2019-07-17T23:40:03.178+02:00
^C
```


Multiple recipients:
```bash
$ fluxxmpp send to1@example.org,to2@example.org "Multiple recipient"
 info client connected
   ⇢  cmd.go:56 main.glob..func1.1
   ⇢  2019-07-17T23:47:57.650+02:00
 info send message
   ⇢  muc=false text="Multiple recipient" to="to1@example.org"
   ⇢  send.go:31 main.send
   ⇢  2019-07-17T23:47:57.651+02:00
 info send message
   ⇢  muc=false text="Multiple recipient" to="to2@example.org"
   ⇢  send.go:31 main.send
   ⇢  2019-07-17T23:47:57.652+02:00
```

Send to MUC:
```bash
journalctl -f | fluxxmpp send testit@conference.chat.sum7.eu - --muc
 info client connected
   ⇢  cmd.go:56 main.glob..func1.1
   ⇢  2019-07-17T23:52:56.269+02:00
 info send message
   ⇢  muc=true text="-- Logs begin at Mon 2019-07-08 22:16:54 CEST. --" to="testit@conference.chat.sum7.eu"
   ⇢  send.go:31 main.send
   ⇢  2019-07-17T23:52:56.270+02:00
 info send message
   ⇢  muc=true text="Jul 17 23:48:58 RECHNERNAME systemd[755]: mail.service: Succeeded." to="testit@conference.chat.sum7.eu"
   ⇢  send.go:31 main.send
   ⇢  2019-07-17T23:52:56.277+02:00
^C
```

## Authentification

### Configuration file

In `/etc/`, `~/.config` and `.` (here).
You could create the file name `fluxxmpp` with you favorite file extenion (e.g. `toml`, `yml`).

e.g. ~/.config/fluxxmpp.toml
```toml
jid      = "bot@example.org"
password = "secret"

addr     = "example.com:5222"
```

### Environment variables

```bash
export FLUXXMPP_JID='bot@example.org';
export FLUXXMPP_PASSWORD='secret';

export FLUXXMPP_ADDR='example.com:5222';

fluxxmpp send to@example.org "Hello Welt";
```

### Parameters

Warning: This should not be used for production systems, as all users on the system
can read the running processes, and their parameters (and thus the password).

```bash
fluxxmpp send to@example.org "Hello World!" --jid bot@example.org --password secret --addr example.com:5222;
```
