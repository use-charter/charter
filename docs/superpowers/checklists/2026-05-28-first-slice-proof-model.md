# First Slice Proof Model

## Required for the first Phase 1 slice

- failing test or failing fixture-driven assertion first
- minimal implementation second
- explicit verification command third
- docs/spec alignment update in the same slice
- no silent expansion of scope

## Definition of done

- the failing test or fixture is committed in the same slice as the fix
- `moon run :test` passes from repo root
- any relevant spec, testing doc, or rule contract changed with the behavior
- no unrelated architecture or naming drift is introduced
- the slice is small enough to review without reading the whole repo

## Evidence shape

- test path or fixture path
- exact verification command
- expected failing signal before implementation
- expected passing signal after implementation
- linked spec or doc surface updated in the same change
