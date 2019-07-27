# sendXMPP

sendxmpp is a tool to send messages from command-line.

## Installation

To install `sendxmpp` in your Go path:

```
$ go get -u gosrc.io/xmpp/cmd/sendxmpp
```

## Usage

```
$ sendxmpp --help
Usage:
  sendxmpp <recipient,> [message] [flags]

Examples:
sendxmpp to@chat.sum7.eu "Hello World!"

Flags:
      --addr string       host[:port]
      --config string     config file (default is ~/.config/fluxxmpp.yml)
  -h, --help              help for sendxmpp
      --jid string        using jid (required)
  -m, --muc               recipient is a muc (join it before sending messages)
      --password string   using password for your jid (required)
```


## Examples

Message from arguments:
```bash
$ sendxmpp to@example.org "Hello World!"
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
$  journalctl -f | sendxmpp to@example.org -
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
$ sendxmpp to1@example.org,to2@example.org "Multiple recipient"
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
journalctl -f | sendxmpp testit@conference.chat.sum7.eu - --muc
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

### Authentification

#### Configuration file

In `/etc/`, `~/.config` and `.` (here).
You could create the file name `fluxxmpp` with you favorite file extenion (e.g. `toml`, `yml`).

e.g. ~/.config/fluxxmpp.toml
```toml
jid      = "bot@example.org"
password = "secret"

addr     = "example.com:5222"
```

#### Environment variables

```bash
export FLUXXMPP_JID='bot@example.org';
export FLUXXMPP_PASSWORD='secret';

export FLUXXMPP_ADDR='example.com:5222';

sendxmpp to@example.org "Hello Welt";
```

#### Parameters

Warning: This should not be used for production systems, as all users on the system
can read the running processes, and their parameters (and thus the password).

```bash
sendxmpp to@example.org "Hello World!" --jid bot@example.org --password secret --addr example.com:5222;
```
