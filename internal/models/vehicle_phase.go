package models

type VehiclePhase string

const (
	PhaseAtBase         VehiclePhase = "AT_BASE"
	PhaseDeadheadBefore VehiclePhase = "DEADHEAD_BEFORE"
	PhaseLayoverBefore  VehiclePhase = "LAYOVER_BEFORE"
	PhaseInProgress     VehiclePhase = "IN_PROGRESS"
	PhaseDeadheadDuring VehiclePhase = "DEADHEAD_DURING"
	PhaseLayoverDuring  VehiclePhase = "LAYOVER_DURING"
	PhaseDeadheadAfter  VehiclePhase = "DEADHEAD_AFTER"
	PhaseUnknown        VehiclePhase = "UNKNOWN"
)

var activeBeforeBlock = map[VehiclePhase]bool{
	PhaseAtBase:         true,
	PhaseDeadheadBefore: true,
	PhaseLayoverBefore:  true,
}

var activeDuringBlock = map[VehiclePhase]bool{
	PhaseInProgress:     true,
	PhaseDeadheadDuring: true,
	PhaseLayoverDuring:  true,
}

var activeAfterBlock = map[VehiclePhase]bool{
	PhaseDeadheadAfter: true,
}

func IsActiveBeforeBlock(phase VehiclePhase) bool {
	return activeBeforeBlock[phase]
}

func IsActiveDuringBlock(phase VehiclePhase) bool {
	return activeDuringBlock[phase]
}

func IsActiveAfterBlock(phase VehiclePhase) bool {
	return activeAfterBlock[phase]
}
