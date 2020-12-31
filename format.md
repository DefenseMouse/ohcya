#
# Cypress SoC File Format
# cyacd and cyacd2
#

# cyacd
#

General info:
 - Short header followed by flash programming data
 - Each data line represents an entire row of flash data
 - data is stored as ASCII data in big-endian 
 - checksum type:
	- 0 is data summation
	- 1 is CRC-16

Header:
 - 4 bytes Silicon ID
 - 1 byte Silicon Revision
 - 1 byte Checksum Type

Flash:
 - ":"
 - 1 byte array ID
 - 2 bytes row number
 - 2 bytes data length
 - N bytes of data
 - 1 byte checksum


# cyacd2
#

General info:
 - bytes in the file are represented by hex text
 - all multi-byte fields are little-endian

Encryption Initial Vector:
 - "@EIV" at start of line
 - Data on the line should be used in the SetEIV command to DFU

Application verification information:
 - "@APPINFO"
 - address of `__cy_app_verify_start`,
 - length in hex
 - e.g. @APPINFO:0x1234,0x20
 - the "signature" (`cy_app_signature`) seems to be a 32bit CRC of the image pointed 
   to by @APPINFO; but it seems that the sig can be changed based on the dev's
   requirements. Don't yet see how to change this to something "secure"

Header:
 - 1 byte: file version
	- should be 1
 - 4 bytes: silicon ID
	- refers to a speciifc cypress unit; can't find public mapping
 - 1 byte: silicon revision
 - 1 byte: checksum type
	- can be 0: checksum (summation?)
	- 1: CRC
 - 1 byte: app ID
	- see Figure 6 to map this number to an object: 
		PSoC 6 MCU Device Firmware Update Software Development Kit Guide
		Page 9
 - 4 bytes: product ID
	- the dev's product; not Cypress'

Data:
 - ":" starts the line
 - 4 byte address
 - N bytes data



