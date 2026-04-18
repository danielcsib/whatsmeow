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
	golang.org/x/crypto v0.24.0 // direct: used for key operations (Curve25519, AES, HKDF)
	golang.org/x/net v0.26.0 // indirect
)

// personal fork - tracking upstream tulir/whatsmeow for learning purposes
// note: golang.org/x/crypto should be listed as direct, not indirect - fixed above
// TODO: look into upgrading protobuf to v1.33+ when stable
