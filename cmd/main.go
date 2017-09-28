package main

import (
    "github.com/crisadamo/gochain"
    "log"
    "net/http"
    "os"
    "strings"
)

func main() {
    serverPort := "8000"
    if len(os.Args) == 2 {
        serverPort = os.Args[1]
    }

    blockchain := gochain.NewBlockchain()
    nodeID := strings.Replace(gochain.PseudoUUID(), "-", "", -1)

    api := gochain.NewGoChainAPI(blockchain, nodeID)

    mux := http.NewServeMux()
    mux.HandleFunc("/nodes/register", api.RegisterNodeHandler)
    mux.HandleFunc("/nodes/resolve", api.ConsensusHandler)
    mux.HandleFunc("/transactions/new", api.TransactionHandler)
    mux.HandleFunc("/mine", api.MineHandler)
    mux.HandleFunc("/chain", api.ChainHandler)

    log.Printf("Starting gochain on port %s\n", serverPort)
    http.ListenAndServe(":"+serverPort, mux)
}
