// Package job defines the conversion job types and a runner that converts
// images sequentially with cancellation support.
package job

// Status is the per-item lifecycle state.
type Status int

const (
	StatusPending Status = iota
	StatusRunning
	StatusDone
	StatusFailed
	StatusSkipped
)

// Mode is the output mode when the input is a directory.
type Mode int

const (
	// ModeZip — when at least one Item.IsDir, all produced PNGs are bundled
	// into a single .zip next to the source directory. With only flat
	// (non-directory) items, ModeZip silently behaves like ModeIndividual.
	ModeZip Mode = iota
	ModeIndividual
)

// OverwritePolicy controls how same-name PNGs are handled at the output dir.
type OverwritePolicy int

const (
	PolicyIncrement OverwritePolicy = iota // photo.png → photo_2.png
	PolicyOverwrite
	PolicySkip
)

// FileItem is one entry in the GUI file list. Path may point to a file or directory.
type FileItem struct {
	Path  string
	IsDir bool
}

// Job is the converted-side request: what to convert and how.
type Job struct {
	Items      []FileItem
	OutputDir  string // empty → derive from each Item per current CLI rules
	OutputMode Mode
	Overwrite  OverwritePolicy
}

// EventKind names a progress event.
type EventKind int

const (
	EventStart EventKind = iota // emitted once before the first item
	EventItem                   // per converted file (or skip)
	EventDone                   // emitted once after the last item
)

// Event carries one progress notification on the runner's output channel.
// SourcePath is the original file (not the produced PNG) for EventItem.
type Event struct {
	Kind       EventKind
	SourcePath string
	OutputPath string // populated on success
	Status     Status // StatusDone / StatusFailed / StatusSkipped for EventItem
	Err        error
	Total      int // total expected files (set on EventStart and EventDone)
	Completed  int // running count of completed (any status) at this event
}
