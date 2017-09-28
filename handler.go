package gochain

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
)

func NewGoChainAPI(blockchain *Blockchain, nodeID string) GoChainAPI {
    return GoChainAPI{blockchain, nodeID}
}

type GoChainAPI struct {
    blockchain *Blockchain
    nodeId     string
}

func (gca *GoChainAPI) TransactionHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    bytes, err := ioutil.ReadAll(r.Body)

    var tx Transaction
    err = json.Unmarshal(bytes, &tx)
    index := gca.blockchain.NewTransaction(tx)

    enc, _ := json.Marshal(map[string]string{
        "message": fmt.Sprintf("Transaction will be added to Block %d", index),
    })

    status := http.StatusCreated
    if err != nil {
        status = http.StatusInternalServerError
        enc, _ = json.Marshal(map[string]string{"error": "fail to add transaction"})
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    w.Write(enc)
}

func (gca *GoChainAPI) MineHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    // We run the proof of work algorithm to get the next proof...
    lastBlock := gca.blockchain.LastBlock()
    lastProof := lastBlock.Proof
    proof := gca.blockchain.ProofOfWork(lastProof)

    // We must receive a reward for finding the proof.
    // The sender is "0" to signify that this node has mined a new coin.
    newTX := Transaction{Sender: "0", Recipient: gca.nodeId, Amount: 1}
    gca.blockchain.NewTransaction(newTX)

    // Forge the new Block by adding it to the chain
    block := gca.blockchain.NewBlock(proof, "")

    enc, err := json.Marshal(map[string]interface{}{
        "message": "New Block Forged",
        "block":   block,
    })

    status := http.StatusOK
    if err != nil {
        status = http.StatusInternalServerError
        enc, _ = json.Marshal(map[string]string{"error": "fail to mine"})
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    w.Write(enc)
}

func (gca *GoChainAPI) ChainHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    enc, err := json.Marshal(map[string]interface{}{
        "chain":  gca.blockchain.chain,
        "length": len(gca.blockchain.chain),
    })

    status := http.StatusOK
    if err != nil {
        status = http.StatusInternalServerError
        enc, _ = json.Marshal(map[string]string{"error": "fail to generate the blockchain"})
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    w.Write(enc)
}

func (gca *GoChainAPI) RegisterNodeHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    bytes, err := ioutil.ReadAll(r.Body)
    var body map[string][]string
    err = json.Unmarshal(bytes, &body)

    for _, node := range body["nodes"] {
        gca.blockchain.RegisterNode(node)
    }

    enc, _ := json.Marshal(map[string]interface{}{
        "message": "New nodes have been added",
        "nodes":   gca.blockchain.nodes.Keys(),
    })

    status := http.StatusCreated
    if err != nil {
        status = http.StatusInternalServerError
        enc, _ = json.Marshal(map[string]string{"error": "fail to register nodes"})
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    w.Write(enc)
}

func (gca *GoChainAPI) ConsensusHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    var resp map[string]interface{}
    if gca.blockchain.ResolveConflicts() {
        resp = map[string]interface{}{
            "message": "Our chain was replaced",
            "chain":   gca.blockchain.chain,
        }
    } else {
        resp = map[string]interface{}{
            "message": "Our chain is authoritative",
            "chain":   gca.blockchain.chain,
        }
    }

    enc, err := json.Marshal(resp)
    status := http.StatusOK
    if err != nil {
        status = http.StatusInternalServerError
        enc, _ = json.Marshal(map[string]string{"error": "fail to generate the blockchain"})
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    w.Write(enc)
}
