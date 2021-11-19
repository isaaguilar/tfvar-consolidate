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
`consolidated.tfvars`. For this example, the output will just show the last
value since the var was repeated in both tfvar files:

```
# consolidated.tfvars
instance = {
  foo = "bar"
}
```

### Other options

#### JSON

JSON files/strings are also valid tfvars. To use a json file, the extension of
the filename must end in `.json`.

#### Environment Variables

This project can also take `TF_VAR` environment variables into account by using
`--use-envs` in the command. Environment variables have the least amount of
precedence, just the the terraform command.

#### Output file

The output file can create a file or overwrite a file when specifying the
`--out` flag. If the flag is omitted, a temp file is generated.

#### Example

The following is a full example:

```bash
echo 'a = "apple"' > a.tfvars
echo 'b = "banana"' > b.tfvars
echo '{"c_map":[{"c":"cake"},{"c":"candy"},{"c":"corn"}]}' > c.tfvars.json
export TF_VAR_name="foo"

tfvar-consolidate -f a.tfvars -f b.tfvars -f c.tfvars.json --use-envs --out consolidated.tfvars
```

The `consolidated.tfvars` is the generated file. By opening this file, the
output will look like the following:

```hcl
a = "apple"

b = "banana"

c_map = [
  {
    c = "cake"
  },
  {
    c = "candy"
  },
  {
    c = "corn"
  }
]

name = "foo"
```
