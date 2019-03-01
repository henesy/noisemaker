package main

import (
	"strings"
	"net"
sc	"strconv"
)

// Silo type
type Silo struct {
	lights	bool	// whether silo lights are on or not
	power	bool	// power on/off
	humid	int		// current humidity %
	temp	int		// current temp °C
	supply	int		// current supply levels bushels
	cont	string	// current contents
	flag	string	// flag string
}

// Printable lights format
func (s *Silo) Lights() string {
	if s.lights {
		return "on"
	}
	
	return "off"
}

// Printable lights format
func (s *Silo) Power() string {
	return b2s(s.power)
}

// Constructor
func NewSilo() (s Silo) {
	// Init silo
	s.humid	= 30
	s.temp	= 20
	s.supply	= 0
	s.cont	= "corn"

	return
}

func (s *Silo) DoCmd(conn net.Conn, connected *bool, write func(string), read func([]byte), invalid func()) {
	ok := func() { write("ok.") }

	for *connected {
		buf := make([]byte, width)
		
		conn.Write([]byte("> "))

		read(buf)

		argv := strings.Fields(string(buf))
		
		if len(argv) > 1 {
			// '\n' counts as a field split, truncate
			argv = argv[:len(argv)-1]
		}

		if len(argv) < 1 {
			write("err: no command specified")
		}
		
		if !s.power && argv[0] != "power" {
			write("err: powered off")
			goto nocmd
		}
		
		if busy && argv[0] != "status" {
			slock.Lock()
			write("err: busy -- " + status)
			slock.Unlock()
			goto nocmd
		}

		if s.supply > 1000 && argv[0] != "supply" {
			write("err: overfull")
			goto nocmd
		}

		// Commands master switch
		switch argv[0] {
		
		// Lights
		case "lights":
			switch len(argv) {
			case 1:
				write(string(s.Lights()))
			case 2:
				if argv[1] == "on" {
					s.lights = true
				} else {
					s.lights = false
				}
				ok()
			default:
				invalid()
			}
			
		// Flag
		case "flag":
			switch len(argv) {
			case 1:
				write(s.flag)
			case 2:
				s.flag = argv[1]
			default:
				invalid()
			}

		// Contents
		case "contents":
			write(s.cont)
		
		// Power
		case "power":
			switch len(argv) {
			case 1:
				write(string(s.Power()))
			case 2:
				if argv[1] == "on" {
					s.power = true
				} else {
					s.power = false
				}
				ok()
			default:
				invalid()
			}
		
		// Supply
		case "supply":
			switch len(argv) {
			case 1:
				write(sc.Itoa(s.supply))
			case 3:
				n, err := sc.Atoi(argv[2])

				if err != nil {
					invalid()
				}

				if argv[1] == "load" {
					// TODO ­ more max/min logic
					s.supply += n
					ok()
					go spin(2, "loading", "idle")
				} else {
					// lower
					s.supply -= n
					ok()
					go spin(2, "unloading", "idle")
				}
			default:
				invalid()
			}
		
		// Heat
		case "heat":
			switch len(argv) {
			case 1:
				write(sc.Itoa(s.temp))
			case 3:
				n, err := sc.Atoi(argv[2])

				if err != nil {
					invalid()
				}

				if argv[1] == "raise" {
					// TODO ­ more max/min logic
					s.temp += n
				} else {
					// lower
					s.temp -= n
				}
				ok()
			default:
				invalid()
			}
		
		// Humidity
		case "humidity":
			switch len(argv) {
			case 1:
				write(sc.Itoa(s.humid))
			case 3:
				n, err := sc.Atoi(argv[2])

				if err != nil {
					invalid()
				}

				if argv[1] == "raise" {
					// TODO ­ more max/min logic
					s.humid += n
				} else {
					// lower
					s.humid -= n
				}
				ok()
			default:
				invalid()
			}
		
		// Status
		case "status":
			slock.Lock()
			write(status)
			slock.Unlock()
		
		// Manual disconnect commands, for convenience
		case "quit":
			fallthrough
		case "exit":
			ok()
			return
			break

		// Command not found
		default:
			write("err: unknown command")
		}
		nocmd:
	}
}
