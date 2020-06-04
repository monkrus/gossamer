package extrinsic

import (
	"github.com/ChainSafe/gossamer/lib/common/optional"
)

// Unchecked implements an Unchecked extrinsic
// https://github.com/paritytech/substrate/blob/master/primitives/runtime/src/generic/unchecked_extrinsic.rs
type Unchecked struct {
	Signature optional.Bytes // pub signature: Option<(Address, Signature, Extra)>,
	Module    byte           // Call::Module
	Function  byte           // Module::Call::Function
}
