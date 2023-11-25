package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Sankey struct {
	file *os.File
}

func NewSankey(path string) *Sankey {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	return &Sankey{
		file: file,
	}
}

func (s *Sankey) Close() {
	err := s.file.Close()
	if err != nil {
		log.Print(err)
	}
}

func (s *Sankey) Trunc() {
	err := s.file.Truncate(0)
	if err != nil {
		log.Print(err)
	}
	_, err = s.file.Seek(0, 0)
	if err != nil {
		log.Print(err)
	}
	_, err = fmt.Fprintf(s.file, "id,name,time\n")
}

func (s *Sankey) Add(id string, name string) {
	_, err := fmt.Fprintf(s.file, "%s,%s,%s\n", id, name, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		log.Print(err)
	}
}
