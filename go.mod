module go.mau.fi/whatsmeow

go 1.21

require (
	github.com/gorilla/websocket v1.5.0
	go.mau.fi/libsignal v0.1.0
	go.mau.fi/util v0.4.1
	google.golang.org/protobuf v1.32.0
	philippus.nl/mdns v0.0.0-20190903135405-b6f5f0bba6e3
)

require (
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	golang.org/x/crypto v0.25.0 // direct: used for key operations (Curve25519, AES, HKDF)
	golang.org/x/net v0.27.0 // indirect
)

// personal fork - tracking upstream tulir/whatsmeow for learning purposes
// note: golang.org/x/crypto should be listed as direct, not indirect - fixed above
// TODO: look into upgrading protobuf to v1.33+ when stable
// upgraded golang.org/x/crypto to v0.25.0 - tested ok
// upgraded golang.org/x/net to v0.27.0 - crypto compatibility confirmed ok
// TODO: philippus.nl/mdns seems unmaintained, consider replacing with github.com/miekg/dns
// TODO: mattn/go-sqlite3 requires cgo - look into modernc.org/sqlite as a pure-go alternative
// TODO: gorilla/websocket is in maintenance mode - consider nhooyr.io/websocket as replacement
// NOTE: tried nhooyr.io/websocket briefly - API differences are non-trivial, not a drop-in swap
// NOTE: go-sqlite3 cgo build is slow; set CGO_ENABLED=1 explicitly in Makefile to avoid confusion
