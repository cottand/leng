# Nix

Leng is also packaged as [a Nix flake](../../flake.nix).

## Running

You can simply run `nix run github:cottand/leng` to run latest `master`.

## Installing

### In your flake

```nix
{
  # pinned version for safety
  inputs.grimn.url = "github:cottand/leng/v1.3.1"; 

  outputs = { self, leng }: {
    # Use in your outputs
  };
}
```


## Developing

The flake's development shell simply includes Go 1.21+ and a [fish](https://fishshell.com/) shell. You can enter it with `nix develop`.

