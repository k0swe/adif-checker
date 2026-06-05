# adif-checker

`adif-checker` is a small Go command-line tool that validates whether an ADIF file is well-formed.

## What it checks

- ADIF tags are terminated correctly
- Tag names and lengths are valid
- Tag data does not run past the end of the file
- Header content is accepted until `<EOH>`
- Unexpected bytes outside valid ADIF content are rejected

## Usage

Run the checker against an ADIF file:

```bash
go run . /path/to/log.adi
```

If the file is valid, the program prints:

```text
ADIF is well-formed
```

If the file is invalid, the program prints an error to stderr and exits with a non-zero status.

## Development

Run the test suite from the repository root:

```bash
go test ./...
```