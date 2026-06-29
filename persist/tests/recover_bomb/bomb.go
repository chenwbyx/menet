//go:generate go install menet/persist/tests/recover_bomb
package main

import (
	"bufio"
	"bytes"
	"errors"
	"io/ioutil"
	"menet/persist/core"
	_ "menet/persist/tests/model"
	"os"
	"time"
)

func main() {
	inputReader := bufio.NewReader(os.Stdin)
	//outputReader := bufio.NewWriter(os.Stdout)
	data, err := ioutil.ReadAll(inputReader)
	if err != nil {
		panic(err)
	}
	pos := bytes.IndexByte(data, byte(' '))
	if pos == -1 {
		panic(errors.New("space separator not found in input"))
	}
	name := string(data[:pos])
	persistData := data[pos+1:]

	manage := core.GetIPersistByName(name)
	if manage != nil {
		manage.RecoverBomb(persistData)
	} else {
		panic(errors.New("manage == nil"))
	}
	time.Sleep(time.Second)
}
