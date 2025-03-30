# Nix

Leng is also packaged as [a Nix flake](../../flake.nix).

## Running

You can simply run `nix run github:cottand/leng` to run latest `master`.

## Installing in NixOS via a Module

The leng flake also exports a NixOS module for easy deployment on NixOS machines.

Please refer to [Configuration](./Configuration.md) for the options you can use under `services.leng.configuration. = ...`.

### In your flake

```nix
{
  inputs = {
    # pinned version for safety
    leng.url = "github:cottand/leng/v1.7.1"; 
    leng.nixpkgs.follows = "nixpkgs";
  };

  outputs = { self, leng, ... }: {
    # Use in your outputs
    nixosConfigurations."this-is-a-server-innit" = nixpkgs.lib.nixosSystem {
      modules = [ 
        ./configuration.nix
        leng.nixosModules.default #  <- import leng module
        {
          services.leng = {       # <-- now you can use services.leng!
            enable = true;
            configuration = {
              api = "127.0.0.1:8080";
              metrics.enabled = true;
              blocking.sourcesStore = "/var/lib/leng-sources";
            };
          };
        }
      ];
    };
  };
}
```


### Legacy Nix

Add the following inside your configuration.nix:
```nix
{pkgs, lib, ... }: {
  imports = [
    # import leng module
    (builtins.getFlake "github:cottand/leng/v1.7.1").nixosModules.default 
  ];
    
  # now you can use services.leng!
  services.leng = {       
    enable = true;
    configuration = {
      api = "127.0.0.1:8080";
      metrics.enabled = true;
      blocking.sourcesStore = "/var/lib/leng-sources";
    };
  };
  
}
```

### Using leng for local DNS

If you want to use leng to resolve the DNS queries of the machine you are installing it on, you should
also add it to the local nameservers:

```nix
networking.nameservers = [ "127.0.0.1" ];
```


## Developing

The flake's development shell simply includes Go 1.21+ and a [fish](https://fishshell.com/) shell. You can enter it with `nix develop`.

