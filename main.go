package main

import (
	"chip8/emulator"
	"os"
)

func main() {
	var cpu = emulator.Init()

	if len(os.Args) < 2 {
		panic("File please!")
	}

	rom, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	cpu.LoadRom(rom)

	cpu.Start()
}