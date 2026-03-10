{
  description = "PxeHub";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }:
  let
    lib = nixpkgs.lib;
    eachSystem = lib.genAttrs [ "x86_64-linux" "aarch64-linux" ];
  in
  {
    packages = eachSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in {
        default = pkgs.buildGo126Module {
          pname = "pxehub";
          version = "0.1.0";
          src = ./.;
          vendorHash = null;

          nativeBuildInputs = with pkgs; [
            makeWrapper
          ];

          postFixup = ''
            wrapProgram $out/bin/pxehub --set PATH ${lib.makeBinPath (with pkgs; [
              dnsmasq
            ])}
          '';
        };
      });

    devShells = eachSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in {
        default = pkgs.mkShell {
          packages = [
            pkgs.go_1_26
            pkgs.gopls
            pkgs.delve
          ];

          shellHook = ''
            export GOPATH=$PWD/.go
            export PATH=$GOPATH/bin:$PATH

            export HTTP_BIND=:8080
            export INTERFACE=lo
            export DHCP_RANGE_START=192.168.50.10
            export DHCP_RANGE_END=192.168.50.254
            export DHCP_MASK=255.255.255.0
            export DHCP_ROUTER=192.168.50.1
            export DNS_SERVERS=1.1.1.1,1.0.0.1
          '';
        };
      });

    nixosModules.pxehub = { config, lib, pkgs, ... }:
        let
        cfg = config.services.pxehub;
        in
        {
        options.services.pxehub = {
            enable = lib.mkEnableOption "PxeHub service";

            package = lib.mkOption {
              type = lib.types.package;
              default = self.packages.${pkgs.system}.default;
            };

            httpBind = lib.mkOption {
              type = lib.types.str;
              default = ":80";
              example = "192.168.1.1:80";
            };

            interface = lib.mkOption {
              type = lib.types.str;
              example = "eth0";
            };

            dhcpRangeStart = lib.mkOption {
              type = lib.types.str;
              example = "192.168.50.10";
            };

            dhcpRangeEnd = lib.mkOption {
              type = lib.types.str;
              example = "192.168.50.254";
            };

            dhcpMask = lib.mkOption {
              type = lib.types.str;
              example = "255.255.255.0";
            };

            dhcpRouter = lib.mkOption {
              type = lib.types.str;
              example = "192.168.50.1";
            };

            dnsServers = lib.mkOption {
              type = lib.types.listOf lib.types.str;
              default = [ "1.1.1.1" "1.0.0.1" ];
            };
        };

        config = lib.mkIf cfg.enable {
            systemd.services.pxehub = {
                description = "PxeHub Server";
                after = [ "network.target" ];
                wantedBy = [ "multi-user.target" ];

                serviceConfig = {
                    ExecStart = ''
                    ${cfg.package}/bin/pxehub
                    '';

                    Environment = [
                    "HTTP_BIND=${cfg.httpBind}"
                    "INTERFACE=${cfg.interface}"
                    "DHCP_RANGE_START=${cfg.dhcpRangeStart}"
                    "DHCP_RANGE_END=${cfg.dhcpRangeEnd}"
                    "DHCP_MASK=${cfg.dhcpMask}"
                    "DHCP_ROUTER=${cfg.dhcpRouter}"
                    "DNS_SERVERS="
                    ];

                    Restart = "always";
                };
            };
            };
        };
    };
}