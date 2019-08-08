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
	"github.com/fatih/color"
)

// ListAccounts returns the accounts stored on disk.
func (am *AccountManager) ListAccounts() (accounts []*StoredAccount, err error) {

	files, err := ioutil.ReadDir(am.accountDir)
	if err != nil {
		return nil, err
	}

	for _, f := range files {

		if f.IsDir() {
			continue
		}

		m, _ := regexp.Match("^[0-9]{10}_[a-zA-Z0-9]{34,}(_default)?$", []byte(f.Name()))
		if !m {
			continue
		}

		nameParts := strings.Split(f.Name(), "_")
		unixTime, _ := strconv.ParseInt(nameParts[0], 10, 64)
		timeCreated := time.Unix(unixTime, 0)
		cipher, _ := ioutil.ReadFile(filepath.Join(am.accountDir, f.Name()))
		if len(cipher) > 0 {
			accounts = append(accounts, &StoredAccount{
				Address:   nameParts[1],
				Cipher:    cipher,
				CreatedAt: timeCreated,
				Default:   strings.HasSuffix(f.Name(), "_default"),
			})
		}
	}

	return
}

// ListCmd fetches and lists all accounts
func (am *AccountManager) ListCmd() error {

	fmt.Println(fmt.Sprintf("\t%s%s%s%s%s",
		color.HiBlackString("Address"),
		strings.Repeat(" ", 32),
		color.HiBlackString("Date Created"),
		strings.Repeat(" ", 10),
		color.HiBlackString("Tag(s)")),
	)

	accts, err := am.ListAccounts()
	if err != nil {
		return err
	}

	for i, a := range accts {
		tagStr := ""
		if a.Default {
			tagStr = "[default]"
		}
		fmt.Println(fmt.Sprintf("[%d]\t%s     %s\t     %s", i, color.CyanString(a.Address),
			humanize.Time(a.CreatedAt), tagStr))
	}

	return nil
}
