package publisher

import (
	"github.com/ethereum/go-ethereum/statediff/builder"
	"time"
	"os"
	"encoding/csv"
	"strconv"
	"strings"
	"path/filepath"
)

var (
	Headers = []string{
		"blockNumber", "blockHash", "accountAction",
		"code", "codeHash",
		"oldNonceValue", "newNonceValue",
		"oldBalanceValue", "newBalanceValue",
		"oldContractRoot", "newContractRoot",
		"storageDiffPaths",
	}

	timeStampFormat = "20060102150405.00000"
	deletedAccountAction = "deleted"
	createdAccountAction = "created"
	updatedAccountAction = "updated"
)

func createCSVFilePath(path string) string {
	now := time.Now()
	timeStamp := now.Format(timeStampFormat)
	filePath := filepath.Join(path, timeStamp)
	filePath = filePath + ".csv"
	return filePath
}

func (p *publisher) publishStateDiffToCSV(sd builder.StateDiff) error {
	filePath := createCSVFilePath(p.Config.Path)

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	var data [][]string
	data = append(data, Headers)
	for _, row := range accumulateCreatedAccountRows(sd) {
		data = append(data, row)
	}
	for _, row := range accumulateUpdatedAccountRows(sd) {
		data = append(data, row)
	}

	for _, row := range accumulateDeletedAccountRows(sd) {
		data = append(data, row)
	}

	for _, value := range data{
		err := writer.Write(value)
		if err != nil {
			return err
		}
	}

	return nil
}

func accumulateUpdatedAccountRows(sd builder.StateDiff) [][]string {
	var updatedAccountRows [][]string
	for _, accountDiff := range sd.UpdatedAccounts {
		formattedAccountData := formatAccountDiffIncremental(accountDiff, sd, updatedAccountAction)

		updatedAccountRows = append(updatedAccountRows, formattedAccountData)
	}

	return updatedAccountRows
}

func accumulateDeletedAccountRows(sd builder.StateDiff) [][]string {
	var deletedAccountRows [][]string
	for _, accountDiff := range sd.DeletedAccounts {
		formattedAccountData := formatAccountDiffEventual(accountDiff, sd, deletedAccountAction)

		deletedAccountRows = append(deletedAccountRows, formattedAccountData)
	}

	return deletedAccountRows
}

func accumulateCreatedAccountRows(sd builder.StateDiff) [][]string {
	var createdAccountRows [][]string
	for _, accountDiff := range sd.CreatedAccounts {
		formattedAccountData := formatAccountDiffEventual(accountDiff, sd, createdAccountAction)

		createdAccountRows = append(createdAccountRows, formattedAccountData)
	}

	return createdAccountRows
}

func formatAccountDiffEventual(accountDiff builder.AccountDiffEventual, sd builder.StateDiff, accountAction string) []string {
	oldContractRoot := accountDiff.ContractRoot.OldValue
	newContractRoot := accountDiff.ContractRoot.NewValue
	var storageDiffPaths []string
	for k := range accountDiff.Storage {
		storageDiffPaths = append(storageDiffPaths, k)
	}
	formattedAccountData := []string{
		strconv.FormatInt(sd.BlockNumber, 10),
		sd.BlockHash.String(),
		accountAction,
		string(accountDiff.Code),
		accountDiff.CodeHash,
		strconv.FormatUint(*accountDiff.Nonce.OldValue, 10),
		strconv.FormatUint(*accountDiff.Nonce.NewValue, 10),
		accountDiff.Balance.OldValue.String(),
		accountDiff.Balance.NewValue.String(),
		*oldContractRoot,
		*newContractRoot,
		strings.Join(storageDiffPaths, ","),
	}
	return formattedAccountData
}

func formatAccountDiffIncremental(accountDiff builder.AccountDiffIncremental, sd builder.StateDiff, accountAction string) []string {
	oldContractRoot := accountDiff.ContractRoot.OldValue
	newContractRoot := accountDiff.ContractRoot.NewValue
	var storageDiffPaths []string
	for k := range accountDiff.Storage {
		storageDiffPaths = append(storageDiffPaths, k)
	}
	formattedAccountData := []string{
		strconv.FormatInt(sd.BlockNumber, 10),
		sd.BlockHash.String(),
		accountAction,
		"",
		accountDiff.CodeHash,
		strconv.FormatUint(*accountDiff.Nonce.OldValue, 10),
		strconv.FormatUint(*accountDiff.Nonce.NewValue, 10),
		accountDiff.Balance.OldValue.String(),
		accountDiff.Balance.NewValue.String(),
		*oldContractRoot,
		*newContractRoot,
		strings.Join(storageDiffPaths, ","),
	}
	return formattedAccountData
}

