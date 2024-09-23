# Bitty

This is a custom BitTorrent client built from scratch using Go. The client is capable of downloading torrents based on `.torrent` files. It utilizes several custom packages to handle torrent metadata parsing, peer connections, and file downloading. Future plans include running the client in a browser using WebAssembly.

## Features

- Decodes `.torrent` files using the `bencode` encoding.
- Connects to peers using the BitTorrent protocol.
- Downloads torrent content and writes it to a specified destination.
- Built with extensibility in mind, with plans for future WebAssembly integration.

## Installation

### Prerequisites

- Go 1.16 or higher
- Git

### Clone the Repository

```bash
git clone https://github.com/peterkwesiansah/bitty
cd bitty
```

### Install Dependencies

Install the required Go packages:

```bash
go mod tidy
```

## Usage

### Running the BitTorrent Client

To run the BitTorrent client, you need to provide a source `.torrent` file and a destination path where the downloaded content will be stored.

1. Compile and run the program:

```bash
go run main.go -src /path/to/source.torrent -dst /path/to/destination
```

Example:

```bash
go run main.go -src debian-12.5.0-amd64-netinst.iso.torrent -dst ./debian-12.5.0-amd64-netinst.iso
```

2. The program will automatically start downloading the content from the torrent and save it in the specified destination folder.

### Command-Line Flags

- `-src`: Path to the source `.torrent` file. (Required)
- `-dst`: Path where the downloaded content should be saved. (Required)

If the `src` or `dst` flags are missing, the program will prompt you to use the correct format.

```bash
run instead: go run main.go -src /path/to/source -dst /path/to/destination
```

## Code Overview

### Main Components

- **Torrent File Parsing**: The client reads the torrent metadata using `bencode` decoding and prepares the necessary information for downloading.
- **Peer Discovery**: The client connects with peers through the tracker and establishes P2P connections.
- **Downloading**: The client downloads the torrent content piece by piece, using the `Torrent` struct to manage the process.

### Key Packages

- `bencodeTorrent`: Handles parsing of the `.torrent` file and decoding the `bencode` format.
- `peers`: Manages peer discovery and P2P communication.
- `torrentfile`: Manages the actual downloading of the torrent data, ensuring that pieces are correctly ordered and written to the destination.

## Plans for WebAssembly

One of the future goals is to compile this BitTorrent client to WebAssembly (Wasm), allowing it to run in a browser environment. The major steps involved in achieving this include:

- **Modifying the network layer**: Browser environments handle WebSockets and HTTP differently from native environments, so the peer connection logic will need to adapt accordingly.
- **File system changes**: Browsers use an in-memory filesystem or the `FileSystem` API for writing downloaded content, so the current file-writing logic will need to be adjusted for WebAssembly.

### To Do

- [ ] Adapt network and peer management for WebAssembly (Wasm).
- [ ] Implement browser-based UI for downloading torrents.
- [ ] Test WebAssembly builds across major browsers.

## Contributing

Contributions are welcome! If you have ideas for new features or improvements, feel free to fork the repository and submit a pull request. Make sure to write unit tests for any new functionality.

## License

This project is licensed under the MIT License. See the `LICENSE` file for more details.
