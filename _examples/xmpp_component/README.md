# xmpp_component

This component will connect to ejabberd and act as a subdomain "service" of your primary XMPP domain
(in that case localhost).

To be able to connect this component, you need to add a listener to your XMPP server.

Here is an example ejabberd configuration for that component listener:

```yaml
listen:
...
  -
    port: 8888
    module: ejabberd_service
    password: "mypass"
```

ejabberd will listen for a component (service) on port 8888 and allows it to connect using the
secret "mypass".