package blocker

import (
  "log"
)

type ControlMsg struct {
  Ip        string
  Block     bool
}

func Blocker(controlChan chan ControlMsg) {
  for msg := range controlChan{
    // we should dispatch to the underlying blocking (iptables)
    log.Println(msg)
  }
}
