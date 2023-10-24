# Contributing

## Getting Started

- With issues:
  - Use the search tool before opening a new issue.
  - Please provide source code and commit sha if you found a bug.
  - Review existing issues and provide feedback or react to them.

- With pull requests:
  - Open your pull request against `main`
  - Your pull request should have no more than two commits, if not you should squash them.
  - It should pass all tests in the available continuous integration systems such as GitHub Actions.
  - You should add/modify tests to cover your proposed code changes.

## Making Changes

- Create a branch from where you want to base your work
  - We typically name branches according to the following format: `helpful_name_<issue_number>`
- Make commits of logical units
- Make sure your commit messages are in a clear and readable format, example:

  ```
  create_files: fixed bug in creating many files.
    
  * creqte files with lock
  * cleanup expired files
  * ...
  ```

- If you're fixing a bug or adding functionality it probably makes sense to write a test
- If you're submitting a new feature, please document it on the README.
- Make sure to run `make fmt` and `make test` in the root of the repo to ensure that your code is
  properly formatted and that tests pass (we use GitHub Actions for continuous integration)

## Submitting Changes

- Push your changes to your branch in your fork of the repository
- Submit a pull request against keepshare's repository `main` branch
- Comment in the pull request when you're ready for the changes to be reviewed: `"ready for review"`
