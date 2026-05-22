# CI Playbook

## Entry
- workflow change or failing CI check

## Freshness
- Check current GitHub Actions docs and action docs first

## Rules
- pinned third-party actions only
- least privilege permissions
- avoid unsafe interpolation in shell

## Verification
- `moon run :actionlint`
- `moon run :zizmor`
