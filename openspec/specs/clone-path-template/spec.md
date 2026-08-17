# clone-path-template Specification

## Purpose
TBD - created by archiving change add-clonepath-and-provider-abstraction. Update Purpose after archive.
## Requirements
### Requirement: Render clone target paths from a template

The system SHALL render a repository's clone target path from the configurable
`clone_pattern_tpl` Go text/template, exposing the documented data fields.

#### Scenario: Default template renders the conventional layout

- **WHEN** rendering a GitHub repo `foo` owned by `TechNative-B-V` with the
  default template and `base_dir` `/home/pim`
- **THEN** the result is `/home/pim/gh.technative-b-v/foo`

#### Scenario: All documented fields are available

- **WHEN** a template references `.BaseDir`, `.Provider`, `.Type`, `.Short`,
  `.Host`, `.Owner`, `.OwnerLower`, `.Repo`, or `.RepoLower`
- **THEN** each is substituted with the corresponding value

### Requirement: Per-provider template override

The system SHALL use a provider's `clone_pattern_tpl` when set, otherwise the
global template.

#### Scenario: Provider override takes precedence

- **WHEN** a provider defines its own `clone_pattern_tpl`
- **THEN** that template is used for its repositories instead of the global one

### Requirement: Safe, absolute clone paths

The system SHALL return cleaned absolute paths and reject templates that escape
the base directory.

#### Scenario: Tilde and base_dir expansion

- **WHEN** the rendered path begins with `~` or a relative segment
- **THEN** it is expanded and cleaned to an absolute path under the home/base
  directory

#### Scenario: Traversal is rejected

- **WHEN** a rendered path resolves outside `base_dir` (e.g. via `..`)
- **THEN** the system returns an error instead of a path

