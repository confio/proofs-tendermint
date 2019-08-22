package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	tmproofs "github.com/confio/ics23-tendermint"
	"github.com/confio/ics23-tendermint/helpers"
	ics23 "github.com/confio/ics23/go"
)

/**
testgen-simple will generate a json struct on stdout (meant to be saved to file for testdata).
this will be an auto-generated existence proof in the form:

{
	"root": "<hex encoded root hash of tree>",
	"key": "<hex encoded key to prove>",
	"value": "<hex encoded value to prove> (empty on non-existence)",
	"proof": "<hex encoded protobuf marshaling of a CommitmentProof>"
}

It accepts two or three arguments (optional size: default 400)

testgen-iavl [exist|nonexist] [left|right|middle] <size>
**/

func main() {
	exist, loc, size, err := parseArgs(os.Args)
	if err != nil {
		fmt.Printf("%+v\n", err)
		fmt.Println("Usage: testgen-iavl [exist|nonexist] [left|right|middle] <size>")
		os.Exit(1)
	}

	data := helpers.BuildMap(size)
	allkeys := helpers.SortedKeys(data)
	root := helpers.CalcRoot(data)

	var key, value []byte
	if exist {
		key = []byte(helpers.GetKey(allkeys, loc))
		value = data[string(key)]
	} else {
		key = []byte(helpers.GetNonKey(allkeys, loc))
	}

	var proof *ics23.CommitmentProof
	if exist {
		proof, err = tmproofs.CreateMembershipProof(data, key)
	} else {
		proof, err = tmproofs.CreateNonMembershipProof(data, key)
	}
	if err != nil {
		fmt.Printf("Error: create proof: %+v\n", err)
		os.Exit(1)
	}

	binary, err := proof.Marshal()
	if err != nil {
		fmt.Printf("Error: protobuf marshal: %+v\n", err)
		os.Exit(1)
	}

	res := map[string]interface{}{
		"root":  hex.EncodeToString(root),
		"key":   hex.EncodeToString(key),
		"value": hex.EncodeToString(value),
		"proof": hex.EncodeToString(binary),
	}
	out, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		fmt.Printf("Error: json encoding: %+v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(out))
}

func parseArgs(args []string) (exist bool, loc helpers.Where, size int, err error) {
	if len(args) != 3 && len(args) != 4 {
		err = fmt.Errorf("Insufficient args")
		return
	}

	switch args[1] {
	case "exist":
		exist = true
	case "nonexist":
		exist = false
	default:
		err = fmt.Errorf("Invalid arg: %s", args[1])
		return
	}

	switch args[2] {
	case "left":
		loc = helpers.Left
	case "middle":
		loc = helpers.Middle
	case "right":
		loc = helpers.Right
	default:
		err = fmt.Errorf("Invalid arg: %s", args[2])
		return
	}

	size = 400
	if len(args) == 4 {
		size, err = strconv.Atoi(args[3])
	}

	return
}
