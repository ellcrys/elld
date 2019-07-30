package accountmgr

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/ellcrys/elld/config"
	"github.com/fatih/color"
)

// ListBurnerAccounts returns the burner accounts stored on disk.
func (am *AccountManager) ListBurnerAccounts() (accounts []*StoredAccount, err error) {

	files, err := ioutil.ReadDir(filepath.Join(am.accountDir, config.BurnerAccountDirName))
	if err != nil {
		return nil, err
	}

	for _, f := range files {

		if f.IsDir() {
			continue
		}

		m, _ := regexp.Match("^[0-9]{10}_[a-zA-Z0-9]{26,34}(_testnet)?$", []byte(f.Name()))
		if !m {
			continue
		}
		nameParts := strings.Split(f.Name(), "_")
		unixTime, _ := strconv.ParseInt(nameParts[0], 10, 64)
		timeCreated := time.Unix(unixTime, 0)
		path := filepath.Join(am.accountDir, config.BurnerAccountDirName, f.Name())
		cipher, _ := ioutil.ReadFile(path)
		if len(cipher) > 0 {
			accounts = append(accounts, &StoredAccount{
				Address:   nameParts[1],
				Cipher:    cipher,
				CreatedAt: timeCreated,
				meta: map[string]interface{}{
					"testnet": len(nameParts) == 3,
				},
			})
		}
	}
	return
}

// ListBurnerCmd fetches and lists all burner accounts
func (am *AccountManager) ListBurnerCmd() error {

	fmt.Println(fmt.Sprintf("\t%s%s%s%s%s",
		color.HiBlackString("Address"),
		strings.Repeat(" ", 32),
		color.HiBlackString("Date Created"),
		strings.Repeat(" ", 10),
		color.HiBlackString("Tag(s)")),
	)

	accounts, err := am.ListBurnerAccounts()
	if err != nil {
		return err
	}

	for i, a := range accounts {

		tags := []string{}

		if isTestnet, ok := a.meta["testnet"].(bool); ok && isTestnet {
			tags = append(tags, color.YellowString("[testnet]"))
		}

		if i == 0 {
			tags = append(tags, "[default]")
		}

		fmt.Println(fmt.Sprintf("[%d]\t%s     %s\t     %s", i, color.CyanString(a.Address),
			humanize.Time(a.CreatedAt), strings.Join(tags, " ")))
	}

	return nil
}
