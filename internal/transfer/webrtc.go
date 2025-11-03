package transfer

import (
	"fmt"
	"log"
	"os"

	"github.com/pion/webrtc/v3"
)

// Transfer handles WebRTC file transfers, translating Python's filexfer.py functionality.
type Transfer struct {
	// WebRTC components will be added here
	// This is a placeholder for the WebRTC implementation
}

// StartTransfer initiates a file transfer, matching Python's start_transfer function.
// This is a direct translation of the filexfer.py start_transfer function.
func StartTransfer(direction string, fileInfo, chunkInfo map[string]interface{}) error {
	// Extract filename from chunkInfo
	filename, ok := chunkInfo["path"].(string)
	if !ok {
		return fmt.Errorf("invalid path in chunkInfo")
	}

	// Hardcoded signaling configuration (matching Python)
	signalingHost := "127.0.0.1"
	signalingPort := 1234
	signalingPath := "aiortc.socket"

	log.Printf("Starting transfer: direction=%s, file=%s, signaling=%s:%d/%s",
		direction, filename, signalingHost, signalingPort, signalingPath)

	// Create peer connection
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	}

	pc, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return fmt.Errorf("failed to create peer connection: %w", err)
	}
	defer pc.Close()

	if direction == "send" {
		return runOffer(pc, filename, signalingHost, signalingPort)
	} else {
		return runAnswer(pc, filename, signalingHost, signalingPort)
	}
}

// runOffer implements the offerer role, translating Python's run_offer function.
func runOffer(pc *webrtc.PeerConnection, filename, signalingHost string, signalingPort int) error {
	// Open file for reading
	fp, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer fp.Close()

	// Create data channel
	dataChannel, err := pc.CreateDataChannel("filexfer", nil)
	if err != nil {
		return fmt.Errorf("failed to create data channel: %w", err)
	}

	// Set up data channel handlers
	doneReading := false
	dataChannel.OnOpen(func() {
		log.Println("Data channel opened, starting file transfer")
		sendFileData(dataChannel, fp, &doneReading)
	})

	dataChannel.OnBufferedAmountLow(func() {
		if !doneReading {
			sendFileData(dataChannel, fp, &doneReading)
		}
	})

	// Create and send offer
	offer, err := pc.CreateOffer(nil)
	if err != nil {
		return fmt.Errorf("failed to create offer: %w", err)
	}

	if err := pc.SetLocalDescription(offer); err != nil {
		return fmt.Errorf("failed to set local description: %w", err)
	}

	// TODO: Send offer via signaling
	log.Printf("Offer created, would send via signaling to %s:%d", signalingHost, signalingPort)

	// TODO: Wait for answer and ICE candidates via signaling
	// This would be implemented in signaling.go

	return nil
}

// runAnswer implements the answerer role, translating Python's run_answer function.
func runAnswer(pc *webrtc.PeerConnection, filename, signalingHost string, signalingPort int) error {
	// Open file for writing
	fp, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer fp.Close()

	var bytesReceived int64

	// Handle incoming data channel
	pc.OnDataChannel(func(channel *webrtc.DataChannel) {
		channel.OnOpen(func() {
			log.Println("Data channel opened, receiving file")
		})

		channel.OnMessage(func(msg webrtc.DataChannelMessage) {
			if len(msg.Data) > 0 {
				bytesReceived += int64(len(msg.Data))
				if _, err := fp.Write(msg.Data); err != nil {
					log.Printf("failed to write to file: %v", err)
				}
			} else {
				// Empty message indicates end of file
				log.Printf("File transfer complete: received %d bytes", bytesReceived)
				// TODO: Send BYE via signaling
			}
		})
	})

	// TODO: Receive offer via signaling and create answer
	log.Printf("Waiting for offer via signaling from %s:%d", signalingHost, signalingPort)

	return nil
}

// sendFileData sends file data in chunks, matching Python's send_data function behavior.
func sendFileData(channel *webrtc.DataChannel, fp *os.File, doneReading *bool) {
	buffer := make([]byte, 16384) // 16KB chunks, matching Python
	for {
		n, err := fp.Read(buffer)
		if err != nil || n == 0 {
			*doneReading = true
			// Send empty message to indicate end
			channel.SendText("")
			return
		}
		if err := channel.Send(buffer[:n]); err != nil {
			log.Printf("failed to send data: %v", err)
			*doneReading = true
			return
		}
		if channel.BufferedAmount() > channel.BufferedAmountLowThreshold() {
			// Wait for buffer to drain
			return
		}
	}
}
