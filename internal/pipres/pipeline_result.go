package pipres

type PipelineResult string

const (
	Undefined          PipelineResult = "UNDEFINED"
	SlotAvailable      PipelineResult = "SLOT_AVAILABLE"
	SlotNotAvailable   PipelineResult = "SLOT_UNAVAILABLE"
	MaybeAlreadySigned PipelineResult = "MAYBE_ALREADY_SIGNED"
	NoRescheduleTasks  PipelineResult = "NO_RESCHEDULE_TASKS"
)
