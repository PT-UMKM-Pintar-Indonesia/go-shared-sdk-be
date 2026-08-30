package sdk_helper

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	"github.com/sirupsen/logrus"
)

type (
	BankType string
	BankCode struct {
		Name   string
		Prefix string
	}

	BankConfig struct {
		Prefix    string
		MaxLength int
	}

	VAGenerator struct {
		configs    map[BankType]BankConfig
		bufferPool sync.Pool
	}
)

const (
	BCA      BankType = "BCA"
	MANDIRI  BankType = "MANDIRI"
	BRI      BankType = "BRI"
	BNI      BankType = "BNI"
	BSI      BankType = "BSI"
	CIMB     BankType = "CIMB"
	PERMATA  BankType = "PERMATA"
	BTN      BankType = "BTN"
	DANAMON  BankType = "DANAMON"
	MAYBANK  BankType = "MAYBANK"
	PANIN    BankType = "PANIN"
	UOB      BankType = "UOB"
	OCBC     BankType = "OCBC"
	BJB      BankType = "BJB"
	BPD_DIY  BankType = "BPD_DIY"
	NAGARI   BankType = "NAGARI"
	SINARMAS BankType = "SINARMAS"
	MEGA     BankType = "MEGA"
	NEO      BankType = "NEO"
)

var bankCodes = []BankCode{
	{Name: "bca", Prefix: "39012"},
	{Name: "mandiri", Prefix: "89508"},
	{Name: "bni", Prefix: "82710"},
	{Name: "bri", Prefix: "85013"},
	{Name: "cimb", Prefix: "70030"},
}

func NewVaGenerator() *VAGenerator {
	return &VAGenerator{
		configs: map[BankType]BankConfig{
			BCA:      {Prefix: "11234", MaxLength: 15},
			MANDIRI:  {Prefix: "89408", MaxLength: 16},
			BRI:      {Prefix: "12345", MaxLength: 18},
			BNI:      {Prefix: "8808", MaxLength: 16},
			BSI:      {Prefix: "900", MaxLength: 16},
			CIMB:     {Prefix: "5919", MaxLength: 16},
			PERMATA:  {Prefix: "8625", MaxLength: 16},
			BTN:      {Prefix: "90333", MaxLength: 16},
			DANAMON:  {Prefix: "7755", MaxLength: 16},
			MAYBANK:  {Prefix: "7812", MaxLength: 16},
			PANIN:    {Prefix: "8877", MaxLength: 16},
			UOB:      {Prefix: "8989", MaxLength: 16},
			OCBC:     {Prefix: "92001", MaxLength: 16},
			BJB:      {Prefix: "4255", MaxLength: 16},
			BPD_DIY:  {Prefix: "550", MaxLength: 16},
			NAGARI:   {Prefix: "9901", MaxLength: 16},
			SINARMAS: {Prefix: "831", MaxLength: 16},
			MEGA:     {Prefix: "9900", MaxLength: 16},
			NEO:      {Prefix: "99202", MaxLength: 16},
		},
		bufferPool: sync.Pool{
			New: func() interface{} {
				b := make([]byte, 18)
				return &b
			},
		},
	}
}

func (h *VAGenerator) Validate(bank BankType, vaNumber string) error {
	config, exists := h.configs[bank]

	if !exists {
		return errors.New("bank is not supported")
	}

	if len(vaNumber) != config.MaxLength {
		return errors.New("wrong va number length")
	}

	prefixLen := len(config.Prefix)

	for i := 0; i < len(vaNumber); i++ {
		if vaNumber[i] < '0' || vaNumber[i] > '9' {
			return errors.New("VA must be numeric")
		}

		if i < prefixLen && vaNumber[i] != config.Prefix[i] {
			return errors.New("wrong va prefix")
		}
	}

	return nil
}

func (h *VAGenerator) Generate(bank BankType, customerID string) (string, error) {
	config, exists := h.configs[bank]
	if !exists {
		return sdk_cons.EMPTY, errors.New("bank is not supported")
	}

	bufPtr := h.bufferPool.Get().(*[]byte)
	buf := *bufPtr
	defer h.bufferPool.Put(bufPtr)

	cleanLen := 0
	for i := 0; i < len(customerID); i++ {
		if customerID[i] >= '0' && customerID[i] <= '9' {
			buf[cleanLen] = customerID[i]
			cleanLen++
		}
	}

	if cleanLen == 0 {
		return sdk_cons.EMPTY, errors.New("customer ID not valid")
	}

	allowedUniqueLength := config.MaxLength - len(config.Prefix)
	result := make([]byte, config.MaxLength)
	copy(result, config.Prefix)

	if cleanLen >= allowedUniqueLength {
		copy(result[len(config.Prefix):], buf[cleanLen-allowedUniqueLength:cleanLen])
	} else {
		paddingLen := allowedUniqueLength - cleanLen

		for i := 0; i < paddingLen; i++ {
			result[len(config.Prefix)+i] = '0'
		}

		copy(result[len(config.Prefix)+paddingLen:], buf[:cleanLen])
	}

	return string(result), nil
}

func generateRandomDigits(length int) (string, error) {
	const digits = "0123456789"

	result := make([]byte, length)
	for i := range result {
		n, err := randomInt(0, 9)

		if err != nil {
			return sdk_cons.EMPTY, err
		}

		result[i] = digits[n]
	}

	return string(result), nil
}

func calculateCheckDigit(number string) (int, error) {
	sum := 0

	for i, c := range number {
		digit := int(c - '0')

		if digit < 0 || digit > 9 {
			return 0, fmt.Errorf("invalid digit in input: %c", c)
		}

		if (i+1)%2 == 0 {
			digit *= 2

			if digit > 9 {
				digit = digit/10 + digit%10
			}
		}

		sum += digit
	}

	return (10 - (sum % 10)) % 10, nil
}

func randomInt(min, max int) (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()) + min, nil
}

func GenerateRandomVA(bankCode string) string {
	prefix := sdk_cons.EMPTY

	for _, bc := range bankCodes {
		if bc.Name == strings.ToLower(bankCode) {
			prefix = bc.Prefix
			break
		}
	}

	if prefix == sdk_cons.EMPTY {
		n, err := randomInt(0, len(bankCodes)-1)

		if err != nil {
			logrus.Error(err)
			return sdk_cons.EMPTY
		}

		prefix = bankCodes[n].Prefix
	}

	customerNumLength, err := randomInt(10, 12)
	if err != nil {
		logrus.Error(err)
		return sdk_cons.EMPTY
	}

	customerNumber, err := generateRandomDigits(customerNumLength)
	if err != nil {
		logrus.Error(err)
		return sdk_cons.EMPTY
	}

	checkDigit, err := calculateCheckDigit(prefix + customerNumber)
	if err != nil {
		logrus.Error(err)
		return sdk_cons.EMPTY
	}

	return prefix + customerNumber + fmt.Sprint(checkDigit)
}
