package main

import (
	"strings"
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/core/chaincode/shim/ext/cid"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("auth_cc0")

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type Wallet struct {
	Name string `json: "name"`
	ID string `json: "id"`
}

// Init initializes the chaincode state
func (s *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	// init 시 인수 입력X 초깃값X
	return shim.Success(nil)
}

// Invoke makes payment of X units from A to B
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### auth_cc Invoke ###########")

	function, args := stub.GetFunctionAndParameters()

	if function == "initWallet" {
		return t.initWallet(stub, args)
	}

	if function == "getWallet" {
		return t.getWallet(stub, args)
	}

	logger.Errorf("Unknown action, check the first argument, must be one of 'initWallet' or 'getWallet'. But got: %v", args[0])
	return shim.Error(fmt.Sprintf("Unknown action, check the first argument, must be one of 'initWallet' or 'getWallet'. But got: %v", args[0]))
}

func (t *SimpleChaincode) initWallet(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// Declare wallets
	var wallet = Wallet{Name: args[0], ID: args[1]}
	// Convert my_wallet to []byte
	walletasJSONBytes, _ := json.Marshal(wallet)
	err := stub.PutState(wallet.ID, walletasJSONBytes)

	if err != nil {
		return shim.Error("Failed to create wallet " + wallet.Name)
	}

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	return shim.Success(nil)
}

// Deletes an entity from state
func (t *SimpleChaincode) getWallet(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// Getting attributes from an idemix credential
	ou, found, err := cid.GetAttributeValue(stub, "hf.Affiliation");
	if err != nil {
		return shim.Error("Failed to get attribute 'ou'")
	}
	if !found {
		return shim.Error("attribute 'ou' not found")
	}
	if !strings.HasSuffix(ou, "visitingServiceProvider") {
		return shim.Error(fmt.Sprintf("Incorrect 'o' returned, got '%s' expecting to end with 'visitingServiceProvider'", ou))
	}

	walletAsBytes, err := stub.GetState(args[0])
	if err != nil {
		fmt.Println(err.Error())
	}

	wallet := Wallet{}
	json.Unmarshal(walletAsBytes, &wallet)

	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false

	if bArrayMemberAlreadyWritten == true {
		buffer.WriteString(",")
	}

	buffer.WriteString("{Validated Employees: ")
	buffer.WriteString(ou)
	buffer.WriteString(" ")

	buffer.WriteString("\"Name\":")
	buffer.WriteString("\"")
	buffer.WriteString(wallet.Name)
	buffer.WriteString("\"")

	buffer.WriteString(", \"ID\":")
	buffer.WriteString("\"")
	buffer.WriteString(wallet.ID)
	buffer.WriteString("\"")

	buffer.WriteString("}")
	bArrayMemberAlreadyWritten = true
	buffer.WriteString("]\n")

	return shim.Success(buffer.Bytes())
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		logger.Errorf("Error starting Simple chaincode: %s", err)
	}
}