package utils

import (
	"encoding/json"
	"log"
	"time"

	"nostr-hero/src/types"

	"github.com/gorilla/websocket"
)

type Mailboxes struct {
	Read  []string
	Write []string
	Both  []string
}

// ToStringSlice combines Read, Write, and Both into a single []string.
func (r Mailboxes) ToStringSlice() []string {
	var urls []string
	urls = append(urls, r.Read...)
	urls = append(urls, r.Write...)
	urls = append(urls, r.Both...)
	return urls
}

func FetchUserRelays(publicKey string, relays []string) (*Mailboxes, error) {
	var mailbox Mailboxes

	for _, url := range relays {
		log.Printf("🔍 Connecting to relay: %s\n", url)

		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			log.Printf("❌ Failed to connect to WebSocket: %v\n", err)
			continue
		}

		subscriptionID := "sub-relay"

		filter := types.SubscriptionFilter{
			Authors: []string{publicKey},
			Kinds:   []int{10002}, // Kind 10002 corresponds to relay list (NIP-65)
		}

		subRequest := []interface{}{"REQ", subscriptionID, filter}
		requestJSON, err := json.Marshal(subRequest)
		if err != nil {
			log.Printf("❌ Failed to marshal subscription request: %v\n", err)
			conn.Close()
			return nil, err
		}

		log.Printf("📡 Sending subscription request to %s\n", url)
		if err := conn.WriteMessage(websocket.TextMessage, requestJSON); err != nil {
			log.Printf("❌ Failed to send subscription request: %v\n", err)
			conn.Close()
			return nil, err
		}

		conn.SetReadDeadline(time.Now().Add(WebSocketTimeout))

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("⚠️ Error reading from relay %s: %v\n", url, err)
				conn.Close()
				return nil, err
			}

			var response []interface{}
			if err := json.Unmarshal(message, &response); err != nil {
				log.Printf("❌ Failed to parse response from %s: %v\n", url, err)
				conn.Close()
				return nil, err
			}

			switch response[0] {
			case "EVENT":
				var event types.NostrEvent
				eventData, _ := json.Marshal(response[2])
				if err := json.Unmarshal(eventData, &event); err != nil {
					log.Printf("❌ Failed to parse event JSON from %s: %v\n", url, err)
					continue
				}

				log.Printf("📜 Received relay list event from %s: %+v\n", url, event)

				for _, tag := range event.Tags {
					if len(tag) > 1 && tag[0] == "r" {
						relayURL := tag[1]
						if len(tag) == 3 {
							switch tag[2] {
							case "read":
								mailbox.Read = append(mailbox.Read, relayURL)
							case "write":
								mailbox.Write = append(mailbox.Write, relayURL)
							}
						} else {
							mailbox.Both = append(mailbox.Both, relayURL)
						}
					}
				}

			case "EOSE":
				log.Printf("✅ Received EOSE from %s. Closing subscription...\n", url)

				// Send CLOSE message
				closeRequest := []interface{}{"CLOSE", subscriptionID}
				closeJSON, _ := json.Marshal(closeRequest)

				if err := conn.WriteMessage(websocket.TextMessage, closeJSON); err != nil {
					log.Printf("❌ Failed to send CLOSE message to %s: %v\n", url, err)
				}

				// Wait for "CLOSED" response with timeout
				closedChan := make(chan struct{})
				go func() {
					for {
						_, message, err := conn.ReadMessage()
						if err != nil {
							break
						}

						var resp []interface{}
						if err := json.Unmarshal(message, &resp); err != nil {
							break
						}

						if len(resp) > 1 && resp[0] == "CLOSED" && resp[1] == subscriptionID {
							log.Printf("🔌 Subscription closed on relay %s\n", url)
							closedChan <- struct{}{}
							return
						}
					}
				}()

				select {
				case <-closedChan:
					// Got a "CLOSED" response, safe to disconnect
					conn.Close()
				case <-time.After(1 * time.Second):
					// No "CLOSED" response, force disconnect
					log.Printf("⚠️ No CLOSED response from %s, disconnecting manually.\n", url)
					conn.Close()
				}

				return &mailbox, nil
			}
		}
	}

	if len(mailbox.Read) == 0 && len(mailbox.Write) == 0 && len(mailbox.Both) == 0 {
		return nil, nil
	}

	return &mailbox, nil
}

