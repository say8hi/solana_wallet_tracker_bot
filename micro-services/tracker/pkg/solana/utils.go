package solana

import (
    "github.com/mr-tron/base58"
    "fmt"
    "regexp"
)

var (
    solanaAddressRegex = regexp.MustCompile(`^[1-9A-HJ-NP-Za-km-z]{32,44}$`)
    
    ErrInvalidAddressFormat = fmt.Errorf("invalid address format")
    ErrInvalidAddressLength = fmt.Errorf("invalid address length")
    ErrInvalidChecksum     = fmt.Errorf("invalid checksum")
)

// ValidateSolanaAddress проверяет валидность Solana адреса
func ValidateSolanaAddress(address string) error {
    // Проверка формата
    if !solanaAddressRegex.MatchString(address) {
        return ErrInvalidAddressFormat
    }

    // Декодируем base58
    decoded, err := base58.Decode(address)
    if err != nil {
        return fmt.Errorf("failed to decode address: %w", err)
    }

    // Проверка длины
    if len(decoded) != 32 {
        return ErrInvalidAddressLength
    }

    return nil
}

