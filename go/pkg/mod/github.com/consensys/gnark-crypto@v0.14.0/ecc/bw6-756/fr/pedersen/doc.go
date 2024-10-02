// Copyright 2020 Consensys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by consensys/gnark-crypto DO NOT EDIT

// Package pedersen allows to compute and verify Pedersen vector commitments
//
// Pedersen vector commitments are a type of homomorphic commitments that allow
// to commit to a vector of values and prove knowledge of the committed values.
// The commitments can be batched and verified in a single operation.
//
// The commitments are computed using a set of basis elements. The proving key
// contains the basis elements and their exponentiations by a random value. The
// verifying key contains the G2 generator and its exponentiation by the inverse
// of the random value.
//
// The setup process is a trusted setup and must be done securely, preferably using MPC.
// After the setup, the proving key does not have to be secret, but the randomness
// used during the setup must be discarded.
package pedersen
