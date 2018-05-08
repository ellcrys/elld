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

// StoredAccount represents an encrypted account stored on disk
type StoredAccount struct {
	Address   string
	Cipher    []byte
	CreatedAt time.Time
}

// GetAccountsOnDisk returns the accounts stored on disk.
func (am *AccountManager) GetAccountsOnDisk() (accounts []*StoredAccount, err error) {

	files, err := ioutil.ReadDir(am.accountDir)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		if !f.IsDir() {
			m, _ := regexp.Match("[0-9]{10}_[a-zA-Z0-9]{34,}", []byte(f.Name()))
			if m {
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
		}
	}
	return
}

// List fetches and lists all accounts
func (am *AccountManager) List() error {

	fmt.Println(fmt.Sprintf("%s%s%s",
		color.HiBlackString("Address"),
		strings.Repeat(" ", 32),
		color.HiBlackString("Date Created")),
	)

	accts, err := am.GetAccountsOnDisk()
	if err != nil {
		return err
	}

	for _, a := range accts {
		fmt.Println(fmt.Sprintf("%s     %s", color.CyanString(a.Address), humanize.Time(a.CreatedAt)))
	}

	return nil
}
