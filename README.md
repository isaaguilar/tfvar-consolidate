# tfvar-consolidate

Consolidate tfvars into a single file.

> **PROOF Of CONCEPT** Note that this project is a simple proof-of-concept and does not work exactly as stacking `-var-file` flags using the `terraform` command yet.

## Installation

Install the binary with go

```
go install github.com/isaaguilar/tfvar-consolidate
```


## Usage:

Run the binary and pass in the `tfvars` files to consolidate. The precedence
order of consolidation will take the last value defined. Currently, the
consolidation does not traverse maps or arrays types. So a map definition will
overwrite a map with the same root key.

Example:
```
tfvar-consolidate -f file1.tfvars -f file2.tfvars --out out.tfvars
```

The result of consolidating `file1.tfvars` and `file2.tfvars` will be saved to `out.tfvars`.

