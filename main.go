package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/VictorLowther/btree"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/nsf/termbox-go"
)

func byBestBid(a, b *OrderBookEntry) bool {
	return a.Size >= b.Size
}

func byBestAsk(a, b *OrderBookEntry) bool {
	return a.Size < b.Size
}

type OrderBook struct {
	Asks *btree.Tree[*OrderBookEntry]
	Bids *btree.Tree[*OrderBookEntry]
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		Asks: btree.New(byBestAsk),
		Bids: btree.New(byBestBid),
	}
}

func (ob *OrderBook) handleBinanceOrderBookResult(res BinanceOrderbookResult) {
	for _, ask := range res.Asks {
		price, _ := strconv.ParseFloat(ask[0], 64)
		size, _ := strconv.ParseFloat(ask[1], 64)
		if size > 0 {
			entry := &OrderBookEntry{
				Price: price,
				Size:  size,
			}

			ob.Asks.Insert(entry)
		} else {
			if thing, ok := ob.Asks.Get(getAskByPrice(price)); ok {
				fmt.Printf("-- deleting level %2f", price)
				ob.Asks.Delete(thing)

			}
		}

	}

	for _, bid := range res.Bids {
		price, _ := strconv.ParseFloat(bid[0], 64)
		size, _ := strconv.ParseFloat(bid[1], 64)
		if size > 0 {
			entry := &OrderBookEntry{
				Price: price,
				Size:  size,
			}

			ob.Bids.Insert(entry)
		}

	}
}

func getAskByPrice(price float64) btree.CompareAgainst[*OrderBookEntry] {
	return func(e *OrderBookEntry) int {
		switch {
		case e.Price < price:
			return -1
		case e.Price > price:
			return 1
		default:
			return 0
		}
	}
}

type OrderBookEntry struct {
	Price float64
	Size  float64
}
type BinanceOrderbookResult struct {
	Asks [][]string `json:"a"`
	Bids [][]string `json:"b"`
}
type BinanceDepthResponse struct {
	Stream string                 `json:"stream"`
	Data   BinanceOrderbookResult `json:"data"`
}

func main() {
	termbox.Init()

loop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				break loop
				// 	case termbox.KeyArrowUp, termbox.KeyArrowRight:
				// 		switch_output_mode(1)
				// 		draw_all()
				// 	case termbox.KeyArrowDown, termbox.KeyArrowLeft:
				// 		switch_output_mode(-1)
				// 		draw_all()

				// 	}
				// case termbox.EventResize:
				// 	draw_all()
			}
		}

		termbox.SetCell(10, 10, 'o', termbox.ColorDefault, termbox.ColorCyan)
		termbox.Flush()

	}
}

func _main() {
	fmt.Println("I think I'm big mitch")

	if err := termbox.Init(); err != nil {
		log.Fatal(err)
	}
	if err := godotenv.Load(); err != nil {
		log.Fatal(".env Load failed", err)
	}
	conn, _, err := websocket.DefaultDialer.Dial(os.Getenv("WSENDPOINT"), nil) //
	if err != nil {
		fmt.Println(err)
	}

	var (
		ob     = NewOrderBook()
		result BinanceDepthResponse
	)

	for {
		if err := conn.ReadJSON(&result); err != nil {
			log.Fatal(err)
		}
		ob.handleBinanceOrderBookResult(result.Data)
		it := ob.Asks.Iterator(nil, nil)
		for it.Next() {

			fmt.Printf("%+v\n", it.Item())
		}

	}
}
