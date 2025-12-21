package websocket

// Event types
const (
	// Matchmaking events
	EventJoinQueue       = "join_queue"
	EventLeaveQueue      = "leave_queue"
	EventMatchFound      = "match_found"
	EventAcceptMatch     = "accept_match"
	EventPartnerAccepted = "partner_accepted"
	EventBothReady       = "both_ready"

	// Room events
	EventJoinRoom  = "join_room"
	EventLeaveRoom = "leave_room"
	EventRoomReady = "room_ready"

	// WebRTC signaling events
	EventWebRTCOffer  = "webrtc_offer"
	EventWebRTCAnswer = "webrtc_answer"
	EventICECandidate = "ice_candidate"

	// Call events
	EventCallStart           = "call_start"
	EventCallEnd             = "call_ended"
	EventMediaStateChange    = "media_state_changed"
	EventPartnerDisconnected = "partner_disconnected"

	// Interview events
	EventInterviewStart     = "interview_start"
	EventInterviewEnd       = "interview_end"
	EventEvaluationComplete = "evaluation_complete"

	// Chat/messaging
	EventMessage = "message"

	// System events
	EventConnected    = "connected"
	EventDisconnected = "disconnected"
	EventError        = "error"
)

// Event represents a WebSocket event
type Event struct {
	Type      string                 `json:"type"`
	From      string                 `json:"from,omitempty"`
	To        string                 `json:"to,omitempty"`
	RoomID    string                 `json:"roomId,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	SDP       interface{}            `json:"sdp,omitempty"`
	Candidate interface{}            `json:"candidate,omitempty"`
}

// MatchFoundEvent is sent when a match is found
type MatchFoundEvent struct {
	Type    string `json:"type"`
	RoomID  string `json:"roomId"`
	Message string `json:"message"`
}

// WebRTCOfferEvent represents a WebRTC offer
type WebRTCOfferEvent struct {
	Type string                 `json:"type"`
	From string                 `json:"from"`
	To   string                 `json:"to"`
	SDP  map[string]interface{} `json:"sdp"`
}

// WebRTCAnswerEvent represents a WebRTC answer
type WebRTCAnswerEvent struct {
	Type string                 `json:"type"`
	From string                 `json:"from"`
	To   string                 `json:"to"`
	SDP  map[string]interface{} `json:"sdp"`
}

// ICECandidateEvent represents an ICE candidate
type ICECandidateEvent struct {
	Type      string                 `json:"type"`
	From      string                 `json:"from"`
	To        string                 `json:"to"`
	Candidate map[string]interface{} `json:"candidate"`
}

// MessageEvent represents a chat message
type MessageEvent struct {
	Type    string `json:"type"`
	From    string `json:"from"`
	RoomID  string `json:"roomId"`
	Message string `json:"message"`
}

// ErrorEvent represents an error message
type ErrorEvent struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}
