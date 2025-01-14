// Copyright 2019 ChainSafe Systems (ON) Corp.
// This file is part of gossamer.
//
// The gossamer library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The gossamer library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the gossamer library. If not, see <http://www.gnu.org/licenses/>.

package runtime

import (
	"bytes"

	"github.com/ChainSafe/gossamer/lib/scale"
)

// Version struct
type Version struct {
	Spec_name         []byte
	Impl_name         []byte
	Authoring_version int32
	Spec_version      int32
	Impl_version      int32
}

// VersionAPI struct that holds Runtime Version info and API array
type VersionAPI struct {
	RuntimeVersion *Version
	API            []*API_Item
}

// API_Item struct to hold runtime API Name and Version
type API_Item struct {
	Name []byte
	Ver  int32
}

// Decode to scale decode []byte to VersionAPI struct
func (v *VersionAPI) Decode(in []byte) error {
	// decode runtime version
	_, err := scale.Decode(in, v.RuntimeVersion)
	if err != nil {
		return err
	}

	// 1 + len(Spec_name) + 1 + len(Impl_name) + 12 for  3 int32's - 1 (zero index)
	index := len(v.RuntimeVersion.Spec_name) + len(v.RuntimeVersion.Impl_name) + 14

	// read byte at index for qty of apis
	sd := scale.Decoder{Reader: bytes.NewReader(in[index : index+1])}
	numApis, err := sd.DecodeInteger()
	if err != nil {
		return err
	}
	// put index on first value
	index++
	// load api_item objects
	for i := 0; i < int(numApis); i++ {
		ver, err := scale.Decode(in[index+8+(i*12):index+12+(i*12)], int32(0))
		if err != nil {
			return err
		}
		v.API = append(v.API, &API_Item{
			Name: in[index+(i*12) : index+8+(i*12)],
			Ver:  ver.(int32),
		})
	}

	return nil
}

var (
	// CoreVersion is the runtime API call Core_version
	CoreVersion = "Core_version"
	// CoreInitializeBlock is the runtime API call Core_initialize_block
	CoreInitializeBlock = "Core_initialize_block"
	// CoreExecuteBlock is the runtime API call Core_execute_block
	CoreExecuteBlock = "Core_execute_block"
	// Metadata_metadata is the runtime API call Metadata_metadata
	Metadata_metadata = "Metadata_metadata"
	// TaggedTransactionQueueValidateTransaction is the runtime API call TaggedTransactionQueue_validate_transaction
	TaggedTransactionQueueValidateTransaction = "TaggedTransactionQueue_validate_transaction"
	// GrandpaAuthorities is the runtime API call GrandpaApi_grandpa_authorities
	GrandpaAuthorities = "GrandpaApi_grandpa_authorities"
	// BabeAPIConfiguration is the runtime API call BabeApi_configuration
	BabeAPIConfiguration = "BabeApi_configuration"
	// BlockBuilderInherentExtrinsics is the runtime API call BlockBuilder_inherent_extrinsics
	BlockBuilderInherentExtrinsics = "BlockBuilder_inherent_extrinsics"
	// BlockBuilderApplyExtrinsic is the runtime API call BlockBuilder_apply_extrinsic
	BlockBuilderApplyExtrinsic = "BlockBuilder_apply_extrinsic"
	// BlockBuilderFinalizeBlock is the runtime API call BlockBuilder_finalize_block
	BlockBuilderFinalizeBlock = "BlockBuilder_finalize_block"
)
