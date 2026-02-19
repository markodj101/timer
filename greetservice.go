package main

import "os"

type GreetService struct{}

func (g *GreetService) Greet(name string) string {
	return "Hello " + name + "!"
}


type FileReading struct{}

func (g *FileReading) ReadFile(path string) ([]byte, error) {
	return  os.ReadFile(path)
}