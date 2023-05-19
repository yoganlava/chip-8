package emulator

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

var FontData = [...]uint8 {
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80,  // F
}

const (
	DisplayWidth = 64
	DisplayHeight = 32
	ScreenWidth = 640
	ScreenHeight = 320
	PixelWidth = ScreenWidth / DisplayWidth
	PixelHeight = ScreenHeight / DisplayHeight
)

type CPU struct {
	display [DisplayHeight][DisplayWidth]bool
	ram [4096]uint8
	registers [16]uint8
	stack [16]uint16
	keys [16]bool
	// delay timer
	dt uint8
	// sound timer
	st uint8
	// program counter
	pc uint16
	// stack pointer
	sp uint8
	// index register
	ir uint16

	// renderer for screen
	renderer *sdl.Renderer
	// last timer dec
	lastTimerDec int64
}

func Init() *CPU {
	return &CPU {
		pc: 0x200,
	}
}

func (c *CPU) loadFontTable() {
	for i, byte := range FontData {
		c.ram[i] = byte
	}
}

func (c *CPU) pollKey() {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch e := event.(type) {
			case *sdl.KeyboardEvent:
				switch e.Keysym.Sym {
					case sdl.K_ESCAPE:
						os.Exit(0)
					case sdl.K_1:
						c.keys[0x1] = e.Type == sdl.KEYDOWN
					case sdl.K_2:
						c.keys[0x2] = e.Type == sdl.KEYDOWN
					case sdl.K_3:
						c.keys[0x3] = e.Type == sdl.KEYDOWN
					case sdl.K_4:
						c.keys[0xC] = e.Type == sdl.KEYDOWN
					case sdl.K_q:
						c.keys[0x4] = e.Type == sdl.KEYDOWN
					case sdl.K_w:
						c.keys[0x5] = e.Type == sdl.KEYDOWN
					case sdl.K_e:
						c.keys[0x6] = e.Type == sdl.KEYDOWN
					case sdl.K_r:
						c.keys[0xD] = e.Type == sdl.KEYDOWN
					case sdl.K_a:
						c.keys[0x7] = e.Type == sdl.KEYDOWN
					case sdl.K_s:
						c.keys[0x8] = e.Type == sdl.KEYDOWN
					case sdl.K_d:
						c.keys[0x9] = e.Type == sdl.KEYDOWN
					case sdl.K_f:
						c.keys[0xE] = e.Type == sdl.KEYDOWN
					case sdl.K_z:
						c.keys[0xA] = e.Type == sdl.KEYDOWN
					case sdl.K_x:
						c.keys[0x0] = e.Type == sdl.KEYDOWN
					case sdl.K_c:
						c.keys[0xB] = e.Type == sdl.KEYDOWN
					case sdl.K_v:
						c.keys[0xF] = e.Type == sdl.KEYDOWN
				}
		}
	}
}

func (c *CPU) Start() {
	// TODO cpu timing

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()
	println("SDL INIT")

	window, err := sdl.CreateWindow("test", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		ScreenWidth, ScreenHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	println("Window INIT")

	
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, 0)
	if err != nil {
		panic(err)
	}
	println("Renderer INIT")

	defer renderer.Destroy()


	c.renderer = renderer
	println("Start")
	c.loadFontTable()

		
	go func() {
		for {
			c.tick()

			if (c.dt | c.st) > 0 {
				now := time.Now().UnixMilli()
				if (now - c.lastTimerDec) > 1000 {
					c.lastTimerDec = now
					if c.dt > 0 {
						c.dt--
					}
					if c.st > 0 {
						c.st--
					}
				}
			}

			
			c.renderer.Present()
			sdl.Delay(1000 / 500)
		}
	}()
	for {
		c.pollKey()
	}
}

func (c *CPU) draw() {
	for i := 0; i < DisplayHeight; i++ {
		for j := 0; j < DisplayWidth; j++ {
			if c.display[i][j] {
				c.renderer.SetDrawColor(255, 255, 255, 255)
			} else {
				c.renderer.SetDrawColor(0, 0, 0, 255)

			}
			c.renderer.FillRect(
				&sdl.Rect{
					X: int32(j * PixelHeight),
					Y: int32(i * PixelWidth),
					W: PixelWidth,
					H: PixelHeight,
				},
			)
		}
	}
}

func (c *CPU) tick() {
	// fetch current opcode
	var currentOpCode = uint16(c.ram[c.pc]) << 8 | uint16(c.ram[c.pc + 1])
	c.pc += 2

	// fmt.Printf("OP: %X\n", currentOpCode)

	switch currentOpCode & 0xF000 {
		case 0x0000:
			switch currentOpCode & 0x00FF {
				// CLS
				case 0xE0:
					c.display = [32][64]bool {}
					c.renderer.SetDrawColor(0, 0, 0, 255)
					c.renderer.Clear()
				// RET
				case 0xEE:
					c.pc = c.stack[c.sp]
					c.sp--
				default:
					panic(fmt.Sprintf("Unknown Instruction: %X\n", currentOpCode))
			}
		// JMP — 1NNN
		// Jump to the address in NNN. Sets the PC to NNN.
		case 0x1000:
			c.pc = currentOpCode & 0x0FFF
		// CALL NNN — 2NNN
		// Call the subroutine at address NNN.
		case 0x2000:
			c.sp++
			c.stack[c.sp] = c.pc
			c.pc = currentOpCode & 0x0FFF
		// SE VX, NN — 3XNN
		// Skip the next instruction if VX == NN. 
		// Compare the value of register VX with NN and if they are equal, increment PC by two.
		case 0x3000:
			vx := c.registers[(currentOpCode & 0x0F00) >> 8]
			if vx == uint8(currentOpCode & 0x00FF) {
				c.pc += 2
			}
		// SNE VX, NN — 4XNN
		// Skip the next instruction if VX != NN.
		// Compare the value of register VX with NN and if they are not equal, increment PC by two.
		case 0x4000:
			vx := c.registers[(currentOpCode & 0x0F00) >> 8]
			if vx != uint8(currentOpCode & 0x00FF) {
				c.pc += 2
			}
		// SE VX, VY — 5XY0
		// Skip the next instruction if VX == VY.
		// Compare the value of register VX with the value of VY and if they are equal, increment PC by two.
		case 0x5000:
			vx := c.registers[(currentOpCode & 0x0F00) >> 8]
			vy := c.registers[(currentOpCode & 0x00F0) >> 4]
			if vx == vy {
				c.pc += 2
			}
		// LD VX, NN
		// Load the value NN into the register VX.
		case 0x6000:
			vx := (currentOpCode & 0x0F00) >> 8
			c.registers[vx] = uint8(currentOpCode & 0x00FF)
		// ADD VX, NN — 7XNN
		// Add the value NN to the value of register VX and store the result in VX.
		case 0x7000:
			vx := (currentOpCode & 0x0F00) >> 8
			c.registers[vx] = c.registers[vx] + uint8(currentOpCode & 0x00FF)
		case 0x8000:
			switch currentOpCode & 0x000F {
				// LD VX, VY — 8XY0
				// Put the value of register VY into VX.
				case 0x0:
					c.registers[(currentOpCode & 0x0F00) >> 8] = c.registers[(currentOpCode & 0x00F0) >> 4]
				// OR VX, VY — 8XY1
				// Perform a bitwise OR between the values of VX and VY and store the result in VX.
				case 0x1:
					c.registers[(currentOpCode & 0x0F00) >> 8] = c.registers[(currentOpCode & 0x0F00) >> 8] | c.registers[(currentOpCode & 0x00F0) >> 4]
				// AND VX, VY — 8XY2
				// Perform a bitwise AND between the values of VX and VY and store the result in VX.
				case 0x2:
					c.registers[(currentOpCode & 0x0F00) >> 8] = c.registers[(currentOpCode & 0x0F00) >> 8] & c.registers[(currentOpCode & 0x00F0) >> 4]
				// XOR VX, VY — 8XY3
				// Perform a bitwise XOR between the values of VX and VY and store the result in VX.
				case 0x3:
					c.registers[(currentOpCode & 0x0F00) >> 8] = c.registers[(currentOpCode & 0x0F00) >> 8] ^ c.registers[(currentOpCode & 0x00F0) >> 4]
				// ADD VX, VY — 8XY4
				// Add the values of VX and VY and store the result in VX. 
				// Put the carry bit in VF (if there is overflow, set VF to 1, otherwise 0).
				case 0x4:
					vx := c.registers[(currentOpCode & 0x0F00) >> 8]
					vy := c.registers[(currentOpCode & 0x00F0) >> 4]
					c.registers[(currentOpCode & 0x0F00) >> 8] = vx + vy
					
					if vy > (0xFF - vx) {
						c.registers[0xF] = 1
					} else {
						c.registers[0xF] = 0
					}
				// SUB VX, VY — 8XY5
				// Subtract the value of VY from VX and store the result in VX.
				// Put the borrow in VF (if there is borrow, VX > VY, set VF to 1, otherwise 0).
				case 0x5:
					vx := c.registers[(currentOpCode & 0x0F00) >> 8]
					vy := c.registers[(currentOpCode & 0x00F0) >> 4]
					c.registers[(currentOpCode & 0x0F00) >> 8] = vx - vy
					
					if vx > vy {
						c.registers[0xF] = 1
					} else {
						c.registers[0xF] = 0
					}
				// SHR VX {, VY} — 8XY6
				// Shift right, or divide VX by two.
				// Store the least significant bit of VX in VF, and then divide VX and store its value in VX
				case 0x6:
					vx := c.registers[(currentOpCode & 0x0F00) >> 8]
					c.registers[(currentOpCode & 0x0F00) >> 8] = vx >> 1
					c.registers[0xF] = vx & 0x01
				// SUBN VX, VY — 8XY7
				// Subtract the value of VY from VX and store the result in VX. 
				// Set VF to 1 if there is no borrow, to 0 otherwise.
				case 0x7:
					vx := c.registers[(currentOpCode & 0x0F00) >> 8]
					vy := c.registers[(currentOpCode & 0x00F0) >> 4]
					c.registers[(currentOpCode & 0x0F00) >> 8] = vy - vx
					if vy > vx {
						c.registers[0xF] = 1
					} else {
						c.registers[0xF] = 0
					}
				// SHL VX {, VY} — 8XYE
				// Shift left, or multiply VX by two.
				// Store the most significant bit of VX in VF, and then multiply VX and store its value in VX
				case 0xE:
					vx := c.registers[(currentOpCode & 0x0F00) >> 8]
					c.registers[(currentOpCode & 0x0F00) >> 8] = vx << 1
					c.registers[0xF] = vx >> 7
				default:
					panic(fmt.Sprintf("Unknown Instruction: %X\n", currentOpCode))
			}
		// SNE VX, VY — 9XY0
		// Skip the next instruction if the values of VX and VY are not equal.
		case 0x9000:
			vx := c.registers[(currentOpCode & 0x0F00) >> 8]
			vy := c.registers[(currentOpCode & 0x00F0) >> 4]
			if vx != vy {
				c.pc += 2
			}
		// LD I, NNN — ANNN
		// Set the value of I to the address NNN.
		case 0xA000:
			c.ir = currentOpCode & 0x0FFF
		// JMP V0, NNN — BNNN
		// Jump to the location NNN + V0.
		case 0xB000:
			c.pc = uint16(c.registers[0x0]) + (currentOpCode & 0x0FFF)
		// RND VX, NN – CXNN
		// Generate a random byte (from 0 to 255), do a bitwise AND with NN and store the result to VX.
		case 0xC000:
			c.registers[(currentOpCode & 0x0F00) >> 8] = uint8(rand.Intn(math.MaxUint8 + 1)) & uint8((currentOpCode & 0x00FF))
		// DRW VX, VY, N — DXYN
		case 0xD000:
			// size of sprite in bytes
			byteNum := currentOpCode & 0x000F
			// VX
			startX := c.registers[(currentOpCode & 0x0F00) >> 8]
			// VY
			startY := c.registers[(currentOpCode & 0x00F0) >> 4]

			for row := uint16(0); row < byteNum; row++ {
				currentSpriteByte := c.ram[c.ir + row]
				currentY := (startY + uint8(row)) % DisplayHeight
				for col := uint8(0); col < uint8(8); col++ {
					currentX := (startX + uint8(col)) % DisplayWidth
					currentDisplayBit := c.display[currentY][currentX]
					currentDisplayPos := (currentSpriteByte & (0x80 >> col)) != 0

					if c.display[currentY][currentX] {
						c.registers[0xF] = 1
					}
					c.display[currentY][currentX] = currentDisplayBit != currentDisplayPos
					
					if currentX == DisplayWidth - 1 {
						break
					}
				}
				if currentY == DisplayHeight - 1 {
					break
				}
			}
			c.draw()
		case 0xE000:
			switch currentOpCode & 0x00FF {
				// SKP VX — EX9E
				// Skip the next instruction if the key with the value of VX is currently pressed.
				case 0x9E:
					if c.keys[c.registers[(currentOpCode & 0x0F00) >> 8]] {
						c.pc += 2
					}
				// SKNP VX — EXA1
				// Skip the next instruction if the key with the value of VX is currently not pressed. 
				case 0xA1:
					if !c.keys[c.registers[(currentOpCode & 0x0F00) >> 8]] {
						c.pc += 2
					}
				default:
					panic(fmt.Sprintf("Unknown Instruction: %X\n", currentOpCode))
			}
		case 0xF000:
			switch currentOpCode & 0x00FF {
				// LD VX, DT — FX07
				// Read the delay timer register value into VX.
				case 0x07:
					c.registers[(currentOpCode & 0x0F00) >> 8] = c.dt
				case 0x0A:
					pressed := false
					for i, key := range c.keys {
						if key {
							c.registers[(currentOpCode & 0x0F00) >> 8] = uint8(i)
							pressed = true
							break
						}
					}
					if !pressed {
						c.pc -= 2
					}
				// LD DT, VX — FX15
				// Load the value of VX into the delay timer DT.
				case 0x15:
					c.dt = c.registers[(currentOpCode & 0x0F00) >> 8]
				// LD ST, VX — FX18
				// Load the value of VX into the sound time ST.
				case 0x18:
					c.st = c.registers[(currentOpCode & 0x0F00) >> 8]
				// ADD I, VX — FX1E
				// Add the values of I and VX, and store the result in I.
				case 0x1E:
					c.ir = c.ir + uint16(c.registers[(currentOpCode & 0x0F00) >> 8])
				// LD F, VX — FX29
				// Set the location of the sprite for the digit VX to I. 
				case 0x29:
					c.ir = uint16(c.registers[(currentOpCode & 0x0F00) >> 8]) * 0x05
				// LD B, VX — FX33
				// Store the binary-coded decimal in VX and put it in three consecutive memory slots starting at I.
				case 0x33:
					vx := c.registers[(currentOpCode & 0x0F00) >> 8]
					hundred := vx / 100
					tens := (vx - hundred * 100) / 10
					ones := vx - hundred * 100 - tens * 10

					c.ram[c.ir] = hundred
					c.ram[c.ir + 1] = tens
					c.ram[c.ir + 2] = ones

				// LD [I], VX — FX55
				// Store registers from V0 to VX in the main memory, starting at location I.
				case 0x55:
					for i := uint16(0); i <= (currentOpCode & 0x0F00) >> 8; i++ {
						c.ram[c.ir + i] = c.registers[i]
					}
				// LD VX, [I] — FX65
				// Load the memory data starting at address I into the registers V0 to VX.
				case 0x65:
					for i := uint16(0); i <= (currentOpCode & 0x0F00) >> 8; i++ {
						c.registers[i] = c.ram[c.ir + i]
					}
				default:
					panic(fmt.Sprintf("Unknown Instruction: %X\n", currentOpCode))
			}
		default:
			panic(fmt.Sprintf("Unknown Instruction: %X\n", currentOpCode))
	}
}

func (c *CPU) LoadRom(rom []byte) {
	// Always reset state before loading
	// c.Reset()

	// TODO check if rom can fit
	for i := 0; i < len(rom); i++ {
		c.ram[i + 512] = rom[i]
	}
}