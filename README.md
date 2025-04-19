There may be numerous version of oui lookup tools, but this one uses Wireshark manufacturer database unlike others using IEEE database.
The advantage of Wireshark DB is it has more entries and supports 28 and 36 bit mac blocks too. The IEEE one only have 24 bit blocks.

e.g. if you search FC:D2:B6:20:11:22 which is mac-address from a 28 bit block assigned to Soma GmbH.
The IEEE DB will match it against a defuakt 24 bit and shows as "IEEE Registration Authority", however Wireshark manufacturer database clearly shows it as FC:D2:B6:20:00:00/28 Soma GmbH

# oui2manuf

A simple and efficient Go-based command-line tool to look up MAC address manufacturers using the **Wireshark OUI/manufacturer database**.

## Why This Tool?

There are several OUI lookup tools available, but **`oui2manuf` stands out by using the Wireshark manufacturer database instead of the IEEE registry**. This offers significant advantages:

- **Wireshark's database supports more granular resolution**, including **28-bit and 36-bit MAC address blocks**.
- **IEEE's database only contains 24-bit blocks**, which can result in incorrect or generic lookups.

### Real-world Example

```bash
MAC: FC:D2:B6:20:11:22
IEEE DB result: IEEE Registration Authority (Generic 24-bit match)
Wireshark DB result: Soma GmbH (Accurate 28-bit match: FC:D2:B6:20/28)
```

This makes oui2manuf much more accurate when identifying hardware vendors, especially for newer or less common manufacturers.

## Features

- Fetches and parses the latest `manuf.gz` file from Wireshark's database.
- Supports MAC blocks of varying lengths: `/24`, `/28`, `/36`.
- Performs longest prefix match in all the blocks (36-bit > 28-bit > 24-bit) for most accurate vendor lookup.
- Caches the parsed database locally to avoid repeated downloads.
- CLI interface for easy vendor lookup.

## Installation
```bash
git clone https://github.com/0xdeafc0de/oui2manuf.git
cd oui2manuf
go build -o oui2manuf
```

## Usage
```bash
./oui2manuf <mac_address>
```

## Examples -

./oui2manuf 00:00:20:33:11:22
# Manufacturer: Dataindustri (24-bit match)

./oui2manuf FC:D2:B6:1A:00:01
# Manufacturer: Link (28-bit match)

./oui2manuf 8C:1F:64:DC:61:99
# Manufacturer: R&K (36-bit match)

