<div align="center">

<img src=".github/images/aegis_squad.png" alt="Logo" width="500"/>

# Squad Aegis

Incompetent Legion's fork of Squad Aegis, a comprehensive control panel for Squad game server administration

</div>

## Incompetent Legion Fork

This repository is the Incompetent Legion fork of [Codycody31/squad-aegis](https://github.com/Codycody31/squad-aegis). It keeps the upstream project's core goal of providing a web control panel for Squad servers, but carries changes needed to run it reliably against our server setup and deployment workflow.

The fork currently focuses on:

- **SAT/SPM/SuperMod compatibility**: Expanded log parsing for SuperMod/SAT/SPM variants, including `PostLogin`, draw events, vehicle possess events, and map names that use prefixes such as `SU_` or suffixes such as `_HalfCap`.
- **More reliable server and player data**: Fixes for connection feed attribution, session counts, combat/player history totals, team balancer winner detection, and chart ranges where different metrics start at different times.
- **Dashboard and UI fixes**: Better chart time labels, consistent server metric ranges, map thumbnail lookup for modded layers, and layout protection for long unbroken chat messages.
- **Container publishing**: A GitHub Actions workflow builds and publishes Docker images to GHCR, and the compose example uses that image path.
- **Authentication hardening**: Sessions now use server-set HttpOnly cookies, hashed session tokens, safer login errors, Valkey-backed failed-login rate limiting, password-policy enforcement, and revocation of other sessions after password changes.

The reason for these changes is practical: the upstream project assumes mostly vanilla Squad log formats and local/manual deployment patterns, while Incompetent Legion runs a modded server stack with SuperMod/SAT/SPM and wants reproducible container deploys with fewer dashboard blind spots.

## Notice

The project is currently in the early stages of development and is not ready for production use. Expect breaking changes and bugs as I work on making this stable and not a bunch of ai written spaghetti code from the PoC.

## Overview

**Squad Aegis** is an all-in-one control panel designed to manage multiple Squad game servers efficiently. Whether you're running a single server or a complex cluster, Squad Aegis provides a centralized interface to keep everything under control.

## Acknowledgments

See [Acknowledgements](https://squad-aegis.com/docs/acknowledgements) for more details.

## Features

### Core Features

- **Multi-Server Management**: Central hub allowing supervision and control of multiple servers from a single interface.
- **Role-Based Access**: Define specific permissions for users to ensure only authorized actions are executed.
- **RCON Interface**: Engage with server commands using a user-friendly web-based interface.
- **Audit System**: Detailed logging of administrative activities for transparency and security.
- **Plugin Architecture**: Extend functionality with a modular plugin system, allowing for custom features and integrations.

## Screenshots

![Dashboard](.github/images/dashboard.png)

## Contributing

We welcome contributions!

## Support

For help and support, you can refer to the resources below:

- **Issue Tracker**: [Submit bug reports and feature requests](https://github.com/Codycody31/squad-aegis/issues)

## License

Squad Aegis is licensed under the Gnu General Public License v3.0 (GPL-3.0). See the [LICENSE](LICENSE) and [NOTICE](NOTICE) files for more details.
