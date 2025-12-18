package utils

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"nostr-hero/types"

	"github.com/gorilla/websocket"
)

const WebSocketTimeout = 2 * time.Second // Increased timeout


func FetchUserMetadata(publicKey string, relays []string) (*types.NostrEvent, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var latestEvent *types.NostrEvent
	var latestCreatedAt int64

	for _, url := range relays {
		wg.Add(1)

		go func(relayURL string) {
			defer wg.Done()
			log.Printf("🔍 Connecting to relay: %s\n", relayURL)

			conn, _, err := websocket.DefaultDialer.Dial(relayURL, nil)
			if err != nil {
				log.Printf("❌ WebSocket connection failed (%s): %v\n", relayURL, err)
				return
			}

			subscriptionID := "sub1"

			filter := types.SubscriptionFilter{
				Authors: []string{publicKey},
				Kinds:   []int{0}, // Kind 0 = Metadata
			}

			requestJSON, err := json.Marshal([]interface{}{"REQ", subscriptionID, filter})
			if err != nil {
				log.Printf("❌ Failed to marshal request: %v\n", err)
				conn.Close()
				return
			}

			log.Printf("📡 Sending request to %s: %s\n", relayURL, requestJSON)
			if err := conn.WriteMessage(websocket.TextMessage, requestJSON); err != nil {
				log.Printf("❌ Failed to send request to %s: %v\n", relayURL, err)
				conn.Close()
				return
			}

			conn.SetReadDeadline(time.Now().Add(WebSocketTimeout))

			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					log.Printf("⚠️ Error reading from relay %s: %v\n", relayURL, err)
					conn.Close()
					return
				}

				var response []interface{}
				if err := json.Unmarshal(message, &response); err != nil {
					log.Printf("❌ Failed to parse response from %s: %v\n", relayURL, err)
					conn.Close()
					return
				}

				switch response[0] {
				case "EVENT":
					var event types.NostrEvent
					eventData, _ := json.Marshal(response[2])
					if err := json.Unmarshal(eventData, &event); err != nil {
						log.Printf("❌ Failed to parse event JSON from %s: %v\n", relayURL, err)
						continue
					}

					log.Printf("📜 Received event from %s: %+v\n", relayURL, event)

					mu.Lock()
					if event.CreatedAt > latestCreatedAt {
						latestCreatedAt = event.CreatedAt
						latestEvent = &event
					}
					mu.Unlock()

				case "EOSE":
					log.Printf("✅ Received EOSE from %s. Closing subscription...\n", relayURL)

					// Send CLOSE message
					closeRequest := []interface{}{"CLOSE", subscriptionID}
					closeJSON, _ := json.Marshal(closeRequest)

					if err := conn.WriteMessage(websocket.TextMessage, closeJSON); err != nil {
						log.Printf("❌ Failed to send CLOSE message to %s: %v\n", relayURL, err)
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
								log.Printf("🔌 Subscription closed on relay %s\n", relayURL)
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
						log.Printf("⚠️ No CLOSED response from %s, disconnecting manually.\n", relayURL)
						conn.Close()
					}

					return
				}
			}
		}(url)
	}

	wg.Wait()

	if latestEvent == nil {
		log.Println("❌ No metadata events received.")
		return nil, nil
	}

	log.Printf("✅ Latest raw metadata event selected: %+v\n", latestEvent)
	return latestEvent, nil
}

