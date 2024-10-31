package util

import (
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/evolutionlandorg/evo-backend/util/log"
)

type AbiInstant struct {
	Inputs []EventMapping `json:"inputs"`
	Name   string         `json:"name"`
	Type   string         `json:"type"`
}

func AnalysisAbiEvent(contractAddress, chainName string) ([]AbiInstant, error) {
	contractMap := GetContractsMap(chainName)
	if contractMap[contractAddress] == "" || contractAddress == "" {
		return nil, errors.New("don't relate this contract")
	}
	contractsListen := Evo.ContractsListen
	fileName := ""
	for _, contractName := range contractsListen {
		if contractMap[contractAddress] == strings.ToLower(contractName) {
			fileName = strings.ToLower(contractName[0:1]) + contractName[1:]
		}
	}
	if fileName == "" {
		return nil, errors.New("can't find abi file")
	}
	b, err := os.ReadFile("contract/" + fileName + ".abi")
	if err != nil {
		log.Debug("read %s file error: %s", "contract/"+fileName+".abi", err)
	}
	var abi []AbiInstant
	_ = json.Unmarshal(b, &abi)
	return abi, nil
}

type EventMapping struct {
	Indexed bool   `json:"indexed"`
	Type    string `json:"type"`
	Name    string `json:"name"`
}

func GetAbiEventInputs(abiInstant []AbiInstant, eventName string) []EventMapping {
	var data []EventMapping
	for _, funcName := range abiInstant {
		if funcName.Name == eventName && funcName.Type == "event" {
			inputs := funcName.Inputs
			return inputs
		}
	}
	return data
}
