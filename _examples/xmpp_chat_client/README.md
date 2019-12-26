# Chat TUI example
This is a simple chat example, with a TUI.   
It shows the library usage and a few of its capabilities. 
## How to run
### Build
You can build the client using : 
```
    go build -o example_client
```
and then run with (on unix for example): 
```
    ./example_client
```
or you can simply build + run in one command while at the example directory root, like this:
```
    go run xmpp_chat_client.go interface.go 
```

### Configuration
The example needs a configuration file to run. A sample file is provided.  
By default, the example will look for a file named "config" in the current directory.
To provide a different configuration file, pass the following argument to the example :
```
    go run xmpp_chat_client.go interface.go -c /path/to/config
```
where /path/to/config is the path to the directory containing the configuration file. The configuration file must be named 
"config" and be using the yaml format.  
  
Required fields are :
```yaml
Server :
  - full_address: "localhost:5222"
Client : # This is you
  - jid: "testuser2@localhost"
  - pass: "pass123" #Password in a config file yay

# Contacts list, ";" separated
Contacts : "testuser1@localhost;testuser3@localhost"
# Should we log stanzas ?  
LogStanzas:
  - logger_on: "true"
  - logfile_path: "./logs" # Path to directory, not file.
```

## How to use
Shortcuts :   
    - ctrl+space  : switch between input window and menu window.  
    - While in input window :  
        - enter : sends a message if in message mode (see menu options)  
        - ctrl+e : sends a raw stanza when in raw mode (see menu options)    
    - ctrl+c : quit  