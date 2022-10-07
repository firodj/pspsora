package models

type MipsOpcode struct {
	Address            uint32
	Encoded            uint32
	IsConditional      bool
	IsConditionMet     bool
	IsBranch           bool
	IsLinkedBranch     bool
	IsLikelyBranch     bool
	IsBranchToRegister bool
	HasDelaySlot       bool
	IsDataAccess       bool
	HasRelevantAddress bool
	BranchTarget       uint32
	BranchRegister     int
	DataSize           int
	DataAddress        uint32
	RelevantAddress    uint32
	Dizz               string
	Log                string
}
