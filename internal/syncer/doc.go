// Package syncer clones missing repositories and fast-forward-pulls existing
// ones into the templated clone paths, skipping dirty trees.
//
// Implemented in milestone "03 CLI sync", epic "03a Sync engine". See
// BRIEFING.md section 10.
package syncer
