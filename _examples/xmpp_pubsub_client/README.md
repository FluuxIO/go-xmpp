# PubSub client example

## Description
This is a simple example of a client that :
* Creates a node on a service
* Subscribes to that node
* Publishes to that node
* Gets the notification from the publication and prints it on screen

## Requirements
You need to hve running jabber server, like [ejabberd](https://www.ejabberd.im/) that supports [XEP-0060](https://xmpp.org/extensions/xep-0060.html).

## How to use
Just run : 
```
    go run xmpp_ps_client.go
```