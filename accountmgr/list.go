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

		m, _ := regexp.Match("^[0-9]{10}_[a-zA-Z0-9]{34,}$", []byte(f.Name()))
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
			})
		}
	}
	return
}

// ListCmd fetches and lists all accounts
func (am *AccountManager) ListCmd() error {

	fmt.Println(fmt.Sprintf("\t%s%s%s",
		color.HiBlackString("Address"),
		strings.Repeat(" ", 32),
		color.HiBlackString("Date Created")),
	)

	accts, err := am.ListAccounts()
	if err != nil {
		return err
	}

	for i, a := range accts {
		defStr := "[default]"
		if i != 0 {
			defStr = ""
		}
		fmt.Println(fmt.Sprintf("[%d]\t%s     %s\t%s", i, color.CyanString(a.Address), humanize.Time(a.CreatedAt), defStr))
	}

	return nil
}
