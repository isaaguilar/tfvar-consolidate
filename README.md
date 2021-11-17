# tfvar-consolidate

Consolidate `tfvars` into a single file. Accepts **HCL** and **JSON** formats. The output is **HCL**.

## Installation

Install the binary with go

```
go get -u github.com/isaaguilar/tfvar-consolidate
```

This will install `tfvar-consolidate` in `$GOPATH/bin`, which usually places it
into your shell `PATH` so you can then run it as `tfvar-consolidate`.


## Usage:

Run the binary and pass in the `tfvars` files to consolidate with the `-f` flag.
The var precedence order of consolidation will take the last value defined.
Like `terraform`, var consolidation does not traverse maps or arrays.

For example, when faced with two files that define the same value:

```hcl2
# file1.tfvars
instance = {
    class = "m5",
    size = "large"
}

# file2.tfvars
instance = {
    foo = "bar"
}
```

Then running the program:

```bash
tfvar-consolidate -f file1.tfvars -f file2.tfvars --out consolidated.tfvars
```

The result of consolidating `file1.tfvars` and `file2.tfvars` will be saved to
`consolidated.tfvars`. And the output will just show the last value:

```
# consolidated.tfvars
instance = {
  foo = "bar"
}
```
