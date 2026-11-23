# Contributing to go-passbolt

## Introduction

Thank you for your interest in go-passbolt. We welcome contributions from everyone, and this guide explains how to make one that we can review and merge.

The team maintaining this SDK is small, so every pull request has a real review cost. The process below exists to make sure your time and ours goes into changes we can actually accept.

### Code of Conduct

First things first, please read our [Code of Conduct](https://www.passbolt.com/code_of_conduct).
Help us keep Passbolt open and inclusive!

## Where to go

### Reporting a security issue

If you've found a security issue in Passbolt, please do not open a public issue. Follow our responsible disclosure process: https://www.passbolt.com/docs/contribute/security/vulnerability/

### Reporting a bug

Open an [issue](https://github.com/passbolt/go-passbolt/issues/new/choose) using the bug report template.
Check first that a similar issue does not already exist, open or closed.

### Requesting a feature, asking a question, getting help

Please use the community forum, not GitHub issues:

- Feature requests and changes: https://community.passbolt.com/c/backlog
- Installation and configuration help: https://community.passbolt.com/c/installation-issues
- Questions: https://community.passbolt.com/c/uncategorized

## Contributing code

### Before you write any code

Please follow these steps, in order, before opening a pull request:

1. **Search the community forum.** For a new feature or a change of behaviour, start or join a thread in the [backlog](https://community.passbolt.com/c/backlog) so that interest and design can be discussed.
2. **Search the [issues](https://github.com/passbolt/go-passbolt/issues?q=is%3Aissue)**, open and closed, to make sure nobody has already reported or proposed it.
3. **Open an issue.** Use the bug report template for a bug, or the code change proposal template for anything else, linking the forum thread.
4. **Wait for a maintainer to reply.** When we agree with the change, we will say so on the issue and add the `accepted` label. We may also ask you to change the scope, or tell you that it does not fit the SDK.
5. **Only then open a pull request**, referencing the issue with `Closes #<number>`.

Pull requests that do not link an accepted issue may be closed without review. This is about review capacity, not about the quality of your work: it keeps us from having to turn down changes after you have spent time on them.

Fixes to typos and documentation do not need an issue first.

### What a good pull request looks like

- **One issue per pull request.** Keep it small and focused; unrelated changes belong in separate issues.
- **A few at a time.** Please don't open many pull requests at once. Wait for the open ones to be reviewed before sending more.
- **Tests.** Bug fixes come with a test that fails without the fix. New behaviour comes with tests.
- **Checks pass locally:**
  ```bash
  go test ./...
  go vet ./...
  golangci-lint run
  ```
- **Compatibility.** The SDK is imported by other projects, so it must keep compiling with the Go version declared in `go.mod`. Avoid breaking changes to exported APIs; if one is unavoidable, discuss it on the issue first.
- **Target the `main` branch** and fill in the pull request template.

### AI-assisted contributions

You may use AI tools, but you remain the author of the contribution and responsible for it:

- **Disclose it.** Say in the pull request which tool you used and what for.
- **Understand it.** You must be able to explain every line of the change and answer review comments yourself.
- **Write the description yourself.** A tool does not know why you want to make this change.
- Pull requests that are clearly unreviewed generated output will be closed.

The same process applies: an AI-assisted change still needs an accepted issue first.

### Licence

Contributions are accepted under the project's [licence](LICENSE). The first time you open a pull request, the CLA assistant will ask you to sign the contributor licence agreement.
