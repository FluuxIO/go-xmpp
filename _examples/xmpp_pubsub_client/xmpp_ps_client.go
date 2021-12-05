package main

import (
	"context"
	"encoding/xml"
	"errors"
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

var invalidResp = errors.New("invalid response")

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
			log.Println("Received a message ! => \n" + string(data))
		})

	client, err := xmpp.NewClient(&config, router, func(err error) { log.Println(err) })
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	createNode(ctx, cancel, client)

	// ================================================================================
	// Configure the node. This can also be done in a single message with the creation
	configureNode(ctx, cancel, client)

	// ====================================
	// Subscribe to this node :
	subToNode(ctx, cancel, client)

	// ==========================
	// Publish to that node
	pubToNode(ctx, cancel, client)

	// =============================
	// Let's purge the node :
	purgeRq, _ := stanza.NewPurgeAllItems(serviceName, nodeName)
	purgeCh, err := client.SendIQ(ctx, purgeRq)
	if err != nil {
		log.Fatalf("could not send purge request: %v", err)
	}
	select {
	case purgeResp := <-purgeCh:

		if purgeResp.Type == stanza.IQTypeError {
			cancel()
			if vld, err := purgeResp.IsValid(); !vld {
				log.Fatalf(invalidResp.Error()+" %v"+" reason: %v", purgeResp, err)
			}
			log.Fatalf("error while purging node : %s", purgeResp.Error.Text)
		}
		log.Println("node successfully purged")
	case <-time.After(1000 * time.Millisecond):
		cancel()
		log.Fatal("No iq response was received in time while purging node")
	}

	cancel()
}

func createNode(ctx context.Context, cancel context.CancelFunc, client *xmpp.Client) {
	rqCreate, err := stanza.NewCreateNode(serviceName, nodeName)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	createCh, err := client.SendIQ(ctx, rqCreate)
	if err != nil {
		log.Fatalf("%+v", err)
	} else {

		if createCh != nil {
			select {
			case respCr := <-createCh:
				// Got response from server
				if respCr.Type == stanza.IQTypeError {
					if vld, err := respCr.IsValid(); !vld {
						log.Fatalf(invalidResp.Error()+" %+v"+" reason: %s", respCr, err)
					}
					if respCr.Error.Reason != "conflict" {
						log.Fatalf("%+v", respCr.Error.Text)
					}
					log.Println(respCr.Error.Text)
				} else {
					fmt.Print("successfully created channel")
				}
			case <-time.After(100 * time.Millisecond):
				cancel()
				log.Fatal("No iq response was received in time while creating node")
			}
		}
	}
}

func configureNode(ctx context.Context, cancel context.CancelFunc, client *xmpp.Client) {
	// First, ask for a form with the config options
	confRq, _ := stanza.NewConfigureNode(serviceName, nodeName)
	confReqCh, err := client.SendIQ(ctx, confRq)
	if err != nil {
		log.Fatalf("could not send iq : %v", err)
	}
	select {
	case confForm := <-confReqCh:
		// If the request was successful, we now have a form with configuration options to update
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

		// Send the modified fields as a form
		submitConf, err := stanza.NewFormSubmissionOwner(serviceName,
			nodeName,
			[]*stanza.Field{
				fields["pubsub#max_payload_size"],
				fields["pubsub#notification_type"],
			})

		c, _ := client.SendIQ(ctx, submitConf)
		select {
		case confResp := <-c:
			if confResp.Type == stanza.IQTypeError {
				cancel()
				if vld, err := confResp.IsValid(); !vld {
					log.Fatalf(invalidResp.Error()+" %v"+" reason: %v", confResp, err)
				}
				log.Fatalf("node configuration failed : %s", confResp.Error.Text)
			}
			log.Println("node configuration was successful")
			return

		case <-time.After(300 * time.Millisecond):
			cancel()
			log.Fatal("No iq response was received in time while configuring the node")
		}

	case <-time.After(300 * time.Millisecond):
		cancel()
		log.Fatal("No iq response was received in time while asking for the config form")
	}
}

func subToNode(ctx context.Context, cancel context.CancelFunc, client *xmpp.Client) {
	rqSubscribe, err := stanza.NewSubRq(serviceName, stanza.SubInfo{
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
			log.Println("Subscribed to the service")
		case <-time.After(300 * time.Millisecond):
			cancel()
			log.Fatal("No iq response was received in time while subscribing")
		}
	}
}

func pubToNode(ctx context.Context, cancel context.CancelFunc, client *xmpp.Client) {
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
			log.Println("Published item to the service")
		case <-time.After(300 * time.Millisecond):
			cancel()
			log.Fatal("No iq response was received in time while publishing")
		}
	}
}
