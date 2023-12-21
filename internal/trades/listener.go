package trades

import (
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/sefikcan/read-time-trade/pkg/config"
	kafkaClient "github.com/sefikcan/read-time-trade/pkg/kafka"
	"github.com/sefikcan/read-time-trade/pkg/logger"
	"github.com/segmentio/kafka-go"
	"net/url"
	"strconv"
	"strings"
)

type TradeListener interface {
	GetConnection() (*websocket.Conn, error)
	CloseConnection()
	CreateConnection() (*websocket.Conn, error)
	UnSubscribeAndClose(conn *websocket.Conn, tradeTopics []string) error
	SubscribeAndListen(topics []string)
}

type tradeListener struct {
	log           logger.Logger
	cfg           *config.Config
	kafkaProducer kafkaClient.Producer
}

func NewTradeListener(log logger.Logger, cfg *config.Config, kafkaProducer kafkaClient.Producer) *tradeListener {
	return &tradeListener{
		log:           log,
		cfg:           cfg,
		kafkaProducer: kafkaProducer,
	}
}

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

func (l *tradeListener) GetConnection() (*websocket.Conn, error) {
	if conn != nil {
		return conn, nil
	}

	u := url.URL{Scheme: "wss", Host: "stream.binance.com:9443", Path: "/ws"}
	l.log.Infof("Connecting to %s", u.String())
	c, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		l.log.Infof("Handshake failed with status %d", resp.StatusCode)
		l.log.Fatal("Dial:", err)
	}

	return c, nil
}

func (l *tradeListener) CloseConnections() {
	err := conn.Close()
	if err != nil {
		l.log.Fatal(err)
	}
}

func (l *tradeListener) CreateConnection() (*websocket.Conn, error) {
	connection, err := l.GetConnection()
	if err != nil {
		l.log.Fatalf("Failed to get connection %s", err.Error())
		return nil, err
	}

	return connection, nil
}

func (l *tradeListener) UnSubscribeAndClose(conn *websocket.Conn, tradeTopics []string) error {
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
		l.log.Fatal("Failed to JSON Encode trade topics")
		return err
	}

	err = conn.WriteMessage(websocket.TextMessage, b)

	return nil
}

func (l *tradeListener) SubscribeAndListen(topics []string) error {
	conn, err := l.GetConnection()
	if err != nil {
		l.log.Fatalf("Failed to get connection %s", err.Error())
		return err
	}

	conn.SetPongHandler(func(appData string) error {
		l.log.Info("Received pong:", appData)
		pingFrame := []byte{1, 2, 3, 4, 5}
		err := conn.WriteMessage(websocket.PingMessage, pingFrame)
		if err != nil {
			l.log.Fatal(err)
		}
		return nil
	})

	tradeTopics := make([]string, 0, len(topics))
	for _, topic := range topics {
		tradeTopics = append(tradeTopics, topic+"@"+"aggTrade")
	}
	l.log.Info("Listening to trades for ", tradeTopics)

	message := RequestParams{
		Id:     subscribeId,
		Method: "SUBSCRIBE",
		Params: tradeTopics,
	}

	b, err := json.Marshal(message)
	if err != nil {
		l.log.Fatal("Failed to JSON Encode trade topics")
		return err
	}

	err = conn.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		l.log.Fatal("Failed to subscribe to topics " + err.Error())
		return err
	}

	defer func(conn *websocket.Conn, tradeTopics []string) {
		err := l.UnSubscribeAndClose(conn, tradeTopics)
		if err != nil {
			l.log.Fatal(err)
		}
	}(conn, tradeTopics)
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			l.log.Fatal(err)
		}
	}(conn)

	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			l.log.Fatal(err)
			return err
		}

		trade := Ticker{}
		err = json.Unmarshal(payload, &trade)
		if err != nil {
			l.log.Fatal(err)
			return err
		}
		l.log.Info(trade.Symbol, trade.Price, trade.Quantity)

		go func() {
			bytes, err := json.Marshal(trade)
			if err != nil {
				l.log.Fatalf("Error marshalling ticker data: %s", err.Error())
			}

			err = l.kafkaProducer.PublishMessage(context.Background(), kafka.Message{
				Key:   []byte(trade.Symbol + "-" + strconv.Itoa(int(trade.Time))),
				Value: bytes,
				Topic: "trades-" + strings.ToLower(trade.Symbol),
			})
			if err != nil {
				l.log.Fatal(err)
			}
		}()
	}
}
