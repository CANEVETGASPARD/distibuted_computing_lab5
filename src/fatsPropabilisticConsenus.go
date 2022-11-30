package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// STRUCTURE
type Communications struct {
	ask_ij      chan int
	ask_ji      chan int
	callback_ji chan int
	callback_ij chan int
	id_i        int
	id_j        int
}

type Node struct {
	id             int
	communications []Communications
	result         chan int
}

// CONSTANTS
const (
	NUMBER_OF_NODES_ASKED = 5
	NUMBER_OF_NODES       = 5
	TERMINATION           = 10
	BETA                  = 0.2
)

func sending_ask(i int, j int, ask_Sent chan int) {
	fmt.Println("[Node", i, "] sending ask to", j)
	ask_Sent <- i

}
func handling_ask(i int, j int, ask_Received chan int, callback_Sent chan int, opinion int) {
	select {
	case <-ask_Received:

		callback_Sent <- opinion
	case <-time.After(1 * time.Second):
		return
	}

	fmt.Println("[Node", i, "] ask handled")
	return

}
func handling_response(i int, j int, callback_Received chan int, opinions chan int) {

	select {
	case reponse := <-callback_Received:
		opinions <- reponse
	case <-time.After(1 * time.Second):
		return

	}
	fmt.Println("[Node", i, "] response handled")
	return
}

func main() {
	rand.Seed(time.Now().UnixNano())

	fmt.Println("STARTING")

	// INITIALIZATION
	node_neighbours_i := make([]int, 25)
	node_neighbours_j := make([]int, 25)

	for i := 0; i < NUMBER_OF_NODES; i++ {
		for j := 0; j < NUMBER_OF_NODES; j++ {

			node_neighbours_i[i*NUMBER_OF_NODES+j] = i
			node_neighbours_j[i*NUMBER_OF_NODES+j] = j
		}
	}

	// INITIALIZATION OF NODES
	nodes := make([]Node, NUMBER_OF_NODES)
	for i := 0; i < NUMBER_OF_NODES; i++ {
		nodes[i].id = i
		nodes[i].communications = make([]Communications, 0)
		nodes[i].result = make(chan int)
	}

	// INITIALIZATION OF COMMUNICATIONS
	for i := 0; i < len(node_neighbours_i); i++ {
		com := Communications{
			ask_ij:      make(chan int, NUMBER_OF_NODES),
			ask_ji:      make(chan int, NUMBER_OF_NODES),
			callback_ji: make(chan int, NUMBER_OF_NODES),
			callback_ij: make(chan int, NUMBER_OF_NODES),
			id_i:        node_neighbours_i[i],
			id_j:        node_neighbours_j[i],
		}
		nodes[node_neighbours_i[i]].communications = append(nodes[node_neighbours_i[i]].communications, com)
		nodes[node_neighbours_j[i]].communications = append(nodes[node_neighbours_j[i]].communications, Communications{
			ask_ij:      com.ask_ji,
			ask_ji:      com.ask_ij,
			callback_ji: com.callback_ij,
			callback_ij: com.callback_ji,
			id_i:        node_neighbours_j[i],
			id_j:        node_neighbours_i[i],
		})
	}

	// INITIALIZATION OF GOROUTINES
	for i := 0; i < NUMBER_OF_NODES; i++ {
		go node(nodes[i])
	}

	// WAITING FOR TERMINATION
	for i := 0; i < NUMBER_OF_NODES; i++ {
		result := <-nodes[i].result
		fmt.Println("[Node", i, "] TERMINATED WITH OPINION", result)
	}

}

func node(node Node) {
	n := 0
	k := 0
	opinion := rand.Intn(2)
	fmt.Println("[Node", node.id, "] STARTED WITH OPINION", opinion)
	for n < TERMINATION {
		temporaryOpinion := 0

		// selecting randomly node id asked
		node_asked := make([]int, NUMBER_OF_NODES_ASKED)
		for i := 0; i < NUMBER_OF_NODES_ASKED; i++ {
			node_asked[i] = rand.Intn(NUMBER_OF_NODES)
		}

		// sending ask to nodes
		for i := 0; i < NUMBER_OF_NODES_ASKED; i++ {
			com := findComToAsk(node_asked[i], node)
			com.ask_ij <- 1
		}

		// handling ask
		for i := 0; i < len(node.communications); i++ {
			go func(currentOpinion *int, index int) {
				for {
					select {
					case <-node.communications[index].ask_ji:
						opinionSend := *currentOpinion
						node.communications[index].callback_ji <- opinionSend
					}
				}
			}(&opinion, i)
		}

		// handling response
		for i := 0; i < NUMBER_OF_NODES_ASKED; i++ {
			com := findComToAsk(node_asked[i], node)
			select {
			case opinion := <-com.callback_ij:
				temporaryOpinion += opinion
			}

		}

		// updating opinion
		nk := float64(temporaryOpinion) / float64(NUMBER_OF_NODES_ASKED)
		uk := float64(U(k))

		newOpinion := opinion
		if nk > uk {
			newOpinion = 1
		}
		if nk < uk {
			newOpinion = 0
		}
		if opinion != newOpinion {
			n = 0
		} else {
			n++
		}
		opinion = newOpinion
		k++
	}
	node.result <- opinion
}

func findComToAsk(id int, node Node) Communications {

	for i := 0; i < len(node.communications); i++ {
		if node.communications[i].id_j == id {
			return node.communications[i]
		}
	}
	return Communications{}
}

func U(iter int) int {

	value_used := iter + 1

	value := math.Round(float64(rand.Intn(value_used)) / float64(value_used))

	if value < BETA {
		return 1
	} else {
		return 0
	}
}
