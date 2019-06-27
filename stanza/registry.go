package stanza

import (
	"encoding/xml"
	"reflect"
	"sync"
)

type MsgExtension interface{}
type PresExtension interface{}

// The Registry for msg and IQ types is a global variable.
// TODO: Move to the client init process to remove the dependency on a global variable.
//   That should make it possible to be able to share the decoder.
// TODO: Ensure that a client can add its own custom namespace to the registry (or overload existing ones).

type PacketType uint8

const (
	PKTPresence PacketType = iota
	PKTMessage
	PKTIQ
)

var TypeRegistry = newRegistry()

// We store different registries per packet type and namespace.
type registryKey struct {
	packetType PacketType
	namespace  string
}

type registryForNamespace map[string]reflect.Type

type registry struct {
	// We store different registries per packet type and namespace.
	msgTypes map[registryKey]registryForNamespace
	// Handle concurrent access
	msgTypesLock *sync.RWMutex
}

func newRegistry() *registry {
	return &registry{
		msgTypes:     make(map[registryKey]registryForNamespace),
		msgTypesLock: &sync.RWMutex{},
	}
}

// MapExtension stores extension type for packet payload.
// The match is done per PacketType (iq, message, or presence) and XML tag name.
// You can use the alias "*" as local XML name to be able to match all unknown tag name for that
// packet type and namespace.
func (r *registry) MapExtension(pktType PacketType, name xml.Name, extension MsgExtension) {
	key := registryKey{pktType, name.Space}
	r.msgTypesLock.RLock()
	store := r.msgTypes[key]
	r.msgTypesLock.RUnlock()

	r.msgTypesLock.Lock()
	defer r.msgTypesLock.Unlock()
	if store == nil {
		store = make(map[string]reflect.Type)
	}
	store[name.Local] = reflect.TypeOf(extension)
	r.msgTypes[key] = store
}

// GetExtensionType returns extension type for packet payload, based on packet type and tag name.
func (r *registry) GetExtensionType(pktType PacketType, name xml.Name) reflect.Type {
	key := registryKey{pktType, name.Space}

	r.msgTypesLock.RLock()
	defer r.msgTypesLock.RUnlock()
	store := r.msgTypes[key]
	result := store[name.Local]
	if result == nil && name.Local != "*" {
		return store["*"]
	}
	return result
}

// GetPresExtension returns an instance of PresExtension, by matching packet type and XML
// tag name against the registry.
func (r *registry) GetPresExtension(name xml.Name) PresExtension {
	if extensionType := r.GetExtensionType(PKTPresence, name); extensionType != nil {
		val := reflect.New(extensionType)
		elt := val.Interface()
		if presExt, ok := elt.(PresExtension); ok {
			return presExt
		}
	}
	return nil
}

// GetMsgExtension returns an instance of MsgExtension, by matching packet type and XML
// tag name against the registry.
func (r *registry) GetMsgExtension(name xml.Name) MsgExtension {
	if extensionType := r.GetExtensionType(PKTMessage, name); extensionType != nil {
		val := reflect.New(extensionType)
		elt := val.Interface()
		if msgExt, ok := elt.(MsgExtension); ok {
			return msgExt
		}
	}
	return nil
}

// GetIQExtension returns an instance of IQPayload, by matching packet type and XML
// tag name against the registry.
func (r *registry) GetIQExtension(name xml.Name) IQPayload {
	if extensionType := r.GetExtensionType(PKTIQ, name); extensionType != nil {
		val := reflect.New(extensionType)
		elt := val.Interface()
		if iqExt, ok := elt.(IQPayload); ok {
			return iqExt
		}
	}
	return nil
}
