services:
{{- range $i, $node := .Nodes}}
  injectived{{$i}}:
    container_name: injstress-{{$node.Moniker}}
    image: "{{$.Image}}"
    environment:
      - DEBUG={{if eq $.Debug true}}1{{else}}0{{end}}
      - NODE_HOME=/root/.injectived
      - CHAIN_ID={{$.ChainID}}
      - MONIKER={{$node.Moniker}}
    cap_add:
      - SYS_PTRACE
    security_opt:
      - seccomp:unconfined
    {{- if $node.PortsExposed}}
    ports:
      - "{{$node.Ports.P2P}}-{{$node.Ports.RPC}}:26656-26657"
      - "{{$node.Ports.API}}:10337"
      - "{{$node.Ports.GRPC}}:9900"
      - "{{$node.Ports.GRPCWeb}}:9091"
      - "{{$node.Ports.PProf}}:2345"
      {{- if $node.Ports.EVMRPC}}
      - "{{$node.Ports.EVMRPC}}-{{$node.Ports.EVMWSPort}}:8545-8546"
      {{- end}}
      {{- if $node.Ports.Prometheus}}
      - "{{$node.Ports.Prometheus}}:26660"
      {{- end}}
    {{- end}}
    volumes:
      - {{$node.Home}}:/root/.injectived:Z
    networks:
      localnet:
        ipv4_address: {{$node.IPAddr}}
    {{- if $node.DependsOn}}
    depends_on:
      {{- range $dep := $node.DependsOn}}
      - injectived{{$dep}}
      {{- end}}
    {{- end}}
    command: injectived --home=/root/.injectived start
{{- end}}

networks:
  localnet:
    driver: bridge
    ipam:
      driver: default
      config:
        - subnet: {{.Network.Subnet}}
