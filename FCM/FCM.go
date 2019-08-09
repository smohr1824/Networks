package FCM

import ("github.com/smohr1824/Networks/Core")

// pending conversion of Network to id nodes with uint instead of string
type FCM struct {
	concepts map[uint32] string
	graph Core.Network
}


