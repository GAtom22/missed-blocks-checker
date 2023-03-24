package types

import (
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	TombstonedEmoji = "💀"
	JailedEmoju     = "❌"
	UnjailedEmoji   = "👌"
)

const (
	TombstonedDesc = "was tombstoned"
	JailedDesc     = "was jailed"
	UnjailedDesc   = "was unjailed"
)

type ValidatorState struct {
	Address          string
	Moniker          string
	ConsensusAddress string
	MissedBlocks     int64
	Jailed           bool
	Active           bool
	Tombstoned       bool
}

func NewValidatorState(
	validator stakingtypes.Validator,
	info slashingtypes.ValidatorSigningInfo,
) ValidatorState {
	return ValidatorState{
		Address:          validator.OperatorAddress,
		Moniker:          validator.Description.Moniker,
		ConsensusAddress: info.Address,
		MissedBlocks:     info.MissedBlocksCounter,
		Jailed:           validator.Jailed,
		Active:           validator.Status == 3, // BOND_STATUS_BONDED
		Tombstoned:       info.Tombstoned,
	}
}

type ValidatorsState map[string]ValidatorState
