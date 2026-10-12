# Dependabot PR triage

Last refreshed 2026-10-12 against `dev` @ `6154d75`.

Working notes from a sweep of the open Dependabot queue: what each PR carries,
what blocks it, and the ordering constraints this monorepo imposes.

## Ordering constraints worth remembering

**Cross-module `replace` coupling.** `app` and `backend` both `replace`
dialect and toolkit with local paths, so a bump inside `dialect` must land with
its consumers. Merging the shared library first leaves them unbuildable:

```
go: updates to go.mod needed; to update it:
	go mod tidy
```

Dependabot opens one PR per module and cannot see the coupling, so every
Go-module bump in a shared dependency arrives as a set that has to be merged as
one commit. The 2026-09-20 sweep hit this on all four Go PRs (#254, #255, #256,
#257); the 2026-08-22 sweep hit it on #125, #128 and #135.

A dependency that is direct in one module and indirect in another is the
sharpest form: Dependabot opens no PR for the module where it is indirect, so
its PR alone never builds. x/crypto is direct in `backend` and `dialect` and
indirect in `app`.

**Wails is three pins, not two.** The version is written in three places that
must agree:

| Pin | Where | Dependabot sees it |
|---|---|---|
| Go library | `app/go.mod` | yes |
| JS runtime | `app/frontend/package.json` | yes |
| `wails3` CLI | `WAILS_VERSION` in `ci.yml` and `ci-app.yml` | no |

The CLI generates the bindings, so a stale CLI pin is what
`Generated bindings are up to date` would catch, but only once the drift
changes generated output. PR CI cannot catch the desktop-build half of the
skew at all, because those jobs only run on `workflow_dispatch` or a push to
`dev` or a tag. By 2026-09-20 the CLI pin had sat at beta.12 while the library
moved to beta.17, five releases of silent drift. Bump all three together.

**Same-module PRs collide.** Within a module, the second and later PRs conflict
on `go.mod` once the first lands. Merge one at a time and let Dependabot rebase
between.

**npm PRs collide on the lockfile.** Git merges two lockfiles cleanly line-wise
but semantically inconsistently, and CI then hard-fails:

```
npm error `npm ci` can only install packages when your package.json and
npm error package-lock.json are in sync.
```

Merge one, let Dependabot rebase the other, then merge it. Four npm PRs landed
that way on 2026-09-20 without incident.

**Regenerating a lockfile needs CI's npm.** `npm install` rewrites the whole
file to whatever the local npm believes. An npm older than the one that last
wrote it silently drops the `libc` arrays that select musl against glibc
optional binaries; a newer one dedupes nested trees. Either way the bump
disappears into hundreds of unrelated lines. When a lockfile needs a version
changed by hand, edit the two entries for that package and leave the rest
alone.

**Red CI on a stale PR is not automatically a stale base.** The 2026-09-20
queue had sat for two days and every Go PR was red, which looked like the base
had been broken when they were opened. Refreshing them against a known-good
`dev` left them just as red: the failure was the `replace` coupling above, and
it was in the log the whole time. Read the failing job before rerunning it.

**A test red on `dev` blocks every PR that runs its leg.** Branch protection
wants `CI OK`, and a PR that touches `dialect` or `.github` runs
`Dialect / Tests`. The 2026-10-02 sweep found that leg already red on `dev`:
#380 added the PostgreSQL case `a routine body that is one RETURN expression`
without the inspector change that makes it pass; #381 landed that change on
2026-10-03. Nothing touching those paths can merge until the fix lands, however
green the bump itself is.

**typescript-eslint is three packages at one version.** `@typescript-eslint/eslint-plugin`
peers on `@typescript-eslint/parser` at the same minor, and the `typescript-eslint`
meta package pins both exactly. Dependabot opens one PR per package, so the plugin
PR alone fails `npm ci` with a peer conflict (#406 on 2026-10-02), and the meta
package often gets no PR at all because the npm queue is capped at five open
PRs. Bump the three together in one commit.

**A Go security release turns every PR red, and Dependabot cannot fix it.**
CI takes each module's toolchain from the `go` directive in its `go.mod`
(`GOTOOLCHAIN=local`), and `Security (*)` runs govulncheck against it. When Go
ships a patch release with standard library fixes, the gate goes red on `dev`
and on every PR at once, including Dependabot's own. Dependabot never bumps the
`go` directive, so the fix is a hand edit of every `go.mod`, and it belongs in
the sweep that is already open. The 2026-10-11 release of 1.26.9 did this to
the 2026-10-09 queue. `toolkit` was still on `go 1.25.13`, a line that no
longer receives security fixes, which is why it had to jump a minor.

**Stack Dependabot's lock patches, do not regenerate the lockfile.** With CI's
exact npm (Node 24.21.0, npm 11.19.0 on 2026-10-12), a lockfile written from
scratch still differed from the committed one by about 1200 lines of
formatting and ordering. What kept the diff to the bumped packages was to
cherry-pick each npm PR in turn (`git fetch origin refs/pull/N/head`),
resolve the one-package conflicts toward the newer version, finish the
package.json edits by hand, and only then run `npm install --package-lock-only`
and `npm dedupe --package-lock-only` under CI's npm, which flattens the nested
copies the one-at-a-time bumps left behind. `npm ci` afterwards is the proof
that package.json and the lock agree.

**An eslint major moves `@eslint/js` with it.** `js.configs.recommended` comes
from `@eslint/js`, not from `eslint`, so an eslint major bumped alone keeps the
old recommended set and CI stays green by accident. Dependabot opened eslint 10
(#512) without `@eslint/js` 10 because the npm queue is capped at five PRs.
eslint 10's recommended set added `preserve-caught-error` and
`no-useless-assignment`, each with one finding here.

**A typescript-eslint patch can add findings.** 8.71.1 reports a `$bindable`
prop that a component only writes as unused, although the parent reads it
through `bind:`. That is what made #509 red alone, not the peer conflict. No
upstream fix on 2026-10-12, and `eslint-plugin-svelte` 3.23.1 with
`svelte-eslint-parser` 1.9.0 does not change the verdict. The two props in
`InputMultiline.svelte` carry a scoped disable with the reason.

## Remaining queue

Empty once #536 lands. It folds every PR of the 2026-10-09 sweep.

## Merged

**2026-10-09** (superseded by #536, which bumps every pin at once) - #513,
#514, #515 and #517 (aws-sdk-go-v2 1.47.1, config 1.33.6, service/s3 1.114.0
and credentials 1.20.6, backend), #516 (x/sync 0.24.0, backend), #508
(wails/v3 beta.28, app), #510 (@wailsio/runtime beta.28), #512 (eslint
10.12.0), #509 and #511 (typescript-eslint and @typescript-eslint/eslint-plugin
8.71.1), #507 (dompurify 3.4.16), #506 (anthropics/claude-code-action 1.0.244).
#536 also moves every `go` directive to 1.26.9, the `WAILS_VERSION` pin to
beta.28, `@eslint/js` to 10.0.1 and `@typescript-eslint/parser` to 8.71.1.

**2026-10-02** (superseded by #414, which bumps every pin at once) - #399, #405
and #407 (sqlite 1.60.1, dialect, backend and app), #400, #401 and #403
(compress 1.20.1, app, backend and dialect), #404 (wails/v3 beta.26, app),
#402 (katex 0.18.9, landed as 0.18.10), #406 and #408 (@typescript-eslint
plugin and parser 8.71.0), #409 (monaco-editor 0.57.0), #410
(marked-katex-extension 5.1.13), #411 and #413 (github/codeql-action init and
analyze 4.38.2), #412 (anthropics/claude-code-action 1.0.237). #414 also moves
the `@wailsio/runtime` and `WAILS_VERSION` pins to beta.26, the
`typescript-eslint` meta package to 8.71.0, and libc 1.77.1 across the three
modules that use sqlite.

**2026-09-20** (direct) - #258 (typescript-eslint 8.70.0), #260
(@typescript-eslint/eslint-plugin 8.70.0), #261 (marked 18.0.13), #262
(@playwright/test 1.63.0), #264 and #266 (github/codeql-action init and
analyze 4.38.0), #265 (astral-sh/setup-uv 10.1.0).

**2026-09-20** (superseded by #271, which bumps every pin at once) - #254 and
#256 (x/crypto 0.57.0, dialect and backend), #255 and #257 (sqlite 1.59.0, app
and dialect), #259 (wails/v3 beta.22, app), #263 (@wailsio/runtime beta.23).
#271 lands x/crypto 0.57.0, sqlite 1.59.0 and libc 1.75.7 across all four Go
modules, and all three Wails pins on beta.23.

**2026-08-22** - #126 (sqlite 1.56.0, app), #129 (compress 1.19.2, app),
#127 (compress 1.19.2, backend), #130 (okms-sdk-go 0.5.4, backend, drops
RSA1_5 from the key-wrapping algorithms), #138 (x/crypto 0.55.0, backend),
#140 (eslint-plugin-svelte 3.23.0).

**2026-08-20** - #118 (dompurify 3.4.12), #122 (@sveltejs/adapter-static
3.0.10), #123 (eslint-config-prettier 10.1.8).

## Closed

**2026-08-22** - #139, #142, #143 (@tanstack/ai-client / -openai / -grok).
`b7d9e32` removed the whole family; these were opened against the pre-removal
`dev`, no longer merged, and would have resurrected the dependency.
