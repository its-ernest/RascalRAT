package protocol

import "time"

// TaskRequest defines an explicit configuration or sequence for the client agent.
type TaskRequest struct {
	TaskID      string        `json:"task_id"`
	PayloadType string        `json:"payload_type"` // e.g., "powershell", "system_query"
	ScriptBlock string        `json:"script_block"` // The command string or script block
	Timeout     time.Duration `json:"timeout"`      // Execution cutoff threshold
}

// TaskResponse encapsulates the execution results returned from the client endpoint.
type TaskResponse struct {
	TaskID       string        `json:"task_id"`
	Success      bool          `json:"success"`
	ExitCode     int           `json:"exit_code"`
	Stdout       string        `json:"stdout"`
	Stderr       string        `json:"stderr"`
	ErrorMessage string        `json:"error_message,omitempty"`
	Duration     time.Duration `json:"duration"`
}
