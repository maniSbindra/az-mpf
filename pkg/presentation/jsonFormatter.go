package presentation

import (
	"encoding/json"
	"io"
	"log"
)

func (d *displayConfig) displayJSON(w io.Writer) error {
	jsonBytes, err := json.Marshal(d.result.RequiredPermissions)
	if err != nil {
		log.Fatalf("Error converting output to JSON :%v \n", err)
	}
	// fmt.Println(string(jsonBytes))
	_, err = w.Write(jsonBytes)
	if err != nil {
		return err
	}
	return nil
}
