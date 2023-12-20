package trades

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
)

type RequestParams struct {
	Id     int      `json:"id"`
	Method string   `json:"method"`
	Params []string `json:"params"`
}

var conn *websocket.Conn

const (
	subscribeId   = 1
	unSubscribeId = 2
)

func getConnection() (*websocket.Conn, error) {
	if conn != nil {
		return conn, nil
	}

	u := url.URL{Scheme: "wss", Host: "stream.binance.com:9443", Path: "/ws"}
	log.Printf("Connecting to %s", u.String())
	c, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Printf("Handshake failed with status %d", resp.StatusCode)
		log.Fatal("Dial:", err)
	}

	return c, nil
}

func closeConnections() {
	err := conn.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func CreateConnection() (*websocket.Conn, error) {
	connection, err := getConnection()
	if err != nil {
		log.Fatalf("Failed to get connection %s", err.Error())
		return nil, err
	}

	return connection, nil
}

func unSubscribeAndClose(conn *websocket.Conn, tradeTopics []string) error {
	message := struct {
		Id     int      `json:"id"`
		Method string   `json:"method"`
		Params []string `json:"params"`
	}{
		Id:     unSubscribeId,
		Method: "UNSUBSCRIBE",
		Params: tradeTopics,
	}

	b, err := json.Marshal(message)
	if err != nil {
		log.Fatal("Failed to JSON Encode trade topics")
		return err
	}

	err = conn.WriteMessage(websocket.TextMessage, b)

	return nil
}

func SubscribeAndListen(topics []string) error {
	conn, err := getConnection()
	if err != nil {
		log.Fatalf("Failed to get connection %s", err.Error())
		return err
	}

	conn.SetPongHandler(func(appData string) error {
		fmt.Println("Received pong:", appData)
		pingFrame := []byte{1, 2, 3, 4, 5}
		err := conn.WriteMessage(websocket.PingMessage, pingFrame)
		if err != nil {
			fmt.Println(err)
		}
		return nil
	})

	tradeTopics := make([]string, 0, len(topics))
	for _, topic := range topics {
		tradeTopics = append(tradeTopics, topic+"@"+"aggTrade")
	}
	log.Println("Listening to trades for ", tradeTopics)

	message := RequestParams{
		Id:     subscribeId,
		Method: "SUBSCRIBE",
		Params: tradeTopics,
	}

	b, err := json.Marshal(message)
	if err != nil {
		log.Fatal("Failed to JSON Encode trade topics")
		return err
	}

	err = conn.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		log.Fatal("Failed to subscribe to topics " + err.Error())
		return err
	}

	defer func(conn *websocket.Conn, tradeTopics []string) {
		err := unSubscribeAndClose(conn, tradeTopics)
		if err != nil {
			log.Fatal(err)
		}
	}(conn, tradeTopics)
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(conn)

	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return err
		}

		trade := Ticker{}
		err = json.Unmarshal(payload, &trade)
		if err != nil {
			fmt.Println(err)
			return err
		}
		log.Println(trade.Symbol, trade.Price, trade.Quantity)
	}
}
