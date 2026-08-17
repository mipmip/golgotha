// Package provider defines the Provider interface, the Repo domain model, and
// the auth resolver shared by the GitHub, Codeberg and GitLab clients.
//
// The abstraction is implemented in milestone "01 Foundations", epic "01d
// Provider & auth abstraction"; concrete clients live in milestone "02
// Provider clients & cache". See BRIEFING.md sections 4 and 6.
package provider
