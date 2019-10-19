package modules

import (
	"fmt"
	"github.com/davidscholberg/go-i3barjson"
	"os/exec"
	"strings"
)

// Volume represents the configuration for the volume display block.
type Volume struct {
	BlockConfigBase `yaml:",inline"`
	SinkDevice      string `yaml:"sink_device"`
}

// UpdateBlock updates the volume display block.
func (c Volume) UpdateBlock(b *i3barjson.Block) {
	b.Color = c.Color
	fullTextFmt := fmt.Sprintf("%s%%s%%%%", c.Label)
	pamixerCmd := "pamixer"
	if c.SinkDevice == "" {
		c.SinkDevice = "0"
	}
	cmdArgs := []string{"--sink", c.SinkDevice, "--get-volume"}
	out, err := exec.Command(pamixerCmd, cmdArgs...).Output()
	if err != nil {
		b.FullText = fmt.Sprintf(fullTextFmt, err.Error())
	} else {
		b.FullText = fmt.Sprintf(fullTextFmt, strings.Trim(string(out), "\n"))
	}
}
