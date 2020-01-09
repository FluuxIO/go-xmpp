# Jukebox example  

## Requirements
- You need mpg123 installed on your computer because the example runs it as a command :
[Official MPG123 website](https://mpg123.de/)  
Most linux distributions have a package for it.  
- You need a soundcloud ID to play a music from the website through mpg123. You currently cannot play music files with this example.  
Your user ID is available in your account settings on the [soundcloud website](https://soundcloud.com/)  
**One is provided for convenience.**
- You need a running jabber server. You can run your local instance of [ejabberd](https://www.ejabberd.im/) for example. 
- You need a registered user on the running jabber server.

## Run
You can edit the soundcloud ID in the example file with your own, or use the provided one :
```go
const scClientID = "dde6a0075614ac4f3bea423863076b22"
``` 

To run the example, build it with (while in the example directory) :
```
go build xmpp_jukebox.go
```

then run it with (update the command arguments accordingly):
```
./xmpp_jukebox -jid=MY_USERE@MY_DOMAIN/jukebox -password=MY_PASSWORD -address=MY_SERVER:MY_SERVER_PORT
```
Make sure to have a resource, for instance "/jukebox", on your jid.

Then you can send the following stanza to "MY_USERE@MY_DOMAIN/jukebox" (with the resource) to play a song (update the soundcloud URL accordingly) :
```xml
<iq id="1" to="MY_USERE@MY_DOMAIN/jukebox" type="set">
    <set xml:lang="en" xmlns="urn:xmpp:iot:control">
        <string name="url" value="https://soundcloud.com/UPDATE/ME"/>
    </set>
</iq>
```   
