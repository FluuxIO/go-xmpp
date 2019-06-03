package xmpp

import (
	"reflect"
	"sync"
)

type MsgExtension interface{}

// The Registry for msg and IQ types is a global variable.
// TODO: Move to the client init process to remove the dependency on a global variable.
//   That should make it possible to be able to share the decoder.
// TODO: Ensure that a client can add its own custom namespace to the registry (or overload existing ones).
var typeRegistry = newRegistry()

type namespace = string

type registry struct {
	// Key is namespace of message extension
	msgTypes     map[namespace]reflect.Type
	msgTypesLock *sync.RWMutex

	iqTypes map[namespace]reflect.Type
}

func newRegistry() registry {
	return registry{
		msgTypes:     make(map[namespace]reflect.Type),
		msgTypesLock: &sync.RWMutex{},
		iqTypes:      make(map[namespace]reflect.Type),
	}
}

// Mutexes are not needed when adding a Message or IQ extension in init function.
// However, forcing the use of the mutex protect the data structure against unexpected use
// of the registry by developers using the library.
func (r registry) RegisterMsgExt(namespace string, extension MsgExtension) {
	r.msgTypesLock.Lock()
	defer r.msgTypesLock.Unlock()
	r.msgTypes[namespace] = reflect.TypeOf(extension)
}

func (r registry) getMsgExtType(namespace string) reflect.Type {
	r.msgTypesLock.RLock()
	defer r.msgTypesLock.RUnlock()
	return r.msgTypes[namespace]
}

func (r registry) getmsgType(namespace string) MsgExtension {
	if extensionType := r.getMsgExtType(namespace); extensionType != nil {
		val := reflect.New(extensionType)
		elt := val.Interface()
		if msgExt, ok := elt.(MsgExtension); ok {
			return msgExt
		}
	}
	return nil
}

// Registry to support message extensions
//var msgTypeRegistry = make(map[string]reflect.Type)

// Registry to instantiate the right IQ payload element
// Key is namespace and key of the payload
var iqTypeRegistry = make(map[string]reflect.Type)
