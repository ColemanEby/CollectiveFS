package transfer

// Signaling handles WebRTC signaling, matching Python's aiortc.contrib.signaling behavior.
// Currently this is a placeholder - the Python code uses hardcoded signaling
// which would need to be implemented for full WebRTC support.
//
// The Python implementation uses:
// - signaling_host: "127.0.0.1"
// - signaling_port: 1234
// - signaling_path: "aiortc.socket"
//
// This would typically use WebSocket or Unix socket for signaling.
type Signaling struct {
	host string
	port int
	path string
}

// NewSignaling creates a new signaling instance.
func NewSignaling(host string, port int, path string) *Signaling {
	return &Signaling{
		host: host,
		port: port,
		path: path,
	}
}

// TODO: Implement signaling client/server
// This would handle:
// - Sending/receiving SDP offers and answers
// - Exchanging ICE candidates
// - Sending BYE messages
// - Connection lifecycle management
