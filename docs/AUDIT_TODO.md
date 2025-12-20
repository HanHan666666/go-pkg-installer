# Audit TODO (SPECIFICATION.md alignment)

## High priority
- Align branch config fields between schema and runtime (prefer `condition/branches/default`, keep compatibility if needed).
- Unify task parameter names across schema/example/runtime; support spec-first names and legacy aliases.
- Support `go:` namespace for Tasks and Screens consistently (strip prefix + registry lookup).
- Implement missing builtin tasks declared in schema (systemdService, dbusService, permission, net_script, removeDesktopEntry, rollback).

## Medium priority
- Add preflight environment detection and populate `InstallContext.Env` (distro/arch/desktop/disk/privilege info).
- Add privilege strategy (sudo/pkexec/none) and re-exec when required by tasks.
- Add headless parameters (`--accept-license`, `--install-dir`, `--install-type`, `--set`) to fill context and satisfy guards.
- Implement license scroll-to-end gating when `requireScrollToEnd: true`.
- Add step sidebar + branding placeholders in GUI layout.

## Low priority
- Enhance Summary/Finish screens with task plan, error summaries, and log export.
- Add basic i18n resource loading (zh/en) for built-in UI strings.
