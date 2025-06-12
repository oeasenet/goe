# 17. Contributing to Goe 🤝

We welcome contributions to the Goe framework! Whether you're fixing a bug, improving documentation, or proposing a new feature, your help is appreciated. This guide outlines how to contribute effectively.

## Code of Conduct

Before contributing, please read our [Code of Conduct](../CODE_OF_CONDUCT.md). We expect all contributors to adhere to it to ensure a welcoming and inclusive environment for everyone.

## Ways to Contribute

*   **Reporting Bugs**: If you find a bug, please open an issue on the GitHub repository. Include:
    *   A clear and descriptive title.
    *   Steps to reproduce the bug.
    *   What you expected to happen.
    *   What actually happened.
    *   Your Goe version, Go version, and OS.
    *   Any relevant code snippets or logs.
*   **Suggesting Enhancements or New Features**: Open an issue to discuss your ideas. Provide:
    *   A clear description of the proposed enhancement or feature.
    *   The problem it solves or the value it adds.
    *   Potential implementation ideas (if any).
*   **Improving Documentation**: If you find errors, omissions, or areas that could be clearer in the documentation, please open an issue or submit a pull request with your improvements. This documentation itself is a great place to contribute!
*   **Writing Code**: Submit pull requests for bug fixes or new features.
*   **Reviewing Pull Requests**: Help review PRs submitted by others.

## Setting Up Your Development Environment

1.  **Prerequisites**:
    *   Go 1.21 or newer.
    *   Git.
    *   A GitHub account.

2.  **Fork the Repository**:
    *   Go to the Goe GitHub repository (let's assume `github.com/oease/goe` for this example).
    *   Click the "Fork" button to create your own copy of the repository.

3.  **Clone Your Fork**:
    ```bash
    git clone https://github.com/YOUR_USERNAME/goe.git
    cd goe
    ```

4.  **Add Upstream Remote**:
    This allows you to keep your fork in sync with the main repository.
    ```bash
    git remote add upstream https://github.com/oease/goe.git
    ```

5.  **Install Dependencies**:
    Goe uses Go modules. Dependencies should be automatically managed.
    ```bash
    go mod tidy
    ```

## Making Changes

1.  **Keep Your Fork Updated**:
    Before starting work on a new contribution, ensure your `main` (or `master`) branch is up-to-date with the upstream repository:
    ```bash
    git checkout main
    git pull upstream main
    git push origin main # Update your fork's main branch
    ```

2.  **Create a New Branch**:
    Create a descriptive branch for your changes:
    ```bash
    git checkout -b feature/my-new-feature  # For new features
    # or
    git checkout -b fix/bug-description     # For bug fixes
    # or
    git checkout -b docs/improve-config-guide # For documentation
    ```

3.  **Write Your Code/Documentation**:
    *   Make your changes.
    *   Follow existing code style and conventions (see below).
    *   Ensure your code is well-commented, especially for public APIs.
    *   If adding new features or fixing bugs, write or update relevant tests.

4.  **Code Style and Conventions**:
    *   **Go Standard Formatting**: Run `gofmt` or `goimports` on your code before committing. Many IDEs do this automatically.
        ```bash
        gofmt -w .
        # or
        goimports -w .
        ```
    *   **Linting**: While not strictly enforced by a specific linter in this guide, consider using a linter like [golangci-lint](https://golangci-lint.run/) to catch common issues. If the project adopts one, its configuration will be in the repository.
    *   **Comments**: Write clear and concise comments. Exported functions, types, and constants should generally have godoc comments.
    *   **Error Handling**: Follow Go's idiomatic error handling. Wrap errors for context.
    *   **Testing**: Strive for good test coverage.

5.  **Run Tests**:
    Before submitting your contribution, ensure all tests pass:
    ```bash
    go test ./...
    ```
    If the project has a `Makefile` with a test target, use that (e.g., `make test`).

6.  **Commit Your Changes**:
    *   Use clear and descriptive commit messages. A good format is:
        ```
        feat: Add support for new cache driver X

        This commit introduces a new cache driver X, providing...
        Fixes #123 (if applicable)
        ```
        Or for smaller changes:
        ```
        fix: Correct typo in configuration documentation
        ```
    *   Prefixes like `feat:`, `fix:`, `docs:`, `test:`, `chore:`, `refactor:` are common (Conventional Commits).
    *   Commit frequently with logical units of work.

7.  **Push to Your Fork**:
    ```bash
    git push origin feature/my-new-feature
    ```

## Submitting a Pull Request (PR)

1.  **Open a Pull Request**:
    *   Go to your fork on GitHub (`https://github.com/YOUR_USERNAME/goe`).
    *   You should see a button to "Compare & pull request" for your recently pushed branch. Click it.
    *   If not, go to the "Pull requests" tab and click "New pull request". Select your fork and branch to compare against the main repository's `main` (or `master`) branch.

2.  **Describe Your PR**:
    *   Provide a clear title for your PR.
    *   In the description, explain the changes you've made and why.
    *   If your PR addresses an existing issue, link to it (e.g., "Closes #456", "Fixes #789").
    *   Include any relevant information for the reviewers.

3.  **Code Review**:
    *   Project maintainers and other contributors will review your PR.
    *   Be prepared to discuss your changes and make adjustments based on feedback.
    *   Respond to comments পেশাদারীভাবে (professionally) and constructively.

4.  **CI Checks**:
    The Goe project likely has Continuous Integration (CI) checks (e.g., GitHub Actions) that automatically build your code, run tests, and perform other checks. Ensure these pass.

5.  **Merging**:
    Once your PR is approved and all checks pass, a maintainer will merge it into the main repository. 🎉

## Licensing

By contributing to Goe, you agree that your contributions will be licensed under its [Apache 2.0 License](../LICENSE).

Thank you for considering contributing to the Goe framework! Your efforts help make it better for everyone.

---

Next, we'll provide a handy [Appendix: Glossary](18-appendix-glossary.md) of terms.
```
