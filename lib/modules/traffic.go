package modules

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/davidscholberg/go-i3barjson"
	"github.com/safchain/ethtool"
)

// Traffic represents the configuration for the traffic block.
type Traffic struct {
	BlockConfigBase `yaml:",inline"`
	IfaceName       string `yaml:"interface_name"`
	Limit           uint64 `yaml:"limit"`
}

type ByteCounter struct {
	Up   uint64 `json:"up"`
	Down uint64 `json:"down"`
}

// UpdateBlock updates the traffic interface block.
func (c Traffic) UpdateBlock(b *i3barjson.Block) {
	b.Color = c.Color
	fullTextFmt := fmt.Sprintf("%s%%d ⇧, %%d ⇩", c.Label)

	ethHandle, err := ethtool.NewEthtool()
	if err != nil {
		panic(err.Error())
	}
	defer ethHandle.Close()

	// Get current byte counter
	status, _ := ethHandle.Stats(c.IfaceName)
	var current ByteCounter
	current.Up = status["tx_bytes"]
	current.Down = status["rx_bytes"]
	// fmt.Printf("current   up %d\n", current.Up)
	// fmt.Printf("current down %d\n", current.Down)

	// Read old values
	filename := fmt.Sprintf("/tmp/goblocks-traffic-%s", c.IfaceName)

	var old ByteCounter
	byteValue, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(byteValue, &old)
	if err != nil {
		old.Up = 0
		old.Down = 0
	}
	// fmt.Printf("old   up %d\n", old.Up)
	// fmt.Printf("old down %d\n", old.Down)

	var speed ByteCounter
	interval := uint64(c.GetUpdateInterval())
	// fmt.Println(interval)
	speed.Up = ((current.Up - old.Up) / interval) / 125000
	speed.Down = ((current.Down - old.Down) / interval) / 125000
	// fmt.Printf("speed   up %d\n", speed.Up)
	// fmt.Printf("speed down %d\n", speed.Down)

	// Save current values
	stateFile, _ := os.OpenFile(filename, os.O_WRONLY, 0600)
	data, _ := json.Marshal(&current)
	_, err = stateFile.Write(data)
	if err != nil {
		panic(err)
	}
	defer stateFile.Close()

	if c.Limit > 0 && (speed.Up > c.Limit || speed.Down > c.Limit) {
		b.Urgent = true
	}

	b.FullText = fmt.Sprintf(fullTextFmt, speed.Up, speed.Down)
}
