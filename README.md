# gopacket-srv6
`gopacket` (v1.1.19) is missing decoder and encoder for `IPv6Routing` Extension with `RoutingType = 4` (SRv6)
This file module is a patch to enable using gopacket with SRv6.

For information, here are some related PRs:
* [703](https://github.com/google/gopacket/pull/703)   : implementation of SRv6 extension header decoding (LastEntry, Flags and Tag are missing)
* [879](https://github.com/google/gopacket/pull/879)   : implementation of SRv6 extension header decoding
* [889](https://github.com/google/gopacket/pull/889)   : implementation of SRv6 extension header decoding
* [1040](https://github.com/google/gopacket/pull/1040) : implementation of SRv6 extension header decoding + serialization

Nota: For the decoding part, this code is based on [PR 1040](https://github.com/google/gopacket/pull/1040), by [Takanori Hirano](github.com/hrntknr) and [Gernot Vormayr](https://github.com/notti).
