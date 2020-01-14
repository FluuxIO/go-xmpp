package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"gosrc.io/xmpp"
	"gosrc.io/xmpp/stanza"
	"log"
	"time"
)

const (
	userJID       = "testuser2@localhost"
	serverAddress = "localhost:5222"
	nodeName      = "lel_node"
	serviceName   = "pubsub.localhost"
)

func main() {

	config := xmpp.Config{
		TransportConfiguration: xmpp.TransportConfiguration{
			Address: serverAddress,
		},
		Jid:        userJID,
		Credential: xmpp.Password("pass123"),
		// StreamLogger: os.Stdout,
		Insecure: true,
	}
	router := xmpp.NewRouter()
	router.NewRoute().Packet("message").
		HandlerFunc(func(s xmpp.Sender, p stanza.Packet) {
			data, _ := xml.Marshal(p)
			fmt.Println("Received a publication ! => \n" + string(data))
		})

	client, err := xmpp.NewClient(config, router, func(err error) { fmt.Println(err) })
	if err != nil {
		log.Fatalf("%+v", err)
	}

	// ==========================
	// Client connection
	err = client.Connect()
	if err != nil {
		log.Fatalf("%+v", err)
	}

	// ==========================
	// Create a node
	rqCreate, err := stanza.NewCreateNode(serviceName, nodeName)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	createCh, err := client.SendIQ(ctx, rqCreate)
	if err != nil {
		log.Fatalf("%+v", err)
	} else {

		if createCh != nil {
			select {
			case respCr := <-createCh:
				// Got response from server
				if respCr.Error != nil {
					if respCr.Error.Reason != "conflict" {
						log.Fatalf("%+v", respCr.Error.Text)
					}
					fmt.Println(respCr.Error.Text)
				} else {
					fmt.Print("successfully created channel")
				}
			case <-time.After(100 * time.Millisecond):
				cancel()
				log.Fatal("No iq response was received in time")
			}
		}
	}

	// ====================================
	// Now let's subscribe to this node :
	rqSubscribe, _ := stanza.NewSubRq(serviceName, stanza.SubInfo{
		Node: nodeName,
		Jid:  userJID,
	})
	if err != nil {
		log.Fatalf("%+v", err)
	}
	subRespCh, _ := client.SendIQ(ctx, rqSubscribe)
	if subRespCh != nil {
		select {
		case <-subRespCh:
			fmt.Println("Subscribed to the service")
		case <-time.After(100 * time.Millisecond):
			cancel()
			log.Fatal("No iq response was received in time")
		}
	}

	// ==========================
	// Publish to that node
	pub, err := stanza.NewPublishItemRq(serviceName, nodeName, "", stanza.Item{
		Publisher: "testuser2",
		Any: &stanza.Node{
			XMLName: xml.Name{
				Space: "http://www.w3.org/2005/Atom",
				Local: "entry",
			},
			Nodes: []stanza.Node{
				{
					XMLName: xml.Name{Space: "", Local: "title"},
					Attrs:   nil,
					Content: "My pub item title",
					Nodes:   nil,
				},
				{
					XMLName: xml.Name{Space: "", Local: "summary"},
					Attrs:   nil,
					Content: "My pub item content summary",
					Nodes:   nil,
				},
				{
					XMLName: xml.Name{Space: "", Local: "link"},
					Attrs: []xml.Attr{
						{
							Name:  xml.Name{Space: "", Local: "rel"},
							Value: "alternate",
						},
						{
							Name:  xml.Name{Space: "", Local: "type"},
							Value: "text/html",
						},
						{
							Name:  xml.Name{Space: "", Local: "href"},
							Value: "http://denmark.lit/2003/12/13/atom03",
						},
					},
				},
				{
					XMLName: xml.Name{Space: "", Local: "id"},
					Attrs:   nil,
					Content: "My pub item content ID",
					Nodes:   nil,
				},
				{
					XMLName: xml.Name{Space: "", Local: "published"},
					Attrs:   nil,
					Content: "2003-12-13T18:30:02Z",
					Nodes:   nil,
				},
				{
					XMLName: xml.Name{Space: "", Local: "updated"},
					Attrs:   nil,
					Content: "2003-12-13T18:30:02Z",
					Nodes:   nil,
				},
			},
		},
	})

	if err != nil {
		log.Fatalf("%+v", err)
	}
	pubRespCh, _ := client.SendIQ(ctx, pub)
	if pubRespCh != nil {
		select {
		case <-pubRespCh:
			fmt.Println("Published item to the service")
		case <-time.After(100 * time.Millisecond):
			cancel()
			log.Fatal("No iq response was received in time")
		}
	}

	// =============================
	// Let's purge the node now :
	purgeRq, _ := stanza.NewPurgeAllItems(serviceName, nodeName)
	client.SendIQ(ctx, purgeRq)

	// =============================
	// Configure the node :
	confRq, _ := stanza.NewConfigureNode(serviceName, nodeName)
	confReqCh, err := client.SendIQ(ctx, confRq)
	select {
	case confForm := <-confReqCh:
		fields, err := confForm.GetFormFields()
		if err != nil {
			log.Fatal("No config fields found !")
		}

		// These are some common fields expected to be present. Change processing to your liking
		if fields["pubsub#max_payload_size"] != nil {
			fields["pubsub#max_payload_size"].ValuesList[0] = "100000"
		}

		if fields["pubsub#notification_type"] != nil {
			fields["pubsub#notification_type"].ValuesList[0] = "headline"
		}

		submitConf, err := stanza.NewFormSubmissionOwner(serviceName,
			nodeName,
			[]*stanza.Field{
				fields["pubsub#max_payload_size"],
				fields["pubsub#notification_type"],
			})

		c, _ := client.SendIQ(ctx, submitConf)
		select {
		case <-c:
			fmt.Println("node configuration was successful")
		case <-time.After(300 * time.Millisecond):
			cancel()
			log.Fatal("No iq response was received in time")

		}

	case <-time.After(300 * time.Millisecond):
		cancel()
		log.Fatal("No iq response was received in time")
	}

	cancel()
}
