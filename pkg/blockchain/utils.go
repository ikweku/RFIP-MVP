package blockchain

import (
    "bytes"
    "encoding/binary"
    "log"
)

// IntToHex converts an int64 to a byte array
func IntToHex(num int64) []byte {
    buff := new(bytes.Buffer)
    err := binary.Write(buff, binary.BigEndian, num)
    if err != nil {
        log.Panic(err)
    }

    return buff.Bytes()
}

// ReverseBytes reverses a byte array
func ReverseBytes(data []byte) {
    for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
        data[i], data[j] = data[j], data[i]
    }
}

// ValidateAddress checks if address is valid
func ValidateAddress(address string) bool {
    pubKeyHash := Base58Decode([]byte(address))
    actualChecksum := pubKeyHash[len(pubKeyHash)-4:]
    version := pubKeyHash[0]
    pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
    targetChecksum := Checksum(append([]byte{version}, pubKeyHash...))

    return bytes.Compare(actualChecksum, targetChecksum) == 0
}

// Checksum generates a checksum for a public key
func Checksum(payload []byte) []byte {
    firstSHA := sha256.Sum256(payload)
    secondSHA := sha256.Sum256(firstSHA[:])

    return secondSHA[:addressChecksumLen]
}

